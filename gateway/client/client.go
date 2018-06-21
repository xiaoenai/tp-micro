// Copyright 2018 github.com/xiaoenai. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

var (
	staticClients *StaticClients
	dynamicCli    *micro.Client
	etcdCli       *etcd.Client
)

// Init initializes a common inner tp-micro client.
func Init(cliCfg micro.CliConfig, protoFunc socket.ProtoFunc, etcdClient *etcd.Client) {
	etcdCli = etcdClient
	dynamicCli = micro.NewClient(
		cliCfg,
		discovery.NewLinkerFromEtcd(etcdCli),
	)
	dynamicCli.SetProtoFunc(protoFunc)
	staticClients = newStaticClients(cliCfg, protoFunc)
}

// StaticClient returns the client whose server address is srvAddr.
// If the client does not exist, set and return it.
func StaticClient(srvAddr string) *micro.Client {
	return staticClients.GetOrSet(srvAddr)
}

// DynamicClient returns the common inner dynamic routing client.
func DynamicClient() *micro.Client {
	return dynamicCli
}

// EtcdClient returns the common ETCD client.
func EtcdClient() *etcd.Client {
	return etcdCli
}
