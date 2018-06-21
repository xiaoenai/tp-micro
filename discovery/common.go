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

package discovery

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/henrylee2cn/goutil"
)

var (
	// serviceNamespace the service prefix of ETCD key
	serviceNamespace = "TP-SRV@"
)

// ServiceNamespace returns the service prefix of ETCD key.
func ServiceNamespace() string {
	return serviceNamespace
}

// SetServiceNamespace sets the service prefix of ETCD key.
// Note: It should be called the first time after importing this package.
func SetServiceNamespace(prefix string) {
	serviceNamespace = prefix
}

// ServiceInfo serivce info
type ServiceInfo struct {
	UriPaths []string `json:"uri_paths"`
	mu       sync.RWMutex
}

// String returns the JSON string.
func (s *ServiceInfo) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(s)
	return goutil.BytesToString(b)
}

// Append appends uri path
func (s *ServiceInfo) Append(uriPath ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UriPaths = append(s.UriPaths, uriPath...)
}

func createServiceKey(addr string) string {
	return serviceNamespace + addr
}

func getHostport(serviceKey string) string {
	return strings.TrimPrefix(serviceKey, serviceNamespace)
}
