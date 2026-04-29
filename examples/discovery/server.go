package main

import (
	"time"

	tp "github.com/swxctx/teleport"
	"github.com/swxctx/teleport/socket/example/pb"
	micro "github.com/xiaoenai/tp-micro/v2"
	"github.com/xiaoenai/tp-micro/v2/discovery"
	"github.com/xiaoenai/tp-micro/v2/model/etcd"
)

func main() {
	// discovery.SetServiceNamespace("test@")
	tp.SetSocketNoDelay(false)
	tp.SetShutdown(time.Second*20, nil, nil)

	cfg := micro.SrvConfig{
		DefaultBodyCodec: "protobuf",
		ListenAddress:    ":0",
		EnableHeartbeat:  true,
	}
	srv := micro.NewServer(cfg, discovery.ServicePlugin(
		cfg.InnerIpPort(),
		etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	))
	{
		group := srv.SubRoute("group")
		group.RoutePull(new(Home))
	}
	srv.ListenAndServe()
}

// Home controller
type Home struct {
	tp.PullCtx
}

// Test handler
func (h *Home) Test(args *pb.PbTest) (*pb.PbTest, *tp.Rerror) {
	return &pb.PbTest{
		A: args.A + args.B,
		B: args.A - args.B,
	}, nil
}
