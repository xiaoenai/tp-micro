package clientele

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

var clusterClient *micro.Client
var cliCfg micro.CliConfig
var etcdCfg etcd.EasyConfig
var peerName string
var incr int64
var mutex sync.Mutex

func init() {
	cfgo.MustReg("cluster_client", &cliCfg)
	cfgo.MustReg("etcd", &etcdCfg)
	peerName = filepath.Base(os.Args[0])
	clusterClient = micro.NewClient(
		cliCfg,
		discovery.NewLinker(etcdCfg),
	)
}

// GetSeq creates a new sequence with some prefix string.
func GetSeq(prefix ...string) string {
	mutex.Lock()
	seq := fmt.Sprintf("%s[%d]", peerName, incr)
	incr++
	mutex.Unlock()
	for _, p := range prefix {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		seq = p + ">" + seq
	}
	return seq
}

// Ctx handler's context
type Ctx interface {
	// Seq returns the input packet sequence.
	Seq() string
	// RealIp returns the the current real remote addr.
	RealIp() string
	// Query returns the input packet uri query object.
	Query() url.Values
}

// ClusterCall sends a packet and receives reply, by etcd discovery.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func ClusterCall(ctx Ctx, uri string, arg interface{}, result interface{}, setting ...socket.PacketSetting) tp.CallCmd {
	return clusterClient.Call(uri, arg, result, settingDecorator(ctx, setting)...)
}

// ClusterPush sends a packet by etcd discovery, but do not receives reply.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func ClusterPush(ctx Ctx, uri string, arg interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	return clusterClient.Push(uri, arg, settingDecorator(ctx, setting)...)
}

func settingDecorator(ctx Ctx, settings []socket.PacketSetting) []socket.PacketSetting {
	if ctx == nil {
		return append([]socket.PacketSetting{
			tp.WithSeq(GetSeq()),
		}, settings...)
	}
	return append([]socket.PacketSetting{
		tp.WithSeq(GetSeq(ctx.Seq())),
		tp.WithRealIp(ctx.RealIp()),
	}, settings...)
}
