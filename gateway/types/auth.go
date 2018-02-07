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

// AccessTokenMgr access token manager
type AccessTokenMgr interface {
	// Apply applies for access token.
	Apply() (AccessToken, *tp.Rerror)
	// Verify checks access token.
	Verify(accessToken string) (AccessToken, *tp.Rerror)
}

// AccessToken access token info
type AccessToken interface {
	// Uid returns the user id.
	Uid() string
	// String returns the access token string.
	String() string
}
