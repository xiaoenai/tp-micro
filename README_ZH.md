# Ants [![GitHub release](https://img.shields.io/github/release/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/releases) [![report card](https://goreportcard.com/badge/github.com/xiaoenai/ants)](http://goreportcard.com/report/xiaoenai/ants) [![Coverage Status](https://coveralls.io/repos/github/xiaoenai/ants/badge.svg&branch=master)](https://coveralls.io/github/xiaoenai/ants?branch=master) [![github issues](https://img.shields.io/github/issues/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/issues?q=is%3Aissue+is%3Aclosed) [![view teleport](https://img.shields.io/badge/based%20on-teleport-00BCD4.svg)](https://github.com/henrylee2cn/teleport) [![view ant](https://img.shields.io/badge/based%20on-ant-00BCD4.svg)](https://github.com/henrylee2cn/ant) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


Ants 是一套基于 [Ant](https://github.com/henrylee2cn/ant) 和 [Teleport](https://github.com/henrylee2cn/teleport) 的高可用的微服务平台解决方案。


## 1. 安装

```
go version ≥ 1.9
```

```sh
go get -u -f -d github.com/xiaoenai/ants/...
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

## 4. 项目管理

### 4.1 安装

```sh
cd $GOPATH/github.com/henrylee2cn/ant/cmd/ant
go install
```

### 4.2 用法

- 生成项目

```
NAME:
   ant gen - Generate an ant project

USAGE:
   ant gen [command options] [arguments...]

OPTIONS:
   --template value, -t value    The template for code generation(relative/absolute)
   --app_path value, -p value  The path(relative/absolute) of the project
```

示例：`ant gen -t ./__ant__tpl__.go -p ./myant`

模板文件 `__ant__tpl__.go` 样板:

```go
// package __ANT__TPL__ is the project template
package __ANT__TPL__

// __API__PULL__ register PULL router:
//  /home
//  /math/divide
type __API__PULL__ interface {
  Home(*struct{}) *HomeReply
  Math
}

// __API__PUSH__ register PUSH router:
//  /stat
type __API__PUSH__ interface {
  Stat(*StatArgs)
}

// Math controller
type Math interface {
  // Divide handler
  Divide(*DivideArgs) *DivideReply
}

// HomeReply home reply
type HomeReply struct {
  Content string // text
}

type (
  // DivideArgs divide api args
  DivideArgs struct {
    // dividend
    A float64
    // divisor
    B float64 `param:"<range: 0.01:100000>"`
  }
  // DivideReply divide api result
  DivideReply struct {
    // quotient
    C float64
  }
)

// StatArgs stat handler args
type StatArgs struct {
  Ts int64 // timestamps
}
```

- 热编译运行

```
NAME:
   ant run - Compile and run gracefully (monitor changes) an any existing go project

USAGE:
   ant run [options] [arguments...]
 or
   ant run [options except -app_path] [arguments...] {app_path}

OPTIONS:
   --watch_exts value, -x value  Specified to increase the listening file suffix (default: ".go", ".ini", ".yaml", ".toml", ".xml")
   --app_path value, -p value    The path(relative/absolute) of the project
```

example: `ant run -x .yaml -p myant` or `ant run -x .yaml myant`

[更多 ant command](https://github.com/henrylee2cn/ant/tree/master/cmd/ant)

## 5. Ant项目示例

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

## 6. 开源协议

Ant 项目采用商业应用友好的 [Apache2.0](https://github.com/xiaoenai/ant/raw/master/LICENSE) 协议发布
