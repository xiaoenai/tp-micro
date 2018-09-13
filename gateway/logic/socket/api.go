package socket

import (
	"strings"
	"sync"
	_ "unsafe"

	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/gateway/logic"
	"github.com/xiaoenai/tp-micro/gateway/logic/hosts"
	"github.com/xiaoenai/tp-micro/gateway/types"
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
//go:linkname totalConn github.com/xiaoenai/tp-micro/gateway.TotalConn
//go:nosplit
func totalConn() int32 {
	if outerPeer == nil {
		return 0
	}
	return int32(outerPeer.CountSession())
}

// SocketTotal returns the long connections total.
func (g *gw) SocketTotal(*types.SocketTotalArgs) (*types.SocketTotalReply, *tp.Rerror) {
	return &types.SocketTotalReply{ConnTotal: totalConn()}, nil
}

// innerPush pushes the message to the specified user.
func innerPush(uid string, uri string, args interface{}, bodyCodec byte) *tp.Rerror {
	sess, rerr := logic.SocketHooks().GetSession(outerPeer, uid)
	if rerr != nil {
		return rerr
	}
	return sess.Push(uri, args, tp.WithBodyCodec(bodyCodec))
}

var socketPushReply = new(types.SocketPushReply)

// SocketPush pushes message to the specified user.
func (g *gw) SocketPush(args *types.SocketPushArgs) (*types.SocketPushReply, *tp.Rerror) {
	rerr := innerPush(args.SessionId, args.Uri, args.Body, byte(args.BodyCodec))
	if rerr != nil {
		return nil, rerr
	}
	return socketPushReply, nil
}

// SocketMpush multi-push messages to the specified users.
func (g *gw) SocketMpush(args *types.SocketMpushArgs) (*types.SocketMpushReply, *tp.Rerror) {
	var (
		wg                sync.WaitGroup
		sep               = "?"
		failureSessionIds = make([]string, 0, len(args.Target))
		lock              sync.Mutex
		body              = args.Body
		bodyCodec         = byte(args.BodyCodec)
	)
	if strings.Contains(args.Uri, "?") {
		sep = "&"
	}
	wg.Add(len(args.Target))
	for _, t := range args.Target {
		var uri string
		if t.AdditionalQuery != "" {
			uri = args.Uri + sep + t.AdditionalQuery
		} else {
			uri = args.Uri
		}
		sessId := t.SessionId
		tp.TryGo(func() {
			defer wg.Done()
			rerr := innerPush(sessId, uri, body, bodyCodec)
			if rerr != nil {
				lock.Lock()
				failureSessionIds = append(failureSessionIds, sessId)
				lock.Unlock()
				tp.Tracef("SocketMpush: %s", rerr.String())
			}
		})
	}
	wg.Wait()
	return &types.SocketMpushReply{
		FailureSessionIds: failureSessionIds,
	}, nil
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
