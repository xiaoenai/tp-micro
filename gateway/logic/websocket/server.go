package websocket

import (
	"github.com/henrylee2cn/erpc/v6"
	ws "github.com/henrylee2cn/erpc/v6/mixer/websocket"
	"github.com/henrylee2cn/erpc/v6/plugin/auth"
	"github.com/henrylee2cn/erpc/v6/plugin/binder"
	"github.com/henrylee2cn/erpc/v6/plugin/heartbeat"
	"github.com/henrylee2cn/erpc/v6/plugin/proxy"
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/gateway/logic"
)

var (
	outerPeer      erpc.Peer
	clientAuthInfo string
)

// Serve starts websocket gateway service.
func Serve(outerSrvCfg micro.SrvConfig, protoFunc erpc.ProtoFunc) {
	// plugins
	globalLeftPlugin := []erpc.Plugin{
		binder.NewStructArgsBinder(nil),
		authChecker,
		webSocketConnTabPlugin,
		proxy.NewPlugin(logic.ProxySelector),
		preWritePushPlugin(),
	}
	if outerSrvCfg.EnableHeartbeat{
		globalLeftPlugin = append(globalLeftPlugin,heartbeat.NewPong())
	}
	// new ws server
	srv := ws.NewServer(
		"/",
		outerSrvCfg.PeerConfig(),
		globalLeftPlugin...,
	)
	// ws outer peer
	outerPeer = srv.Peer
	go srv.ListenAndServe(protoFunc)

	select {}
}

// auth plugin
var authChecker = auth.NewCheckerPlugin(
	func(sess auth.Session, fn auth.RecvOnce) (ret interface{}, stat *erpc.Status) {
		var authInfo string
		stat = fn(&authInfo)
		if !stat.OK() {
			return
		}
		erpc.Tracef("auth info: %v", authInfo)
		stat = webSocketConnTabPlugin.authAndLogon(authInfo, sess)
		if !stat.OK() {
			return
		}
		return "", nil
	},
	erpc.WithBodyCodec('s'),
)

// preWritePushPlugin returns PreWritePushPlugin.
func preWritePushPlugin() erpc.Plugin {
	return &perPusher{fn: logic.WebSocketHooks().PreWritePush}
}

type perPusher struct {
	fn func(erpc.WriteCtx) *erpc.Status
}

func (p *perPusher) Name() string {
	return "PUSH-LOGIC"
}

var (
	_ erpc.PreWritePushPlugin = (*perPusher)(nil)
)

func (p *perPusher) PreWritePush(ctx erpc.WriteCtx) *erpc.Status {
	return p.fn(ctx)
}
