package long

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/types"
)

// longConn long connection controller.
type longConn struct {
	tp.PullCtx
}

// TotalConn returns the long connections total.
func TotalConn() int32 {
	return int32(srv.Peer().CountSession())
}

// Total returns the long connections total.
func (l *longConn) Total(*types.TotalLongConnArgs) (*types.TotalLongConnReply, *tp.Rerror) {
	return &types.TotalLongConnReply{ConnTotal: TotalConn()}, nil
}

// InnerPush pushs the message to the designated uid.
func InnerPush(uid string, uri string, args interface{}, bodyCodec byte) *tp.Rerror {
	sess, rerr := logic.LongConnHooks().GetSession(peer, uid)
	if rerr != nil {
		return rerr
	}
	return sess.Push(uri, args, ant.WithBodyCodec(bodyCodec))
}

var pushLongConnReply = new(types.PushLongConnReply)

// Push returns the long connections total.
func (l *longConn) Push(args *types.PushLongConnArgs) (*types.PushLongConnReply, *tp.Rerror) {
	rerr := InnerPush(args.Uid, args.Uri, args.Body, byte(args.BodyCodec))
	if rerr != nil {
		return nil, rerr
	}
	return pushLongConnReply, nil
}
