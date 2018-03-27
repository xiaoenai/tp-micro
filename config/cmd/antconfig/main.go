package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	conf "github.com/xiaoenai/ants/config"
)

type config struct {
	Server micro.SrvConfig `yaml:"server"`
	Etcd   etcd.EasyConfig `yaml:"etcd"`
}

func (e *config) Reload(bindFunc cfgo.BindFunc) error {
	return bindFunc()
}

func main() {
	cfg := &config{
		Server: micro.SrvConfig{
			ListenAddress: ":4040",
		},
		Etcd: etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	}
	cfgo.MustReg("antconfig", cfg)

	etcdClient, err := etcd.EasyNew(cfg.Etcd)
	if err != nil {
		tp.Fatalf("%v", err)
	}

	conf.InitMgr(etcdClient)

	srv := micro.NewServer(
		cfg.Server,
		discovery.ServicePluginFromEtcd(
			cfg.Server.InnerIpPort(),
			etcdClient,
		),
	)
	srv.RoutePull(conf.PullCtrl())
	srv.ListenAndServe()
}
