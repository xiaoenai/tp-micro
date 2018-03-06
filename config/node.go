package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery/etcd"
)

// InitNode initializes a config node.
func InitNode(etcdClient *etcd.Client) {
	globalNodes = &Nodes{
		nodeMap: make(map[string]*Node),
	}

	var err error
	globalNodes.etcdSession, err = etcd.NewSession(etcdClient)
	if err != nil {
		ant.Fatalf("Initialization of the global node failed: %s", err.Error())
	}
}

// SyncNode registers a configuration template to etcd,
// and update it when monitoring the configuration changes.
func SyncNode(service, version string, cfg Config) {
	globalNodes.mustAdd(service, version, cfg)
}

// Nodes config node handlers
type Nodes struct {
	nodeMap     map[string]*Node
	etcdSession *etcd.Session
	rwMutex     sync.RWMutex
}

var globalNodes *Nodes

func (n *Nodes) mustAdd(service, version string, cfg Config) {
	must(n.add(service, version, cfg))
}

func (n *Nodes) add(service, version string, cfg Config) (err error) {
	n.rwMutex.Lock()
	defer n.rwMutex.Unlock()

	key := NewKey(service, version)
	if _, ok := n.nodeMap[key]; ok {
		return fmt.Errorf("Repeat the registration configuration: %s", key)
	}

	node := &Node{
		key:         key,
		object:      cfg,
		etcdMutex:   etcd.NewLocker(n.etcdSession, key),
		Initialized: false,
		Config:      cfg.String(),
		doInitCh:    make(chan error, 1),
	}
	n.nodeMap[key] = node

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("etcd concurrency lock fail: %v", p)
		}
	}()

	node.etcdMutex.Lock()
	defer node.etcdMutex.Unlock()

	resp, err := n.etcdSession.Client().Get(context.TODO(), key)
	if err != nil {
		return err
	}

	if len(resp.Kvs) > 0 {
		err = node.bind(resp.Kvs[0].Value)
		if node.Initialized {
			go node.watch(n.etcdSession.Client())
			return err
		}

	} else {
		_, err = n.etcdSession.Client().Put(context.TODO(), key, node.String())
		if err != nil {
			return err
		}
	}

	ant.Warnf("Wait for the configuration in the ETCD to be set: %s", key)
	go node.watch(n.etcdSession.Client())
	return node.waitInit()
}

// Node config node handler
type Node struct {
	key    string
	object Config
	// Config string
	Config string `json:"config"`
	// Is it initialized?
	Initialized bool `json:"initialized"`
	doInitCh    chan error
	etcdMutex   sync.Locker
}

// func parseNode(data []byte) (*Node, error) {
// 	var n = new(Node)
// 	err := json.Unmarshal(data, n)
// 	if err == nil {
// 		n.doInitCh = make(chan error, 1)
// 	}
// 	return n, err
// }

func (n *Node) bind(data []byte) error {
	inited := n.Initialized
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}

	if inited {
		err = n.object.Reload([]byte(n.Config))
	} else {
		err = n.object.Load([]byte(n.Config))
		if n.Initialized {
			select {
			case n.doInitCh <- err:
			default:
			}
		}
	}

	return err
}

func (n *Node) waitInit() error {
	return <-n.doInitCh
}

// String returns the encoding string
func (n *Node) String() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (n *Node) watch(etcdClient *etcd.Client) {
	watcher := etcdClient.Watch(
		context.TODO(),
		n.key,
	)
	for wresp := range watcher {
		for _, event := range wresp.Events {
			if event.Type != etcd.EventTypePut {
				continue
			}
			err := n.bind(event.Kv.Value)
			if err != nil {
				ant.Errorf("Binding configuration from etcd failed: %s", err)
			}
		}
	}
}
