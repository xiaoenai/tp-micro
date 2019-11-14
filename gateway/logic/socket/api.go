package socket

import (
	"strings"
	"sync"
	_ "unsafe"

	"github.com/henrylee2cn/erpc/v6"
	"github.com/xiaoenai/tp-micro/v6/gateway/logic"
	"github.com/xiaoenai/tp-micro/v6/gateway/logic/hosts"
	"github.com/xiaoenai/tp-micro/v6/gateway/types"
)

// gw long connection controller.
type gw struct {
	erpc.CallCtx
}

// Hosts returns the gateway seriver hosts.
func (g *gw) Hosts(*struct{}) (*types.GwHosts, *erpc.Status) {
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
func (g *gw) SocketTotal(*types.SocketTotalArgs) (*types.SocketTotalReply, *erpc.Status) {
	return &types.SocketTotalReply{ConnTotal: totalConn()}, nil
}

// innerPush pushes the message to the specified user.
func innerPush(uid string, uri string, args interface{}, bodyCodec byte) *erpc.Status {
	sess, stat := logic.SocketHooks().GetSession(outerPeer, uid)
	if stat != nil {
		return stat
	}
	return sess.Push(uri, args, erpc.WithBodyCodec(bodyCodec))
}

var socketPushReply = new(types.SocketPushReply)

// SocketPush pushes message to the specified user.
func (g *gw) SocketPush(args *types.SocketPushArgs) (*types.SocketPushReply, *erpc.Status) {
	stat := innerPush(args.SessionId, args.Uri, args.Body, byte(args.BodyCodec))
	if stat != nil {
		return nil, stat
	}
	return socketPushReply, nil
}

// SocketMpush multi-push messages to the specified users.
func (g *gw) SocketMpush(args *types.SocketMpushArgs) (*types.SocketMpushReply, *erpc.Status) {
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
		erpc.TryGo(func() {
			defer wg.Done()
			stat := innerPush(sessId, uri, body, bodyCodec)
			if stat != nil {
				lock.Lock()
				failureSessionIds = append(failureSessionIds, sessId)
				lock.Unlock()
				erpc.Tracef("SocketMpush: %s", stat.String())
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
func (g *gw) SocketKick(args *types.SocketKickArgs) (*types.SocketKickReply, *erpc.Status) {
	existed, _ := Kick(args.SessionId)
	return &types.SocketKickReply{
		Existed: existed,
	}, nil
}
