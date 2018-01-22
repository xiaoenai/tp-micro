package tcp

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
)

var srv *ant.Server

// Serve starts TCP gateway service.
func Serve(srvCfg ant.SrvConfig, protoFunc socket.ProtoFunc, etcdPlugin tp.Plugin) {
	srv = ant.NewServer(srvCfg, etcdPlugin)
	srv.Listen(protoFunc)
}
