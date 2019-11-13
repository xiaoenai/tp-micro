# TP-Micro [![GitHub release](https://img.shields.io/github/release/xiaoenai/tp-micro.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/releases) [![report card](https://goreportcard.com/badge/github.com/xiaoenai/tp-micro?style=flat-square)](http://goreportcard.com/report/xiaoenai/tp-micro) [![github issues](https://img.shields.io/github/issues/xiaoenai/tp-micro.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/xiaoenai/tp-micro.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/xiaoenai/tp-micro) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/xiaoenai/tp-micro/tree/v3/examples) [![view teleport](https://img.shields.io/badge/based%20on-teleport-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/teleport/v6) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


TP-Micro master(v6) is a simple, powerful micro service framework based on [Teleport v6](https://github.com/henrylee2cn/teleport/v6/tree/master).

[简体中文](https://github.com/xiaoenai/tp-micro/blob/master/README_ZH.md)

![tp-micro flow chart](https://github.com/xiaoenai/tp-micro/raw/master/doc/tp-micro_flow_chart.png)

## Install


```
go version ≥ 1.12
```

```sh
go get -u -f -d github.com/xiaoenai/tp-micro/...
cd $GOPATH/src/github.com/xiaoenai/tp-micro/cmd/micro
go install
```

## Feature

- Support auto service-discovery
- Supports custom service linker
- Support load balancing
- Support Asynchronous multiplexing IO
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
- Circuit breaker for overload protection

## Example

- server

```go
package main

import (
    micro "github.com/xiaoenai/tp-micro/v6"
    tp "github.com/henrylee2cn/teleport/v6"
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
func (p *P) Divide(arg *Arg) (int, *tp.Status) {
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

- client

```go
package main

import (
    micro "github.com/xiaoenai/tp-micro/v6"
    tp "github.com/henrylee2cn/teleport/v6"
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
    stat := cli.Call("/p/divide", &Arg{
        A: 10,
        B: 2,
    }, &result).Status()
    if stat != nil {
        tp.Fatalf("%v", stat)
    }
    tp.Infof("10/2=%d", result)
    stat = cli.Call("/p/divide", &Arg{
        A: 10,
        B: 0,
    }, &result).Status()
    if stat == nil {
        tp.Fatalf("%v", stat)
    }
    tp.Infof("test binding error: ok: %v", stat)
}
```

[More Examples](https://github.com/xiaoenai/tp-micro/tree/v3/examples)

## Learn `micro` Command

Command micro is deployment tools of tp-micro frameware.

- Quickly create a tp-micro project
- Run tp-micro project with hot compilation

### Generate project

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

- The initial template file `__tp-micro__tpl__.go`:

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
├── api
│   ├── handler.go
│   ├── pull_handler.gen.go
│   ├── push_handler.gen.go
│   ├── router.gen.go
│   └── router.go
├── args
│   ├── const.gen.go
│   ├── const.go
│   ├── type.gen.go
│   ├── type.go
│   └── var.go
├── config
│   └── config.yaml
├── config.go
├── doc
│   ├── APIDoc.md
│   ├── README.md
│   └── databases.md
├── log
│   └── PID
├── logic
│   ├── model
│   │   ├── init.go
│   │   ├── mongo_meta.gen.go
│   │   ├── mysql_device.gen.go
│   │   ├── mysql_log.gen.go
│   │   └── mysql_user.gen.go
│   └── tmp_code.gen.go
├── main.go
├── rerrs
│   └── rerrs.go
└── sdk
    ├── rpc.gen.go
    ├── rpc.gen_test.go
    ├── rpc.go
    └── rpc_test.go
```

**Desc:**

- This `micro gen` command only covers files with the ".gen.go" suffix if the `__tp-micro__gen__.lock` file exists
- Add `.gen` suffix to the file name of the automatically generated file, DO NOT EDIT!
- `.tmp` is temporary code used to ensure successful compilation!<br>It will be overwritten when `micro gen` is run!<br>When the project is completed, it should be removed!
- The type of handler's parameter and result must be struct!
- You can modify the created template file `__tp-micro__tpl__.go`, and run the `micro gen` command again to update the project

[Generated Default Sample](https://github.com/xiaoenai/tp-micro/tree/v3/examples/project)

### Create README.md(only)

`micro newdoc` command help:

```
NAME:
   micro newdoc - Generate a tp-micro project README.md

USAGE:
   micro newdoc [command options] [arguments...]

OPTIONS:
   --app_path value, -p value  The path(relative/absolute) of the project
```

### Run project

`micro run` command help:

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

### Add model

Add mysql model struct code to project template.

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


[More Micro Command](https://github.com/xiaoenai/tp-micro/tree/master/cmd/micro)

## Usage

### Peer(server or client) Demo

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

### Call-Controller-Struct API template

```go
type Aaa struct {
    tp.CallCtx
}
func (x *Aaa) XxZz(arg *<T>) (<T>, *tp.Status) {
    ...
    return r, nil
}
```

- register it to root router:

```go
// register the call route: /aaa/xx_zz
peer.RouteCall(new(Aaa))

// or register the call route: /xx_zz
peer.RouteCallFunc((*Aaa).XxZz)
```

### Call-Handler-Function API template

```go
func XxZz(ctx tp.CallCtx, arg *<T>) (<T>, *tp.Status) {
    ...
    return r, nil
}
```

- register it to root router:

```go
// register the call route: /xx_zz
peer.RouteCallFunc(XxZz)
```

### Push-Controller-Struct API template

```go
type Bbb struct {
    tp.PushCtx
}
func (b *Bbb) YyZz(arg *<T>) *tp.Status {
    ...
    return nil
}
```

- register it to root router:

```go
// register the push route: /bbb/yy_zz
peer.RoutePush(new(Bbb))

// or register the push route: /yy_zz
peer.RoutePushFunc((*Bbb).YyZz)
```

### Push-Handler-Function API template

```go
// YyZz register the route: /yy_zz
func YyZz(ctx tp.PushCtx, arg *<T>) *tp.Status {
    ...
    return nil
}
```

- register it to root router:

```go
// register the push route: /yy_zz
peer.RoutePushFunc(YyZz)
```

### Unknown-Call-Handler-Function API template

```go
func XxxUnknownCall (ctx tp.UnknownCallCtx) (interface{}, *tp.Status) {
    ...
    return r, nil
}
```

- register it to root router:

```go
// register the unknown call route: /*
peer.SetUnknownCall(XxxUnknownCall)
```

### Unknown-Push-Handler-Function API template

```go
func XxxUnknownPush(ctx tp.UnknownPushCtx) *tp.Status {
    ...
    return nil
}
```

- register it to root router:

```go
// register the unknown push route: /*
peer.SetUnknownPush(XxxUnknownPush)
```

### The mapping rule of struct(func) name to URI path:

- `AaBb` -> `/aa_bb`
- `Aa_Bb` -> `/aa/bb`
- `aa_bb` -> `/aa/bb`
- `Aa__Bb` -> `/aa_bb`
- `aa__bb` -> `/aa_bb`
- `ABC_XYZ` -> `/abc/xyz`
- `ABcXYz` -> `/abc_xyz`
- `ABC__XYZ` -> `/abc_xyz`

### Plugin Demo

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

func (i *ignoreCase) PostReadCallHeader(ctx tp.ReadCtx) *tp.Status {
    // Dynamic transformation path is lowercase
    ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
    return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx tp.ReadCtx) *tp.Status {
    // Dynamic transformation path is lowercase
    ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
    return nil
}
```

### Register above handler and plugin

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

### Config

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

#### Param-Tags

tag   |   key    | required |     value     |   desc
------|----------|----------|---------------|----------------------------------
param |   meta    | no |  (name e.g.`param:"<meta:id>"`)  | It indicates that the parameter is from the meta.
param |   swap    | no |   name (e.g.`param:"<swap:id>"`)  | It indicates that the parameter is from the context swap.
param |   desc   |      no      |     (e.g.`param:"<desc:id>"`)   | Parameter Description
param |   len    |      no      |   (e.g.`param:"<len:3:6>"`)  | Length range [a,b] of parameter's value
param |   range  |      no      |   (e.g.`param:"<range:0:10>"`)   | Numerical range [a,b] of parameter's value
param |  nonzero |      no      |    -    | Not allowed to zero
param |  regexp  |      no      |   (e.g.`param:"<regexp:^\\w+$>"`)  | Regular expression validation
param |   stat   |      no      |(e.g.`param:"<stat:100002:wrong password format>"`)| Custom error code and message

NOTES:

* `param:"-"` means ignore
* Encountered untagged exportable anonymous structure field, automatic recursive resolution
* Parameter name is the name of the structure field converted to snake format
* If the parameter is not from `meta` or `swap`, it is the default from the body

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

#### Example

```go
package main

import (
    tp "github.com/henrylee2cn/teleport/v6"
    micro "github.com/xiaoenai/tp-micro/v6"
)

type (
    // Arg arg
    Arg struct {
        A int
        B int `param:"<range:1:100>"`
        Query
        XyZ string `param:"<meta><nonzero><stat: 100002: Parameter cannot be empty>"`
    }
    Query struct {
        X string `param:"<meta>"`
    }
)

// P handler
type P struct {
    tp.CallCtx
}

// Divide divide API
func (p *P) Divide(arg *Arg) (int, *tp.Status) {
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

[Detail Example](https://github.com/xiaoenai/tp-micro/tree/master/examples/binder)

### Optimize

- SetMessageSizeLimit sets max packet size.
  If maxSize<=0, set it to max uint32.

```go
func SetMessageSizeLimit(maxMessageSize uint32)
```

- SetSocketKeepAlive sets whether the operating system should send
  keepalive messages on the connection.

```go
func SetSocketKeepAlive(keepalive bool)
```

- SetSocketKeepAlivePeriod sets period between keep alives.

```go
func SetSocketKeepAlivePeriod(d time.Duration)
```

- SetSocketNoDelay controls whether the operating system should delay
  packet transmission in hopes of sending fewer packets (Nagle's
  algorithm).  The default is true (no delay), meaning that data is
  sent as soon as possible after a Write.

```go
func SetSocketNoDelay(_noDelay bool)
```

- SetSocketReadBuffer sets the size of the operating system's
  receive buffer associated with the connection.

```go
func SetSocketReadBuffer(bytes int)
```

- SetSocketWriteBuffer sets the size of the operating system's
  transmit buffer associated with the connection.

```go
func SetSocketWriteBuffer(bytes int)
```

[More Usage](https://github.com/henrylee2cn/teleport/v6)

## License

Micro is under Apache v2 License. See the [LICENSE](https://github.com/xiaoenai/tp-micro/raw/master/LICENSE) file for the full license text
