// Package discovery is the service discovery module implemented by ETCD.
//
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
//
package discovery

import (
	"context"
	"time"

	tp "github.com/henrylee2cn/teleport"
	heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

const (
	// minimum lease TTL is 5-second
	minLeaseTTL = 5
)

// Service automatically registered api info to etcd
type Service struct {
	hostport    string
	serviceKey  string
	allApis     []string
	excludeApis []string
	serviceInfo *ServiceInfo
	client      *etcd.Client
	leaseid     etcd.LeaseID
}

var (
	_ tp.PostRegPlugin    = new(Service)
	_ tp.PostListenPlugin = new(Service)
)

// ServicePlugin creates a teleport plugin which automatically registered api info to etcd.
// Note:
// excludeApis must not be registered to etcd.
// If etcdConfig.DialTimeout<0, it means unlimit;
// If etcdConfig.DialTimeout=0, use the default value(15s).
func ServicePlugin(hostport string, etcdConfig etcd.EasyConfig, excludeApis ...string) *Service {
	s := ServicePluginFromEtcd(hostport, nil, excludeApis...)
	var err error
	s.client, err = etcd.EasyNew(etcdConfig)
	if err != nil {
		tp.Fatalf("%v: %v", err, s.Name())
		return s
	}
	return s
}

// ServicePluginFromEtcd creates a teleport plugin which automatically registered api info to etcd.
// Note:
// excludeApis must not be registered to etcd.
func ServicePluginFromEtcd(hostport string, etcdClient *etcd.Client, excludeApis ...string) *Service {
	s := &Service{
		hostport:    hostport,
		serviceKey:  createServiceKey(hostport),
		client:      etcdClient,
		serviceInfo: new(ServiceInfo),
	}
	s.ExcludeApi(heartbeat.HeartbeatUri)
	s.ExcludeApi(excludeApis...)
	return s
}

// Etcd returns the etcd client.
func (s *Service) Etcd() *etcd.Client {
	return s.client
}

// ExcludeApi appends apis that must not be registered to etcd.
func (s *Service) ExcludeApi(excludeApi ...string) {
	s.excludeApis = append(s.excludeApis, excludeApi...)
}

// Name returns name.
func (s *Service) Name() string {
	return "ETCD(" + s.serviceKey + ")"
}

// PostReg registers URI path.
func (s *Service) PostReg(handler *tp.Handler) error {
	s.allApis = append(s.allApis, handler.Name())
	return nil
}

// PostListen adds serviceInfo, and starts etcd keep alive.
func (s *Service) PostListen() error {
L:
	for _, api := range s.allApis {
		for _, a := range s.excludeApis {
			if a == api {
				continue L
			}
		}
		s.serviceInfo.Append(api)
	}
	ch, err := s.keepAlive()
	if err != nil {
		return err
	}
	go func() {
		name := s.Name()
		for {
			select {
			case <-s.client.Ctx().Done():
				tp.Warnf("%s: etcd server closed", name)
				s.revoke()
				tp.Warnf("%s: stop\n", name)
				return
			case ka, ok := <-ch:
				if !ok {
					tp.Debugf("%s: etcd keep alive channel closed, and restart it", name)
					s.revoke()
					ch = s.anywayKeepAlive()
				} else {
					tp.Tracef("%s: recv etcd ttl:%d", name, ka.TTL)
				}
			}
		}
	}()
	return nil
}

func (s *Service) anywayKeepAlive() <-chan *etcd.LeaseKeepAliveResponse {
	ch, err := s.keepAlive()
	for err != nil {
		time.Sleep(minLeaseTTL)
		ch, err = s.keepAlive()
	}
	return ch
}

func (s *Service) keepAlive() (<-chan *etcd.LeaseKeepAliveResponse, error) {
	resp, err := s.client.Grant(context.TODO(), minLeaseTTL)
	if err != nil {
		return nil, err
	}

	info := s.serviceInfo.String()

	_, err = s.client.Put(
		context.TODO(),
		s.serviceKey,
		info,
		etcd.WithLease(resp.ID),
	)
	if err != nil {
		return nil, err
	}

	s.leaseid = resp.ID

	ch, err := s.client.KeepAlive(context.TODO(), resp.ID)
	if err == nil {
		tp.Infof("%s: PUT %q : %q", s.Name(), s.serviceKey, info)
	}
	return ch, err
}

func (s *Service) revoke() {
	_, err := s.client.Revoke(context.TODO(), s.leaseid)
	if err != nil {
		tp.Errorf("%s: revoke service error: %s", s.Name(), err.Error())
		return
	}
}
