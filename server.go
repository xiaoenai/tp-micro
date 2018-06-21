// Copyright 2018 HenryLee. All Rights Reserved.
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

package micro

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	binder "github.com/henrylee2cn/tp-ext/plugin-binder"
	heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"
)

// SrvConfig server config
// Note:
//  yaml tag is used for github.com/henrylee2cn/cfgo
//  ini tag is used for github.com/henrylee2cn/ini
type SrvConfig struct {
	Network           string        `yaml:"network"              ini:"network"              comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
	ListenAddress     string        `yaml:"listen_address"       ini:"listen_address"       comment:"Listen address; for server role"`
	TlsCertFile       string        `yaml:"tls_cert_file"        ini:"tls_cert_file"        comment:"TLS certificate file path"`
	TlsKeyFile        string        `yaml:"tls_key_file"         ini:"tls_key_file"         comment:"TLS key file path"`
	DefaultSessionAge time.Duration `yaml:"default_session_age"  ini:"default_session_age"  comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	DefaultContextAge time.Duration `yaml:"default_context_age"  ini:"default_context_age"  comment:"Default PULL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	SlowCometDuration time.Duration `yaml:"slow_comet_duration"  ini:"slow_comet_duration"  comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
	DefaultBodyCodec  string        `yaml:"default_body_codec"   ini:"default_body_codec"   comment:"Default body codec type id"`
	PrintDetail       bool          `yaml:"print_detail"         ini:"print_detail"         comment:"Is print body and metadata or not"`
	CountTime         bool          `yaml:"count_time"           ini:"count_time"           comment:"Is count cost time or not"`
	EnableHeartbeat   bool          `yaml:"enable_heartbeat"     ini:"enable_heartbeat"     comment:"enable heartbeat"`
}

// Reload Bi-directionally synchronizes config between YAML file and memory.
func (s *SrvConfig) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if len(s.Network) == 0 {
		s.Network = "tcp"
	}
	if len(s.ListenAddress) == 0 {
		s.ListenAddress = "0.0.0.0:9090"
	}
	return err
}

// ListenPort returns the listened port, such as '9090'.
func (s *SrvConfig) ListenPort() string {
	_, port, err := net.SplitHostPort(s.ListenAddress)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	return port
}

// InnerIpPort returns the service's intranet address, such as '192.168.1.120:8080'.
func (s *SrvConfig) InnerIpPort() string {
	hostPort, err := InnerIpPort(s.ListenPort())
	if err != nil {
		tp.Fatalf("%v", err)
	}
	return hostPort
}

// OuterIpPort returns the service's extranet address, such as '113.116.141.121:8080'.
func (s *SrvConfig) OuterIpPort() string {
	hostPort, err := OuterIpPort(s.ListenPort())
	if err != nil {
		tp.Fatalf("%v", err)
	}
	return hostPort
}

func (s *SrvConfig) peerConfig() tp.PeerConfig {
	host, port, err := net.SplitHostPort(s.ListenAddress)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	listenPort, _ := strconv.Atoi(port)
	return tp.PeerConfig{
		DefaultSessionAge: s.DefaultSessionAge,
		DefaultContextAge: s.DefaultContextAge,
		SlowCometDuration: s.SlowCometDuration,
		DefaultBodyCodec:  s.DefaultBodyCodec,
		PrintDetail:       s.PrintDetail,
		CountTime:         s.CountTime,
		Network:           s.Network,
		LocalIP:           host,
		ListenPort:        uint16(listenPort),
	}
}

// Server server peer
type Server struct {
	peer   tp.Peer
	binder *binder.StructArgsBinder
}

// NewServer creates a server peer.
func NewServer(cfg SrvConfig, globalLeftPlugin ...tp.Plugin) *Server {
	doInit()
	if cfg.EnableHeartbeat {
		globalLeftPlugin = append(globalLeftPlugin, heartbeat.NewPong())
	}
	peer := tp.NewPeer(cfg.peerConfig(), globalLeftPlugin...)
	binder := binder.NewStructArgsBinder(nil)
	peer.PluginContainer().AppendRight(binder)
	if len(cfg.TlsCertFile) > 0 && len(cfg.TlsKeyFile) > 0 {
		err := peer.SetTlsConfigFromFile(cfg.TlsCertFile, cfg.TlsKeyFile)
		if err != nil {
			tp.Fatalf("%v", err)
		}
	}
	s := &Server{
		peer:   peer,
		binder: binder,
	}
	s.SetBindErrorFunc(nil)
	return s
}

// Peer returns the peer
func (s *Server) Peer() tp.Peer {
	return s.peer
}

// PluginContainer returns the global plugin container.
func (s *Server) PluginContainer() *tp.PluginContainer {
	return s.peer.PluginContainer()
}

// SetBindErrorFunc sets the binding or balidating error function.
// Note: If fn=nil, set as default.
func (s *Server) SetBindErrorFunc(fn binder.ErrorFunc) {
	if fn != nil {
		s.binder.SetErrorFunc(fn)
		return
	}
	s.binder.SetErrorFunc(func(handlerName, paramName, reason string) *tp.Rerror {
		return tp.NewRerror(
			400,
			"Invalid Parameter",
			fmt.Sprintf(`{"handler": %q, "param": %q, "reason": %q}`, handlerName, paramName, reason),
		)
	})
}

// Router returns the root router of pull or push handlers.
func (s *Server) Router() *tp.Router {
	return s.peer.Router()
}

// SubRoute adds handler group.
func (s *Server) SubRoute(pathPrefix string, plugin ...tp.Plugin) *tp.SubRouter {
	return s.peer.SubRoute(pathPrefix, plugin...)
}

// RoutePull registers PULL handlers, and returns the paths.
func (s *Server) RoutePull(ctrlStruct interface{}, plugin ...tp.Plugin) []string {
	return s.peer.RoutePull(ctrlStruct, plugin...)
}

// RoutePullFunc registers PULL handler, and returns the path.
func (s *Server) RoutePullFunc(pullHandleFunc interface{}, plugin ...tp.Plugin) string {
	return s.peer.RoutePullFunc(pullHandleFunc, plugin...)
}

// RoutePush registers PUSH handlers, and returns the paths.
func (s *Server) RoutePush(ctrlStruct interface{}, plugin ...tp.Plugin) []string {
	return s.peer.RoutePush(ctrlStruct, plugin...)
}

// RoutePushFunc registers PUSH handler, and returns the path.
func (s *Server) RoutePushFunc(pushHandleFunc interface{}, plugin ...tp.Plugin) string {
	return s.peer.RoutePushFunc(pushHandleFunc, plugin...)
}

// SetUnknownPull sets the default handler,
// which is called when no handler for PULL is found.
func (s *Server) SetUnknownPull(fn func(tp.UnknownPullCtx) (interface{}, *tp.Rerror), plugin ...tp.Plugin) {
	s.peer.SetUnknownPull(fn, plugin...)
}

// SetUnknownPush sets the default handler,
// which is called when no handler for PUSH is found.
func (s *Server) SetUnknownPush(fn func(tp.UnknownPushCtx) *tp.Rerror, plugin ...tp.Plugin) {
	s.peer.SetUnknownPush(fn, plugin...)
}

// Close closes server.
func (s *Server) Close() error {
	return s.peer.Close()
}

// CountSession returns the number of sessions.
func (s *Server) CountSession() int {
	return s.peer.CountSession()
}

// GetSession gets the session by id.
func (s *Server) GetSession(sessionId string) (tp.Session, bool) {
	return s.peer.GetSession(sessionId)
}

// ListenAndServe turns on the listening service.
func (s *Server) ListenAndServe(protoFunc ...socket.ProtoFunc) error {
	return s.peer.ListenAndServe(protoFunc...)
}

// RangeSession ranges all sessions. If fn returns false, stop traversing.
func (s *Server) RangeSession(fn func(sess tp.Session) bool) {
	s.peer.RangeSession(fn)
}

// ServeConn serves the connection and returns a session.
func (s *Server) ServeConn(conn net.Conn, protoFunc ...socket.ProtoFunc) (tp.Session, error) {
	return s.peer.ServeConn(conn, protoFunc...)
}
