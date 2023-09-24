package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/serf/serf"

	"github.com/otakakot/service-discovery-by-serf/gen/api"
)

func main() {
	slog.Info("Starting service-discovery")

	// Port for serf
	serfPort := os.Getenv("SERF_PORT")

	// Address to bind Serf on
	bindAddr := fmt.Sprintf(":%s", serfPort)

	// Port for raft
	rpcPort := os.Getenv("RPC_PORT")

	// Unique server ID
	nodeName, _ := os.Hostname()

	// Serf addresses to join
	startJoinAddrs := os.Getenv("START_JOIN_ADDRS")
	if startJoinAddrs == "" {
		slog.Info("No start join addresses specified, will join self")
		startJoinAddrs = fmt.Sprintf("127.0.0.1:%s", serfPort)
	}

	existing := strings.Split(startJoinAddrs, ",")

	ms, err := NewMembership(bindAddr, rpcPort, nodeName, existing)
	if err != nil {
		panic(err)
	}

	go ms.EventHandler()

	if err := Retry(
		func() error {
			if err := ms.Join(existing); err != nil {
				return fmt.Errorf("failed to join serf: %v", err)
			}

			return nil
		},
		3,
	); err != nil {
		panic(err)
	}

	hdl, err := api.NewServer(ms)
	if err != nil {
		panic(err)
	}

	slog.Info("Starting HTTP server")

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", rpcPort),
		Handler: hdl,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-ctx.Done()

	slog.Info("start shutdown")

	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cansel()

	if err := ms.Leave(); err != nil {
		panic(err)
	}

	if err := ms.Serf.Shutdown(); err != nil {
		panic(err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}

	slog.Info("done shutdown")
}

type Membership struct {
	Serf           *serf.Serf
	Events         chan serf.Event
	BindAddr       string
	NodeName       string
	Tags           map[string]string
	StartJoinAddrs []string
}

func NewMembership(
	bindAddr string,
	rpcPort string,
	nodeName string,
	startJoinAddrs []string,
) (*Membership, error) {
	host, _, err := net.SplitHostPort(bindAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to split bind address: %v", err)
	}

	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tcp addr: %v", err)
	}

	cfg := serf.DefaultConfig()

	cfg.Init()

	cfg.MemberlistConfig.BindAddr = addr.IP.String()
	cfg.MemberlistConfig.BindPort = addr.Port
	events := make(chan serf.Event)
	cfg.EventCh = events

	sf, err := serf.Create(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create serf: %v", err)
	}

	return &Membership{
		Serf:     sf,
		Events:   events,
		NodeName: nodeName,
		Tags: map[string]string{
			"rpc_addr": net.JoinHostPort(host, rpcPort),
		},
		StartJoinAddrs: startJoinAddrs,
	}, nil
}

func (ms *Membership) EventHandler() {
	for event := range ms.Events {
		switch event.EventType() {
		case serf.EventMemberJoin:
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s joined", member.Name))
				if ms.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
			}
		case serf.EventMemberLeave:
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s left", member.Name))
				if ms.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
			}
		}
	}
}

func (ms *Membership) IsLocal(
	member serf.Member,
) bool {
	return ms.Serf.LocalMember().Name == member.Name
}

func (ms *Membership) Members() []serf.Member {
	return ms.Serf.Members()
}

func (ms *Membership) Leave() error {
	return ms.Serf.Leave()
}

func (ms *Membership) Join(
	existing []string,
) error {
	slog.Info(fmt.Sprintf("Joining Serf: %v", existing))

	if _, err := ms.Serf.Join(existing, true); err != nil {
		return fmt.Errorf("failed to join: %v", err)
	}

	return nil
}

var _ api.Handler = (*Membership)(nil)

func (ms *Membership) ListCluster(
	ctx context.Context,
) (api.ListClusterRes, error) {
	slog.InfoContext(ctx, "call list cluster")

	clusters := make([]api.Cluster, len(ms.Members()))

	for i, member := range ms.Members() {
		slog.InfoContext(ctx, fmt.Sprintf("member: %s", member.Name))
		clusters[i] = api.Cluster{
			ID:       member.Name,
			RpcAddr:  "",
			IsLeader: false,
		}
	}

	return &api.ListClusterResponse{
		Clusters: clusters,
	}, nil
}

func (ms *Membership) Health(
	ctx context.Context,
) (api.HealthRes, error) {
	slog.InfoContext(ctx, "call health")

	return &api.HealthOK{}, nil
}

func Retry(
	fn func() error,
	retries int,
) error {
	attempt := 0

	for {
		if attempt > retries {
			return fmt.Errorf("failed after %d attempts", attempt)
		}

		slog.Info(fmt.Sprintf("Attempt %d", attempt))

		if err := fn(); err == nil {
			return nil
		} else {
			slog.Warn(err.Error())
		}

		interval := int(math.Pow(2, float64(attempt)))

		time.Sleep(time.Duration(interval) * time.Second)

		attempt++
	}
}
