// Package gateway is the main program for TCP and HTTP services.
//
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
//
package gateway

import (
	"github.com/henrylee2cn/ant/discovery"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/client"
	"github.com/xiaoenai/ants/gateway/logic/http"
	"github.com/xiaoenai/ants/gateway/logic/tcp"
	"github.com/xiaoenai/ants/gateway/types"
)

// Run the gateway main program.
func Run(cfg *Config, biz *types.Business) error {
	// config
	err := cfg.check()
	if err != nil {
		return err
	}

	// etcd
	etcdClient, err := discovery.NewEtcdClient(cfg.Etcd)
	if err != nil {
		return err
	}

	// business
	logic.SetBusiness(biz)

	// client
	client.Init(
		cfg.InnerClient,
		socket.NewFastProtoFunc,
		discovery.NewLinkerFromEtcd(etcdClient),
	)

	if cfg.EnableOuterHttp {
		go http.Serve(cfg.OuterHttpServer)
	}

	if cfg.EnableOuterTcp {
		go tcp.Serve(
			cfg.OuterTcpServer,
			socket.NewFastProtoFunc,
			discovery.ServicePluginFromEtcd(cfg.innerAddr, etcdClient),
		)
	}

	select {}
}
