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
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/xiaoenai/ants/gateway/logic"
)

type socketConnTab struct{}

var (
	socketConnTabPlugin                         = new(socketConnTab)
	_                   tp.PostDisconnectPlugin = socketConnTabPlugin
)

func (c *socketConnTab) logon(accessToken string, sess plugin.AuthSession) *tp.Rerror {
	token, rerr := logic.AccessTokenMgr().Verify(accessToken)
	if rerr != nil {
		return rerr
	}
	rerr = logic.SocketHooks().OnLogon(sess, token)
	if rerr == nil {
		tp.Tracef("[+SOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.(tp.BaseSession).Id())
	}
	return rerr
}

func (c *socketConnTab) logoff(sess tp.BaseSession) *tp.Rerror {
	tp.Tracef("[-SOCKET_CONN] addr: %s, id: %s", sess.RemoteAddr().String(), sess.Id())
	return logic.SocketHooks().OnLogoff(sess)
}

func (c *socketConnTab) Name() string {
	return "SocketConnTab"
}

func (c *socketConnTab) PostDisconnect(sess tp.BaseSession) *tp.Rerror {
	return c.logoff(sess)
}
