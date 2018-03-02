# Ants [![GitHub release](https://img.shields.io/github/release/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/releases) [![report card](https://goreportcard.com/badge/github.com/xiaoenai/ants)](http://goreportcard.com/report/xiaoenai/ants) [![Coverage Status](https://coveralls.io/repos/github/xiaoenai/ants/badge.svg&branch=master)](https://coveralls.io/github/xiaoenai/ants?branch=master) [![github issues](https://img.shields.io/github/issues/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/issues?q=is%3Aissue+is%3Aclosed) [![view teleport](https://img.shields.io/badge/based%20on-teleport-00BCD4.svg)](https://github.com/henrylee2cn/teleport) [![view ant](https://img.shields.io/badge/based%20on-ant-00BCD4.svg)](https://github.com/henrylee2cn/ant) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


Ants 是一套基于 [Ant](https://github.com/henrylee2cn/ant) 和 [Teleport](https://github.com/henrylee2cn/teleport) 的高可用的微服务平台解决方案。


## 1. 安装

```
go version ≥ 1.7
```

```sh
go get -u github.com/xiaoenai/ants/...
```

## 2. 特性

- 支持服务自动发现
- 支持自定义服务链接选择器
- 支持负载均衡
- 支持多路复用IO及其连接池
- 支持自定义协议
- 支持自定义Body的编解码类型
- 支持插件扩展
- 支持心跳机制
- 日志信息详尽，支持打印输入、输出消息的详细信息（状态码、消息头、消息体）
- 支持设置慢操作报警阈值
- 支持自定义日志
- 支持平滑关闭与更新
- 支持推送
- 支持的网络类型：`tcp`、`tcp4`、`tcp6`、`unix`、`unixpacket`等
- 客户端支持断线后自动重连

## 3. 项目结构

（推荐）

```
├── README.md
├── main.go
├── api
│   ├── handlers.gen.go
│   ├── handlers.go
│   ├── router.gen.go
│   └── router.go
├── logic
│   └── xxx.go
├── sdk
│   ├── rpc.gen.go
│   ├── rpc.gen_test.go
│   ├── rpc.go
│   └── rpc_test.go
└── types
    ├── types.gen.go
    └── types.go
```

说明：

- 在自动生成的文件的文件名中增加 `.gen` 后缀进行标记

## 4. Ant项目示例

- 服务端

```go
package main

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
)

// Args args
type Args struct {
	A int
	B int `param:"<range:1:>"`
}

// P handler
type P struct {
	tp.PullCtx
}

// Divide divide API
func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func main() {
	srv := ant.NewServer(ant.SrvConfig{
		ListenAddress: ":9090",
	})
	srv.RoutePull(new(P))
	srv.Listen()
}
```

- 客户端

```go
package main

import (
	"github.com/henrylee2cn/ant"
)

func main() {
	cli := ant.NewClient(
		ant.CliConfig{},
		ant.NewStaticLinker(":9090"),
	)
	defer	cli.Close()

	type Args struct {
		A int
		B int
	}

	var reply int
	rerr := cli.Pull("/p/divide", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if rerr != nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("10/2=%d", reply)
	rerr = cli.Pull("/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if rerr == nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("test binding error: ok: %v", rerr)
}
```

[更多](https://github.com/henrylee2cn/ant/tree/master/samples)

## 5. 开源协议

Ant 项目采用商业应用友好的 [Apache2.0](https://github.com/xiaoenai/ant/raw/master/LICENSE) 协议发布
