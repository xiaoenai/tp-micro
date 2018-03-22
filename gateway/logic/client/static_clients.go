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

package client

import (
	"sync"

	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/henrylee2cn/tp-micro"
)

// StaticClients static clients map
type StaticClients struct {
	clients   map[string]*micro.Client
	cfg       micro.CliConfig
	protoFunc socket.ProtoFunc
	mu        sync.RWMutex
}

var staticClients *StaticClients

// StaticClient returns the client whose server address is srvAddr.
// If the client does not exist, set and return it.
func StaticClient(srvAddr string) *micro.Client {
	return staticClients.GetOrSet(srvAddr)
}

// newStaticClients creates a static clients map.
func newStaticClients(cfg micro.CliConfig, protoFunc socket.ProtoFunc) *StaticClients {
	return &StaticClients{
		clients:   make(map[string]*micro.Client),
		cfg:       cfg,
		protoFunc: protoFunc,
	}
}

// Set sets the client whose server address is srvAddr.
func (s *StaticClients) Set(srvAddr string) {
	s.mu.Lock()
	cli := micro.NewClient(s.cfg, micro.NewStaticLinker(srvAddr))
	cli.SetProtoFunc(s.protoFunc)
	s.clients[srvAddr] = cli
	s.mu.Unlock()
}

// GetOrSet returns the client whose server address is srvAddr.
// If the client does not exist, set and return it.
func (s *StaticClients) GetOrSet(srvAddr string) *micro.Client {
	s.mu.RLock()
	cli, ok := s.clients[srvAddr]
	s.mu.RUnlock()
	if ok {
		return cli
	}
	s.mu.Lock()
	cli, ok = s.clients[srvAddr]
	defer s.mu.Unlock()
	if ok {
		return cli
	}
	cli = micro.NewClient(s.cfg, micro.NewStaticLinker(srvAddr))
	cli.SetProtoFunc(s.protoFunc)
	s.clients[srvAddr] = cli
	return cli
}

// Get returns the client whose server address is srvAddr.
func (s *StaticClients) Get(srvAddr string) (*micro.Client, bool) {
	s.mu.RLock()
	cli, ok := s.clients[srvAddr]
	s.mu.RUnlock()
	return cli, ok
}
