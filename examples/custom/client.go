package main

import (
	tp "github.com/henrylee2cn/teleport/v6"
	"github.com/henrylee2cn/teleport/v6/plugin/auth"
	micro "github.com/xiaoenai/tp-micro/v6"
)

func main() {
	cli := micro.NewClient(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		micro.NewStaticLinker(":5020"),
		auth.LaunchAuth(generateAuthInfo),
	)
	// test call
	var reply int
	stat := cli.Call("/math/divide?access_token=sdfghj", Msg{A: 10, B: 2}, &reply).Status()
	if stat != nil {
		tp.Fatalf("%v", stat)
	}
	tp.Infof("10/2=%d", reply)

	// test push
	cli.RoutePushFunc(push)
	for msg := range c {
		// your business
		tp.Infof("received: %v", msg)
	}
}

func generateAuthInfo() string {
	return "client-auth-info-12345"
}

var c = make(msgChan, 1000)

type msgChan chan *Msg

type Msg struct {
	A int
	B int
}

func push(ctx tp.PushCtx, arg *Msg) *tp.Status {
	c <- arg
	return nil
}
