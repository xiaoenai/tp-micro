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
	"net"
	"net/http"

	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
	"github.com/valyala/fasthttp"
)

// HttpSrvConfig config of HTTP server
type HttpSrvConfig struct {
	ListenAddress string `yaml:"listen_address"`
	OuterHost     string `yaml:"outer_host"`
	TlsCertFile   string `yaml:"tls_cert_file"`
	TlsKeyFile    string `yaml:"tls_key_file"`
	AllowCross    bool   `yaml:"allow_cross"`
}

// ListenPort returns the listened port, such as '8080'.
func (h *HttpSrvConfig) ListenPort() string {
	_, port, err := net.SplitHostPort(h.ListenAddress)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	return port
}

// InnerIpPort returns the service's intranet address, such as '192.168.1.120:8080'.
func (h *HttpSrvConfig) InnerIpPort() string {
	hostPort, err := ant.InnerIpPort(h.ListenPort())
	if err != nil {
		tp.Fatalf("%v", err)
	}
	return hostPort
}

// OuterIpPort returns the service's extranet address, such as '113.116.141.121:8080'.
func (h *HttpSrvConfig) OuterIpPort() string {
	if len(h.OuterHost) == 0 {
		h.OuterHost, _ = ant.OuterIpPort(h.ListenPort())
	}
	return h.OuterHost
}

// Serve starts HTTP gateway service.
func Serve(srvCfg HttpSrvConfig) {
	var tlsConfig *tls.Config
	var err error
	if len(srvCfg.TlsCertFile) > 0 && len(srvCfg.TlsKeyFile) > 0 {
		tlsConfig, err = tp.NewTlsConfigFromFile(srvCfg.TlsCertFile, srvCfg.TlsKeyFile)
		if err != nil {
			tp.Fatalf("%v", err)
		}
	}

	lis, err := tp.NewInheritListener("tcp", srvCfg.ListenAddress, tlsConfig)
	if err != nil {
		tp.Fatalf("%v", err)
	}

	allowCross = srvCfg.AllowCross

	var network = "http"
	if tlsConfig != nil {
		network = "https"
	}
	addr := lis.Addr().String()
	tp.Printf("register http handler: %s", gwHostsUri)
	tp.Printf("listen ok (network:%s, addr:%s)", network, addr)

	err = (&fasthttp.Server{
		Name:    "ants-gateway",
		Handler: handler,
	}).Serve(lis)

	if err != nil && err != http.ErrServerClosed {
		tp.Fatalf("%v", err)
	}
}
