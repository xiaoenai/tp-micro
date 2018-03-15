package socket

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/hosts"
	"github.com/xiaoenai/ants/gateway/types"
)

// gw long connection controller.
type gw struct {
	tp.PullCtx
}

// Hosts returns the gateway seriver hosts.
func (g *gw) Hosts(*struct{}) (*types.GwHosts, *tp.Rerror) {
	return hosts.GwHosts(), nil
}

// TotalConn returns the long connections total.
func TotalConn() int32 {
	return int32(outerPeer.CountSession())
}

// SocketTotal returns the long connections total.
func (g *gw) SocketTotal(*types.SocketTotalArgs) (*types.SocketTotalReply, *tp.Rerror) {
	return &types.SocketTotalReply{ConnTotal: TotalConn()}, nil
}

// InnerPush pushs the message to the designated uid.
func InnerPush(uid string, uri string, args interface{}, bodyCodec byte) *tp.Rerror {
	sess, rerr := logic.SocketHooks().GetSession(outerPeer, uid)
	if rerr != nil {
		return rerr
	}
	return sess.Push(uri, args, ant.WithBodyCodec(bodyCodec))
}

var pushSocketReply = new(types.SocketPushReply)

// SocketPush returns the long connections total.
func (g *gw) SocketPush(args *types.SocketPushArgs) (*types.SocketPushReply, *tp.Rerror) {
	rerr := InnerPush(args.Uid, args.Uri, args.Body, byte(args.BodyCodec))
	if rerr != nil {
		return nil, rerr
	}
	return pushSocketReply, nil
}
