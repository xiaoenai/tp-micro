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

package types

import (
	"github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/erpc/v6/plugin/auth"
	"github.com/valyala/fasthttp"
	micro "github.com/xiaoenai/tp-micro/v6"
)

// SocketHooks TCP socket connecting event hooks
type SocketHooks interface {
	// OnLogon is called when the client goes online.
	OnLogon(auth.Session, AccessToken) *erpc.Status
	// OnLogoff is called when the client goes offline.
	OnLogoff(erpc.BaseSession) *erpc.Status
	// GetSession returns session from peer by uid.
	GetSession(peer erpc.Peer, uid string) (erpc.Session, *erpc.Status)
	//PreWritePush is executed before writing PUSH packet.
	PreWritePush(erpc.WriteCtx) *erpc.Status
}

type (
	// HttpHooks HTTP connecting event hooks
	HttpHooks interface {
		// OnRequest is called when the client requests.
		OnRequest(params RequestArgs, body []byte, authFunc AuthFunc) (AccessToken, []erpc.MessageSetting, *erpc.Status)
	}
	// RequestArgs http query parameters
	RequestArgs interface {
		// Query returns query arguments from request URI.
		QueryArgs() *fasthttp.Args
		// Header returns the header object of request.
		Header() *fasthttp.RequestHeader
	}
)

// WebSocketHooks WEB socket connecting event hooks
type WebSocketHooks interface {
	// OnLogon is called when the client goes online.
	OnLogon(auth.Session, AccessToken) *erpc.Status
	// OnLogoff is called when the client goes offline.
	OnLogoff(erpc.BaseSession) *erpc.Status
	// GetSession returns session from peer by uid.
	GetSession(peer erpc.Peer, uid string) (erpc.Session, *erpc.Status)
	//PreWritePush is executed before writing PUSH packet.
	PreWritePush(erpc.WriteCtx) *erpc.Status
}

// DefaultSocketHooks creates a new default SocketHooks object.
func DefaultSocketHooks() SocketHooks {
	return new(defSocketHooks)
}

// DefaultWebSocketHooks creates a new default WebSocketHooks object.
func DefaultWebSocketHooks() WebSocketHooks {
	return new(defSocketHooks)
}

type defSocketHooks struct{}

func (d *defSocketHooks) OnLogon(sess auth.Session, accessToken AccessToken) *erpc.Status {
	sess.SetID(accessToken.SessionId())
	return nil
}

func (d *defSocketHooks) OnLogoff(erpc.BaseSession) *erpc.Status {
	return nil
}

func (d *defSocketHooks) GetSession(peer erpc.Peer, uid string) (erpc.Session, *erpc.Status) {
	sess, ok := peer.GetSession(uid)
	if !ok {
		return nil, micro.RerrNotOnline
	}
	return sess, nil
}

func (d *defSocketHooks) PreWritePush(erpc.WriteCtx) *erpc.Status {
	return nil
}

// DefaultHttpHooks creates a new default HttpHooks object.
func DefaultHttpHooks() HttpHooks {
	return new(defHttpHooks)
}

type defHttpHooks struct{}

func (d *defHttpHooks) OnRequest(params RequestArgs, body []byte, authFunc AuthFunc) (AccessToken, []erpc.MessageSetting, *erpc.Status) {
	accessToken, stat := authFunc(string(params.QueryArgs().Peek("access_token")))
	return accessToken, nil, stat
}
