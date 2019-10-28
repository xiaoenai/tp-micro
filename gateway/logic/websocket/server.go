package websocket

import (
	tp "github.com/henrylee2cn/teleport"
	ws "github.com/henrylee2cn/teleport/mixer/websocket"
	"github.com/henrylee2cn/teleport/plugin/auth"
	"github.com/henrylee2cn/teleport/plugin/binder"
	"github.com/henrylee2cn/teleport/plugin/heartbeat"
	"github.com/henrylee2cn/teleport/plugin/proxy"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/gateway/logic"
)

var (
	outerPeer      tp.Peer
	clientAuthInfo string
)

// Serve starts websocket gateway service.
func Serve(outerSrvCfg micro.SrvConfig, protoFunc tp.ProtoFunc) {
	// new ws server
	srv := newWsServer(
		outerSrvCfg,
		binder.NewStructArgsBinder(nil),
		authChecker,
		webSocketConnTabPlugin,
		proxy.NewPlugin(logic.ProxySelector),
		preWritePushPlugin(),
	)
	// ws outer peer
	outerPeer = srv.Peer
	go srv.ListenAndServe(protoFunc)

	select {}
}

// newWsServer
func newWsServer(outerSrvCfg micro.SrvConfig, globalLeftPlugin ...tp.Plugin) *ws.Server {
	// Heartbeat
	if outerSrvCfg.EnableHeartbeat {
		globalLeftPlugin = append(globalLeftPlugin, heartbeat.NewPong())
	}
	// new ws server
	return ws.NewServer(
		"/",
		outerSrvCfg.PeerConfig(),
		globalLeftPlugin...,
	)
}

// auth plugin
var authChecker = auth.NewCheckerPlugin(
	func(sess auth.Session, fn auth.RecvOnce) (ret interface{}, stat *tp.Status) {
		var authInfo string
		stat = fn(&authInfo)
		if !stat.OK() {
			return
		}
		tp.Tracef("auth info: %v", authInfo)
		stat = webSocketConnTabPlugin.authAndLogon(authInfo, sess)
		if !stat.OK() {
			return
		}
		return "", nil
	},
	tp.WithBodyCodec('s'),
)

// preWritePushPlugin returns PreWritePushPlugin.
func preWritePushPlugin() tp.Plugin {
	return &perPusher{fn: logic.WebSocketHooks().PreWritePush}
}

type perPusher struct {
	fn func(tp.WriteCtx) *tp.Status
}

func (p *perPusher) Name() string {
	return "PUSH-LOGIC"
}

var (
	_ tp.PreWritePushPlugin = (*perPusher)(nil)
)

func (p *perPusher) PreWritePush(ctx tp.WriteCtx) *tp.Status {
	return p.fn(ctx)
}
