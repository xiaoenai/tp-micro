package main

import (
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
)

// Home HTML home page
func Home(ctx tp.PullCtx, args *struct{}) ([]byte, *tp.Rerror) {
	ctx.SetMeta("Content-Type", "text/html; charset=utf-8")
	return []byte(`<!DOCTYPE html>
	<html>
	<head>
	    <title>home page</title>
	</head>
	<body>
	<br/>
	<h2><center>HTML Test!</center></h2>
	</body>
	</html>`), nil
}

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
	cfg := micro.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	}
	srv := micro.NewServer(cfg, discovery.ServicePlugin(
		cfg.InnerIpPort(),
		etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	))
	srv.RoutePullFunc(Home)
	srv.RoutePull(new(Math))
	srv.ListenAndServe()
}
