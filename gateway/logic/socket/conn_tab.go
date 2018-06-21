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
	"strings"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/xiaoenai/tp-micro/gateway/logic"
)

type socketConnTab struct{}

var (
	socketConnTabPlugin = (*socketConnTab)(nil)
)

var (
	_ tp.PostReadPullBodyPlugin = socketConnTabPlugin
	_ tp.PostReadPushBodyPlugin = socketConnTabPlugin
	_ tp.PostDisconnectPlugin   = socketConnTabPlugin
)

func (c *socketConnTab) Name() string {
	return "SocketConnTab"
}

func (c *socketConnTab) authAndLogon(authInfo string, sess plugin.AuthSession) *tp.Rerror {
	token, rerr := logic.AuthFunc()(authInfo)
	if rerr != nil {
		return rerr
	}
	info := token.AddedQuery()
	if info != nil && info.Len() > 0 {
		sess.Swap().Store(socketConnTabPlugin, info.String())
	}
	rerr = logic.SocketHooks().OnLogon(sess, token)
	if rerr == nil {
		tp.Tracef("[+SOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.(tp.BaseSession).Id())
	}
	return rerr
}

func (c *socketConnTab) PostReadPullBody(ctx tp.ReadCtx) *tp.Rerror {
	_appendQuery, _ := ctx.Swap().Load(socketConnTabPlugin)
	appendQuery, _ := _appendQuery.(string)
	u := ctx.UriObject()
	u.RawQuery += "&" + appendQuery
	u.RawQuery = strings.Trim(u.RawQuery, "&")
	return nil
}

func (c *socketConnTab) PostReadPushBody(ctx tp.ReadCtx) *tp.Rerror {
	return c.PostReadPullBody(ctx)
}

func (c *socketConnTab) PostDisconnect(sess tp.BaseSession) *tp.Rerror {
	tp.Tracef("[-SOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.Id())
	return logic.SocketHooks().OnLogoff(sess)
}
