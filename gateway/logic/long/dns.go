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

package long

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/henrylee2cn/ant/discovery"
	"github.com/henrylee2cn/goutil/coarsetime"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/codec"
	"github.com/xiaoenai/ants/gateway/logic/client"
	"github.com/xiaoenai/ants/gateway/sdk"
	"github.com/xiaoenai/ants/gateway/types"
)

// DNS gateway ip:port list
type DNS struct {
	ips           atomic.Value
	ipsLock       sync.Mutex
	weightIps     map[string]*WeightIp
	weightIpsLock sync.RWMutex
}

const servicePrefix = discovery.ServiceNamespace + "gateway"

var dns = new(DNS)

// GatewayDNS returns the gateway ip list.
func GatewayDNS() []string {
	a, ok := dns.ips.Load().([]string)
	if !ok {
		return []string{}
	}
	return a
}

// Name returns the plugin name.
func (d *DNS) Name() string {
	return "gateway_dns"
}

// PostListen starts the DNS program.
func (d *DNS) PostListen() error {
	d.ips.Store([]string{})
	d.resetGatewayIps(true)
	go d.watchEtcd()
	return nil
}

func (d *DNS) watchEtcd() {
	var (
		updateCh = make(chan int8)
	)
	go func() {
		for flag := range updateCh {
			switch flag {
			case 1:
				d.resetGatewayIps(false)
			default:
				d.weightIpsLock.RLock()
				d.sortAndStoreIpsLocked()
				d.weightIpsLock.RUnlock()
			}
		}
	}()
	const interval = time.Second * 30
	var (
		after   *time.Timer
		now     = coarsetime.CeilingTimeNow()
		last    = now
		watcher = client.EtcdClient().Watch(context.Background(), servicePrefix, discovery.WithPrefix())
	)
	for {
		after = time.NewTimer(interval)
		select {
		case wresp := <-watcher:
			for range wresp.Events {
				now = coarsetime.CeilingTimeNow()
				if now.Sub(last) > time.Second*5 {
					select {
					case updateCh <- 1:
						last = now
					default:
					}
				}
			}
		case <-after.C:
			if now.Sub(last) > time.Second*5 {
				select {
				case updateCh <- 2:
				default:
				}
			}
		}
		after.Stop()
	}
}

func (d *DNS) resetGatewayIps(goSort bool) {
	resp, err := client.EtcdClient().Get(
		context.Background(),
		servicePrefix,
		discovery.WithPrefix(),
		discovery.WithSort(discovery.SortByKey, discovery.SortDescend),
	)
	if err != nil || resp.Kvs == nil || len(resp.Kvs) == 0 {
		d.ips.Store([]string{})
		return
	}
	var (
		q  url.Values
		ip string
		m  = make(map[string]*WeightIp, len(resp.Kvs))
	)
	for _, n := range resp.Kvs {
		q, err = url.ParseQuery(string(n.Value))
		if err != nil {
			tp.Errorf("gateway ip error: %s", err.Error())
			continue
		}
		ip = q.Get("ip")
		m[ip] = &WeightIp{
			ip:        ip,
			innerAddr: q.Get("inner_addr"),
		}
	}
	if goSort {
		go func() {
			time.Sleep(1e9)
			d.weightIpsLock.Lock()
			d.weightIps = m
			d.sortAndStoreIpsLocked()
			d.weightIpsLock.Unlock()
		}()
	} else {
		d.weightIpsLock.Lock()
		d.weightIps = m
		d.sortAndStoreIpsLocked()
		d.weightIpsLock.Unlock()
	}
}

func (d *DNS) sortAndStoreIpsLocked() {
	cnt := len(d.weightIps)
	if cnt == 0 {
		return
	}
	var (
		reply   *types.TotalLongConnReply
		t       time.Time
		sortIps = make(SortWeightIps, 0, cnt)
		rerr    *tp.Rerror
	)
	for ip, w := range d.weightIps {
		t = time.Now()
		reply, rerr = sdk.TotalLongConn(
			w.innerAddr,
			tp.WithBodyCodec(codec.ID_PROTOBUF),
		)
		if rerr != nil {
			tp.Warnf("gateway is not available: ip: %s, inner_addr: %s, error: %s", ip, w.innerAddr, rerr)
			continue
		}
		w.weight = -int64(reply.ConnTotal) - int64(time.Since(t)/time.Millisecond)
		sortIps = append(sortIps, w)
	}
	sort.Sort(sortIps)
	ips := make([]string, 0, len(sortIps))
	for i, w := range sortIps {
		// 只保留前5个
		if i > 4 {
			break
		}
		ips = append(ips, w.ip)
	}
	d.ipsLock.Lock()
	d.ips.Store(ips)
	d.ipsLock.Unlock()
	b, _ := json.MarshalIndent(ips, "", "  ")
	tp.Tracef("[UPDATE GATEWAYS] %s", b)
}

type (
	// SortWeightIps IP weight information list for sorting
	SortWeightIps []*WeightIp
	// WeightIp IP information with weight
	WeightIp struct {
		ip, innerAddr string
		weight        int64
	}
)

// Len is the number of elements in the collection.
func (s SortWeightIps) Len() int {
	return len(s)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s SortWeightIps) Less(i, j int) bool {
	return s[i].weight > s[j].weight
}

// Swap swaps the elements with indexes i and j.
func (s SortWeightIps) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
