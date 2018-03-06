package config

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery/etcd"
)

// Config config interface
type Config interface {
	Load([]byte) error
	Reload([]byte) error
	String() string
}

// Client handling the specified microservice configuration
type Client struct {
	service, version string
	etcdClient       *etcd.Client
}

var etcdCli = new(etcd.Client)

// Init initializes a common etcd client.
func Init(etcdClient *etcd.Client) {
	*etcdCli = *etcdClient
	initGlobalNodes()
}

const (
	// KEY_PREFIX the prifix of config data key in etcd
	KEY_PREFIX = "ANTS-CONF"
)

// NewKey creates a config data key
// Note: service and version can not contant "@"!
func NewKey(service string, version string) string {
	return KEY_PREFIX + "@" + service + "@" + version
}

func must(err error) {
	if err != nil {
		ant.Fatalf("%v", err)
	}
}
