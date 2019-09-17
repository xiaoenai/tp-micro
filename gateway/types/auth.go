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
	"github.com/henrylee2cn/teleport/utils"
)

type (
	// AuthFunc Verifies access token
	AuthFunc func(authInfo string) (AccessToken, *tp.Status)
	// AccessToken access token info
	AccessToken interface {
		// String returns the access token string.
		String() string
		// SessionId specifies the string as the session ID.
		SessionId() string
		// Uid returns the user id.
		Uid() string
		// DeviceId returns the device id.
		DeviceId() string
		// AddedQuery the user information will be appended to the URI query part.
		AddedQuery() *utils.Args
	}
)

// DefaultAuthFunc returns the default authorization function for access behavior.
func DefaultAuthFunc() AuthFunc {
	return defAuthFunc
}

func defAuthFunc(authInfo string) (AccessToken, *tp.Status) {
	return defAccessToken(authInfo), nil
}

type defAccessToken string

// String returns the access token string.
func (d defAccessToken) String() string {
	return string(d)
}

// SessionId specifies the string as the session ID.
func (d defAccessToken) SessionId() string {
	return string(d)
}

// Uid returns the user id.
func (d defAccessToken) Uid() string {
	return string(d)
}

// DeviceId returns the device id.
func (d defAccessToken) DeviceId() string {
	return string(d)
}

type ctx struct{ tp.UnknownPushCtx }

// AddedQuery the user information will be appended to the URI query part.
func (d defAccessToken) AddedQuery() *utils.Args {
	args := utils.AcquireArgs()
	args.Set("_uid", d.Uid())
	args.Set("_device_id", d.DeviceId())
	return args
}
