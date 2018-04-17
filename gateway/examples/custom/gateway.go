package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/ants/gateway"
	"github.com/xiaoenai/ants/gateway/helper/agent"
	"github.com/xiaoenai/ants/gateway/types"
	"github.com/xiaoenai/ants/model/redis"
)

type config struct {
	Gw    gateway.Config
	Redis redis.Config
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	c.Gw.OuterHttpServer.AllowCross = true
	err := bind()
	return err
}

func main() {
	cfg := config{
		Gw:    *gateway.NewConfig(),
		Redis: *redis.NewConfig(),
	}
	cfgo.AllowAppsShare(true)
	cfgo.MustReg("gateway", &cfg)
	// Run a gateway instance with default business logic and default socket protocol.
	biz := types.DefaultBusiness()
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	agent.Init(10001, redisClient, redisClient)
	biz.SocketHooks = agent.GetSocketHooks()
	gateway.Run(cfg.Gw, biz, nil)
}
