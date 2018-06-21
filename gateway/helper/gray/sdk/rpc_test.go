package sdk_test

import (
	"testing"

	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/model/etcd"

	"github.com/xiaoenai/tp-micro/gateway/helper/gray"
	"github.com/xiaoenai/tp-micro/gateway/helper/gray/types"
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
		reply, rerr := gray.IsGray(new(types.IsGrayArgs))
		if rerr != nil {
			tp.Errorf("IsGray: rerr: %v", rerr)
		} else {
			tp.Infof("IsGray: reply: %#v", reply)
		}
	}
	{
		reply, rerr := gray.Get(new(types.GetArgs))
		if rerr != nil {
			tp.Errorf("Get: rerr: %v", rerr)
		} else {
			tp.Infof("Get: reply: %#v", reply)
		}
	}
	{
		reply, rerr := gray.Delete(new(types.DeleteArgs))
		if rerr != nil {
			tp.Errorf("Delete: rerr: %v", rerr)
		} else {
			tp.Infof("Delete: reply: %#v", reply)
		}
	}
	{
		reply, rerr := gray.Set(new(types.SetArgs))
		if rerr != nil {
			tp.Errorf("Set: rerr: %v", rerr)
		} else {
			tp.Infof("Set: reply: %#v", reply)
		}
	}
}
