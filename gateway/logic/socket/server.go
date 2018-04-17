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

package socket

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/xiaoenai/ants/gateway/client"
	"github.com/xiaoenai/ants/gateway/logic"
)

var (
	outerPeer tp.Peer
)

// Serve starts TCP gateway service.
func Serve(outerSrvCfg, innerSrvCfg micro.SrvConfig, protoFunc socket.ProtoFunc) {
	outerServer := micro.NewServer(
		outerSrvCfg,
		plugin.VerifyAuth(socketConnTabPlugin.authAndLogon),
		socketConnTabPlugin,
		plugin.Proxy(logic.ProxySelector),
		preWritePushPlugin(),
	)

	outerPeer = outerServer.Peer()

	innerPlugins := logic.InnerServerPlugins()
	discoveryService := discovery.ServicePluginFromEtcd(
		innerSrvCfg.InnerIpPort(),
		client.EtcdClient(),
	)
	innerPlugins = append(innerPlugins, discoveryService)
	innerServer := micro.NewServer(
		innerSrvCfg,
		innerPlugins...,
	)

	gwGroup := innerServer.SubRoute("/gw")
	{
		verGroup := gwGroup.SubRoute(logic.ApiVersion())
		{
			verGroup.RoutePullFunc((*gw).Hosts)
			discoveryService.ExcludeApi(verGroup.RoutePullFunc((*gw).SocketTotal))
			discoveryService.ExcludeApi(verGroup.RoutePullFunc((*gw).SocketPush))
			discoveryService.ExcludeApi(verGroup.RoutePullFunc((*gw).SocketKick))
		}
	}

	go outerServer.ListenAndServe(protoFunc)
	go innerServer.ListenAndServe(protoFunc)

	select {}
}

// preWritePushPlugin returns PreWritePushPlugin.
func preWritePushPlugin() tp.Plugin {
	return &perPusher{fn: logic.SocketHooks().PreWritePush}
}

type perPusher struct {
	fn func(tp.WriteCtx) *tp.Rerror
}

func (p *perPusher) Name() string {
	return "PUSH-LOGIC"
}

var (
	_ tp.PreWritePushPlugin = (*perPusher)(nil)
)

func (p *perPusher) PreWritePush(ctx tp.WriteCtx) *tp.Rerror {
	return p.fn(ctx)
}
