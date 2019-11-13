package websocket

import (
	"strings"
	"sync"
	_ "unsafe"

	tp "github.com/henrylee2cn/teleport/v6"
	"github.com/xiaoenai/tp-micro/v6/gateway/logic"
	"github.com/xiaoenai/tp-micro/v6/gateway/types"
)

// Gw long connection controller.
type Gw struct {
	tp.CallCtx
}

// totalConn returns the long connections total.
//go:linkname totalConn github.com/xiaoenai/tp-micro/gateway.TotalConn
//go:nosplit
func wsTotalConn() int32 {
	if outerPeer == nil {
		return 0
	}
	return int32(outerPeer.CountSession())
}

// WsTotal returns the long connections total.
func (g *Gw) WsTotal(*types.WsTotalArgs) (*types.WsTotalReply, *tp.Status) {
	return &types.WsTotalReply{ConnTotal: wsTotalConn()}, nil
}

// innerPush pushes the message to the specified user.
func innerPush(uid string, uri string, args interface{}, bodyCodec byte) *tp.Status {
	sess, stat := logic.WebSocketHooks().GetSession(outerPeer, uid)
	if stat != nil {
		return stat
	}
	return sess.Push(uri, args, tp.WithBodyCodec(bodyCodec))
}

var wsPushReply = new(types.WsPushReply)

// WsPush pushes message to the specified user.
func (g *Gw) WsPush(args *types.WsPushArgs) (*types.WsPushReply, *tp.Status) {
	stat := innerPush(args.SessionId, args.Uri, args.Body, byte(args.BodyCodec))
	if stat != nil {
		return nil, stat
	}
	return wsPushReply, nil
}

// WsMpush multi-push messages to the specified users.
func (g *Gw) WsMpush(args *types.WsMpushArgs) (*types.WsMpushReply, *tp.Status) {
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
			stat := innerPush(sessId, uri, body, bodyCodec)
			if stat != nil {
				lock.Lock()
				failureSessionIds = append(failureSessionIds, sessId)
				lock.Unlock()
				tp.Tracef("SocketMpush: %s", stat.String())
			}
		})
	}
	wg.Wait()
	return &types.WsMpushReply{
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
func (g *Gw) WsKick(args *types.SocketKickArgs) (*types.SocketKickReply, *tp.Status) {
	existed, _ := Kick(args.SessionId)
	return &types.SocketKickReply{
		Existed: existed,
	}, nil
}
