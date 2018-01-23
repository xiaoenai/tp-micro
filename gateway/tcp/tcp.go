package tcp

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/client"
)

var srv *ant.Server

// Serve starts TCP gateway service.
func Serve(srvCfg ant.SrvConfig, protoFunc socket.ProtoFunc, etcdPlugin tp.Plugin) {
	srv = ant.NewServer(
		srvCfg,
		etcdPlugin,
		plugin.VerifyAuth(verifyAuthInfo),
		plugin.Proxy(&proxyClient{client.Client()}),
	)
	srv.Listen(protoFunc)
}

func verifyAuthInfo(authInfo string, sess plugin.AuthSession) *tp.Rerror {
	tp.Debugf("verify-auth: id: %s, info: %s", sess.Id(), authInfo)
	// TODO some business code
	// ...
	return nil
}
