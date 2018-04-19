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
)

type (
	// AuthFunc Verifies access token
	AuthFunc func(authInfo string) (AccessToken, *tp.Rerror)
	// AccessToken access token info
	AccessToken interface {
		// Uid returns the user id.
		Uid() string
		// String returns the access token string.
		String() string
	}
)

// DefaultAuthFunc returns the default authorization function for access behavior.
func DefaultAuthFunc() AuthFunc {
	return defAuthFunc
}

func defAuthFunc(authInfo string) (AccessToken, *tp.Rerror) {
	return defAccessToken(authInfo), nil
}

type defAccessToken string

func (d defAccessToken) Uid() string {
	return string(d)
}

func (d defAccessToken) String() string {
	return string(d)
}
