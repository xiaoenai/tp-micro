package sdk

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/types"
)

var client *ant.Client

// Init init client with config and linker.
func Init(cliConfig ant.CliConfig, linker ant.Linker) {
	client = ant.NewClient(
		cliConfig,
		linker,
	)
}

// InitWithClient init client with current client.
func InitWithClient(cli *ant.Client) {
	client = cli
}

// PullMathDivide division form mathematical calculation controller.
func PullMathDivide(args *types.MathDivideArgs, setting ...socket.PacketSetting) (*types.MathDivideReply, *tp.Rerror) {
	reply := new(types.MathDivideReply)
	rerr := client.Pull("/root/math/divide", args, reply, setting...).Rerror()
	return reply, rerr
}
