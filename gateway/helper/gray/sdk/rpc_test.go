package sdk_test

import (
	"testing"

	tp "github.com/henrylee2cn/teleport/v6"
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/model/etcd"

	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray"
	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/types"
)

// TestSdk test SDK.
func TestSdk(t *testing.T) {
	gray.Init(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	)

	{
		reply, stat := gray.IsGray(new(types.IsGrayArgs))
		if stat != nil {
			tp.Errorf("IsGray: rerr: %v", stat)
		} else {
			tp.Infof("IsGray: reply: %#v", reply)
		}
	}
	{
		reply, stat := gray.Get(new(types.GetArgs))
		if stat != nil {
			tp.Errorf("Get: rerr: %v", stat)
		} else {
			tp.Infof("Get: reply: %#v", reply)
		}
	}
	{
		reply, stat := gray.Delete(new(types.DeleteArgs))
		if stat != nil {
			tp.Errorf("Delete: rerr: %v", stat)
		} else {
			tp.Infof("Delete: reply: %#v", reply)
		}
	}
	{
		reply, stat := gray.Set(new(types.SetArgs))
		if stat != nil {
			tp.Errorf("Set: rerr: %v", stat)
		} else {
			tp.Infof("Set: reply: %#v", reply)
		}
	}
}
