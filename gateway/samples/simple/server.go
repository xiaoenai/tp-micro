package main

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery"
	"github.com/henrylee2cn/ant/discovery/etcd"
	tp "github.com/henrylee2cn/teleport"
)

// Args args
type Args struct {
	A int
	B int `param:"<range:1:>"`
}

// Math handler
type Math struct {
	tp.PullCtx
}

// Divide divide API
func (m *Math) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func main() {
	cfg := ant.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	}
	srv := ant.NewServer(cfg, discovery.ServicePlugin(
		cfg.InnerIpPort(),
		etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	))
	srv.RoutePull(new(Math))
	srv.Listen()
}
