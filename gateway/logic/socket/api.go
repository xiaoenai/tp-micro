package socket

import (
	_ "unsafe"

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

// totalConn returns the long connections total.
//go:linkname totalConn github.com/xiaoenai/ants/gateway.TotalConn
//go:nosplit
func totalConn() int32 {
	return int32(outerPeer.CountSession())
}

// SocketTotal returns the long connections total.
func (g *gw) SocketTotal(*types.SocketTotalArgs) (*types.SocketTotalReply, *tp.Rerror) {
	return &types.SocketTotalReply{ConnTotal: totalConn()}, nil
}

// innerPush pushs the message to the designated uid.
func innerPush(uid string, uri string, args interface{}, bodyCodec byte) *tp.Rerror {
	sess, rerr := logic.SocketHooks().GetSession(outerPeer, uid)
	if rerr != nil {
		return rerr
	}
	return sess.Push(uri, args, tp.WithBodyCodec(bodyCodec))
}

var socketPushReply = new(types.SocketPushReply)

// SocketPush returns the long connections total.
func (g *gw) SocketPush(args *types.SocketPushArgs) (*types.SocketPushReply, *tp.Rerror) {
	rerr := innerPush(args.SessionId, args.Uri, args.Body, byte(args.BodyCodec))
	if rerr != nil {
		return nil, rerr
	}
	return socketPushReply, nil
}

// Kick kicks the uid offline.
func Kick(uid string) (existed bool, err error) {
	sess, existed := outerPeer.GetSession(uid)
	if existed {
		err = sess.Close()
	}
	return existed, err
}

// SocketKick kicks the uid offline.
func (g *gw) SocketKick(args *types.SocketKickArgs) (*types.SocketKickReply, *tp.Rerror) {
	existed, _ := Kick(args.SessionId)
	return &types.SocketKickReply{
		Existed: existed,
	}, nil
}
