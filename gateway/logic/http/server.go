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
	"crypto/tls"
	"net/http"

	tp "github.com/henrylee2cn/teleport"
	"github.com/valyala/fasthttp"
)

// OuterHttpSrvConfig config of HTTP server
type OuterHttpSrvConfig struct {
	ListenAddress string `yaml:"listen_address"`
	TlsCertFile   string `yaml:"tls_cert_file"`
	TlsKeyFile    string `yaml:"tls_key_file"`
	AllowCross    bool   `yaml:"allow_cross"`
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
	allowCross = srvCfg.AllowCross
	s := &fasthttp.Server{
		Name:    "ants-gateway",
		Handler: handler,
	}
	err = s.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		tp.Fatalf("%v", err)
	}
}
