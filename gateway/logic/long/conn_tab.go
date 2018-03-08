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

package long

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/xiaoenai/ants/gateway/logic"
)

type connTab struct{}

var (
	connTabPlugin                         = new(connTab)
	_             tp.PostDisconnectPlugin = connTabPlugin
)

func (c *connTab) logon(accessToken string, sess plugin.AuthSession) *tp.Rerror {
	tp.Debugf("verify-auth: id: %s, info: %s", sess.Id(), accessToken)
	token, rerr := logic.AccessTokenVerifier()(accessToken)
	if rerr != nil {
		return rerr
	}
	return logic.LongConnHooks().OnLogon(token, sess)
}

func (c *connTab) logoff(sess tp.BaseSession) *tp.Rerror {
	tp.Tracef("[-CONN] ip: %s, id: %s", sess.RemoteIp(), sess.Id())
	return logic.LongConnHooks().OnLogoff(sess)
}

func (c *connTab) Name() string {
	return "connTab"
}

func (c *connTab) PostDisconnect(sess tp.BaseSession) *tp.Rerror {
	return c.logoff(sess)
}
