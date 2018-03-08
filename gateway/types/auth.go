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
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

// AccessTokenMgr access token manager
type AccessTokenMgr interface {
	// Apply applies for access token.
	Apply(uid string) (AccessToken, *tp.Rerror)
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

// DefaultAccessTokenMgr creates a new default AccessTokenMgr object.
func DefaultAccessTokenMgr() AccessTokenMgr {
	return &defAccessTokenMgr{
		applyFunc: DefaultAccessToken,
	}
}

type defAccessTokenMgr struct {
	applyFunc func(string) (AccessToken, *tp.Rerror)
}

func (d *defAccessTokenMgr) Apply(uid string) (AccessToken, *tp.Rerror) {
	return d.applyFunc(uid)
}

var (
	rerrInvalidAccessToken = tp.NewRerror(401, "Authorization Failed", "Invalid Access Token")
)

func (d *defAccessTokenMgr) Verify(accessToken string) (AccessToken, *tp.Rerror) {
	if len(accessToken) < 24 {
		return nil, rerrInvalidAccessToken
	}
	return defAccessToken(accessToken), nil
}

// DefaultAccessToken creates a new default AccessToken object.
func DefaultAccessToken(uid string) (AccessToken, *tp.Rerror) {
	return defAccessToken(goutil.URLRandomString(24) + uid), nil
}

type defAccessToken string

func (d defAccessToken) Uid() string {
	return string(d[24+1:])
}

func (d defAccessToken) String() string {
	return string(d)
}
