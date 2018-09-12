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
	"github.com/xiaoenai/tp-micro/clientele"
	"github.com/xiaoenai/tp-micro/gateway/types"
)

// Init initializes a common inner ant client.
func Init(apiVersion string, protoFunc tp.ProtoFunc) {
	clientele.SetProtoFunc(protoFunc)
	SetApiVersion(apiVersion)
}

var _apiVersion string

// SetApiVersion sets API version
func SetApiVersion(apiVersion string) {
	_apiVersion = strings.TrimRight(path.Join("/", apiVersion), "/")
}

// GwHosts returns the gateway host list.
func GwHosts(setting ...tp.MessageSetting) (*types.GwHosts, *tp.Rerror) {
	var reply = new(types.GwHosts)
	rerr := clientele.DynamicCall(nil, "/gw"+_apiVersion+"/hosts", nil, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}

// SocketTotal returns the long connections total of the remote server.
func SocketTotal(srvAddr string, setting ...tp.MessageSetting) (*types.SocketTotalReply, *tp.Rerror) {
	var reply = new(types.SocketTotalReply)
	rerr := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/socket_total", nil, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}

// SocketPush pushes message to the specified user.
func SocketPush(srvAddr string, args *types.SocketPushArgs, setting ...tp.MessageSetting) (*types.SocketPushReply, *tp.Rerror) {
	var reply = new(types.SocketPushReply)
	rerr := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/socket_push", args, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}

// SocketMpush multi-push messages to the specified users.
func SocketMpush(srvAddr string, args *types.SocketMpushArgs, setting ...tp.MessageSetting) (*types.SocketMpushReply, *tp.Rerror) {
	var reply = new(types.SocketMpushReply)
	rerr := clientele.StaticCall(nil, srvAddr, "/gw"+_apiVersion+"/socket_mpush", args, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}
