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
	"net"
	"strings"
	"sync"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

// InnerIpPort returns the service's intranet address, such as '192.168.1.120:8080'.
func InnerIpPort(port string) (string, error) {
	host, err := goutil.IntranetIP()
	if err != nil {
		return "", err
	}
	hostPort := net.JoinHostPort(host, port)
	return hostPort, nil
}

// OuterIpPort returns the service's extranet address, such as '113.116.141.121:8080'.
func OuterIpPort(port string) (string, error) {
	host, err := goutil.ExtranetIP()
	if err != nil {
		return "", err
	}
	hostPort := net.JoinHostPort(host, port)
	return hostPort, nil
}

var initOnce sync.Once

func doInit() {
	initOnce.Do(func() {
		go tp.GraceSignal()
	})
}

func getUriPath(uri string) string {
	if idx := strings.Index(uri, "?"); idx != -1 {
		return uri[:idx]
	}
	return uri
}
