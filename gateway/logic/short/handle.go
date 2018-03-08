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

package short

import (
	"bytes"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/valyala/fasthttp"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/client"
)

func handler(ctx *fasthttp.RequestCtx) {
	(&requestHandler{ctx: ctx}).handle()
}

type requestHandler struct {
	ctx       *fasthttp.RequestCtx
	queryArgs *fasthttp.Args
}

func (r *requestHandler) handle() {
	// verify access token
	accessToken := accessTokenGetter(r)
	token, rerr := logic.AccessTokenVerifier()(accessToken)
	if rerr != nil {
		r.replyError(rerr)
		return
	}
	settings, rerr := logic.ShortConnHooks().OnRequest(token, r)
	if rerr != nil {
		r.replyError(rerr)
		return
	}

	var ctx = r.ctx
	var h = ctx.Request.Header

	// set header
	h.VisitAll(func(key, value []byte) {
		settings = append(settings, socket.WithAddMeta(string(key), string(value)))
	})

	// set body codec
	var contentType = h.ContentType()
	bodyCodec := GetBodyCodec(h.ContentType())
	settings = append(settings, socket.WithBodyCodec(bodyCodec))

	// set real ip
	var realIp string
	if xRealIp := h.Peek("X-Real-IP"); len(xRealIp) > 0 {
		realIp = string(xRealIp)
	} else if xForwardedFor := h.Peek("X-Forwarded-For"); len(xForwardedFor) > 0 {
		realIp = string(bytes.Split(xForwardedFor, []byte{','})[0])
	}
	if len(realIp) == 0 {
		realIp = ctx.RemoteAddr().String()
	}
	settings = append(settings, socket.WithAddMeta(tp.MetaRealIp, realIp))

	var bodyBytes = ctx.Request.Body()
	var reply []byte
	var uri = goutil.BytesToString(ctx.Path())

	pullcmd := client.ProxyClient().Pull(uri, bodyBytes, &reply, settings...)

	// fail
	if rerr := pullcmd.Rerror(); rerr != nil {
		r.replyError(rerr)
		return
	}

	// succ
	pullcmd.InputMeta().VisitAll(func(key, value []byte) {
		ctx.Response.Header.Add(goutil.BytesToString(key), goutil.BytesToString(value))
	})
	ctx.Response.Header.SetBytesV("Content-Type", contentType)
	ctx.SetBody(reply)
}

func (r *requestHandler) replyError(rerr *tp.Rerror) {
	r.ctx.Response.Header.Set("Content-Type", "application/json")
	b, _ := rerr.MarshalJSON()
	msg := goutil.BytesToString(b)
	if rerr.Code < 200 {
		r.ctx.Error(msg, 500)
	} else {
		r.ctx.Error(msg, int(rerr.Code))
	}
}

// Query returns query arguments from request URI.
func (r *requestHandler) Query(key string) string {
	if r.queryArgs == nil {
		r.queryArgs = r.ctx.QueryArgs()
	}
	v := r.queryArgs.Peek(key)
	if len(v) == 0 {
		return ""
	}
	return string(v)
}

// Header returns header value for the given key.
func (r *requestHandler) Header(key string) string {
	v := r.ctx.Request.Header.Peek(key)
	if len(v) == 0 {
		return ""
	}
	return string(v)
}

// Cookie returns cookie for the given key.
func (r *requestHandler) Cookie(key string) string {
	v := r.ctx.Request.Header.Cookie(key)
	if len(v) == 0 {
		return ""
	}
	return string(v)
}
