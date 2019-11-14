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

	"github.com/henrylee2cn/erpc/v6"
	"github.com/xiaoenai/tp-micro/v6/clientele"
	"github.com/xiaoenai/tp-micro/v6/gateway/types"
)

// Init initializes a common inner ant client.
func Init(apiVersion string, protoFunc erpc.ProtoFunc) {
	clientele.SetProtoFunc(protoFunc)
	SetApiVersion(apiVersion)
}

var _apiVersion string

// SetApiVersion sets API version
func SetApiVersion(apiVersion string) {
	_apiVersion = strings.TrimRight(path.Join("/", apiVersion), "/")
}

// GwHosts returns the gateway host list.
func GwHosts(setting ...erpc.MessageSetting) (*types.GwHosts, *erpc.Status) {
	var reply = new(types.GwHosts)
	stat := clientele.DynamicCall(nil, "/gw"+_apiVersion+"/hosts", nil, reply, setting...).Status()
	if stat != nil {
		return nil, stat
	}
	return reply, nil
}

// SocketTotal returns the long connections total of the remote server.
func SocketTotal(srvAddr string, setting ...erpc.MessageSetting) (*types.SocketTotalReply, *erpc.Status) {
	var reply = new(types.SocketTotalReply)
	stat := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/socket_total", nil, reply, setting...).Status()
	if stat != nil {
		return nil, stat
	}
	return reply, nil
}

// SocketPush pushes message to the specified user.
func SocketPush(srvAddr string, args *types.SocketPushArgs, setting ...erpc.MessageSetting) (*types.SocketPushReply, *erpc.Status) {
	var reply = new(types.SocketPushReply)
	stat := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/socket_push", args, reply, setting...).Status()
	if stat != nil {
		return nil, stat
	}
	return reply, nil
}

// SocketMpush multi-push messages to the specified users.
func SocketMpush(srvAddr string, args *types.SocketMpushArgs, setting ...erpc.MessageSetting) (*types.SocketMpushReply, *erpc.Status) {
	var reply = new(types.SocketMpushReply)
	stat := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/socket_mpush", args, reply, setting...).Status()
	if stat != nil {
		return nil, stat
	}
	return reply, nil
}

// WsTotal returns the long connections total of the remote server.
func WsTotal(srvAddr string, setting ...erpc.MessageSetting) (*types.WsTotalReply, *erpc.Status) {
	var reply = new(types.WsTotalReply)
	stat := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/ws_total", nil, reply, setting...).Status()
	if stat != nil {
		return nil, stat
	}
	return reply, nil
}

// WsPush pushes message to the specified user.
func WsPush(srvAddr string, args *types.SocketPushArgs, setting ...erpc.MessageSetting) (*types.WsPushReply, *erpc.Status) {
	var reply = new(types.WsPushReply)
	stat := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/ws_push", args, reply, setting...).Status()
	if stat != nil {
		return nil, stat
	}
	return reply, nil
}

// WsMpush multi-push messages to the specified users.
func WsMpush(srvAddr string, args *types.WsMpushArgs, setting ...erpc.MessageSetting) (*types.WsMpushReply, *erpc.Status) {
	var reply = new(types.WsMpushReply)
	stat := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/ws_mpush", args, reply, setting...).Status()
	if stat != nil {
		return nil, stat
	}
	return reply, nil
}
