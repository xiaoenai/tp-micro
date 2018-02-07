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

package main

import (
	"net"
	"strconv"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery"
	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/goutil"
	"github.com/xiaoenai/ants/gateway/http"
	"github.com/xiaoenai/redis"
)

// Config app config
type Config struct {
	EnableOuterHttp bool                    `yaml:"enable_outer_http"`
	EnableOuterTcp  bool                    `yaml:"enable_outer_tcp"`
	OuterHttpServer http.OuterHttpSrvConfig `yaml:"outer_http_server"`
	OuterTcpServer  ant.SrvConfig           `yaml:"outer_tpc_server"`
	InnerServer     ant.SrvConfig           `yaml:"inner_server"`
	InnerClient     ant.CliConfig           `yaml:"inner_client"`
	Etcd            discovery.EtcdConfig    `yaml:"etcd"`
	Redis           redis.Config            `yaml:"redis"`
	outerPort       int
	innerPort       int
	innerAddr       string
}

// Reload load or reload config
func (c *Config) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err != nil {
		return err
	}
	err = c.InnerClient.Check()
	if err != nil {
		return err
	}
	c.outerPort, err = getPort(c.OuterTcpServer.ListenAddress)
	if err != nil {
		return err
	}
	c.innerPort, err = getPort(c.InnerServer.ListenAddress)
	innerIp, err := goutil.IntranetIP()
	if err != nil {
		return err
	}
	c.innerAddr = net.JoinHostPort(innerIp, strconv.Itoa(c.innerPort))
	return err
}

func getPort(addr string) (int, error) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(port)
}

var cfg = &Config{
	EnableOuterHttp: true,
	EnableOuterTcp:  true,
	OuterHttpServer: http.OuterHttpSrvConfig{
		ListenAddress: "0.0.0.0:5000",
	},
	OuterTcpServer: ant.SrvConfig{
		ListenAddress:   "0.0.0.0:5020",
		EnableHeartbeat: true,
		PrintBody:       true,
		CountTime:       true,
	},
	InnerServer: ant.SrvConfig{
		ListenAddress:   "0.0.0.0:5030",
		EnableHeartbeat: true,
		PrintBody:       true,
		CountTime:       true,
	},
	InnerClient: ant.CliConfig{
		Failover:        3,
		HeartbeatSecond: 60,
	},
	Etcd: discovery.EtcdConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	},
	Redis: *redis.NewConfig(),
}

func init() {
	cfgo.MustReg("gateway", cfg)
}
