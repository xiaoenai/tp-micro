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
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/ants/gateway/logic/client"
	"github.com/xiaoenai/ants/gateway/types"
)

// Init initializes a common inner ant client.
var Init = client.Init

// TotalLongConn returns the long connections total of the remote server.
func TotalLongConn(srvAddr string, setting ...socket.PacketSetting) (*types.TotalLongConnReply, *tp.Rerror) {
	var reply = new(types.TotalLongConnReply)
	rerr := client.StaticClient(srvAddr).Pull("/gateway/long_conn/total", nil, reply, setting...).Rerror()
	if rerr != nil {
		return nil, rerr
	}
	return reply, nil
}
