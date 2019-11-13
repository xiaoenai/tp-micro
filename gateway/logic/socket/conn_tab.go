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
	tp "github.com/henrylee2cn/teleport/v6"
	"github.com/henrylee2cn/teleport/v6/plugin/auth"
	"github.com/xiaoenai/tp-micro/v6/gateway/logic"
)

type socketConnTab struct{}

var (
	socketConnTabPlugin = (*socketConnTab)(nil)
)

var (
	_ tp.PostReadCallBodyPlugin = socketConnTabPlugin
	_ tp.PostReadPushBodyPlugin = socketConnTabPlugin
	_ tp.PostDisconnectPlugin   = socketConnTabPlugin
)

func (c *socketConnTab) Name() string {
	return "SocketConnTab"
}

func (c *socketConnTab) authAndLogon(authInfo string, sess auth.Session) *tp.Status {
	token, stat := logic.AuthFunc()(authInfo)
	if stat != nil {
		return stat
	}
	info := token.AddedQuery()
	if info != nil && info.Len() > 0 {
		sess.Swap().Store(socketConnTabPlugin, info.String())
	}
	stat = logic.SocketHooks().OnLogon(sess, token)
	if stat == nil {
		tp.Tracef("[+SOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.(tp.BaseSession).ID())
	}
	return stat
}

func (c *socketConnTab) PostReadCallBody(ctx tp.ReadCtx) *tp.Status {
	_appendQuery, _ := ctx.Swap().Load(socketConnTabPlugin)
	appendQuery, _ := _appendQuery.(string)
	tp.WithAddMeta("", appendQuery)
	return nil
}

func (c *socketConnTab) PostReadPushBody(ctx tp.ReadCtx) *tp.Status {
	return c.PostReadCallBody(ctx)
}

func (c *socketConnTab) PostDisconnect(sess tp.BaseSession) *tp.Status {
	tp.Tracef("[-SOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.ID())
	return logic.SocketHooks().OnLogoff(sess)
}
