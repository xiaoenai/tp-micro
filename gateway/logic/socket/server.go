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
	"net"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin/auth"
	"github.com/henrylee2cn/teleport/plugin/proxy"
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/gateway/client"
	"github.com/xiaoenai/tp-micro/gateway/logic"
)

var (
	outerPeer   tp.Peer
	outerServer *micro.Server
)

// OuterServeConn serves connetion.
func OuterServeConn(conn net.Conn) {
	sess, err := outerServer.ServeConn(conn)
	if err != nil {
		tp.Errorf("Serve net.Conn error: %v", err)
	}
	<-sess.CloseNotify()
}

// Serve starts TCP gateway service.
func Serve(outerSrvCfg, innerSrvCfg micro.SrvConfig, protoFunc socket.ProtoFunc) {
	outerServer = micro.NewServer(
		outerSrvCfg,
		auth.VerifyAuth(socketConnTabPlugin.authAndLogon),
		socketConnTabPlugin,
		proxy.Proxy(logic.ProxySelector),
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
			verGroup.RouteCallFunc((*gw).Hosts)
			discoveryService.ExcludeApi(verGroup.RouteCallFunc((*gw).SocketTotal))
			discoveryService.ExcludeApi(verGroup.RouteCallFunc((*gw).SocketPush))
			discoveryService.ExcludeApi(verGroup.RouteCallFunc((*gw).SocketMpush))
			discoveryService.ExcludeApi(verGroup.RouteCallFunc((*gw).SocketKick))
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
