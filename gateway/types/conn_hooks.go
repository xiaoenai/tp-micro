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
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/socket"
)

// LongConnHooks TCP socket connecting event hooks
type LongConnHooks interface {
	// OnLogon is called when the client goes online.
	OnLogon(AccessToken, plugin.AuthSession) *tp.Rerror
	// OnLogoff is called when the client goes offline.
	OnLogoff(tp.BaseSession) *tp.Rerror
	// GetSession returns session from peer by uid.
	GetSession(peer tp.Peer, uid string) (tp.Session, *tp.Rerror)
	//PreWritePush is executed before writing PUSH packet.
	PreWritePush(tp.WriteCtx) *tp.Rerror
}

// ShortConnHooks HTTP connecting event hooks
type ShortConnHooks interface {
	// OnRequest is called when the client requests.
	OnRequest(AccessToken, RequestArgs) ([]socket.PacketSetting, *tp.Rerror)
}

// DefaultLongConnHooks creates a new default LongConnHooks object.
func DefaultLongConnHooks() LongConnHooks {
	return new(defLongConnHooks)
}

type defLongConnHooks struct{}

func (d *defLongConnHooks) OnLogon(accessToken AccessToken, sess plugin.AuthSession) *tp.Rerror {
	sess.SetId(accessToken.Uid())
	return nil
}

func (d *defLongConnHooks) OnLogoff(tp.BaseSession) *tp.Rerror {
	return nil
}

var (
	rerrNotOnline = tp.NewRerror(404, "Not Found", "User is not online")
)

func (d *defLongConnHooks) GetSession(peer tp.Peer, uid string) (tp.Session, *tp.Rerror) {
	sess, ok := peer.GetSession(uid)
	if !ok {
		return nil, rerrNotOnline
	}
	return sess, nil
}

func (d *defLongConnHooks) PreWritePush(tp.WriteCtx) *tp.Rerror {
	return nil
}

// DefaultShortConnHooks creates a new default ShortConnHooks object.
func DefaultShortConnHooks() ShortConnHooks {
	return new(defShortConnHooks)
}

type defShortConnHooks struct{}

func (d *defShortConnHooks) OnRequest(AccessToken, RequestArgs) ([]socket.PacketSetting, *tp.Rerror) {
	return nil, nil
}
