// Copyright 2018 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package micro

import (
	"sync"
	"time"

	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin/heartbeat"
)

type (
	// CliConfig client config
	// Note:
	//  yaml tag is used for github.com/henrylee2cn/cfgo
	//  ini tag is used for github.com/henrylee2cn/ini
	CliConfig struct {
		Network            string               `yaml:"network"                ini:"network"                comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
		LocalIP            string               `yaml:"local_ip"               ini:"local_ip"               comment:"Local IP"`
		TlsCertFile        string               `yaml:"tls_cert_file"          ini:"tls_cert_file"          comment:"TLS certificate file path"`
		TlsKeyFile         string               `yaml:"tls_key_file"           ini:"tls_key_file"           comment:"TLS key file path"`
		DefaultSessionAge  time.Duration        `yaml:"default_session_age"    ini:"default_session_age"    comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
		DefaultContextAge  time.Duration        `yaml:"default_context_age"    ini:"default_context_age"    comment:"Default CALL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
		DefaultDialTimeout time.Duration        `yaml:"default_dial_timeout"   ini:"default_dial_timeout"   comment:"Default maximum duration for dialing; for client role; ns,µs,ms,s,m,h"`
		Failover           int                  `yaml:"failover"               ini:"failover"               comment:"The maximum times of failover"`
		SlowCometDuration  time.Duration        `yaml:"slow_comet_duration"    ini:"slow_comet_duration"    comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
		DefaultBodyCodec   string               `yaml:"default_body_codec"     ini:"default_body_codec"     comment:"Default body codec type id"`
		PrintDetail        bool                 `yaml:"print_detail"           ini:"print_detail"           comment:"Is print body and metadata or not"`
		CountTime          bool                 `yaml:"count_time"             ini:"count_time"             comment:"Is count cost time or not"`
		HeartbeatSecond    int                  `yaml:"heartbeat_second"       ini:"heartbeat_second"       comment:"When the heartbeat interval(second) is greater than 0, heartbeat is enabled; if it's smaller than 3, change to 3 default"`
		CircuitBreaker     CircuitBreakerConfig `yaml:"circuit_breaker" ini:"circuit_breaker" comment:"Circuit breaker config"`
	}
	// CircuitBreakerConfig circuit breaker config
	CircuitBreakerConfig struct {
		Enable          bool          `yaml:"enable" ini:"enable" comment:"Whether to use circuit breaker"`
		ErrorPercentage int           `yaml:"error_percentage" ini:"error_percentage" comment:"break linker when the error rate exceeds the threshold during a statistical period; default 50"`
		BreakDuration   time.Duration `yaml:"break_duration" ini:"break_duration" comment:"The period of one-cycle break in milliseconds; must ≥ 1ms"`
	}
)

// Reload Bi-directionally synchronizes config between YAML file and memory.
func (c *CliConfig) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err != nil {
		return err
	}
	return c.Check()
}

// Check check and correct config.
func (c *CliConfig) Check() error {
	if len(c.Network) == 0 {
		c.Network = "tcp"
	}
	if len(c.LocalIP) == 0 {
		c.LocalIP = "0.0.0.0"
	}
	if c.Failover < 0 {
		c.Failover = 0
	}
	if c.HeartbeatSecond <= 0 {
		c.HeartbeatSecond = 0
	} else if c.HeartbeatSecond < 3 {
		c.HeartbeatSecond = 3
	}
	if c.CircuitBreaker.ErrorPercentage <= 0 || c.CircuitBreaker.ErrorPercentage > 100 {
		c.CircuitBreaker.ErrorPercentage = defaultErrorPercentage
	}
	if c.CircuitBreaker.BreakDuration < time.Millisecond {
		c.CircuitBreaker.BreakDuration = defaultBreakDuration
	}
	return nil
}

func (c *CliConfig) peerConfig() tp.PeerConfig {
	return tp.PeerConfig{
		DefaultSessionAge:  c.DefaultSessionAge,
		DefaultContextAge:  c.DefaultContextAge,
		DefaultDialTimeout: c.DefaultDialTimeout,
		RedialTimes:        1,
		SlowCometDuration:  c.SlowCometDuration,
		DefaultBodyCodec:   c.DefaultBodyCodec,
		PrintDetail:        c.PrintDetail,
		CountTime:          c.CountTime,
		Network:            c.Network,
		LocalIP:            c.LocalIP,
	}
}

// Client client peer
type Client struct {
	peer           tp.Peer
	circuitBreaker *circuitBreaker
	protoFunc      tp.ProtoFunc
	closeCh        chan struct{}
	closeMu        sync.Mutex
	maxTry         int
	heartbeatPing  heartbeat.Ping
}

// NewClient creates a client peer.
func NewClient(cfg CliConfig, linker Linker, globalLeftPlugin ...tp.Plugin) *Client {
	doInit()
	if err := cfg.Check(); err != nil {
		tp.Fatalf("%v", err)
	}
	var heartbeatPing heartbeat.Ping
	if cfg.HeartbeatSecond > 0 {
		heartbeatPing = heartbeat.NewPing(cfg.HeartbeatSecond, false)
		globalLeftPlugin = append(globalLeftPlugin, heartbeatPing)
	}
	peer := tp.NewPeer(cfg.peerConfig(), globalLeftPlugin...)
	if len(cfg.TlsCertFile) > 0 && len(cfg.TlsKeyFile) > 0 {
		err := peer.SetTLSConfigFromFile(cfg.TlsCertFile, cfg.TlsKeyFile)
		if err != nil {
			tp.Fatalf("%v", err)
		}
	}
	cli := &Client{
		peer:          peer,
		protoFunc:     tp.DefaultProtoFunc(),
		closeCh:       make(chan struct{}),
		maxTry:        cfg.Failover + 1,
		heartbeatPing: heartbeatPing,
	}
	cli.circuitBreaker = newCircuitBreaker(
		cfg.CircuitBreaker.Enable,
		cfg.CircuitBreaker.ErrorPercentage,
		cfg.CircuitBreaker.BreakDuration,
		linker,
		func(addr string) (tp.Session, *tp.Status) {
			return cli.peer.Dial(addr, cli.protoFunc)
		})
	cli.circuitBreaker.start()
	return cli
}

// SetProtoFunc sets tp.ProtoFunc.
func (c *Client) SetProtoFunc(protoFunc tp.ProtoFunc) {
	if protoFunc == nil {
		protoFunc = tp.DefaultProtoFunc()
	}
	c.protoFunc = protoFunc
}

// Peer returns the peer.
func (c *Client) Peer() tp.Peer {
	return c.peer
}

// PluginContainer returns the global plugin container.
func (c *Client) PluginContainer() *tp.PluginContainer {
	return c.peer.PluginContainer()
}

// UseCallHeartbeat uses CALL method to ping.
func (c *Client) UseCallHeartbeat() {
	c.heartbeatPing.UseCall()
}

// UsePushHeartbeat uses PUSH method to ping.
func (c *Client) UsePushHeartbeat() {
	c.heartbeatPing.UsePush()
}

// SubRoute adds handler group.
func (c *Client) SubRoute(pathPrefix string, plugin ...tp.Plugin) *tp.SubRouter {
	return c.peer.SubRoute(pathPrefix, plugin...)
}

// RoutePush registers PUSH handlers, and returns the paths.
func (c *Client) RoutePush(ctrlStruct interface{}, plugin ...tp.Plugin) []string {
	return c.peer.RoutePush(ctrlStruct, plugin...)
}

// RoutePushFunc registers PUSH handler, and returns the path.
func (c *Client) RoutePushFunc(pushHandleFunc interface{}, plugin ...tp.Plugin) string {
	return c.peer.RoutePushFunc(pushHandleFunc, plugin...)
}

// AsyncCall sends a packet and receives reply asynchronously.
// Note:
//  If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
//  If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure;
//  Do not support failover to try again.
func (c *Client) AsyncCall(
	serviceMethod string,
	arg interface{},
	result interface{},
	callCmdChan chan<- tp.CallCmd,
	setting ...tp.MessageSetting,
) tp.CallCmd {
	if callCmdChan == nil {
		callCmdChan = make(chan tp.CallCmd, 10) // buffered.
	} else {
		// If caller passes callCmdChan != nil, it must arrange that
		// callCmdChan has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(callCmdChan) == 0 {
			tp.Panicf("*Client.AsyncCall(): callCmdChan channel is unbuffered")
		}
	}
	select {
	case <-c.closeCh:
		callCmd := tp.NewFakeCallCmd(serviceMethod, arg, result, RerrClientClosed)
		callCmdChan <- callCmd
		return callCmd
	default:
	}

	cliSess, stat := c.circuitBreaker.selectSession(serviceMethod)
	if stat != nil {
		callCmd := tp.NewFakeCallCmd(serviceMethod, arg, result, stat)
		callCmdChan <- callCmd
		return callCmd
	}
	callCmd := cliSess.AsyncCall(serviceMethod, arg, result, callCmdChan, setting...)
	cliSess.feedback(!tp.IsConnError(callCmd.Status()))
	return callCmd
}

// Call sends a packet and receives reply.
// Note:
//  If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
//  If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *Client) Call(serviceMethod string, arg interface{}, result interface{}, setting ...tp.MessageSetting) tp.CallCmd {
	select {
	case <-c.closeCh:
		return tp.NewFakeCallCmd(serviceMethod, arg, result, RerrClientClosed)
	default:
	}
	var (
		cliSess     *cliSession
		callCmd     tp.CallCmd
		stat        *tp.Status
		healthy     bool
		callCmdChan = make(chan tp.CallCmd, 1)
	)
	for i := 0; i < c.maxTry; i++ {
		cliSess, stat = c.circuitBreaker.selectSession(serviceMethod)
		if stat != nil {
			return tp.NewFakeCallCmd(serviceMethod, arg, result, stat)
		}
		cliSess.AsyncCall(serviceMethod, arg, result, callCmdChan, setting...)
		callCmd = <-callCmdChan
		healthy = !tp.IsConnError(callCmd.Status())
		cliSess.feedback(healthy)
		if healthy {
			return callCmd
		}
		if i > 0 {
			tp.Debugf("the %dth failover is triggered because: %s", i, callCmd.Status().String())
		}
	}
	return callCmd
}

// Push sends a packet, but do not receives reply.
// Note:
//  If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
//  If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *Client) Push(serviceMethod string, arg interface{}, setting ...tp.MessageSetting) *tp.Status {
	select {
	case <-c.closeCh:
		return RerrClientClosed
	default:
	}
	var (
		cliSess *cliSession
		stat    *tp.Status
		healthy bool
	)
	for i := 0; i < c.maxTry; i++ {
		cliSess, stat = c.circuitBreaker.selectSession(serviceMethod)
		if stat != nil {
			return stat
		}
		stat = cliSess.Push(serviceMethod, arg, setting...)
		healthy = !tp.IsConnError(stat)
		cliSess.feedback(healthy)
		if healthy {
			return stat
		}
		if i > 0 {
			tp.Debugf("the %dth failover is triggered because: %s", i, stat.String())
		}
	}
	return stat
}

// Close closes client.
func (c *Client) Close() {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	select {
	case <-c.closeCh:
		return
	default:
		close(c.closeCh)
		c.peer.Close()
		c.circuitBreaker.close()
	}
}
