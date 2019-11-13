package main

import (
	tp "github.com/henrylee2cn/teleport/v6"
	micro "github.com/xiaoenai/tp-micro/v6"
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
	stat := cli.Call("/static/p/divide?x=testquery_x&xy_z=testquery_xy_z", &Arg{
		A: 10,
		B: 2,
	}, &result).Status()
	if tp.IsConnError(stat) {
		tp.Fatalf("has conn rerror: %v", stat)
	}
	if stat != nil {
		tp.Fatalf("%v", stat)
	}
	tp.Printf("test 10/2=%d", result)

	stat = cli.Call("/static/p/divide?x=testquery_x&xy_z=testquery_xy_z", &Arg{
		A: 10,
		B: 0,
	}, &result).Status()
	if tp.IsConnError(stat) {
		tp.Fatalf("has conn rerror: %v", stat)
	}
	if stat == nil {
		tp.Fatalf("%v", stat)
	}
	tp.Printf("test 10/0:%v", stat)

	stat = cli.Call("/static/p/divide", &Arg{
		A: 10,
		B: 5,
	}, &result).Status()
	if stat == nil {
		tp.Fatalf("%v", stat)
	}
	tp.Printf("test 10/5:%v", stat)
}
