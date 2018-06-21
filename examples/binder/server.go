package main

import (
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
)

type (
	// Arg arg
	Arg struct {
		A int
		B int `param:"<range:1:100>"`
		Query
		XyZ string `param:"<query><nonzero><rerr: 100002: Parameter cannot be empty>"`
	}
	Query struct {
		X string `param:"<query>"`
	}
)

// P handler
type P struct {
	tp.PullCtx
}

// Divide divide API
func (p *P) Divide(arg *Arg) (int, *tp.Rerror) {
	tp.Infof("query arg x: %s, xy_z: %s", arg.Query.X, arg.XyZ)
	return arg.A / arg.B, nil
}

func main() {
	srv := micro.NewServer(micro.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	})
	group := srv.SubRoute("/static")
	group.RoutePull(new(P))
	srv.ListenAndServe()
}
