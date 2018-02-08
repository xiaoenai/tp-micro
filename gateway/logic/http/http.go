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
	"crypto/tls"
	"net/http"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/valyala/fasthttp"
	"github.com/xiaoenai/ants/gateway/logic/client"
)

// OuterHttpSrvConfig config of HTTP server
type OuterHttpSrvConfig struct {
	ListenAddress string `yaml:"listen_address"`
	TlsCertFile   string `yaml:"tls_cert_file"`
	TlsKeyFile    string `yaml:"tls_key_file"`
}

// Serve starts HTTP gateway service.
func Serve(srvCfg OuterHttpSrvConfig) {
	var tlsConfig *tls.Config
	var err error
	if len(srvCfg.TlsCertFile) > 0 && len(srvCfg.TlsKeyFile) > 0 {
		tlsConfig, err = tp.NewTlsConfigFromFile(srvCfg.TlsCertFile, srvCfg.TlsKeyFile)
		if err != nil {
			tp.Fatalf("%v", err)
		}
	}
	ln, err := tp.NewInheritListener("tcp", srvCfg.ListenAddress, tlsConfig)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	err = fasthttp.Serve(ln, handler)
	if err != nil && err != http.ErrServerClosed {
		tp.Fatalf("%v", err)
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	var h = ctx.Request.Header
	var settings = make([]socket.PacketSetting, 0, h.Len()+2)

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
		ctx.Response.Header.Set("Content-Type", "application/json")
		b, _ := rerr.MarshalJSON()
		msg := goutil.BytesToString(b)
		if rerr.Code < 200 {
			ctx.Error(msg, 500)
		} else {
			ctx.Error(msg, int(rerr.Code))
		}
		return
	}

	// succ
	pullcmd.InputMeta().VisitAll(func(key, value []byte) {
		ctx.Response.Header.Add(goutil.BytesToString(key), goutil.BytesToString(value))
	})
	ctx.Response.Header.SetBytesV("Content-Type", contentType)
	ctx.SetBody(reply)
}
