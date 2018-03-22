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

package sdk

import (
	"path"
	"strings"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	"github.com/xiaoenai/ants/gateway/logic/client"
	"github.com/xiaoenai/ants/gateway/types"
)

// Init initializes a common inner ant client.
func Init(apiVersion string, cliCfg ant.CliConfig, protoFunc socket.ProtoFunc, etcdClient *etcd.Client) {
	client.Init(cliCfg, protoFunc, etcdClient)
	SetApiVersion(apiVersion)
}

var _apiVersion string

// SetApiVersion sets API version
func SetApiVersion(apiVersion string) {
	_apiVersion = strings.TrimRight(path.Join("/", apiVersion), "/")
}

// GwHosts returns the gateway host list.
func GwHosts(setting ...socket.PacketSetting) (*types.GwHosts, *tp.Rerror) {
	var reply = new(types.GwHosts)
	rerr := client.AntClient().Pull("/gw"+_apiVersion+"/hosts", nil, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}

// SocketTotal returns the long connections total of the remote server.
func SocketTotal(srvAddr string, setting ...socket.PacketSetting) (*types.SocketTotalReply, *tp.Rerror) {
	var reply = new(types.SocketTotalReply)
	rerr := client.StaticClient(srvAddr).Pull("/gw"+_apiVersion+"/socket_total", nil, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}

// SocketPush pushs the message to the long connection's client user.
func SocketPush(srvAddr string, args *types.SocketPushArgs, setting ...socket.PacketSetting) (*types.SocketPushReply, *tp.Rerror) {
	var reply = new(types.SocketPushReply)
	rerr := client.StaticClient(srvAddr).Pull("/gw"+_apiVersion+"/socket_push", args, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}
