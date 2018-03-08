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
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/henrylee2cn/ant/discovery/etcd"
	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/goutil/coarsetime"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/codec"
	"github.com/xiaoenai/ants/gateway/logic/client"
	"github.com/xiaoenai/ants/gateway/sdk"
	"github.com/xiaoenai/ants/gateway/types"
)

// Hosts gateway ip:port list
type Hosts struct {
	ips                  atomic.Value
	ipsLock              sync.Mutex
	weightIps            map[string]*WeightIp
	weightIpsLock        sync.Mutex
	serviceKey           string
	leaseid              etcd.LeaseID
	outerAddr, innerAddr string
}

const (
	servicePrefix = "ANTS-GATEWAY"
	// minimum lease TTL is 5-second
	minLeaseTTL = 5
)

var (
	hosts                     = new(Hosts)
	_     tp.PostListenPlugin = hosts
)

func initHosts(outerAddr, innerAddr string) {
	hosts.outerAddr, hosts.innerAddr = outerAddr, innerAddr
	hosts.serviceKey = servicePrefix + "@" + outerAddr + "@" + innerAddr
}

func splitServiceKey(valueBytes []byte) (outerAddr, innerAddr string, ok bool) {
	valueBytes = bytes.TrimPrefix(valueBytes, goutil.StringToBytes(servicePrefix))
	a := bytes.Split(valueBytes, []byte{'@'})
	if len(a) != 3 {
		return
	}
	outerAddr = string(a[1])
	innerAddr = string(a[2])
	ok = true
	return
}

// gatewayList returns the gateway ip list.
func gatewayList() []string {
	a, ok := hosts.ips.Load().([]string)
	if !ok {
		return []string{}
	}
	return a
}

// Name returns the plugin name.
func (d *Hosts) Name() string {
	return "gateway_hosts"
}

// PostListen starts the Hosts program.
func (d *Hosts) PostListen() error {
	ch, err := d.keepAlive()
	if err != nil {
		return err
	}
	go func() {
		name := d.Name()
		for {
			select {
			case <-client.EtcdClient().Ctx().Done():
				tp.Warnf("%s: etcd server closed", name)
				d.revoke()
				tp.Warnf("%s: stop\n", name)
				return
			case ka, ok := <-ch:
				if !ok {
					tp.Debugf("%s: etcd keep alive channel closed, and restart it", name)
					d.revoke()
					ch = d.anywayKeepAlive()
				} else {
					tp.Tracef("%s: recv etcd ttl:%d", name, ka.TTL)
				}
			}
		}
	}()
	d.ips.Store([]string{})
	d.resetGatewayIps(true)
	go d.watchEtcd()
	return nil
}

func (d *Hosts) anywayKeepAlive() <-chan *etcd.LeaseKeepAliveResponse {
	ch, err := d.keepAlive()
	for err != nil {
		time.Sleep(minLeaseTTL)
		ch, err = d.keepAlive()
	}
	return ch
}

func (d *Hosts) keepAlive() (<-chan *etcd.LeaseKeepAliveResponse, error) {
	resp, err := client.EtcdClient().Grant(context.TODO(), minLeaseTTL)
	if err != nil {
		return nil, err
	}

	_, err = client.EtcdClient().Put(
		context.TODO(),
		d.serviceKey,
		"",
		etcd.WithLease(resp.ID),
	)
	if err != nil {
		return nil, err
	}

	d.leaseid = resp.ID

	ch, err := client.EtcdClient().KeepAlive(context.TODO(), resp.ID)
	if err == nil {
		tp.Infof("%s: PUT KEY: %q", d.Name(), d.serviceKey)
	}
	return ch, err
}

func (d *Hosts) revoke() {
	_, err := client.EtcdClient().Revoke(context.TODO(), d.leaseid)
	if err != nil {
		tp.Errorf("%s: revoke service error: %s", d.Name(), err.Error())
		return
	}
}

func (d *Hosts) watchEtcd() {
	const (
		interval = time.Second * 20
		wait     = time.Second * 5
	)
	var (
		updateCh = make(chan struct{})
	)
	go func() {
		var ticker = time.NewTicker(interval)
		for {
			select {
			case <-updateCh:
				ticker.Stop()
				d.resetGatewayIps(false)
				ticker = time.NewTicker(interval)
			case <-ticker.C:
				d.weightIpsLock.Lock()
				d.sortAndStoreIpsLocked()
				d.weightIpsLock.Unlock()
			}
		}
	}()

	var (
		now     = coarsetime.CeilingTimeNow()
		last    = now
		watcher = client.EtcdClient().Watch(
			context.Background(),
			servicePrefix,
			etcd.WithPrefix(),
		)
	)

	for wresp := range watcher {
		for range wresp.Events {
			now = coarsetime.CeilingTimeNow()
			if now.Sub(last) > wait {
				select {
				case updateCh <- struct{}{}:
					last = now
				default:
				}
			}
		}
	}
}

func (d *Hosts) resetGatewayIps(goSort bool) {
	resp, err := client.EtcdClient().Get(
		context.Background(),
		servicePrefix,
		etcd.WithPrefix(),
		etcd.WithSort(etcd.SortByKey, etcd.SortDescend),
	)
	if err != nil || resp.Kvs == nil || len(resp.Kvs) == 0 {
		d.ips.Store([]string{})
		return
	}
	var (
		outerAddr, innerAddr string
		ok                   bool
		m                    = make(map[string]*WeightIp, len(resp.Kvs))
	)
	for _, n := range resp.Kvs {
		outerAddr, innerAddr, ok = splitServiceKey(n.Key)
		if !ok {
			tp.Warnf("invalid gateway service etcd key: %s", n.Key)
			continue
		}
		m[outerAddr] = &WeightIp{
			outerAddr: outerAddr,
			innerAddr: innerAddr,
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

func (d *Hosts) sortAndStoreIpsLocked() {
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
	for _, w := range d.weightIps {
		t = time.Now()
		reply, rerr = sdk.LongConnTotal(
			w.innerAddr,
			tp.WithBodyCodec(codec.ID_PROTOBUF),
		)
		if rerr != nil {
			tp.Warnf("Not available gateway: innerAddr: %s, error: %s", w.innerAddr, rerr)
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
		ips = append(ips, w.outerAddr)
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
		outerAddr, innerAddr string
		weight               int64
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
