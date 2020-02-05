package main

import (
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/discovery"

	"github.com/xiaoenai/tp-micro/examples/project/api"
)

func main() {
	srv := micro.NewServer(
		cfg.Srv,
		discovery.ServicePlugin(cfg.Srv.InnerIpPort(), cfg.Etcd),
	)
	api.Route("/project", srv.Router())
	srv.ListenAndServe()
}
