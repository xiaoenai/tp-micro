package sdk

import (
	"testing"

	"github.com/henrylee2cn/ant"
	"github.com/xiaoenai/ants/gateway/types"
)

// TestSdk test SDK.
func TestSdk(t *testing.T) {
	Init(
		ant.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		ant.NewStaticLinker(":9090"),
	)
	reply, rerr := PullMathDivide(&types.MathDivideArgs{A: 10, B: 5})
	if rerr != nil {
		t.Logf("rerr: %v", rerr)
	} else {
		t.Logf("10 / 5 = %d", *reply)
	}
}
