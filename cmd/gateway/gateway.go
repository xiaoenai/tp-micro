package main

import (
	"time"

	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/erpc/v6"
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/gateway"
	agent "github.com/xiaoenai/tp-micro/v6/gateway/helper/agent/socket"
	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray"
	"github.com/xiaoenai/tp-micro/v6/gateway/logic"
	"github.com/xiaoenai/tp-micro/v6/gateway/types"
	"github.com/xiaoenai/tp-micro/v6/model/etcd"
	"github.com/xiaoenai/tp-micro/v6/model/mysql"
	"github.com/xiaoenai/tp-micro/v6/model/redis"
)

type config struct {
	Gw               gateway.Config  `yaml:"gw"`
	GraySocketClient micro.CliConfig `yaml:"gray_socket_client"`
	GrayEtcd         etcd.EasyConfig `yaml:"gray_etcd"`
	Redis            redis.Config    `yaml:"redis"`
	Mysql            mysql.Config    `yaml:"mysql"`
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	c.Gw.OuterHttpServer.AllowCross = true
	err := bind()
	return err
}

func main() {
	logic.SetApiVersion("v1")
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
		Mysql: *mysql.NewConfig(),
	}
	cfgo.AllowAppsShare(true)
	cfgo.MustReg("gateway", &cfg)

	biz := types.DefaultBusiness()
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		erpc.Fatalf("%v", err)
	}
	agent.Init(redisClient, redisClient)
	biz.SocketHooks = agent.GetSocketHooks()
	_, err = gray.SetGray(biz, cfg.GraySocketClient, cfg.GrayEtcd, cfg.Mysql, cfg.Redis, nil)
	if err != nil {
		erpc.Fatalf("%v", err)
	}

	go func() {
		t := time.NewTicker(time.Second)
		for {
			<-t.C
			erpc.Infof("total conn: %d", gateway.TotalConn())
		}
	}()

	err = gateway.Run(cfg.Gw, biz, nil)
	if err != nil {
		erpc.Fatalf("%v", err)
	}
}
