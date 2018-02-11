package long

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/logic/client"
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

// RemoteTotalLongConn returns the long connections total of the remote server.
func RemoteTotalLongConn(srvAddr string, setting ...socket.PacketSetting) (*types.TotalLongConnReply, *tp.Rerror) {
	var reply = new(types.TotalLongConnReply)
	rerr := client.StaticClient(srvAddr).Pull("/gateway/long_conn/total", nil, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}
