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
	_ "unsafe"

	tp "github.com/henrylee2cn/teleport"
	// "github.com/henrylee2cn/teleport/proto/httproto"
	"github.com/henrylee2cn/teleport/mixer/websocket/jsonSubProto"
	"github.com/xiaoenai/tp-micro/clientele"
	"github.com/xiaoenai/tp-micro/gateway/logic"
	"github.com/xiaoenai/tp-micro/gateway/logic/hosts"
	short "github.com/xiaoenai/tp-micro/gateway/logic/http"
	long "github.com/xiaoenai/tp-micro/gateway/logic/socket"
	ws "github.com/xiaoenai/tp-micro/gateway/logic/websocket"
	"github.com/xiaoenai/tp-micro/gateway/sdk"
	"github.com/xiaoenai/tp-micro/gateway/types"
)

// Run the gateway main program.
// If protoFunc=nil, rawproto.NewRawProtoFunc is used by default.
// If biz=nil, types.DefaultBusiness() is used by default.
func Run(cfg Config, biz *types.Business, protoFunc tp.ProtoFunc) error {
	// config
	err := cfg.check()
	if err != nil {
		return err
	}

	// protocol
	// if protoFunc == nil {
	// 	protoFunc = httproto.NewHTTProtoFunc(true)
	// }

	// client
	clientele.SetProtoFunc(protoFunc)

	// business
	if biz == nil {
		biz = types.DefaultBusiness()
	}
	logic.SetBusiness(biz)

	// sdk
	sdk.SetApiVersion(logic.ApiVersion())

	var (
		httpAddr        string
		outerSocketAddr string
		innerSocketAddr string
		webSocketAddr   string
	)

	// HTTP server
	if cfg.EnableHttp {
		httpAddr = cfg.OuterHttpServer.OuterIpPort()
		go short.Serve(cfg.OuterHttpServer)
	}

	// TCP socket server
	if cfg.EnableSocket {
		outerSocketAddr = cfg.OuterSocketServer.OuterIpPort()
		innerSocketAddr = cfg.InnerSocketServer.InnerIpPort()
		go long.Serve(
			cfg.OuterSocketServer,
			cfg.InnerSocketServer,
			protoFunc,
		)
	}

	// web socket server
	if cfg.EnableWebSocket {
		webSocketAddr = cfg.WebSocketServer.OuterIpPort()
		go ws.Serve(
			cfg.WebSocketServer,
			jsonSubProto.NewJSONSubProtoFunc(),
		)
	}

	hosts.Start(
		httpAddr,
		outerSocketAddr,
		innerSocketAddr,
		webSocketAddr,
	)

	select {}
}

// RegBodyCodecForHTTP registers a mapping of content type to body coder (for http).
func RegBodyCodecForHTTP(contentType string, codecId byte) {
	short.RegBodyCodec(contentType, codecId)
}

// TotalConn returns the current connections amount.
func TotalConn() int32

// HostsNamespace returns the gateway hosts prefix of ETCD key.
func HostsNamespace() string {
	return hosts.HostsNamespace()
}

// SetHostsNamespace sets the gateway hosts prefix of ETCD key.
func SetHostsNamespace(prefix string) {
	hosts.SetHostsNamespace(prefix)
}
