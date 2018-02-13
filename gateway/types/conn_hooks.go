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
