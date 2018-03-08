package main

import (
	"github.com/henrylee2cn/cfgo"
	"github.com/xiaoenai/ants/gateway"
)

func main() {
	cfg := gateway.NewConfig()
	cfgo.MustReg("gateway", cfg)
	// Run a gateway instance with default business logic and default socket protocol.
	gateway.Run(*cfg, nil, nil)
}
