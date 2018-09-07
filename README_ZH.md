# TP-Micro [![GitHub release](https://img.shields.io/github/release/xiaoenai/tp-micro.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/releases) [![report card](https://goreportcard.com/badge/github.com/xiaoenai/tp-micro?style=flat-square)](http://goreportcard.com/report/xiaoenai/tp-micro) [![github issues](https://img.shields.io/github/issues/xiaoenai/tp-micro.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/xiaoenai/tp-micro.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/xiaoenai/tp-micro) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/tree/v3/examples) [![view teleport](https://img.shields.io/badge/based%20on-teleport-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/teleport) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


TP-Micro v3 是一个基于 [Teleport v4](https://github.com/henrylee2cn/teleport/tree/v4) 定制的、简约而强大的微服务框架。


## 安装

```
go version ≥ 1.9
```

```sh
go get -u -f -d github.com/xiaoenai/tp-micro/...
cd $GOPATH/src/github.com/xiaoenai/tp-micro/cmd/micro
go install
```

## 特性

- 支持服务自动发现
- 支持自定义服务链接选择器
- 支持负载均衡
- 支持异步多路复用IO
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
- 支持过载保护（断路器）

## 代码示例

- 服务端

```go
package main

import (
    micro "github.com/xiaoenai/tp-micro"
    tp "github.com/henrylee2cn/teleport"
)

// Arg arg
type Arg struct {
    A int
    B int `param:"<range:1:>"`
}

// P handler
type P struct {
    tp.CallCtx
}

// Divide divide API
func (p *P) Divide(arg *Arg) (int, *tp.Rerror) {
    return arg.A / arg.B, nil
}

func main() {
    srv := micro.NewServer(micro.SrvConfig{
        ListenAddress: ":9090",
    })
    srv.RouteCall(new(P))
    srv.ListenAndServe()
}
```

- 客户端

```go
package main

import (
    micro "github.com/xiaoenai/tp-micro"
    tp "github.com/henrylee2cn/teleport"
)

func main() {
    cli := micro.NewClient(
        micro.CliConfig{},
        micro.NewStaticLinker(":9090"),
    )
    defer cli.Close()

    type Arg struct {
        A int
        B int
    }

    var result int
    rerr := cli.Call("/p/divide", &Arg{
        A: 10,
        B: 2,
    }, &result).Rerror()
    if rerr != nil {
        tp.Fatalf("%v", rerr)
    }
    tp.Infof("10/2=%d", result)
    rerr = cli.Call("/p/divide", &Arg{
        A: 10,
        B: 0,
    }, &result).Rerror()
    if rerr == nil {
        tp.Fatalf("%v", rerr)
    }
    tp.Infof("test binding error: ok: %v", rerr)
}
```

[更多示例](https://github.com/xiaoenai/tp-micro/tree/v3/examples)


## 学习 `micro` 命令

- 快速创建项目
- 热编译模式运行项目

### 生成项目

`micro gen` command help:

```
NAME:
     micro gen - Generate a tp-micro project

USAGE:
     micro gen [command options] [arguments...]

OPTIONS:
     --template value, -t value    The template for code generation(relative/absolute)
     --app_path value, -p value  The path(relative/absolute) of the project
```

example: `micro gen -p ./myapp` or default `micro gen myapp`

- 初始模版文件 `__tp-micro__tpl__.go`:

```go
// package __TPL__ is the project template
package __TPL__

// __API_CALL__ register CALL router:
//  /home
//  /math/divide
type __API_CALL__ interface {
    Home(*struct{}) *HomeResult
    Math
}

// __API_PUSH__ register PUSH router:
//  /stat
type __API_PUSH__ interface {
    Stat(*StatArg)
}

// __MYSQL_MODEL__ create mysql model
type __MYSQL_MODEL__ struct {
    User
    Log
    Device
}

// __MONGO_MODEL__ create mongodb model
type __MONGO_MODEL__ struct {
    Meta
}

// Math controller
type Math interface {
    // Divide handler
    Divide(*DivideArg) *DivideResult
}

// HomeResult home result
type HomeResult struct {
    Content string // text
}

type (
    // DivideArg divide api arg
    DivideArg struct {
        // dividend
        A float64
        // divisor
        B float64 `param:"<range: 0.01:100000>"`
    }
    // DivideResult divide api result
    DivideResult struct {
        // quotient
        C float64
    }
)

// StatArg stat handler arg
type StatArg struct {
    Ts int64 // timestamps
}

// User user info
type User struct {
    Id   int64  `key:"pri"`
    Name string `key:"uni"`
    Age  int32
}

type Log struct {
    Text string
}

type Device struct {
    UUID string `key:"pri"`
}

type Meta struct {
    Hobby []string
    Tags  []string
}
```

- The template generated by `micro gen` command.

```
├── README.md
├── __tp-micro__gen__.lock
├── __tp-micro__tpl__.go
├── config
│   └── config.yaml
├── config.go
├── internal
│   ├── handler
│   │   ├── call.tmp.go
│   │   └── push.tmp.go
│   └── model
│       ├── init.go
│       ├── mongo_meta.gen.go
│       ├── mysql_device.gen.go
│       ├── mysql_log.gen.go
│       └── mysql_user.gen.go
├── log
│   └── PID
├── main.go
├── router.gen.go
└── sdk
    ├── rerr.go
    ├── rpc.gen.go
    ├── rpc.gen_test.go
    ├── type.gen.go
    └── val.gen.go
```

**说明：**

- 如果 `__tp-micro__gen__.lock` 文件存在，`micro gen` 命令只覆盖带有 ".gen.go" 后缀的文件
- 在自动生成的文件的文件名中增加 `.gen` 后缀进行标记
- `.tmp` 是为了通过编译而生成的临时文件，项目完成后应该移除它
- handler的参数和返回值必须是结构体类型
- 你可以修改默认创建的模板文件 `__tp-micro__tpl __.go`，并再次运行 `micro gen` 命令来更新项目

[生成的默认示例](https://github.com/xiaoenai/tp-micro/tree/v3/examples/project)

### 只创建 README.md

`micro newdoc` command help:

```
NAME:
   micro newdoc - Generate a tp-micro project README.md

USAGE:
   micro newdoc [command options] [arguments...]

OPTIONS:
   --app_path value, -p value  The path(relative/absolute) of the project
```

### 热编译运行

`micro run` 命令帮助：

```
NAME:
     micro run - Compile and run gracefully (monitor changes) an any existing go project

USAGE:
     micro run [options] [arguments...]
 or
     micro run [options except -app_path] [arguments...] {app_path}

OPTIONS:
     --watch_exts value, -x value  Specified to increase the listening file suffix (default: ".go", ".ini", ".yaml", ".toml", ".xml")
     --notwatch value, -n value    Not watch files or directories
     --app_path value, -p value    The path(relative/absolute) of the project
```

example: `micro run -x .yaml -p myapp` or `micro run`

### 添加数据模型

从 mysql 数据库表单添加相应结构体到项目模板文件。

`micro tpl` command help:

 ```
 NAME:
   micro tpl - Add mysql model struct code to project template

USAGE:
   micro tpl [command options] [arguments...]

OPTIONS:
   --app_path value, -p value      The path(relative/absolute) of the project
   --host value                    mysql host ip (default: "localhost")
   --port value                    mysql host port (default: "3306")
   --username value, --user value  mysql username (default: "root")
   --password value, --pwd value   mysql password
   --db value                      mysql database (default: "test")
   --table value                   mysql table
   --ssh_user value                ssh user
   --ssh_host value                ssh host ip
   --ssh_port value                ssh host port
 ```


[更多 Micro 命令](https://github.com/xiaoenai/tp-micro/tree/v3/cmd/micro)


## 用法

### Peer端点（服务端或客户端）示例

```go
// Start a server
var peer1 = tp.NewPeer(tp.PeerConfig{
        ListenAddress: "0.0.0.0:9090", // for server role
})
peer1.Listen()

...

// Start a client
var peer2 = tp.NewPeer(tp.PeerConfig{})
var sess, err = peer2.Dial("127.0.0.1:8080")
```


### Call-Controller-Struct 接口模板

```go
type Aaa struct {
        tp.CallCtx
}
func (x *Aaa) XxZz(arg *<T>) (<T>, *tp.Rerror) {
        ...
        return r, nil
}
```

- 注册到根路由：

```go
// register the call route: /aaa/xx_zz
peer.RouteCall(new(Aaa))

// or register the call route: /xx_zz
peer.RouteCallFunc((*Aaa).XxZz)
```

### Call-Handler-Function 接口模板

```go
func XxZz(ctx tp.CallCtx, arg *<T>) (<T>, *tp.Rerror) {
        ...
        return r, nil
}
```

- 注册到根路由：

```go
// register the call route: /xx_zz
peer.RouteCallFunc(XxZz)
```

### Push-Controller-Struct 接口模板

```go
type Bbb struct {
        tp.PushCtx
}
func (b *Bbb) YyZz(arg *<T>) *tp.Rerror {
        ...
        return nil
}
```

- 注册到根路由：

```go
// register the push route: /bbb/yy_zz
peer.RoutePush(new(Bbb))

// or register the push route: /yy_zz
peer.RoutePushFunc((*Bbb).YyZz)
```

### Push-Handler-Function 接口模板

```go
// YyZz register the route: /yy_zz
func YyZz(ctx tp.PushCtx, arg *<T>) *tp.Rerror {
        ...
        return nil
}
```

- 注册到根路由：

```go
// register the push route: /yy_zz
peer.RoutePushFunc(YyZz)
```

### Unknown-Call-Handler-Function 接口模板

```go
func XxxUnknownCall (ctx tp.UnknownCallCtx) (interface{}, *tp.Rerror) {
        ...
        return r, nil
}
```

- 注册到根路由：

```go
// register the unknown call route: /*
peer.SetUnknownCall(XxxUnknownCall)
```

### Unknown-Push-Handler-Function 接口模板

```go
func XxxUnknownPush(ctx tp.UnknownPushCtx) *tp.Rerror {
        ...
        return nil
}
```

- 注册到根路由：

```go
// register the unknown push route: /*
peer.SetUnknownPush(XxxUnknownPush)
```

### 结构体（函数）名称映射到URI路径的规则：

- `AaBb` -> `/aa_bb`
- `Aa_Bb` -> `/aa/bb`
- `aa_bb` -> `/aa/bb`
- `Aa__Bb` -> `/aa_bb`
- `aa__bb` -> `/aa_bb`
- `ABC_XYZ` -> `/abc/xyz`
- `ABcXYz` -> `/abc_xyz`
- `ABC__XYZ` -> `/abc_xyz`

### 插件示例

```go
// NewIgnoreCase Returns a ignoreCase plugin.
func NewIgnoreCase() *ignoreCase {
        return &ignoreCase{}
}

type ignoreCase struct{}

var (
        _ tp.PostReadCallHeaderPlugin = new(ignoreCase)
        _ tp.PostReadPushHeaderPlugin = new(ignoreCase)
)

func (i *ignoreCase) Name() string {
        return "ignoreCase"
}

func (i *ignoreCase) PostReadCallHeader(ctx tp.ReadCtx) *tp.Rerror {
        // Dynamic transformation path is lowercase
        ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
        return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
        // Dynamic transformation path is lowercase
        ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
        return nil
}
```

### 注册以上操作和插件示例到路由

```go
// add router group
group := peer.SubRoute("test")
// register to test group
group.RouteCall(new(Aaa), NewIgnoreCase())
peer.RouteCallFunc(XxZz, NewIgnoreCase())
group.RoutePush(new(Bbb))
peer.RoutePushFunc(YyZz)
peer.SetUnknownCall(XxxUnknownCall)
peer.SetUnknownPush(XxxUnknownPush)
```

### 配置信息

```go
// SrvConfig server config
type SrvConfig struct {
    Network           string        `yaml:"network"              ini:"network"              comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
    ListenAddress     string        `yaml:"listen_address"       ini:"listen_address"       comment:"Listen address; for server role"`
    TlsCertFile       string        `yaml:"tls_cert_file"        ini:"tls_cert_file"        comment:"TLS certificate file path"`
    TlsKeyFile        string        `yaml:"tls_key_file"         ini:"tls_key_file"         comment:"TLS key file path"`
    DefaultSessionAge time.Duration `yaml:"default_session_age"  ini:"default_session_age"  comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
    DefaultContextAge time.Duration `yaml:"default_context_age"  ini:"default_context_age"  comment:"Default CALL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
    SlowCometDuration time.Duration `yaml:"slow_comet_duration"  ini:"slow_comet_duration"  comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
    DefaultBodyCodec  string        `yaml:"default_body_codec"   ini:"default_body_codec"   comment:"Default body codec type id"`
    PrintDetail       bool          `yaml:"print_detail"         ini:"print_detail"         comment:"Is print body and metadata or not"`
    CountTime         bool          `yaml:"count_time"           ini:"count_time"           comment:"Is count cost time or not"`
    EnableHeartbeat   bool          `yaml:"enable_heartbeat"     ini:"enable_heartbeat"     comment:"enable heartbeat"`
}

// CliConfig client config
type CliConfig struct {
    Network             string               `yaml:"network"                ini:"network"                comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
    LocalIP             string               `yaml:"local_ip"               ini:"local_ip"               comment:"Local IP"`
    TlsCertFile         string               `yaml:"tls_cert_file"          ini:"tls_cert_file"          comment:"TLS certificate file path"`
    TlsKeyFile          string               `yaml:"tls_key_file"           ini:"tls_key_file"           comment:"TLS key file path"`
    DefaultSessionAge   time.Duration        `yaml:"default_session_age"    ini:"default_session_age"    comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
    DefaultContextAge   time.Duration        `yaml:"default_context_age"    ini:"default_context_age"    comment:"Default CALL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
    DefaultDialTimeout  time.Duration        `yaml:"default_dial_timeout"   ini:"default_dial_timeout"   comment:"Default maximum duration for dialing; for client role; ns,µs,ms,s,m,h"`
    RedialTimes         int                  `yaml:"redial_times"           ini:"redial_times"           comment:"The maximum times of attempts to redial, after the connection has been unexpectedly broken; for client role"`
    Failover            int                  `yaml:"failover"               ini:"failover"               comment:"The maximum times of failover"`
    SlowCometDuration   time.Duration        `yaml:"slow_comet_duration"    ini:"slow_comet_duration"    comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
    DefaultBodyCodec    string               `yaml:"default_body_codec"     ini:"default_body_codec"     comment:"Default body codec type id"`
    PrintDetail         bool                 `yaml:"print_detail"           ini:"print_detail"           comment:"Is print body and metadata or not"`
    CountTime           bool                 `yaml:"count_time"             ini:"count_time"             comment:"Is count cost time or not"`
    HeartbeatSecond     int                  `yaml:"heartbeat_second"       ini:"heartbeat_second"       comment:"When the heartbeat interval(second) is greater than 0, heartbeat is enabled; if it's smaller than 3, change to 3 default"`
    SessMaxQuota        int                  `yaml:"sess_max_quota"         ini:"sess_max_quota"         comment:"The maximum number of sessions in the connection pool"`
    SessMaxIdleDuration time.Duration        `yaml:"sess_max_idle_duration" ini:"sess_max_idle_duration" comment:"The maximum time period for the idle session in the connection pool; ns,µs,ms,s,m,h"`
    CircuitBreaker      CircuitBreakerConfig `yaml:"circuit_breaker" ini:"circuit_breaker" comment:"Circuit breaker config"`
}

// CircuitBreakerConfig circuit breaker config
type CircuitBreakerConfig struct {
    Enable          bool          `yaml:"enable" ini:"enable" comment:"Whether to use circuit breaker"`
    ErrorPercentage int           `yaml:"error_percentage" ini:"error_percentage" comment:"break linker when the error rate exceeds the threshold during a statistical period; default 50"`
    BreakDuration   time.Duration `yaml:"break_duration" ini:"break_duration" comment:"The period of one-cycle break in milliseconds; must ≥ 1ms"`
}
```


### Binder

#### Param-Tags


tag   |   key    | required |     value     |   desc
------|----------|----------|---------------|----------------------------------
param |   query    | no |  name (e.g.`param:"<query>"` or `param:"<query:id>"`)   | It indicates that the parameter is from the URI query part. e.g. `/a/b?x={query}`
param |   swap    | no |   name (e.g.`param:"<swap:id>"`)  | It indicates that the parameter is from the context swap.
param |   desc   |      no      |     (e.g.`param:"<desc:id>"`)   | Parameter Description
param |   len    |      no      |   (e.g.`param:"<len:3:6>"`)  | Length range [a,b] of parameter's value
param |   range  |      no      |   (e.g.`param:"<range:0:10>"`)   | Numerical range [a,b] of parameter's value
param |  nonzero |      no      |    -    | Not allowed to zero
param |  regexp  |      no      |   (e.g.`param:"<regexp:^\\w+$>"`)  | Regular expression validation
param |   rerr   |      no      |(e.g.`param:"<rerr:100002:wrong password format>"`)| Custom error code and message

NOTES:

* `param:"-"` means ignore
* Encountered untagged exportable anonymous structure field, automatic recursive resolution
* Parameter name is the name of the structure field converted to snake format
* If the parameter is not from `query` or `swap`, it is the default from the body
* Support for multiple rule combinations, e.g.`param:"<regexp:^\\w+$><len:6:8><rerr:100002:wrong password format>"`

#### Field-Types

base    |   slice    | special
--------|------------|------------
string  |  []string  | [][]byte
byte    |  []byte    | [][]uint8
uint8   |  []uint8   | struct
bool    |  []bool    |
int     |  []int     |
int8    |  []int8    |
int16   |  []int16   |
int32   |  []int32   |
int64   |  []int64   |
uint8   |  []uint8   |
uint16  |  []uint16  |
uint32  |  []uint32  |
uint64  |  []uint64  |
float32 |  []float32 |
float64 |  []float64 |

#### 示例

```go
package main

import (
    tp "github.com/henrylee2cn/teleport"
    micro "github.com/xiaoenai/tp-micro"
)

type (
    // Arg arg
    Arg struct {
        A int
        B int `param:"<range:1:100>"`
        Query
        XyZ string `param:"<query><nonzero><rerr: 100002: Parameter cannot be empty>"`
    }
    Query struct {
        X string `param:"<query>"`
    }
)

// P handler
type P struct {
    tp.CallCtx
}

// Divide divide API
func (p *P) Divide(arg *Arg) (int, *tp.Rerror) {
    tp.Infof("query arg x: %s, xy_z: %s", arg.Query.X, arg.XyZ)
    return arg.A / arg.B, nil
}

func main() {
    srv := micro.NewServer(micro.SrvConfig{
        ListenAddress:   ":9090",
        EnableHeartbeat: true,
    })
    group := srv.SubRoute("/static")
    group.RouteCall(new(P))
    srv.ListenAndServe()
}
```

[示例详情](https://github.com/xiaoenai/tp-micro/tree/v3/examples/binder)


### 通信优化

- SetPacketSizeLimit 设置包大小的上限，
    如果 maxSize<=0，上限默认为最大 uint32

```go
func SetPacketSizeLimit(maxPacketSize uint32)
```

- SetSocketKeepAlive 是否允许操作系统的发送TCP的keepalive探测包

```go
func SetSocketKeepAlive(keepalive bool)
```

- SetSocketKeepAlivePeriod 设置操作系统的TCP发送keepalive探测包的频度

```go
func SetSocketKeepAlivePeriod(d time.Duration)
```

- SetSocketNoDelay 是否禁用Nagle算法，禁用后将不在合并较小数据包进行批量发送，默认为禁用

```go
func SetSocketNoDelay(_noDelay bool)
```

- SetSocketReadBuffer 设置操作系统的TCP读缓存区的大小

```go
func SetSocketReadBuffer(bytes int)
```

- SetSocketWriteBuffer 设置操作系统的TCP写缓存区的大小

```go
func SetSocketWriteBuffer(bytes int)
```

[More Usage](https://github.com/henrylee2cn/teleport)


## 开源协议

Micro 项目采用商业应用友好的 [Apache2.0](https://github.com/xiaoenai/tp-micro/raw/v3/LICENSE) 协议发布
