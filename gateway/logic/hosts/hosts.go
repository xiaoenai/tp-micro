package hosts

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

const (
	servicePrefix = "ANTS-GW_HOSTS"
	// minimum lease TTL is 10-second
	minLeaseTTL      = 10
	maxGatewayAmount = 5
)

// Hosts gateway ip:port list
type Hosts struct {
	serviceKey    string
	ips           atomic.Value
	ipsLock       sync.Mutex
	weightIps     map[string]*WeightIp
	weightIpsLock sync.Mutex
	leaseid       etcd.LeaseID
}

var (
	hosts = new(Hosts)
)

// Start initializes and starts the hosts program.
func Start(httpAddr, outerSocketAddr, innerSocketAddr string) error {
	hosts.init(httpAddr, outerSocketAddr, innerSocketAddr)
	return hosts.start()
}

// GwHosts returns the gateway ip list.
func GwHosts() *types.GwHosts {
	r, _ := hosts.ips.Load().(*types.GwHosts)
	return r
}

func (h *Hosts) init(httpAddr, outerSocketAddr, innerSocketAddr string) {
	h.serviceKey = servicePrefix + "@" + httpAddr + "@" + outerSocketAddr + "@" + innerSocketAddr
}

// start starts the SocketHosts program.
func (h *Hosts) start() error {
	h.ips.Store(new(types.GwHosts))
	ch, err := h.keepAlive()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-client.EtcdClient().Ctx().Done():
				tp.Warnf("GwHosts: etcd server closed")
				h.revoke()
				tp.Warnf("GwHosts: stop\n")
				return
			case ka, ok := <-ch:
				if !ok {
					tp.Debugf("GwHosts: etcd keep alive channel closed, and restart it")
					h.revoke()
					ch = h.anywayKeepAlive()
				} else {
					tp.Tracef("GwHosts: recv etcd ttl:%d", ka.TTL)
				}
			}
		}
	}()
	h.resetGatewayIps(true)
	go h.watchEtcd()
	return nil
}

func (h *Hosts) anywayKeepAlive() <-chan *etcd.LeaseKeepAliveResponse {
	ch, err := h.keepAlive()
	for err != nil {
		time.Sleep(minLeaseTTL * time.Second)
		ch, err = h.keepAlive()
	}
	return ch
}

func (h *Hosts) keepAlive() (<-chan *etcd.LeaseKeepAliveResponse, error) {
	resp, err := client.EtcdClient().Grant(context.TODO(), minLeaseTTL)
	if err != nil {
		return nil, err
	}
	_, err = client.EtcdClient().Put(
		context.TODO(),
		h.serviceKey,
		"",
		etcd.WithLease(resp.ID),
	)
	if err != nil {
		return nil, err
	}
	ch, err := client.EtcdClient().KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return nil, err
	}
	h.leaseid = resp.ID
	return ch, nil
}

func (h *Hosts) revoke() {
	_, err := client.EtcdClient().Revoke(context.TODO(), h.leaseid)
	if err != nil {
		tp.Errorf("GwHosts: revoke service error: %s", err.Error())
		return
	}
}

func (h *Hosts) watchEtcd() {
	const (
		interval = minLeaseTTL * 2 * time.Second
		wait     = minLeaseTTL * time.Second
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
				h.resetGatewayIps(false)
				ticker = time.NewTicker(interval)
			case <-ticker.C:
				h.weightIpsLock.Lock()
				h.sortAndStoreIpsLocked()
				h.weightIpsLock.Unlock()
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

func (h *Hosts) resetGatewayIps(goSort bool) {
	resp, err := client.EtcdClient().Get(
		context.Background(),
		servicePrefix,
		etcd.WithPrefix(),
		etcd.WithSort(etcd.SortByKey, etcd.SortDescend),
	)
	if err != nil || resp.Kvs == nil || len(resp.Kvs) == 0 {
		h.ips.Store(new(types.GwHosts))
		return
	}
	var (
		httpAddr        string
		outerSocketAddr string
		innerSocketAddr string
		ok              bool
		m               = make(map[string]*WeightIp, len(resp.Kvs))
	)
	for _, n := range resp.Kvs {
		httpAddr, outerSocketAddr, innerSocketAddr, ok = splitHostsKey(n.Key)
		if !ok {
			tp.Warnf("invalid gateway service etcd key: %s", n.Key)
			continue
		}
		m[string(n.Key)] = &WeightIp{
			httpAddr:        httpAddr,
			outerSocketAddr: outerSocketAddr,
			innerSocketAddr: innerSocketAddr,
		}
	}
	if goSort {
		go func() {
			time.Sleep(1e9)
			h.weightIpsLock.Lock()
			h.weightIps = m
			h.sortAndStoreIpsLocked()
			h.weightIpsLock.Unlock()
		}()
	} else {
		h.weightIpsLock.Lock()
		h.weightIps = m
		h.sortAndStoreIpsLocked()
		h.weightIpsLock.Unlock()
	}
}

func (h *Hosts) sortAndStoreIpsLocked() {
	cnt := len(h.weightIps)
	if cnt == 0 {
		return
	}
	var (
		reply   *types.SocketTotalReply
		t       time.Time
		sortIps = make(SortWeightIps, 0, cnt)
		rerr    *tp.Rerror
	)
	for _, w := range h.weightIps {
		if len(w.innerSocketAddr) > 0 {
			t = time.Now()
			reply, rerr = sdk.SocketTotal(
				w.innerSocketAddr,
				tp.WithBodyCodec(codec.ID_PROTOBUF),
			)
			if rerr != nil {
				tp.Warnf("Not available gateway: innerSocketAddr: %s, error: %s", w.innerSocketAddr, rerr)
				continue
			}
			w.weight = -int64(reply.ConnTotal) - int64(time.Since(t)/time.Millisecond)
		}
		sortIps = append(sortIps, w)
	}
	sort.Sort(sortIps)
	ips := &types.GwHosts{
		Http:   make([]string, 0, len(sortIps)),
		Socket: make([]string, 0, len(sortIps)),
	}
	for i, w := range sortIps {
		// 只保留前5个
		if i >= maxGatewayAmount {
			break
		}
		if len(w.httpAddr) > 0 {
			ips.Http = append(ips.Http, w.httpAddr)
		}
		if len(w.outerSocketAddr) > 0 {
			ips.Socket = append(ips.Socket, w.outerSocketAddr)
		}
	}
	h.ipsLock.Lock()
	h.ips.Store(ips)
	h.ipsLock.Unlock()
	b, _ := json.MarshalIndent(ips, "", "  ")
	tp.Tracef("[UPDATE GW_HOSTS] %s", b)
}

type (
	// SortWeightIps IP weight information list for sorting
	SortWeightIps []*WeightIp
	// WeightIp IP information with weight
	WeightIp struct {
		httpAddr        string
		outerSocketAddr string
		innerSocketAddr string
		weight          int64
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

func splitHostsKey(valueBytes []byte) (httpAddr, outerSocketAddr, innerSocketAddr string, ok bool) {
	valueBytes = bytes.TrimPrefix(valueBytes, goutil.StringToBytes(servicePrefix))
	a := bytes.Split(valueBytes, []byte{'@'})
	if len(a) != 4 {
		return
	}
	httpAddr = string(a[1])
	outerSocketAddr = string(a[2])
	innerSocketAddr = string(a[3])
	ok = true
	return
}
