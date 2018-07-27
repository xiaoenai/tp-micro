package main

import (
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket/example/pb"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

func main() {
	// discovery.SetServiceNamespace("test@")
	tp.SetSocketNoDelay(false)
	tp.SetShutdown(time.Second*20, nil, nil)

	cfg := micro.SrvConfig{
		DefaultBodyCodec: "protobuf",
		ListenAddress:    ":9090",
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
		group.RouteCall(new(Home))
	}
	srv.ListenAndServe()
}

// Home controller
type Home struct {
	tp.CallCtx
}

// Test handler
func (h *Home) Test(args *pb.PbTest) (*pb.PbTest, *tp.Rerror) {
	return &pb.PbTest{
		A: args.A + args.B,
		B: args.A - args.B,
	}, nil
}
