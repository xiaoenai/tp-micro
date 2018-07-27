package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/configer"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/model/etcd"
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
	cfgo.MustReg("configer", cfg)

	etcdClient, err := etcd.EasyNew(cfg.Etcd)
	if err != nil {
		tp.Fatalf("%v", err)
	}

	configer.InitMgr(etcdClient)

	srv := micro.NewServer(
		cfg.Server,
		discovery.ServicePluginFromEtcd(
			cfg.Server.InnerIpPort(),
			etcdClient,
		),
	)
	srv.RouteCall(configer.CallCtrl())
	srv.ListenAndServe()
}
