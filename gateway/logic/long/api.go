package long

import (
	tp "github.com/henrylee2cn/teleport"
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
func (l *longConn) Total(args *types.LongConnArgs) (*types.TotalLongConnReply, *tp.Rerror) {
	return &types.TotalLongConnReply{ConnTotal: TotalConn()}, nil
}
