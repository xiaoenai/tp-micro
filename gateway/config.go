// Copyright 2018 github.com/xiaoenai. All Rights Reserved.
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

package gateway

import (
	"time"

	"github.com/henrylee2cn/cfgo"
	micro "github.com/xiaoenai/tp-micro"
	short "github.com/xiaoenai/tp-micro/gateway/logic/http"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

// Config app config
type Config struct {
	EnableHttp        bool                `yaml:"enable_http"`
	EnableSocket      bool                `yaml:"enable_socket"`
	EnableWebSocket   bool                `yaml:"enable_web_socket"`
	OuterHttpServer   short.HttpSrvConfig `yaml:"outer_http_server"`
	OuterSocketServer micro.SrvConfig     `yaml:"outer_socket_server"`
	InnerSocketServer micro.SrvConfig     `yaml:"inner_socket_server"`
	InnerSocketClient micro.CliConfig     `yaml:"inner_socket_client"`
	Etcd              etcd.EasyConfig     `yaml:"etcd"`
	WebSocketServer   micro.SrvConfig     `yaml:"web_socket_server"`
}

// NewConfig creates a default config.
func NewConfig() *Config {
	return &Config{
		EnableHttp:      true,
		EnableSocket:    true,
		EnableWebSocket: true,
		OuterHttpServer: short.HttpSrvConfig{
			ListenAddress: "0.0.0.0:5000",
			AllowCross:    false,
		},
		OuterSocketServer: micro.SrvConfig{
			ListenAddress:     "0.0.0.0:5020",
			EnableHeartbeat:   true,
			PrintDetail:       true,
			CountTime:         true,
			SlowCometDuration: time.Millisecond * 500,
		},
		InnerSocketServer: micro.SrvConfig{
			ListenAddress:     "0.0.0.0:5030",
			EnableHeartbeat:   true,
			PrintDetail:       true,
			CountTime:         true,
			SlowCometDuration: time.Millisecond * 500,
		},
		WebSocketServer: micro.SrvConfig{
			ListenAddress:     "0.0.0.0:5040",
			EnableHeartbeat:   true,
			PrintDetail:       true,
			CountTime:         true,
			SlowCometDuration: time.Millisecond * 500,
		},
	}
}

// Reload Bi-directionally synchronizes config between YAML file and memory.
func (c *Config) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err == nil {
		c.OuterHttpServer.OuterIpPort()
	}
	c.OuterHttpServer.PrintDetail = c.OuterSocketServer.PrintDetail
	c.OuterHttpServer.CountTime = c.OuterSocketServer.CountTime
	c.OuterHttpServer.SlowCometDuration = c.OuterSocketServer.SlowCometDuration
	return err
}

// check the config
func (c *Config) check() error {
	return nil
}
