package main

import (
	micro "github.com/xiaoenai/tp-micro/v2"
	"github.com/xiaoenai/tp-micro/v2/discovery"

	"github.com/xiaoenai/tp-micro/v2/examples/project/api"
)

func main() {
	srv := micro.NewServer(
		cfg.Srv,
		discovery.ServicePlugin(cfg.Srv.InnerIpPort(), cfg.Etcd),
	)
	api.Route("/project", srv.Router())
	srv.ListenAndServe()
}
