package tcp

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
)

// proxyClient the client for proxy.
type proxyClient struct {
	*ant.Client
}

func (p *proxyClient) Pull(uri string, args interface{}, reply interface{}, setting ...socket.PacketSetting) tp.PullCmd {
	// TODO some business code
	// ...
	return p.Pull(uri, args, reply, setting...)
}

func (p *proxyClient) Push(uri string, args interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	// TODO some business code
	// ...
	return p.Push(uri, args, setting...)
}
