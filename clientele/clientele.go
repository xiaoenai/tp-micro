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

var dynamicClient *micro.Client
var staticClient *staticClients
var protoFunc socket.ProtoFunc
var cliCfg micro.CliConfig
var etcdCfg etcd.EasyConfig
var etcdClient *etcd.Client
var peerName string
var incr int64
var mutex sync.Mutex

func init() {
	cfgo.MustReg("cluster_client", &cliCfg)
	cfgo.MustReg("etcd", &etcdCfg)
	peerName = filepath.Base(os.Args[0])
	protoFunc = socket.DefaultProtoFunc()
	var err error
	etcdClient, err = etcd.EasyNew(etcdCfg)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	dynamicClient = micro.NewClient(
		cliCfg,
		discovery.NewLinkerFromEtcd(etcdClient),
	)
	staticClient = &staticClients{
		clients:   make(map[string]*micro.Client),
		cfg:       cliCfg,
		protoFunc: protoFunc,
	}
}

// GetEtcdClient returns the common ETCD client.
func GetEtcdClient() *etcd.Client {
	return etcdClient
}

// GetProtoFunc sets the socket communication protocol.
func GetProtoFunc() socket.ProtoFunc {
	return protoFunc
}

// SetProtoFunc sets the socket communication protocol.
func SetProtoFunc(_protoFunc socket.ProtoFunc) {
	protoFunc = _protoFunc
	dynamicClient.SetProtoFunc(protoFunc)
	staticClient.protoFunc = protoFunc
}

// GetEtcdCfg returns the etcd config.
func GetEtcdCfg() etcd.EasyConfig {
	return etcdCfg
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

// DynamicCall sends a packet and receives reply, by etcd discovery.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func DynamicCall(ctx Ctx, uri string, arg interface{}, result interface{}, setting ...socket.PacketSetting) tp.CallCmd {
	return dynamicClient.Call(uri, arg, result, settingDecorator(ctx, setting)...)
}

// DynamicPush sends a packet by etcd discovery, but do not receives reply.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func DynamicPush(ctx Ctx, uri string, arg interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	return dynamicClient.Push(uri, arg, settingDecorator(ctx, setting)...)
}

// StaticCall sends a packet and receives reply, by address.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func StaticCall(ctx Ctx, addr string, uri string, arg interface{}, result interface{}, setting ...socket.PacketSetting) tp.CallCmd {
	return staticClient.GetOrSet(addr).Call(uri, arg, result, settingDecorator(ctx, setting)...)
}

// StaticPush sends a packet by address, but do not receives reply.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func StaticPush(ctx Ctx, addr string, uri string, arg interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	return staticClient.GetOrSet(addr).Push(uri, arg, settingDecorator(ctx, setting)...)
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

// staticClients static clients map
type staticClients struct {
	clients   map[string]*micro.Client
	cfg       micro.CliConfig
	protoFunc socket.ProtoFunc
	mu        sync.RWMutex
}

// Set sets the client whose server address is srvAddr.
func (s *staticClients) Set(srvAddr string) {
	s.mu.Lock()
	cli := micro.NewClient(s.cfg, micro.NewStaticLinker(srvAddr))
	cli.SetProtoFunc(s.protoFunc)
	s.clients[srvAddr] = cli
	s.mu.Unlock()
}

// GetOrSet returns the client whose server address is srvAddr.
// If the client does not exist, set and return it.
func (s *staticClients) GetOrSet(srvAddr string) *micro.Client {
	s.mu.RLock()
	cli, ok := s.clients[srvAddr]
	s.mu.RUnlock()
	if ok {
		return cli
	}
	s.mu.Lock()
	cli, ok = s.clients[srvAddr]
	defer s.mu.Unlock()
	if ok {
		return cli
	}
	cli = micro.NewClient(s.cfg, micro.NewStaticLinker(srvAddr))
	cli.SetProtoFunc(s.protoFunc)
	s.clients[srvAddr] = cli
	return cli
}

// Get returns the client whose server address is srvAddr.
func (s *staticClients) Get(srvAddr string) (*micro.Client, bool) {
	s.mu.RLock()
	cli, ok := s.clients[srvAddr]
	s.mu.RUnlock()
	return cli, ok
}
