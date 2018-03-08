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

package logic

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/ants/gateway/types"
)

var globalBusiness *types.Business

// SetBusiness sets business object.
func SetBusiness(biz *types.Business) {
	if biz == nil {
		biz = types.DefaultBusiness()
	} else {
		biz.Init()
	}
	globalBusiness = biz
}

// ApiVersion returns long connection API version.
func ApiVersion() string {
	return globalBusiness.ApiVersion
}

// AccessTokenMgr returns access token manager.
func AccessTokenMgr() types.AccessTokenMgr {
	return globalBusiness.AccessTokenMgr
}

// LongConnHooks returns TCP socket connecting event hooks.
func LongConnHooks() types.LongConnHooks {
	return globalBusiness.LongConnHooks
}

// ShortConnHooks returns HTTP connecting event hooks.
func ShortConnHooks() types.ShortConnHooks {
	return globalBusiness.ShortConnHooks
}

// ProxyHooks returns proxy hooks.
func ProxyHooks() types.ProxyHooks {
	return globalBusiness.ProxyHooks
}

type perPusher struct {
	fn func(tp.WriteCtx) *tp.Rerror
}

func (p *perPusher) Name() string {
	return "PUSH-LOGIC"
}

func (p *perPusher) PreWritePush(ctx tp.WriteCtx) *tp.Rerror {
	return p.fn(ctx)
}

// PreWritePushPlugin returns PreWritePushPlugin.
func PreWritePushPlugin() tp.PreWritePushPlugin {
	return &perPusher{fn: LongConnHooks().PreWritePush}
}
