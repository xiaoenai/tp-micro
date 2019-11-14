package main

import (
	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/erpc/v6"
	"github.com/xiaoenai/tp-micro/v6/gateway"
	"github.com/xiaoenai/tp-micro/v6/gateway/types"
)

//go:generate go build $GOFILE

func main() {
	cfg := gateway.NewConfig()
	cfg.OuterHttpServer.AllowCross = true
	cfgo.MustReg("gateway", cfg)
	// Run a gateway instance with default business logic and default socket protocol.
	biz := types.DefaultBusiness()
	err := gateway.Run(*cfg, biz, nil, nil)
	if err != nil {
		erpc.Fatalf("%v", err)
	}
}
