package config

import (
	"sync"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery/etcd"
	"github.com/henrylee2cn/cfgo"
)

// Config config interface
type Config interface {
	UnmarshalJSON([]byte) error
	MarshalJSON() ([]byte, error)
	Reload([]byte) error
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

var (
	onceRegYaml    sync.Once
	etcdEasyConfig *etcd.EasyConfig
)

type etcdConfig etcd.EasyConfig

func (e *etcdConfig) Reload(bindFunc cfgo.BindFunc) error {
	return bindFunc()
}

// NewEtcdClientFromYaml uses config/etcd.yaml to create a etcd client.
func NewEtcdClientFromYaml() (*etcd.Client, error) {
	onceRegYaml.Do(func() {
		etcdEasyConfig = new(etcd.EasyConfig)
		cfgo.
			MustGet("./config/etcd.yaml").
			MustReg("ETCD", (*etcdConfig)(etcdEasyConfig))
	})
	return etcd.EasyNew(*etcdEasyConfig)
}
