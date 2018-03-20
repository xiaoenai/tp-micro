# Ants [![GitHub release](https://img.shields.io/github/release/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/releases) [![report card](https://goreportcard.com/badge/github.com/xiaoenai/ants)](http://goreportcard.com/report/xiaoenai/ants) [![Coverage Status](https://coveralls.io/repos/github/xiaoenai/ants/badge.svg?branch=master)](https://coveralls.io/github/xiaoenai/ants?branch=master) [![github issues](https://img.shields.io/github/issues/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/xiaoenai/ants.svg)](https://github.com/xiaoenai/ants/issues?q=is%3Aissue+is%3Aclosed) [![view teleport](https://img.shields.io/badge/based%20on-teleport-00BCD4.svg)](https://github.com/henrylee2cn/teleport) [![view ant](https://img.shields.io/badge/based%20on-ant-00BCD4.svg)](https://github.com/henrylee2cn/ant) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


Ants is a highly available microservice platform based on [Ant](https://github.com/henrylee2cn/ant) and [Teleport](https://github.com/henrylee2cn/teleport).

[简体中文](https://github.com/xiaoenai/ants/blob/master/README_ZH.md)


## 1. Install


```
go version ≥ 1.9
```

```sh
go get -u -f -d github.com/xiaoenai/ants/...
```

## 2. Feature

- Support auto service-discovery
- Supports custom service linker
- Support load balancing
- Support NIO and connection pool
- Support custom protocol
- Support custom body codec
- Support plug-in expansion
- Support heartbeat mechanism
- Detailed log information, support print input and output details
- Support for setting slow operation alarm thresholds
- Support for custom log
- Support smooth shutdown and update
- Support push handler
- Support network list: `tcp`, `tcp4`, `tcp6`, `unix`, `unixpacket` and so on
- Client support automatically redials after disconnection


## 3. Project Structure

(recommend)

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

Desc:

- add `.gen` suffix to the file name of the automatically generated file

## 4. Project Management

### 4.1 Install

```sh
cd $GOPATH/github.com/henrylee2cn/ant/cmd/ant
go install
```

### 4.2 Usage


- generate project

```
NAME:
   ant gen - Generate an ant project

USAGE:
   ant gen [command options] [arguments...]

OPTIONS:
   --template value, -t value    The template for code generation(relative/absolute)
   --app_path value, -p value  The path(relative/absolute) of the project
```

example: `ant gen -t ./__ant__tpl__.go -p ./myant` or default `ant gen myant`

template file `__ant__tpl__.go` demo:

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

- run project

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

[More Ant Command](https://github.com/henrylee2cn/ant/tree/master/cmd/ant)

## 5. Demo of Ant

- server

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

- client

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

[More](https://github.com/henrylee2cn/ant/tree/master/samples)

## 6. License

Ants is under Apache v2 License. See the [LICENSE](https://github.com/xiaoenai/ants/raw/master/LICENSE) file for the full license text
