package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	"github.com/xiaoenai/ants/gateway"
	"github.com/xiaoenai/ants/gateway/helper/agent"
	"github.com/xiaoenai/ants/gateway/helper/gray"
	"github.com/xiaoenai/ants/gateway/types"
	"github.com/xiaoenai/ants/model"
	"github.com/xiaoenai/ants/model/redis"
)

type config struct {
	Gw               gateway.Config  `yaml:"gw"`
	GraySocketClient micro.CliConfig `yaml:"gray_socket_client"`
	GrayEtcd         etcd.EasyConfig `yaml:"gray_etcd"`
	Redis            redis.Config    `yaml:"redis"`
	Mysql            model.Config    `yaml:"mysql"`
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	c.Gw.OuterHttpServer.AllowCross = true
	err := bind()
	return err
}

func main() {
	cfg := config{
		Gw: *gateway.NewConfig(),
		GraySocketClient: micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 60,
		},
		GrayEtcd: etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
		Redis: *redis.NewConfig(),
		Mysql: *model.NewConfig(),
	}
	cfgo.AllowAppsShare(true)
	cfgo.MustReg("gateway", &cfg)

	biz := types.DefaultBusiness()
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	agent.Init(10001, redisClient, redisClient)
	biz.SocketHooks = agent.GetSocketHooks()
	err = gray.SetGray(biz, cfg.GraySocketClient, cfg.GrayEtcd, cfg.Mysql, cfg.Redis, nil)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	gateway.Run(cfg.Gw, biz, nil)
}
