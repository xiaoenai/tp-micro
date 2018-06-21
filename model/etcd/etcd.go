// etcd package is the [ETCD](https://github.com/coreos/etcd) client v3 mirror.
// Copyright 2018 HenryLee. All Rights Reserved.
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
//
package etcd

import (
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

// EasyConfig ETCD client config
type EasyConfig struct {
	Endpoints   []string      `yaml:"endpoints"    ini:"endpoints"    comment:"list of URLs"`
	DialTimeout time.Duration `yaml:"dial_timeout" ini:"dial_timeout" comment:"timeout for failing to establish a connection"`
	Username    string        `yaml:"username"     ini:"username"     comment:"user name for authentication"`
	Password    string        `yaml:"password"     ini:"password"     comment:"password for authentication"`
}

// EasyNew creates ETCD client.
// Note:
// If etcdConfig.DialTimeout<0, it means unlimit;
// If etcdConfig.DialTimeout=0, use the default value(15s).
func EasyNew(etcdConfig EasyConfig) (*clientv3.Client, error) {
	if etcdConfig.DialTimeout == 0 {
		etcdConfig.DialTimeout = 15 * time.Second
	} else if etcdConfig.DialTimeout < 0 {
		etcdConfig.DialTimeout = 0
	}
	return clientv3.New(clientv3.Config{
		Endpoints:   etcdConfig.Endpoints,
		DialTimeout: etcdConfig.DialTimeout,
		Username:    etcdConfig.Username,
		Password:    etcdConfig.Password,
	})
}

// migrated from etcd 'github.com/coreos/etcd/clientv3'

// Event types
const (
	EventTypePut    = clientv3.EventTypePut
	EventTypeDelete = clientv3.EventTypeDelete
)

// Client ETCD v3 client
type Client = clientv3.Client

// New creates a new etcdv3 client from a given configuration.
//  func New(cfg clientv3.Config) (*clientv3.Client, error)
var NewClient = clientv3.New

// Config etcd config
type Config = clientv3.Config

// OpOption configures Operations like Get, Put, Delete.
type OpOption = clientv3.OpOption

// LeaseID etcd lease ID
type LeaseID = clientv3.LeaseID

type (
	CompactResponse = clientv3.CompactResponse
	PutResponse     = clientv3.PutResponse
	GetResponse     = clientv3.GetResponse
	DeleteResponse  = clientv3.DeleteResponse
	TxnResponse     = clientv3.TxnResponse
)

// WithLease attaches a lease ID to a key in 'Put' request.
//  func WithLease(leaseID clientv3.LeaseID) clientv3.OpOption
var WithLease = clientv3.WithLease

// WithLimit limits the number of results to return from 'Get' request.
// If WithLimit is given a 0 limit, it is treated as no limit.
//  func WithLimit(n int64) clientv3.OpOption
var WithLimit = clientv3.WithLimit

// WithRev specifies the store revision for 'Get' request.
// Or the start revision of 'Watch' request.
//  func WithRev(rev int64) clientv3.OpOption
var WithRev = clientv3.WithRev

// WithSort specifies the ordering in 'Get' request. It requires
// 'WithRange' and/or 'WithPrefix' to be specified too.
// 'target' specifies the target to sort by: key, version, revisions, value.
// 'order' can be either 'SortNone', 'SortAscend', 'SortDescend'.
//  func WithSort(target SortTarget, order SortOrder) clientv3.OpOption
var WithSort = clientv3.WithSort

// WithPrefix enables 'Get', 'Delete', or 'Watch' requests to operate
// on the keys with matching prefix. For example, 'Get(foo, WithPrefix())'
// can return 'foo1', 'foo2', and so on.
//  func WithPrefix() clientv3.OpOption
var WithPrefix = clientv3.WithPrefix

// WithRange specifies the range of 'Get', 'Delete', 'Watch' requests.
// For example, 'Get' requests with 'WithRange(end)' returns
// the keys in the range [key, end).
// endKey must be lexicographically greater than start key.
//  func WithRange(endKey string) clientv3.OpOption
var WithRange = clientv3.WithRange

// WithFromKey specifies the range of 'Get', 'Delete', 'Watch' requests
// to be equal or greater than the key in the argument.
//  func WithFromKey() clientv3.OpOption
var WithFromKey = clientv3.WithFromKey

// WithSerializable makes 'Get' request serializable. By default,
// it's linearizable. Serializable requests are better for lower latency
// requirement.
//  func WithSerializable() clientv3.OpOption
var WithSerializable = clientv3.WithSerializable

// WithKeysOnly makes the 'Get' request return only the keys and the corresponding
// values will be omitted.
//  func WithKeysOnly() clientv3.OpOption
var WithKeysOnly = clientv3.WithKeysOnly

// WithCountOnly makes the 'Get' request return only the count of keys.
//  func WithCountOnly() clientv3.OpOption
var WithCountOnly = clientv3.WithCountOnly

// WithMinModRev filters out keys for Get with modification revisions less than the given revision.
//  func WithMinModRev(rev int64) clientv3.OpOption
var WithMinModRev = clientv3.WithMinModRev

// WithMaxModRev filters out keys for Get with modification revisions greater than the given revision.
//  func WithMaxModRev(rev int64) clientv3.OpOption
var WithMaxModRev = clientv3.WithMaxModRev

// WithMinCreateRev filters out keys for Get with creation revisions less than the given revision.
//  func WithMinCreateRev(rev int64) clientv3.OpOption
var WithMinCreateRev = clientv3.WithMinCreateRev

// WithMaxCreateRev filters out keys for Get with creation revisions greater than the given revision.
//  func WithMaxCreateRev(rev int64) clientv3.OpOption
var WithMaxCreateRev = clientv3.WithMaxCreateRev

// WithFirstCreate gets the key with the oldest creation revision in the request range.
//  func WithFirstCreate() []clientv3.OpOption
var WithFirstCreate = clientv3.WithFirstCreate

// WithLastCreate gets the key with the latest creation revision in the request range.
//  func WithLastCreate() []clientv3.OpOption
var WithLastCreate = clientv3.WithLastCreate

// WithFirstKey gets the lexically first key in the request range.
//  func WithFirstKey() []clientv3.OpOption
var WithFirstKey = clientv3.WithFirstKey

// WithLastKey gets the lexically last key in the request range.
//  func WithLastKey() []clientv3.OpOption
var WithLastKey = clientv3.WithLastKey

// WithFirstRev gets the key with the oldest modification revision in the request range.
//  func WithFirstRev() []clientv3.OpOption
var WithFirstRev = clientv3.WithFirstRev

// WithLastRev gets the key with the latest modification revision in the request range.
//  func WithLastRev
var WithLastRev = clientv3.WithLastRev

// every 10 minutes when there is no incoming events.
// Progress updates have zero events in WatchResponse.
//  func WithProgressNotify() clientv3.OpOption
var WithProgressNotify = clientv3.WithProgressNotify

// WithCreatedNotify makes watch server sends the created event.
//  func WithCreatedNotify() clientv3.OpOption
var WithCreatedNotify = clientv3.WithCreatedNotify

// WithFilterPut discards PUT events from the watcher.
//  func WithFilterPut() clientv3.OpOption
var WithFilterPut = clientv3.WithFilterPut

// WithFilterDelete discards DELETE events from the watcher.
//  func WithFilterDelete() clientv3.OpOption
var WithFilterDelete = clientv3.WithFilterDelete

// WithPrevKV gets the previous key-value pair before the event happens.
// If the previous KV is already compacted, nothing will be returned.
//  func WithPrevKV() clientv3.OpOption
var WithPrevKV = clientv3.WithPrevKV

// This option can not be combined with non-empty values.
// Returns an error if the key does not exist.
//  func WithIgnoreValue() clientv3.OpOption
var WithIgnoreValue = clientv3.WithIgnoreValue

// This option can not be combined with WithLease.
// Returns an error if the key does not exist.
//  func WithIgnoreLease() clientv3.OpOption
var WithIgnoreLease = clientv3.WithIgnoreLease

// WithAttachedKeys makes TimeToLive list the keys attached to the given lease ID.
//  func WithAttachedKeys() LeaseOption
var WithAttachedKeys = clientv3.WithAttachedKeys

// WithCompactPhysical makes Compact wait until all compacted entries are
// removed from the etcd server's storage.
//  func WithCompactPhysical() CompactOption
var WithCompactPhysical = clientv3.WithCompactPhysical

// WithRequireLeader requires client requests to only succeed
// when the cluster has a leader.
//  func WithRequireLeader(ctx context.Context) context.Context
var WithRequireLeader = clientv3.WithRequireLeader

// SortOrder etcd SortOrder type
type SortOrder = clientv3.SortOrder

// SortOrder types
const (
	SortNone    SortOrder = clientv3.SortNone
	SortAscend            = clientv3.SortAscend
	SortDescend           = clientv3.SortDescend
)

// SortTarget etcd SortTarget type
type SortTarget = clientv3.SortTarget

// SortTarget types
const (
	SortByKey            SortTarget = clientv3.SortByKey
	SortByVersion                   = clientv3.SortByVersion
	SortByCreateRevision            = clientv3.SortByCreateRevision
	SortByModRevision               = clientv3.SortByModRevision
	SortByValue                     = clientv3.SortByValue
)

// SortOption etcd SortOption type
type SortOption = clientv3.SortOption

// LeaseGrantResponse wraps the protobuf message LeaseGrantResponse.
type LeaseGrantResponse = clientv3.LeaseGrantResponse

// LeaseKeepAliveResponse wraps the protobuf message LeaseKeepAliveResponse.
type LeaseKeepAliveResponse = clientv3.LeaseKeepAliveResponse

// LeaseTimeToLiveResponse wraps the protobuf message LeaseTimeToLiveResponse.
type LeaseTimeToLiveResponse = clientv3.LeaseTimeToLiveResponse

// LeaseStatus represents a lease status.
type LeaseStatus = clientv3.LeaseStatus

// Session represents a lease kept alive for the lifetime of a client.
// Fault-tolerant applications may use sessions to reason about liveness.
type Session = concurrency.Session

// SessionOption configures Session.
type SessionOption concurrency.SessionOption

// NewSession gets the leased session for a client.
//  func NewSession(client *v3.Client, opts ...concurrency.SessionOption) (*Session, error)
var NewSession = concurrency.NewSession

// WithSessionTTL configures the session's TTL in seconds.
// If TTL is <= 0, the default 60 seconds TTL will be used.
//  func WithSessionTTL(ttl int) concurrency.SessionOption
var WithSessionTTL = concurrency.WithTTL

// WithSessionLease specifies the existing leaseID to be used for the session.
// This is useful in process restart scenario, for example, to reclaim
// leadership from an election prior to restart.
//  func WithSessionLease(leaseID v3.LeaseID) concurrency.SessionOption
var WithSessionLease = concurrency.WithLease

// WithSessionContext assigns a context to the session instead of defaulting to
// using the client context. This is useful for canceling NewSession and
// Close operations immediately without having to close the client. If the
// context is canceled before Close() completes, the session's lease will be
// abandoned and left to expire instead of being revoked.
//  func WithSessionContext(ctx context.Context) concurrency.SessionOption
var WithSessionContext = concurrency.WithContext

// Mutexutex implements the sync Locker interface with etcd
type Mutex = concurrency.Mutex

// NewMutex creates a sync Locker interface with etcd.
// func NewMutex(s *concurrency.Session, pfx string) *Mutex
var NewMutex = concurrency.NewMutex

// NewLocker creates a sync.Locker backed by an etcd mutex.
//  func NewLocker(s *concurrency.Session, pfx string) sync.Locker
var NewLocker = concurrency.NewLocker
