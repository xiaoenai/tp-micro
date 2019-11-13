// Copyright 2019 github.com/xiaoenai. All Rights Reserved.
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

package websocket

import (
	tp "github.com/henrylee2cn/teleport/v6"
	"github.com/henrylee2cn/teleport/v6/plugin/auth"
	"github.com/xiaoenai/tp-micro/v6/gateway/logic"
)

type webSocketConnTab struct{}

var (
	webSocketConnTabPlugin = (*webSocketConnTab)(nil)
)

var (
	_ tp.PostReadCallBodyPlugin = webSocketConnTabPlugin
	_ tp.PostReadPushBodyPlugin = webSocketConnTabPlugin
	_ tp.PostDisconnectPlugin   = webSocketConnTabPlugin
)

func (c *webSocketConnTab) Name() string {
	return "WebSocketConnTab"
}

func (c *webSocketConnTab) authAndLogon(authInfo string, sess auth.Session) *tp.Status {
	token, stat := logic.AuthFunc()(authInfo)
	if stat != nil {
		return stat
	}
	info := token.AddedQuery()
	if info != nil && info.Len() > 0 {
		sess.Swap().Store(webSocketConnTabPlugin, info.String())
	}
	stat = logic.WebSocketHooks().OnLogon(sess, token)
	if stat == nil {
		tp.Tracef("[+SOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.(tp.BaseSession).ID())
	}
	return stat
}

func (c *webSocketConnTab) PostReadCallBody(ctx tp.ReadCtx) *tp.Status {
	_appendQuery, _ := ctx.Swap().Load(webSocketConnTabPlugin)
	appendQuery, _ := _appendQuery.(string)
	tp.WithAddMeta("", appendQuery)
	return nil
}

func (c *webSocketConnTab) PostReadPushBody(ctx tp.ReadCtx) *tp.Status {
	return c.PostReadCallBody(ctx)
}

func (c *webSocketConnTab) PostDisconnect(sess tp.BaseSession) *tp.Status {
	tp.Tracef("[-WEBSOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.ID())
	return logic.WebSocketHooks().OnLogoff(sess)
}
