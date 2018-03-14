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
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery/etcd"
	"github.com/henrylee2cn/cfgo"
	short "github.com/xiaoenai/ants/gateway/logic/http"
	"github.com/xiaoenai/redis"
)

// Config app config
type Config struct {
	EnableHttp        bool                     `yaml:"enable_http"`
	EnableSocket      bool                     `yaml:"enable_socket"`
	OuterHttpServer   short.OuterHttpSrvConfig `yaml:"outer_http_server"`
	OuterSocketServer ant.SrvConfig            `yaml:"outer_socket_server"`
	InnerSocketServer ant.SrvConfig            `yaml:"inner_socket_server"`
	InnerSocketClient ant.CliConfig            `yaml:"inner_socket_client"`
	Etcd              etcd.EasyConfig          `yaml:"etcd"`
	Redis             redis.Config             `yaml:"redis"`
}

// NewConfig creates a default config.
func NewConfig() *Config {
	return &Config{
		EnableHttp:   true,
		EnableSocket: true,
		OuterHttpServer: short.OuterHttpSrvConfig{
			ListenAddress: "0.0.0.0:5000",
			AllowCross:    false,
		},
		OuterSocketServer: ant.SrvConfig{
			ListenAddress:   "0.0.0.0:5020",
			EnableHeartbeat: true,
			PrintBody:       true,
			CountTime:       true,
		},
		InnerSocketServer: ant.SrvConfig{
			ListenAddress:   "0.0.0.0:5030",
			EnableHeartbeat: true,
			PrintBody:       true,
			CountTime:       true,
		},
		InnerSocketClient: ant.CliConfig{
			Failover:        3,
			HeartbeatSecond: 60,
		},
		Etcd: etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
		Redis: *redis.NewConfig(),
	}
}

// Reload Bi-directionally synchronizes config between YAML file and memory.
func (c *Config) Reload(bind cfgo.BindFunc) error {
	return bind()
}

// check the config
func (c *Config) check() error {
	err := c.InnerSocketClient.Check()
	if err != nil {
		return err
	}
	return nil
}
