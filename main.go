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
	serfPort := cmp.Or(os.Getenv("SERF_PORT"), "8080")

	// Address to bind Serf on
	bindAddr := ":" + serfPort

	// Port for raft
	rpcPort := cmp.Or(os.Getenv("RPC_PORT"), "8888")

	// Unique server ID
	nodeName, _ := os.Hostname()

	// Serf addresses to join
	startJoinAddrsStr := cmp.Or(os.Getenv("START_JOIN_ADDRS"), "127.0.0.1:"+serfPort)

	startJoinAddrs := strings.Split(startJoinAddrsStr, ",")

	mem, err := NewMembership(bindAddr, rpcPort, nodeName, startJoinAddrs)
	if err != nil {
		panic(err)
	}

	go mem.EventHandler()

	retryCount := 3

	if err := Retry(
		func() error {
			if err := mem.Join(mem.StartJoinAddrs); err != nil {
				return fmt.Errorf("join serf: %w", err)
			}

			return nil
		},
		retryCount,
	); err != nil {
		panic(err)
	}

	hdl, err := api.NewServer(mem)
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

	if err := mem.Leave(); err != nil {
		panic(err)
	}

	if err := mem.Serf.Shutdown(); err != nil {
		panic(err)
	}

	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}

	slog.Info("done shutdown")
}

type Membership struct {
	Serf           *serf.Serf
	Events         chan serf.Event // イベントチャネル:ノードがクラスタに参加または離脱したときに Serf のイベントを受信する手段
	NodeName       string          // ノード名:クラスタ全体におけるノードの一意な識別子
	BindAddr       string          // ゴシッププロトコルのためのアドレス
	StartJoinAddrs []string        // 新たなノードが既存のクラスタに参加するように設定する方法
	Tags           map[string]string
}

func NewMembership(
	bindAddr string,
	rpcPort string,
	nodeName string,
	startJoinAddrs []string,
) (*Membership, error) {
	host, _, err := net.SplitHostPort(bindAddr)
	if err != nil {
		return nil, fmt.Errorf("split bind address: %w", err)
	}

	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve tcp addr: %w", err)
	}

	cfg := serf.DefaultConfig()

	cfg.Init()

	cfg.MemberlistConfig.BindAddr = addr.IP.String()

	cfg.MemberlistConfig.BindPort = addr.Port

	events := make(chan serf.Event)

	cfg.EventCh = events

	tags := map[string]string{
		"rpc_addr": net.JoinHostPort(host, rpcPort),
	}

	cfg.Tags = tags

	cfg.NodeName = nodeName

	sf, err := serf.Create(cfg)
	if err != nil {
		return nil, fmt.Errorf("create serf: %w", err)
	}

	return &Membership{
		Serf:           sf,
		Events:         events,
		NodeName:       nodeName,
		Tags:           tags,
		StartJoinAddrs: startJoinAddrs,
	}, nil
}

func (mem *Membership) EventHandler() {
	for event := range mem.Events {
		switch event.EventType() {
		case serf.EventMemberJoin:
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s joined", member.Name))
				if mem.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
				// TODO
			}
		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s left", member.Name))
				if mem.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
				// TODO
			}
		default:
			slog.Info(fmt.Sprintf("handle event %s", event))
			for _, member := range event.(serf.MemberEvent).Members {
				slog.Info(fmt.Sprintf("member %s default", member.Name))
				if mem.IsLocal(member) {
					slog.Info("member is local")
					continue
				}
			}
		}
	}
}

func (mem *Membership) IsLocal(
	member serf.Member,
) bool {
	return mem.Serf.LocalMember().Name == member.Name
}

func (mem *Membership) Members() []serf.Member {
	return mem.Serf.Members()
}

func (mem *Membership) Leave() error {
	return mem.Serf.Leave()
}

func (mem *Membership) Join(
	existing []string,
) error {
	slog.Info(fmt.Sprintf("Joining Serf: %v", existing))

	if _, err := mem.Serf.Join(existing, true); err != nil {
		return fmt.Errorf("join: %w", err)
	}

	return nil
}

var _ api.Handler = (*Membership)(nil)

func (mem *Membership) ListCluster(
	ctx context.Context,
) (api.ListClusterRes, error) {
	slog.InfoContext(ctx, "call list cluster")

	clusters := make([]api.Cluster, len(mem.Members()))

	for i, member := range mem.Members() {
		slog.InfoContext(ctx, fmt.Sprintf("member: %s", member.Name))
		clusters[i] = api.Cluster{
			ID:       member.Name,
			RpcAddr:  member.Addr.String(),
			IsLeader: false,
		}
	}

	return &api.ListClusterResponse{
		Clusters: clusters,
	}, nil
}

func (dis *Membership) Health(
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
			return fmt.Errorf("after %d attempts", attempt)
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
