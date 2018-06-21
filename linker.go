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
	tp "github.com/henrylee2cn/teleport"
)

// NotFoundService error: not found service in linker.
var NotFoundService = tp.NewRerror(tp.CodeNotFound, "Not Found Service", "not found service")

// Linker linker for client.
type Linker interface {
	// Select selects a service address by URI path.
	Select(uriPath string, exclude map[string]struct{}) (addr string, rerr *tp.Rerror)
	// Len returns the number of nodes corresponding to the URI.
	Len(uriPath string) int
	// WatchOffline pushs service node offline notification.
	WatchOffline() <-chan string
	// Close closes the linker.
	Close()
}

// static linker

// NewStaticLinker creates a static linker.
func NewStaticLinker(srvAddr string) Linker {
	return &staticLinker{
		srvAddr: srvAddr,
		ch:      make(chan string),
	}
}

type staticLinker struct {
	srvAddr string
	ch      chan string
}

// Select selects a service address by URI path.
func (d *staticLinker) Select(uriPath string, exclude map[string]struct{}) (string, *tp.Rerror) {
	return d.srvAddr, nil
}

// Len returns the number of nodes corresponding to the URI.
func (d *staticLinker) Len(uriPath string) int {
	return 1
}

// WatchOffline pushs service node offline notification.
func (d *staticLinker) WatchOffline() <-chan string {
	return d.ch
}

// Close closes the linker.
func (d *staticLinker) Close() {
	close(d.ch)
}
