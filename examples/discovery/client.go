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

	cli := micro.NewClient(
		micro.CliConfig{
			DefaultBodyCodec:   "protobuf",
			DefaultDialTimeout: time.Second * 5,
			Failover:           3,
			CircuitBreaker: micro.CircuitBreakerConfig{
				Enable:          true,
				ErrorPercentage: 50,
			},
			HeartbeatSecond: 3,
		},
		discovery.NewLinker(etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		}),
	)
	defer cli.Close()
	var reply = new(pb.PbTest)
	stat := cli.Call(
		"/group/home/test",
		&pb.PbTest{A: 10, B: 2},
		reply,
	).Rerror()
	if stat != nil {
		tp.Errorf("call error: %v", stat)
	} else {
		tp.Infof("call reply: %v", reply)
	}

	// test heartbeat
	time.Sleep(10e9)
	cli.UseCallHeartbeat()
	time.Sleep(10e9)
}
