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

package tcp

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/socket"
)

var srv *ant.Server

// Serve starts TCP gateway service.
func Serve(srvCfg ant.SrvConfig, protoFunc socket.ProtoFunc, etcdPlugin tp.Plugin) {
	srv = ant.NewServer(
		srvCfg,
		etcdPlugin,
		plugin.VerifyAuth(connTabPlugin.logon),
		connTabPlugin,
		plugin.Proxy(caller),
	)
	srv.Listen(protoFunc)
}
