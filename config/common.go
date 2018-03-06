package config

import (
	"github.com/henrylee2cn/ant"
)

// Config config interface
type Config interface {
	Load([]byte) error
	Reload([]byte) error
	String() string
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
