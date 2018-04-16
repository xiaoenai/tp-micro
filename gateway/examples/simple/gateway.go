package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/ants/gateway"
	"github.com/xiaoenai/ants/gateway/helper/agent"
	"github.com/xiaoenai/ants/gateway/types"
	"github.com/xiaoenai/ants/model/redis"
)

func main() {
	cfg := gateway.NewConfig()
	cfg.OuterHttpServer.AllowCross = true
	cfgo.MustReg("gateway", cfg)
	// Run a gateway instance with default business logic and default socket protocol.
	biz := types.DefaultBusiness()
	redisClient, err := redis.NewClient(redis.NewConfig())
	if err != nil {
		tp.Fatalf("%v", err)
	}
	agent.Init(10001, redisClient, redisClient)
	biz.SocketHooks = agent.GetSocketHooks()
	tp.Go(func() {
		agentNewsChan := agent.Subscribe()
		for news := range agentNewsChan {
			tp.Infof("agent news: uid:%s, event:%s",
				news.Uid, news.Event,
			)
		}
	})
	gateway.Run(*cfg, biz, nil)
}
