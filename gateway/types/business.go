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
	"github.com/henrylee2cn/teleport/plugin/proxy"
	"github.com/xiaoenai/tp-micro/clientele"
)

// Business implement your real business logic
type Business struct {
	// AuthFunc Verifies access token
	AuthFunc func(accessToken string) (AccessToken, *tp.Rerror)
	// SocketHooks TCP socket connecting event hooks
	SocketHooks
	// HttpHooks HTTP connecting event hooks
	HttpHooks
	// ProxySelector returns proxy caller by label.
	ProxySelector func(*proxy.ProxyLabel) proxy.Forwarder
	// InnerServerPlugins inner server plugins
	InnerServerPlugins []tp.Plugin
}

// DefaultBusiness creates a new default Business object.
func DefaultBusiness() *Business {
	biz := new(Business)
	biz.Init()
	return biz
}

func (biz *Business) Init() {
	if biz.AuthFunc == nil {
		biz.AuthFunc = DefaultAuthFunc()
	}
	if biz.SocketHooks == nil {
		biz.SocketHooks = DefaultSocketHooks()
	}
	if biz.HttpHooks == nil {
		biz.HttpHooks = DefaultHttpHooks()
	}
	if biz.ProxySelector == nil {
		biz.ProxySelector = DefaultProxySelector()
	}
}

// DefaultProxySelector creates a new default proxy caller selector.
func DefaultProxySelector() func(*proxy.ProxyLabel) proxy.Forwarder {
	return func(*proxy.ProxyLabel) proxy.Forwarder {
		return clientele.GetDynamicClient()
	}
}
