package client

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/teleport/socket"
)

var cli *ant.Client

// Init initializes a common inner ant client.
func Init(cliCfg ant.CliConfig, protoFunc socket.ProtoFunc, linker ant.Linker) {
	cli = ant.NewClient(
		cliCfg,
		linker,
	)
	cli.SetProtoFunc(protoFunc)
}

// Client returns the common inner ant client.
func Client() *ant.Client {
	return cli
}
