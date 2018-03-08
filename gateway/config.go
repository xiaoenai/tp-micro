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
	"github.com/xiaoenai/ants/gateway/logic/short"
	"github.com/xiaoenai/redis"
)

// Config app config
type Config struct {
	EnableOuterHttp bool                     `yaml:"enable_outer_http"`
	EnableOuterTcp  bool                     `yaml:"enable_outer_tcp"`
	OuterHttpServer short.OuterHttpSrvConfig `yaml:"outer_http_server"`
	OuterTcpServer  ant.SrvConfig            `yaml:"outer_tpc_server"`
	InnerTcpServer  ant.SrvConfig            `yaml:"inner_tcp_server"`
	InnerTcpClient  ant.CliConfig            `yaml:"inner_tcp_client"`
	Etcd            etcd.EasyConfig          `yaml:"etcd"`
	Redis           redis.Config             `yaml:"redis"`
	// outerPort, innerPort int
	// outerAddr, innerAddr string
}

// NewConfig creates a default config.
func NewConfig() *Config {
	return &Config{
		EnableOuterHttp: true,
		EnableOuterTcp:  true,
		OuterHttpServer: short.OuterHttpSrvConfig{
			ListenAddress: "0.0.0.0:5000",
		},
		OuterTcpServer: ant.SrvConfig{
			ListenAddress:   "0.0.0.0:5020",
			EnableHeartbeat: true,
			PrintBody:       true,
			CountTime:       true,
		},
		InnerTcpServer: ant.SrvConfig{
			ListenAddress:   "0.0.0.0:5030",
			EnableHeartbeat: true,
			PrintBody:       true,
			CountTime:       true,
		},
		InnerTcpClient: ant.CliConfig{
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
	err := c.InnerTcpClient.Check()
	if err != nil {
		return err
	}
	return nil
}
