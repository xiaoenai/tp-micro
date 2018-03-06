package config

import (
	"github.com/henrylee2cn/ant/discovery/etcd"
	tp "github.com/henrylee2cn/teleport"
)

var mgr = struct {
	etcdClient *etcd.Client
}{}

// InitMgr initializes a config manager.
func InitMgr(etcdClient *etcd.Client) {
	mgr.etcdClient = etcdClient
}

type cfg struct {
	tp.PullCtx
}

func (c *cfg) List(*struct{}) ([]string, *tp.Rerror) {
	return nil, nil
}

// func (c *cfg) Get(*struct{}) (*Node, *tp.Rerror) {
// 	key := c.Query().Get("config-key")

// }

// func (c *cfg) Update(node *Node) (*struct{}, *tp.Rerror) {

// }
