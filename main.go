package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	bindAddr := ":" + serfPort

	// Port for raft
	rpcPort := os.Getenv("RPC_PORT")

	// Unique server ID
	nodeName, _ := os.Hostname()

	// Serf addresses to join
	startJoinAddrs := cmp.Or(os.Getenv("START_JOIN_ADDRS"), "127.0.0.1:"+serfPort)

	existing := strings.Split(startJoinAddrs, ",")

	dis, err := NewDiscovery(bindAddr, rpcPort, nodeName, existing)
	if err != nil {
		panic(err)
	}

	go dis.EventHandler()

	if err := Retry(
		func() error {
			if err := dis.Join(existing); err != nil {
				return fmt.Errorf("failed to join serf: %v", err)
			}

			return nil
		},
		3,
	); err != nil {
		panic(err)
	}

	hdl, err := api.NewServer(dis)
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

	if err := dis.Leave(); err != nil {
		panic(err)
	}

	if err := dis.Serf.Shutdown(); err != nil {
		panic(err)
	}

	if err := srv.Shutdown(ctx); err != nil && errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}

	slog.Info("done shutdown")
}

type Discovery struct {
	Serf           *serf.Serf
	Events         chan serf.Event // イベントチャネル:ノードがクラスタに参加または離脱したときに Serf のイベントを受信する手段
	NodeName       string          // ノード名:クラスタ全体におけるノードの一意な識別子
	BindAddr       string          // ゴシッププロトコルのためのアドレス
	StartJoinAddrs []string        // 新たなノードが既存のクラスタに参加するように設定する方法
	Tags           map[string]string
}

func NewDiscovery(
	bindAddr string,
	rpcPort string,
	nodeName string,
	startJoinAddrs []string,
) (*Discovery, error) {
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

	return &Discovery{
		Serf:     sf,
		Events:   events,
		NodeName: nodeName,
		Tags: map[string]string{
			"rpc_addr": net.JoinHostPort(host, rpcPort),
		},
		StartJoinAddrs: startJoinAddrs,
	}, nil
}

func (dis *Discovery) EventHandler() {
	for event := range dis.Events {
		switch event.EventType() {
		case serf.EventMemberJoin:
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s joined", member.Name))
				if dis.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
			}
		case serf.EventMemberLeave:
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s left", member.Name))
				if dis.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
			}
		default:
			slog.Info(fmt.Sprintf("handle event %s", event))
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s", member.Name))
				if dis.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
			}
		}
	}
}

func (dis *Discovery) IsLocal(
	member serf.Member,
) bool {
	return dis.Serf.LocalMember().Name == member.Name
}

func (dis *Discovery) Members() []serf.Member {
	return dis.Serf.Members()
}

func (dis *Discovery) Leave() error {
	return dis.Serf.Leave()
}

func (dis *Discovery) Join(
	existing []string,
) error {
	slog.Info(fmt.Sprintf("Joining Serf: %v", existing))

	if _, err := dis.Serf.Join(existing, true); err != nil {
		return fmt.Errorf("failed to join: %v", err)
	}

	return nil
}

var _ api.Handler = (*Discovery)(nil)

func (dis *Discovery) ListCluster(
	ctx context.Context,
) (api.ListClusterRes, error) {
	slog.InfoContext(ctx, "call list cluster")

	clusters := make([]api.Cluster, len(dis.Members()))

	for i, member := range dis.Members() {
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

func (dis *Discovery) Health(
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

		slog.Info("attempt: " + strconv.Itoa(attempt))

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
