package main

import (
	"github.com/henrylee2cn/cfgo"
	"github.com/xiaoenai/ants/gateway"
	"github.com/xiaoenai/ants/gateway/types"
)

func main() {
	cfg := gateway.NewConfig()
	cfg.OuterHttpServer.AllowCross = true
	cfgo.MustReg("gateway", cfg)
	// Run a gateway instance with default business logic and default socket protocol.
	biz := types.DefaultBusiness()
	biz.ApiVersion = "v1.0"
	gateway.Run(*cfg, biz, nil)
}
