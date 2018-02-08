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
// If protoFunc=nil, socket.NewFastProtoFunc is used by default.
func Run(cfg *Config, biz *types.Business, protoFunc socket.ProtoFunc) error {
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

	// protocol
	if protoFunc == nil {
		protoFunc = socket.NewFastProtoFunc
	}

	// client
	client.Init(
		cfg.InnerClient,
		protoFunc,
		discovery.NewLinkerFromEtcd(etcdClient),
	)

	// HTTP server
	if cfg.EnableOuterHttp {
		go http.Serve(cfg.OuterHttpServer)
	}

	// TCP server
	if cfg.EnableOuterTcp {
		go tcp.Serve(
			cfg.OuterTcpServer,
			protoFunc,
			discovery.ServicePluginFromEtcd(cfg.innerAddr, etcdClient),
		)
	}

	select {}
}

// RegBodyCodec registers a mapping of content type to body coder.
func RegBodyCodec(contentType string, codecId byte) {
	http.RegBodyCodec(contentType, codecId)
}
