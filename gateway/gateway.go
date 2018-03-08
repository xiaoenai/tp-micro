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
	"github.com/henrylee2cn/ant/discovery/etcd"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/client"
	"github.com/xiaoenai/ants/gateway/logic/long"
	"github.com/xiaoenai/ants/gateway/logic/short"
	"github.com/xiaoenai/ants/gateway/types"
)

// Run the gateway main program.
// If protoFunc=nil, socket.NewFastProtoFunc is used by default.
// If biz=nil, types.DefaultBusiness() is used by default.
func Run(cfg Config, biz *types.Business, protoFunc socket.ProtoFunc) error {
	// config
	err := cfg.check()
	if err != nil {
		return err
	}

	// etcd
	etcdClient, err := etcd.EasyNew(cfg.Etcd)
	if err != nil {
		return err
	}

	// business
	if biz == nil {
		biz = types.DefaultBusiness()
	}
	logic.SetBusiness(biz)

	// protocol
	if protoFunc == nil {
		protoFunc = socket.NewFastProtoFunc
	}

	// client
	client.Init(
		cfg.InnerTcpClient,
		protoFunc,
		etcdClient,
	)

	// HTTP server
	if cfg.EnableOuterHttp {
		go short.Serve(cfg.OuterHttpServer)
	}

	// TCP socket server
	if cfg.EnableOuterTcp {
		go long.Serve(
			cfg.OuterTcpServer,
			cfg.InnerTcpServer,
			protoFunc,
		)
	}

	select {}
}

// RegBodyCodecForShort registers a mapping of content type to body coder (for http).
func RegBodyCodecForShort(contentType string, codecId byte) {
	short.RegBodyCodec(contentType, codecId)
}

// SetAccessTokenGetterForShort sets the function to get access token (for http).
func SetAccessTokenGetterForShort(fn func(args types.RequestArgs) string) {
	short.SetAccessTokenGetter(fn)
}
