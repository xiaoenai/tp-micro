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

package long

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/logic/client"
)

var srv *ant.Server

// Serve starts TCP gateway service.
func Serve(srvCfg ant.SrvConfig, protoFunc socket.ProtoFunc, innerAddr string) {
	srv = ant.NewServer(
		srvCfg,
		discovery.ServicePluginFromEtcd(innerAddr, client.EtcdClient()),
		plugin.VerifyAuth(connTabPlugin.logon),
		connTabPlugin,
		plugin.Proxy(client.ProxyClient()),
		new(DNS),
	)
	srv.Listen(protoFunc)
}
