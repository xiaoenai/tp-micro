package main

import (
	"github.com/henrylee2cn/cfgo"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	conf "github.com/xiaoenai/ants/config"
)

type config struct {
	Server ant.SrvConfig   `yaml:"server"`
	Etcd   etcd.EasyConfig `yaml:"etcd"`
}

func (e *config) Reload(bindFunc cfgo.BindFunc) error {
	return bindFunc()
}

func main() {
	cfg := &config{
		Server: ant.SrvConfig{
			ListenAddress: ":4040",
		},
		Etcd: etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	}
	cfgo.MustReg("antconfig", cfg)

	etcdClient, err := etcd.EasyNew(cfg.Etcd)
	if err != nil {
		ant.Fatalf("%v", err)
	}

	conf.InitMgr(etcdClient)

	srv := ant.NewServer(
		cfg.Server,
		discovery.ServicePluginFromEtcd(
			cfg.Server.InnerIpPort(),
			etcdClient,
		),
	)
	srv.RoutePull(conf.PullCtrl())
	srv.Listen()
}
