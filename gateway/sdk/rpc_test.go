package sdk

import (
	"testing"

	"github.com/henrylee2cn/ant"
	"github.com/xiaoenai/ants/gateway/types"
)

// TestSdk test SDK.
func TestSdk(t *testing.T) {
	etcdClient, err := discovery.NewEtcdClient(cfg.Etcd)
	if err != nil {
		t.Fatalf("%v", err)
	}
	Init(
		ant.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		socket.NewFastProtoFunc,
		etcdClient,
	)
	reply, rerr := LongConnTotal()
	if rerr != nil {
		t.Logf("rerr: %v", rerr)
	} else {
		t.Logf("long connections total: %d", reply.ConnTotal)
	}
}
