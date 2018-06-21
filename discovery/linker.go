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
	"context"
	"encoding/json"
	"net"
	"sync"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

const (
	linkerName = "TP-LINKER(ETCD)"
)

// Node a service node info.
type Node struct {
	Addr string
	Info *ServiceInfo
	mu   sync.RWMutex
}

type linker struct {
	client      *etcd.Client
	nodes       goutil.Map
	uriPaths    goutil.Map
	offlineChan chan string
	innerIp     string
}

// NewLinker creates a etct service linker.
// Note:
// If etcdConfig.DialTimeout<0, it means unlimit;
// If etcdConfig.DialTimeout=0, use the default value(15s).
func NewLinker(etcdConfig etcd.EasyConfig) micro.Linker {
	etcdClient, err := etcd.EasyNew(etcdConfig)
	if err != nil {
		tp.Fatalf("%s: %v", linkerName, err)
		return nil
	}
	return NewLinkerFromEtcd(etcdClient)
}

// NewLinkerFromEtcd creates a etct service linker.
func NewLinkerFromEtcd(etcdClient *etcd.Client) micro.Linker {
	innerIp, err := goutil.IntranetIP()
	if err != nil {
		tp.Fatalf("%s: %v", linkerName, err)
	}
	l := &linker{
		client:      etcdClient,
		nodes:       goutil.AtomicMap(),
		uriPaths:    goutil.AtomicMap(),
		offlineChan: make(chan string, 256),
		innerIp:     innerIp,
	}
	if err := l.initNodes(); err != nil {
		tp.Fatalf("%s: %v", linkerName, err)
	}
	go l.watchNodes()
	return l
}

func (l *linker) getHostport(key string) (string, error) {
	addr := getHostport(key)
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	// Use the loopback address when on the same host
	if host == l.innerIp {
		return "127.0.0.1:" + port, nil
	}
	return addr, nil
}

func (l *linker) addNode(key string, info *ServiceInfo) {
	addr, err := l.getHostport(key)
	if err != nil {
		return
	}
	node := &Node{
		Addr: addr,
		Info: info,
	}
	l.nodes.Store(addr, node)
	var (
		v          interface{}
		ok         bool
		uriPathMap goutil.Map
	)
	for _, uriPath := range info.UriPaths {
		if v, ok = l.uriPaths.Load(uriPath); !ok {
			uriPathMap = goutil.RwMap(1)
			uriPathMap.Store(addr, node)
			l.uriPaths.Store(uriPath, uriPathMap)
		} else {
			uriPathMap = v.(goutil.Map)
			uriPathMap.Store(addr, node)
		}
	}
}

func (l *linker) delNode(key string) {
	addr, _ := l.getHostport(key)
	_node, ok := l.nodes.Load(addr)
	if !ok {
		return
	}
	l.nodes.Delete(addr)
	for _, uriPath := range _node.(*Node).Info.UriPaths {
		_uriPathMap, ok := l.uriPaths.Load(uriPath)
		if !ok {
			continue
		}
		uriPathMap := _uriPathMap.(goutil.Map)
		if _, ok := uriPathMap.Load(addr); ok {
			uriPathMap.Delete(addr)
			if uriPathMap.Len() == 0 {
				l.uriPaths.Delete(uriPath)
			}
		}
	}
	l.offlineChan <- addr
}

func (l *linker) initNodes() error {
	resp, err := l.client.Get(context.TODO(), serviceNamespace, etcd.WithPrefix())
	if err != nil || len(resp.Kvs) == 0 {
		return err
	}
	for _, kv := range resp.Kvs {
		l.addNode(string(kv.Key), getServiceInfo(kv.Value))
		tp.Infof("%s: INIT %q : %q\n", linkerName, kv.Key, kv.Value)
	}
	return nil
}

func (l *linker) watchNodes() {
	rch := l.client.Watch(context.TODO(), serviceNamespace, etcd.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case etcd.EventTypePut:
				l.addNode(string(ev.Kv.Key), getServiceInfo(ev.Kv.Value))
				tp.Infof("%s: %s %q : %q\n", linkerName, ev.Type, ev.Kv.Key, ev.Kv.Value)
			case etcd.EventTypeDelete:
				l.delNode(string(ev.Kv.Key))
				tp.Infof("%s: %s %q\n", linkerName, ev.Type, ev.Kv.Key)
			}
		}
	}
}

func getServiceInfo(value []byte) *ServiceInfo {
	info := &ServiceInfo{}
	err := json.Unmarshal(value, info)
	if err != nil {
		tp.Errorf("%s", err.Error())
	}
	return info
}

// Select selects a service address by URI path.
func (l *linker) Select(uriPath string, exclude map[string]struct{}) (string, *tp.Rerror) {
	iface, exist := l.uriPaths.Load(uriPath)
	if !exist {
		return "", micro.NotFoundService
	}
	nodes := iface.(goutil.Map)
	var addr string
	for i := 0; i < nodes.Len(); i++ {
		if _, iface, exist = nodes.Random(); exist {
			addr = iface.(*Node).Addr
			if _, exist = exclude[addr]; !exist {
				return addr, nil
			}
		}
	}
	return "", micro.NotFoundService
}

// WatchOffline pushs service node offline notification.
func (l *linker) WatchOffline() <-chan string {
	return l.offlineChan
}

// Len returns the number of nodes corresponding to the URI.
func (l *linker) Len(uriPath string) int {
	nodes, exist := l.uriPaths.Load(uriPath)
	if !exist {
		return 0
	}
	return nodes.(goutil.Map).Len()
}

// Close closes the linker.
func (l *linker) Close() {
	close(l.offlineChan)
	l.client.Close()
}
