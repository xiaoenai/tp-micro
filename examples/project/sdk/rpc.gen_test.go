package sdk_test

import (
	"encoding/json"
	"fmt"

	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/model/etcd"

	"github.com/xiaoenai/tp-micro/examples/project/sdk"
)

func init() {
	sdk.Init(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	)
}

func toJsonBytes(i interface{}) []byte {
	b, _ := json.MarshalIndent(i, "", "  ")
	return b
}

func ExampleStat() {
	rerr := sdk.Stat(&sdk.StatArg{})
	if rerr != nil {
		tp.Errorf("Stat: rerr: %s", toJsonBytes(rerr))
	}
	fmt.Printf("")
	// Output:
}

func ExampleHome() {
	result, rerr := sdk.Home(&sdk.EmptyStruct{})
	if rerr != nil {
		tp.Errorf("Home: rerr: %s", toJsonBytes(rerr))
	} else {
		tp.Infof("Home: result: %s", toJsonBytes(result))
	}
	fmt.Printf("")
	// Output:
}

func ExampleMath_Divide() {
	result, rerr := sdk.Math_Divide(&sdk.DivideArg{})
	if rerr != nil {
		tp.Errorf("Math_Divide: rerr: %s", toJsonBytes(rerr))
	} else {
		tp.Infof("Math_Divide: result: %s", toJsonBytes(result))
	}
	fmt.Printf("")
	// Output:
}
