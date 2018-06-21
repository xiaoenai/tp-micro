package main

import (
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	micro "github.com/xiaoenai/tp-micro"
)

func main() {
	cli := micro.NewClient(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		micro.NewStaticLinker(":5020"),
		plugin.LaunchAuth(generateAuthInfo),
	)

	var arg = &struct {
		A int
		B int
	}{
		A: 10,
		B: 2,
	}

	var reply int

	rerr := cli.Pull("/math/divide?access_token=sdfghj", arg, &reply).Rerror()
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("10/2=%d", reply)

	tp.Debugf("waiting for 10s...")
	time.Sleep(time.Second * 10)

	arg.B = 5
	rerr = cli.Pull("/math/divide?access_token=sdfghj", arg, &reply).Rerror()
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("10/5=%d", reply)

	tp.Debugf("waiting for 10s...")
	time.Sleep(time.Second * 10)
}

func generateAuthInfo() string {
	return "client-auth-info-12345"
}
