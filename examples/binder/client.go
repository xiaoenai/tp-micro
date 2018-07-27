package main

import (
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
)

func main() {
	tp.SetLoggerLevel("ERROR")
	cli := micro.NewClient(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		micro.NewStaticLinker(":9090"),
	)
	defer cli.Close()

	type Arg struct {
		A int
		B int
	}

	var result int
	rerr := cli.Call("/static/p/divide?x=testquery_x&xy_z=testquery_xy_z", &Arg{
		A: 10,
		B: 2,
	}, &result).Rerror()
	if tp.IsConnRerror(rerr) {
		tp.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Printf("test 10/2=%d", result)

	rerr = cli.Call("/static/p/divide?x=testquery_x&xy_z=testquery_xy_z", &Arg{
		A: 10,
		B: 0,
	}, &result).Rerror()
	if tp.IsConnRerror(rerr) {
		tp.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr == nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Printf("test 10/0:%v", rerr)

	rerr = cli.Call("/static/p/divide", &Arg{
		A: 10,
		B: 5,
	}, &result).Rerror()
	if rerr == nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Printf("test 10/5:%v", rerr)
}
