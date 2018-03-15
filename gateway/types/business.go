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

// Business implement your real business logic
type Business struct {
	// AccessTokenMgr access token manager
	AccessTokenMgr
	// SocketHooks TCP socket connecting event hooks
	SocketHooks
	// HttpHooks HTTP connecting event hooks
	HttpHooks
	// ProxyHooks proxy hooks
	ProxyHooks
}

// DefaultBusiness creates a new default Business object.
func DefaultBusiness() *Business {
	biz := new(Business)
	biz.Init()
	return biz
}

func (biz *Business) Init() {
	if biz.AccessTokenMgr == nil {
		biz.AccessTokenMgr = DefaultAccessTokenMgr()
	}
	if biz.SocketHooks == nil {
		biz.SocketHooks = DefaultSocketHooks()
	}
	if biz.HttpHooks == nil {
		biz.HttpHooks = DefaultHttpHooks()
	}
	if biz.ProxyHooks == nil {
		biz.ProxyHooks = DefaultProxyHooks()
	}
}
