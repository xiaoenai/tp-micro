package main

import (
	"time"

	"github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/erpc/v6/socket/example/pb"
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/discovery"
	"github.com/xiaoenai/tp-micro/v6/model/etcd"
)

func main() {
	// discovery.SetServiceNamespace("test@")
	erpc.SetSocketNoDelay(false)
	erpc.SetShutdown(time.Second*20, nil, nil)

	cfg := micro.SrvConfig{
		DefaultBodyCodec: "protobuf",
		ListenAddress:    ":",
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
	erpc.CallCtx
}

// Test handler
func (h *Home) Test(args *pb.PbTest) (*pb.PbTest, *erpc.Status) {
	return &pb.PbTest{
		A: args.A + args.B,
		B: args.A - args.B,
	}, nil
}
