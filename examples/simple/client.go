package main

import (
	"time"

	"github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/erpc/v6/plugin/auth"
	micro "github.com/xiaoenai/tp-micro/v6"
)

//go:generate go build $GOFILE

func main() {
	cli := micro.NewClient(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		micro.NewStaticLinker(":5020"),
		authBearer,
	)

	var arg = &struct {
		A int
		B int
	}{
		A: 10,
		B: 2,
	}

	var reply int

	stat := cli.Call("/math/divide?access_token=sdfghj", arg, &reply).Status()
	if !stat.OK() {
		erpc.Fatalf("%v", stat)
	}
	erpc.Infof("10/2=%d", reply)

	erpc.Debugf("waiting for 10s...")
	time.Sleep(time.Second * 10)

	arg.B = 5
	stat = cli.Call("/math/divide?access_token=sdfghj", arg, &reply).Status()
	if !stat.OK() {
		erpc.Fatalf("%v", stat)
	}
	erpc.Infof("10/5=%d", reply)

	erpc.Debugf("waiting for 10s...")
	time.Sleep(time.Second * 10)
}

const clientAuthInfo = "client-auth-info-12345"

var authBearer = auth.NewBearerPlugin(
	func(sess auth.Session, fn auth.SendOnce) (stat *erpc.Status) {
		var ret string
		stat = fn(clientAuthInfo, &ret)
		if !stat.OK() {
			return
		}
		erpc.Infof("auth info: %s, result: %s", clientAuthInfo, ret)
		return
	},
	erpc.WithBodyCodec('s'),
)
