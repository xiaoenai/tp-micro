package clientele

import (
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

var dynamicClient *micro.Client
var staticClient *StaticClients
var protoFunc tp.ProtoFunc
var cliCfg micro.CliConfig
var etcdCfg etcd.EasyConfig
var etcdClient *etcd.Client
var peerName string
var incr int32
var mutex sync.Mutex

func init() {
	cfgo.AllowAppsShare(true)
	cfgo.MustReg("cluster_client", &cliCfg)
	cfgo.MustReg("etcd", &etcdCfg)
	peerName = filepath.Base(os.Args[0])
	protoFunc = tp.DefaultProtoFunc()
	var err error
	etcdClient, err = etcd.EasyNew(etcdCfg)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	dynamicClient = micro.NewClient(
		cliCfg,
		discovery.NewLinkerFromEtcd(etcdClient),
	)
	staticClient = &StaticClients{
		clients:   make(map[string]*micro.Client),
		cfg:       cliCfg,
		protoFunc: protoFunc,
	}
}

// GetEtcdClient returns the common ETCD client.
func GetEtcdClient() *etcd.Client {
	return etcdClient
}

// GetDynamicClient returns the common client with etcd discovery.
func GetDynamicClient() *micro.Client {
	return dynamicClient
}

// GetStaticClients returns the common static client set.
func GetStaticClients() *StaticClients {
	return staticClient
}

// GetProtoFunc sets the socket communication protocol.
func GetProtoFunc() tp.ProtoFunc {
	return protoFunc
}

// SetProtoFunc sets the socket communication protocol.
func SetProtoFunc(_protoFunc tp.ProtoFunc) {
	protoFunc = _protoFunc
	dynamicClient.SetProtoFunc(protoFunc)
	staticClient.protoFunc = protoFunc
}

// GetEtcdCfg returns the etcd config.
func GetEtcdCfg() etcd.EasyConfig {
	return etcdCfg
}

// Ctx handler's context
type Ctx interface {
	// Seq returns the input packet sequence.
	Seq() int32
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
func DynamicCall(ctx Ctx, serviceMethod string, arg interface{}, result interface{}, setting ...tp.MessageSetting) tp.CallCmd {
	return dynamicClient.Call(serviceMethod, arg, result, setting...)
}

// DynamicPush sends a packet by etcd discovery, but do not receives reply.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func DynamicPush(ctx Ctx, serviceMethod string, arg interface{}, setting ...tp.MessageSetting) *tp.Status {
	return dynamicClient.Push(serviceMethod, arg, setting...)
}

// StaticCall sends a packet and receives reply, by address.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func StaticCall(ctx Ctx, addr string, serviceMethod string, arg interface{}, result interface{}, setting ...tp.MessageSetting) tp.CallCmd {
	return staticClient.GetOrSet(addr).Call(serviceMethod, arg, result, setting...)
}

// StaticPush sends a packet by address, but do not receives reply.
// Note:
// The ctx can be nil;
// If the arg is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func StaticPush(ctx Ctx, addr string, serviceMethod string, arg interface{}, setting ...tp.MessageSetting) *tp.Status {
	return staticClient.GetOrSet(addr).Push(serviceMethod, arg, setting...)
}

// StaticClients static clients map
type StaticClients struct {
	clients   map[string]*micro.Client
	cfg       micro.CliConfig
	protoFunc tp.ProtoFunc
	mu        sync.RWMutex
}

// Set sets the client whose server address is srvAddr.
func (s *StaticClients) Set(srvAddr string) {
	s.mu.Lock()
	cli := micro.NewClient(s.cfg, micro.NewStaticLinker(srvAddr))
	cli.SetProtoFunc(s.protoFunc)
	s.clients[srvAddr] = cli
	s.mu.Unlock()
}

// GetOrSet returns the client whose server address is srvAddr.
// If the client does not exist, set and return it.
func (s *StaticClients) GetOrSet(srvAddr string) *micro.Client {
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
func (s *StaticClients) Get(srvAddr string) (*micro.Client, bool) {
	s.mu.RLock()
	cli, ok := s.clients[srvAddr]
	s.mu.RUnlock()
	return cli, ok
}
