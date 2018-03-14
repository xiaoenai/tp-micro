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
	"time"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/client"
	"github.com/xiaoenai/ants/gateway/sdk"
)

var (
	outerPeer tp.Peer
)

// Serve starts TCP gateway service.
func Serve(outerSrvCfg, innerSrvCfg ant.SrvConfig, protoFunc socket.ProtoFunc) {
	sdk.SetApiVersion(logic.ApiVersion())

	outerServer := ant.NewServer(
		outerSrvCfg,
		plugin.VerifyAuth(socketConnTabPlugin.logon),
		socketConnTabPlugin,
		plugin.Proxy(client.ProxyClient()),
		logic.PreWritePushPlugin(),
	)
	outerPeer = outerServer.Peer()

	outerAddr := outerSrvCfg.OuterIpPort()
	innerAddr := innerSrvCfg.InnerIpPort()
	initHosts(outerAddr, innerAddr)

	innerServer := ant.NewServer(
		innerSrvCfg,
		discovery.ServicePluginFromEtcd(
			innerAddr,
			client.EtcdClient(),
			"/gw_hosts",
		),
		hosts,
	)

	innerServer.RoutePullFunc(GwHosts)

	gwGroup := innerServer.SubRoute("/gw")
	{
		verGroup := gwGroup.SubRoute(logic.ApiVersion())
		{
			verGroup.RoutePullFunc((*gw).SocketTotal)
			verGroup.RoutePullFunc((*gw).SocketPush)
		}
	}

	go outerServer.Listen(protoFunc)
	time.Sleep(time.Second)
	go innerServer.Listen(protoFunc)
	select {}
}
