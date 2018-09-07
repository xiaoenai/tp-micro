package main

import (
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
)

func main() {
	srv := micro.NewServer(
		cfg.Srv,
		discovery.ServicePlugin(cfg.Srv.InnerIpPort(), cfg.Etcd),
	)
	route("/project", srv.Router())
	srv.ListenAndServe()
}
