package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/gateway"
	"github.com/xiaoenai/tp-micro/gateway/types"
)

func main() {
	cfg := gateway.NewConfig()
	cfg.OuterHttpServer.AllowCross = true
	cfgo.MustReg("gateway", cfg)
	// Run a gateway instance with default business logic and default socket protocol.
	biz := types.DefaultBusiness()
	err := gateway.Run(*cfg, biz, nil)
	if err != nil {
		tp.Fatalf("%v", err)
	}
}
