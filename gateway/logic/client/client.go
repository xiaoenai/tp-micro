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
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
)

var (
	antCli  *ant.Client
	etcdCli *etcd.Client
)

// Init initializes a common inner ant client.
func Init(cliCfg ant.CliConfig, protoFunc socket.ProtoFunc, etcdClient *etcd.Client) {
	etcdCli = etcdClient
	antCli = ant.NewClient(
		cliCfg,
		discovery.NewLinkerFromEtcd(etcdCli),
	)
	caller = &proxyClient{antCli}
	antCli.SetProtoFunc(protoFunc)
	staticClients = newStaticClients(cliCfg, protoFunc)
}

// AntClient returns the common inner ant client.
func AntClient() *ant.Client {
	return antCli
}

// EtcdClient returns the common ETCD client.
func EtcdClient() *etcd.Client {
	return etcdCli
}
