package websocket

import (
	tp "github.com/henrylee2cn/teleport"
	ws "github.com/henrylee2cn/teleport/mixer/websocket"
	"github.com/henrylee2cn/teleport/plugin/auth"
	"github.com/henrylee2cn/teleport/plugin/proxy"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/gateway/logic"
)

// Serve starts websocket gateway service.
func Serve(outerSrvCfg micro.SrvConfig, protoFunc tp.ProtoFunc) {
	srv := ws.NewServer(
		"/",
		outerSrvCfg.PeerConfig(),
		// authChecker,
		proxy.NewPlugin(logic.ProxySelector),
	)
	go srv.ListenAndServe(protoFunc)

	select {}
}

const clientAuthInfo = "client-auth-info-12345"

var authChecker = auth.NewCheckerPlugin(
	func(sess auth.Session, fn auth.RecvOnce) (ret interface{}, stat *tp.Status) {
		var authInfo string
		stat = fn(&authInfo)
		if !stat.OK() {
			return
		}
		tp.Infof("auth info: %v", authInfo)
		if clientAuthInfo != authInfo {
			return nil, tp.NewStatus(403, "auth fail", "auth fail detail")
		}
		return "pass", nil
	},
	tp.WithBodyCodec('s'),
)
