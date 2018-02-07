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
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/client"
)

// proxyClient the client for proxy.
type proxyClient struct {
	*ant.Client
}

var caller plugin.Caller = &proxyClient{client.Client()}

func (p *proxyClient) Pull(uri string, args interface{}, reply interface{}, setting ...socket.PacketSetting) tp.PullCmd {
	rerr := logic.ProxyHooks().BeforePull(uri, args, reply, setting...)
	if rerr != nil {
		return tp.NewFakePullCmd(uri, args, reply, rerr)
	}

	return p.Pull(uri, args, reply, setting...)
}

func (p *proxyClient) Push(uri string, args interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	rerr := logic.ProxyHooks().BeforePush(uri, args, setting...)
	if rerr != nil {
		return rerr
	}

	return p.Push(uri, args, setting...)
}
