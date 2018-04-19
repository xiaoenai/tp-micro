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

package http

import (
	"bytes"
	"encoding/json"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/codec"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/valyala/fasthttp"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/hosts"
)

const (
	// SEQ the packet sequence query parameter specified by the client
	SEQ = "seq_"
)

var (
	allowCross bool
	gwHostsUri = "/gw/" + logic.ApiVersion() + "/hosts"
)

func handler(ctx *fasthttp.RequestCtx) {
	(&requestHandler{ctx: ctx}).handle()
}

type requestHandler struct {
	ctx *fasthttp.RequestCtx
}

func (r *requestHandler) handle() {
	// cross
	if allowCross && r.crossDomainFilter() {
		return
	}
	var ctx = r.ctx
	var h = r.Header()
	var contentType = goutil.BytesToString(h.ContentType())
	var bodyCodec = GetBodyCodec(contentType, codec.ID_PLAIN)
	var acceptBodyCodec = GetBodyCodec(goutil.BytesToString(h.Peek("Accept")), bodyCodec)

	var label plugin.ProxyLabel
	label.Uri = goutil.BytesToString(ctx.Path())

	// gw hosts
	if label.Uri == gwHostsUri {
		switch acceptBodyCodec {
		case codec.ID_PROTOBUF:
			b, _ := codec.ProtoMarshal(hosts.GwHosts())
			r.ctx.Success("application/x-protobuf", b)
		default:
			b, _ := json.Marshal(hosts.GwHosts())
			r.ctx.Success("application/json", b)
		}
		return
	}

	var bodyBytes = ctx.Request.Body()

	// verify access token
	accessToken, settings, rerr := logic.HttpHooks().OnRequest(r, bodyBytes, logic.AuthFunc())
	if rerr != nil {
		r.replyError(rerr)
		return
	}

	// set header
	h.VisitAll(func(key, value []byte) {
		settings = append(settings, tp.WithAddMeta(string(key), string(value)))
	})

	// set body codec
	settings = append(settings, tp.WithBodyCodec(bodyCodec))

	// set accept body codec
	if acceptBodyCodec != bodyCodec {
		settings = append(settings, tp.WithAcceptBodyCodec(acceptBodyCodec))
	}

	query := r.ctx.QueryArgs()

	// set session id
	if accessToken == nil {
		label.SessionId = ctx.RemoteAddr().String()
	} else {
		label.SessionId = accessToken.Uid()
		if info := accessToken.Info(); info != nil {
			info.VisitAll(func(key, value []byte) {
				query.AddBytesKV(key, value)
			})
		}
	}

	// set real ip
	if xRealIp := h.Peek("X-Real-IP"); len(xRealIp) > 0 {
		label.RealIp = string(xRealIp)
	} else if xForwardedFor := h.Peek("X-Forwarded-For"); len(xForwardedFor) > 0 {
		label.RealIp = string(bytes.Split(xForwardedFor, []byte{','})[0])
	}
	if len(label.RealIp) == 0 {
		label.RealIp = ctx.RemoteAddr().String()
	}

	settings = append(settings, tp.WithAddMeta(tp.MetaRealIp, label.RealIp))

	// set seq
	if seqBytes := query.Peek(SEQ); len(seqBytes) > 0 {
		settings = append(settings, tp.WithSeq(label.RealIp+"@"+goutil.BytesToString(seqBytes)))
	}

	var reply []byte
	label.Uri += "?" + query.String()

	pullcmd := logic.
		ProxySelector(&label).
		Pull(label.Uri, bodyBytes, &reply, settings...)

	// fail
	if rerr := pullcmd.Rerror(); rerr != nil {
		r.replyError(rerr)
		return
	}

	// succ

	var hasRespContentType bool
	pullcmd.InputMeta().VisitAll(func(key, value []byte) {
		k := goutil.BytesToString(key)
		v := goutil.BytesToString(value)
		if k == "Content-Type" {
			hasRespContentType = true
			ctx.Response.Header.Set(k, v)
		} else {
			ctx.Response.Header.Add(k, v)
		}
	})
	if !hasRespContentType {
		ctx.Response.Header.Add(
			"Content-Type",
			GetContentType(pullcmd.InputBodyCodec(), contentType),
		)
	}
	ctx.SetBody(reply)
}

var optionsBytes = []byte("OPTIONS")

func (r *requestHandler) crossDomainFilter() bool {
	r.ctx.Response.Header.Set("Access-Control-Allow-Origin", string(r.ctx.Request.Header.Peek("Origin")))
	r.ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	r.ctx.Response.Header.Set("Access-Control-Allow-Methods", string(r.ctx.Request.Header.Peek("Access-Control-Request-Method")))
	// r.ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	r.ctx.Response.Header.Set("Access-Control-Allow-Headers", string(r.ctx.Request.Header.Peek("Access-Control-Request-Headers")))
	// r.ctx.Response.Header.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	r.ctx.Response.Header.Set("Access-Control-Max-Age", "-1")
	// r.ctx.Response.Header.Set("Access-Control-Max-Age", "172800")
	if !bytes.Equal(r.ctx.Method(), optionsBytes) {
		return false
	}
	r.ctx.SetStatusCode(204)
	return true
}

func (r *requestHandler) replyError(rerr *tp.Rerror) {
	var statusCode int
	if rerr.Code < 200 {
		// Internal communication error
		statusCode = 500
	} else if rerr.Code < 600 {
		// Custom HTTP error
		statusCode = int(rerr.Code)
	} else {
		// Business error
		statusCode = 299
	}
	msg, _ := rerr.MarshalJSON()
	r.ctx.SetStatusCode(statusCode)
	r.ctx.SetContentType("application/json")
	r.ctx.SetBody(msg)
}

// QueryArgs returns the query arguments object of request.
func (r *requestHandler) QueryArgs() *fasthttp.Args {
	return r.ctx.QueryArgs()
}

// Header returns the header object of request.
func (r *requestHandler) Header() *fasthttp.RequestHeader {
	return &r.ctx.Request.Header
}
