// redis (cluster) client package.
package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/henrylee2cn/cfgo"
)

type (
	// Config redis (cluster) client config
	Config struct {
		// redis deploy type, [single, cluster]
		DeployType string `yaml:"deploy_type"`
		// only for single node config, valid when DeployType=single.
		ForSingle SingleConfig `yaml:"for_single"`
		// only for cluster config, valid when DeployType=cluster.
		ForCluster ClusterConfig `yaml:"for_cluster"`

		// An optional password. Must match the password specified in the
		// requirepass server configuration option.
		Password string `yaml:"password,omitempty"`

		// The maximum number of retries before giving up.
		// Default is to not retry failed commands.
		MaxRetries int `yaml:"max_retries,omitempty"`

		// Dial timeout for establishing new connections.
		// Default is 5 seconds.
		DialTimeout int64 `yaml:"dial_timeout,omitempty"`
		// Timeout for socket reads. If reached, commands will fail
		// with a timeout instead of blocking.
		// Default is 3 seconds.
		ReadTimeout int64 `yaml:"read_timeout,omitempty"`
		// Timeout for socket writes. If reached, commands will fail
		// with a timeout instead of blocking.
		// Default is ReadTimeout.
		WriteTimeout int64 `yaml:"write_timeout,omitempty"`

		// PoolSizePerNode applies per cluster node and not for the whole cluster.
		// Maximum number of socket connections.
		// Default is 10 connections per every CPU as reported by runtime.NumCPU.
		PoolSizePerNode int `yaml:"pool_size_per_node"`
		// Amount of time client waits for connection if all connections
		// are busy before returning an error.
		// Default is ReadTimeout + 1 second.
		PoolTimeout int64 `yaml:"pool_timeout,omitempty"`
		// Amount of time after which client closes idle connections.
		// Should be less than server's timeout.
		// Default is 300 seconds.
		IdleTimeout int64 `yaml:"idle_timeout"`
		// Frequency of idle checks.
		// Default is 60 seconds.
		// When minus value is set, then idle check is disabled.
		IdleCheckFrequency int64 `yaml:"idle_check_frequency,omitempty"`

		// Enables read only queries on slave nodes.
		// Only for cluster.
		ReadOnly bool `yaml:"read_only,omitempty"`

		init bool
	}

	// SingleConfig redis single node client config.
	SingleConfig struct {
		// host:port address.
		Addr string `yaml:"addr"`

		// Maximum backoff between each retry.
		// Default is 512 seconds; -1 disables backoff.
		MaxRetryBackoff int64 `yaml:"max_retry_backoff,omitempty"`
	}

	// ClusterConfig redis cluster client config.
	ClusterConfig struct {
		// A seed list of host:port addresses of cluster nodes.
		Addrs []string `yaml:"addrs"`

		// The maximum number of retries before giving up. Command is retried
		// on network errors and MOVED/ASK redirects.
		// Default is 16.
		MaxRedirects int `yaml:"max_redirects,omitempty"`

		// Enables routing read-only queries to the closest master or slave node.
		RouteByLatency bool `yaml:"route_by_latency,omitempty"`
	}
)

// deploy types
const (
	TypeSingle  = "single"
	TypeCluster = "cluster"
)

// Reload reloads config.
func (cfg *Config) Reload(bind cfgo.BindFunc) error {
	if cfg.init {
		return nil
	}
	err := bind()
	if err != nil {
		return err
	}
	cfg.init = true
	if cfg.DeployType != TypeSingle && cfg.DeployType != TypeCluster {
		return fmt.Errorf("redis config: deploy_type optional enumeration list: %s, %s", TypeSingle, TypeCluster)
	}
	return nil
}

// ReadConfig read config from specified yaml section.
func ReadConfig(configSection string) (*Config, error) {
	var cfg = NewConfig()
	var err error
	if cfgo.IsReg(configSection) {
		err = cfgo.BindSection(configSection, cfg)
	} else {
		err = cfgo.Reg(configSection, cfg)
	}
	return cfg, err
}

// NewConfig creates a default config.
func NewConfig() *Config {
	return &Config{
		DeployType: TypeSingle,
		ForSingle:  SingleConfig{Addr: "127.0.0.1:6379"},
		// ForCluster: ClusterConfig{Addrs: []string{"127.0.0.1:6379"}},
	}
}

// Client redis (cluster) client.
type (
	Client struct {
		cfg *Config
		Cmdable
	}
	Cmdable interface {
		redis.Cmdable
		Subscribe(channels ...string) *redis.PubSub
	}
	// Alias
	PubSub             = redis.PubSub
	Message            = redis.Message
	GeoLocation        = redis.GeoLocation
	GeoRadiusQuery     = redis.GeoRadiusQuery
	ZRangeBy           = redis.ZRangeBy
	Z                  = redis.Z
	Pipeliner          = redis.Pipeliner
	RedisCmdable       = redis.Cmdable
	SliceCmd           = redis.SliceCmd
	StatusCmd          = redis.StatusCmd
	Cmder              = redis.Cmder
	IntCmd             = redis.IntCmd
	DurationCmd        = redis.DurationCmd
	BoolCmd            = redis.BoolCmd
	StringCmd          = redis.StringCmd
	FloatCmd           = redis.FloatCmd
	StringSliceCmd     = redis.StringSliceCmd
	BoolSliceCmd       = redis.BoolSliceCmd
	StringStringMapCmd = redis.StringStringMapCmd
	StringIntMapCmd    = redis.StringIntMapCmd
	ZSliceCmd          = redis.ZSliceCmd
	ScanCmd            = redis.ScanCmd
	ClusterSlotsCmd    = redis.ClusterSlotsCmd
)

// NewClient creates a redis(cluster) client from yaml config, and pings the client.
func NewClient(cfg *Config) (*Client, error) {
	var c = &Client{
		cfg: cfg,
	}
	switch cfg.DeployType {
	case TypeSingle:
		c.Cmdable = redis.NewClient(&redis.Options{
			Addr:               cfg.ForSingle.Addr,
			Password:           cfg.Password,
			MaxRetries:         cfg.MaxRetries,
			MaxRetryBackoff:    time.Duration(cfg.ForSingle.MaxRetryBackoff) * time.Second,
			DialTimeout:        time.Duration(cfg.DialTimeout) * time.Second,
			ReadTimeout:        time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:       time.Duration(cfg.WriteTimeout) * time.Second,
			PoolSize:           cfg.PoolSizePerNode,
			PoolTimeout:        time.Duration(cfg.PoolTimeout) * time.Second,
			IdleTimeout:        time.Duration(cfg.IdleTimeout) * time.Second,
			IdleCheckFrequency: time.Duration(cfg.IdleCheckFrequency) * time.Second,
		})

	case TypeCluster:
		c.Cmdable = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:              cfg.ForCluster.Addrs,
			MaxRedirects:       cfg.ForCluster.MaxRedirects,
			ReadOnly:           cfg.ReadOnly,
			RouteByLatency:     cfg.ForCluster.RouteByLatency,
			Password:           cfg.Password,
			MaxRetries:         cfg.MaxRetries,
			DialTimeout:        time.Duration(cfg.DialTimeout) * time.Second,
			ReadTimeout:        time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:       time.Duration(cfg.WriteTimeout) * time.Second,
			PoolSize:           cfg.PoolSizePerNode,
			PoolTimeout:        time.Duration(cfg.PoolTimeout) * time.Second,
			IdleTimeout:        time.Duration(cfg.IdleTimeout) * time.Second,
			IdleCheckFrequency: time.Duration(cfg.IdleCheckFrequency) * time.Second,
		})

	default:
		return nil, fmt.Errorf("redis.Config.DeployType: optional enumeration list: %s, %s", TypeSingle, TypeCluster)
	}

	if _, err := c.Ping().Result(); err != nil {
		return nil, err
	}
	return c, nil
}

// Config returns config.
func (c *Client) Config() *Config {
	return c.cfg
}

// IsCluster returns whether it is a cluster.
func (c *Client) IsCluster() bool {
	return c.cfg.DeployType == TypeCluster
}

// ToSingle tries to convert it to *redis.Client.
func (c *Client) ToSingle() (*redis.Client, bool) {
	cli, ok := c.Cmdable.(*redis.Client)
	return cli, ok
}

// ToCluster tries to convert it to *redis.ClusterClient.
func (c *Client) ToCluster() (*redis.ClusterClient, bool) {
	clu, ok := c.Cmdable.(*redis.ClusterClient)
	return clu, ok
}

// LockCallback 使用分布式锁执行回调函数
// 注意：每10毫秒尝试1次上锁，且上锁后默认锁定1分钟
func (c *Client) LockCallback(lockKey string, callback func(), maxLock ...time.Duration) error {
	var d = time.Minute
	if len(maxLock) > 0 {
		d = maxLock[0]
	}
	// lock
	for lockOk, err := c.SetNX(lockKey, "", d).Result(); !lockOk; lockOk, err = c.SetNX(lockKey, "", d).Result() {
		if err != nil && !IsRedisNil(err) {
			return err
		}
		time.Sleep(time.Millisecond * 10)
	}
	// unlock
	defer c.Del(lockKey)
	// do
	callback()
	return nil
}

// Redis nil reply, .e.g. when key does not exist.
const Nil = redis.Nil

// IsRedisNil Is the redis nil reply? .e.g. when key does not exist.
func IsRedisNil(err error) bool {
	return redis.Nil == err
}

/*
List of `*Client` methods, inherited from `github.com/go-redis/redis.Cmdable` interface:
	Pipeline() redis.Pipeliner
	Pipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error)

	ClientGetName() *redis.StringCmd
	Echo(message interface{}) *redis.StringCmd
	Ping() *redis.StatusCmd
	Quit() *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
	Unlink(keys ...string) *redis.IntCmd
	Dump(key string) *redis.StringCmd
	Exists(keys ...string) *redis.IntCmd
	Expire(key string, expiration time.Duration) *redis.BoolCmd
	ExpireAt(key string, tm time.Time) *redis.BoolCmd
	Keys(pattern string) *redis.StringSliceCmd
	Migrate(host, port, key string, db int64, timeout time.Duration) *redis.StatusCmd
	Move(key string, db int64) *redis.BoolCmd
	ObjectRefCount(key string) *redis.IntCmd
	ObjectEncoding(key string) *redis.StringCmd
	ObjectIdleTime(key string) *redis.DurationCmd
	Persist(key string) *redis.BoolCmd
	PExpire(key string, expiration time.Duration) *redis.BoolCmd
	PExpireAt(key string, tm time.Time) *redis.BoolCmd
	PTTL(key string) *redis.DurationCmd
	RandomKey() *redis.StringCmd
	Rename(key, newkey string) *redis.StatusCmd
	RenameNX(key, newkey string) *redis.BoolCmd
	Restore(key string, ttl time.Duration, value string) *redis.StatusCmd
	RestoreReplace(key string, ttl time.Duration, value string) *redis.StatusCmd
	Sort(key string, sort Sort) *redis.StringSliceCmd
	SortInterfaces(key string, sort Sort) *redis.SliceCmd
	TTL(key string) *redis.DurationCmd
	Type(key string) *redis.StatusCmd
	Scan(cursor uint64, match string, count int64) *redis.ScanCmd
	SScan(key string, cursor uint64, match string, count int64) *redis.ScanCmd
	HScan(key string, cursor uint64, match string, count int64) *redis.ScanCmd
	ZScan(key string, cursor uint64, match string, count int64) *redis.ScanCmd
	Append(key, value string) *redis.IntCmd
	BitCount(key string, bitCount *redis.BitCount) *redis.IntCmd
	BitOpAnd(destKey string, keys ...string) *redis.IntCmd
	BitOpOr(destKey string, keys ...string) *redis.IntCmd
	BitOpXor(destKey string, keys ...string) *redis.IntCmd
	BitOpNot(destKey string, key string) *redis.IntCmd
	BitPos(key string, bit int64, pos ...int64) *redis.IntCmd
	Decr(key string) *redis.IntCmd
	DecrBy(key string, decrement int64) *redis.IntCmd
	Get(key string) *redis.StringCmd
	GetBit(key string, offset int64) *redis.IntCmd
	GetRange(key string, start, end int64) *redis.StringCmd
	GetSet(key string, value interface{}) *redis.StringCmd
	Incr(key string) *redis.IntCmd
	IncrBy(key string, value int64) *redis.IntCmd
	IncrByFloat(key string, value float64) *redis.FloatCmd
	MGet(keys ...string) *redis.SliceCmd
	MSet(pairs ...interface{}) *redis.StatusCmd
	MSetNX(pairs ...interface{}) *redis.BoolCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetBit(key string, offset int64, value int) *redis.IntCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	SetXX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	SetRange(key string, offset int64, value string) *redis.IntCmd
	StrLen(key string) *redis.IntCmd
	HDel(key string, fields ...string) *redis.IntCmd
	HExists(key, field string) *redis.BoolCmd
	HGet(key, field string) *redis.StringCmd
	HGetAll(key string) *redis.StringStringMapCmd
	HIncrBy(key, field string, incr int64) *redis.IntCmd
	HIncrByFloat(key, field string, incr float64) *redis.FloatCmd
	HKeys(key string) *redis.StringSliceCmd
	HLen(key string) *redis.IntCmd
	HMGet(key string, fields ...string) *redis.SliceCmd
	HMSet(key string, fields map[string]interface{}) *redis.StatusCmd
	HSet(key, field string, value interface{}) *redis.BoolCmd
	HSetNX(key, field string, value interface{}) *redis.BoolCmd
	HVals(key string) *redis.StringSliceCmd
	BLPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd
	BRPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd
	BRPopLPush(source, destination string, timeout time.Duration) *redis.StringCmd
	LIndex(key string, index int64) *redis.StringCmd
	LInsert(key, op string, pivot, value interface{}) *redis.IntCmd
	LInsertBefore(key string, pivot, value interface{}) *redis.IntCmd
	LInsertAfter(key string, pivot, value interface{}) *redis.IntCmd
	LLen(key string) *redis.IntCmd
	LPop(key string) *redis.StringCmd
	LPush(key string, values ...interface{}) *redis.IntCmd
	LPushX(key string, value interface{}) *redis.IntCmd
	LRange(key string, start, stop int64) *redis.StringSliceCmd
	LRem(key string, count int64, value interface{}) *redis.IntCmd
	LSet(key string, index int64, value interface{}) *redis.StatusCmd
	LTrim(key string, start, stop int64) *redis.StatusCmd
	RPop(key string) *redis.StringCmd
	RPopLPush(source, destination string) *redis.StringCmd
	RPush(key string, values ...interface{}) *redis.IntCmd
	RPushX(key string, value interface{}) *redis.IntCmd
	SAdd(key string, members ...interface{}) *redis.IntCmd
	SCard(key string) *redis.IntCmd
	SDiff(keys ...string) *redis.StringSliceCmd
	SDiffStore(destination string, keys ...string) *redis.IntCmd
	SInter(keys ...string) *redis.StringSliceCmd
	SInterStore(destination string, keys ...string) *redis.IntCmd
	SIsMember(key string, member interface{}) *redis.BoolCmd
	SMembers(key string) *redis.StringSliceCmd
	SMove(source, destination string, member interface{}) *redis.BoolCmd
	SPop(key string) *redis.StringCmd
	SPopN(key string, count int64) *redis.StringSliceCmd
	SRandMember(key string) *redis.StringCmd
	SRandMemberN(key string, count int64) *redis.StringSliceCmd
	SRem(key string, members ...interface{}) *redis.IntCmd
	SUnion(keys ...string) *redis.StringSliceCmd
	SUnionStore(destination string, keys ...string) *redis.IntCmd
	ZAdd(key string, members ...Z) *redis.IntCmd
	ZAddNX(key string, members ...Z) *redis.IntCmd
	ZAddXX(key string, members ...Z) *redis.IntCmd
	ZAddCh(key string, members ...Z) *redis.IntCmd
	ZAddNXCh(key string, members ...Z) *redis.IntCmd
	ZAddXXCh(key string, members ...Z) *redis.IntCmd
	ZIncr(key string, member Z) *redis.FloatCmd
	ZIncrNX(key string, member Z) *redis.FloatCmd
	ZIncrXX(key string, member Z) *redis.FloatCmd
	ZCard(key string) *redis.IntCmd
	ZCount(key, min, max string) *redis.IntCmd
	ZIncrBy(key string, increment float64, member string) *redis.FloatCmd
	ZInterStore(destination string, store ZStore, keys ...string) *redis.IntCmd
	ZRange(key string, start, stop int64) *redis.StringSliceCmd
	ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	ZRangeByScore(key string, opt ZRangeBy) *redis.StringSliceCmd
	ZRangeByLex(key string, opt ZRangeBy) *redis.StringSliceCmd
	ZRangeByScoreWithScores(key string, opt ZRangeBy) *redis.ZSliceCmd
	ZRank(key, member string) *redis.IntCmd
	ZRem(key string, members ...interface{}) *redis.IntCmd
	ZRemRangeByRank(key string, start, stop int64) *redis.IntCmd
	ZRemRangeByScore(key, min, max string) *redis.IntCmd
	ZRemRangeByLex(key, min, max string) *redis.IntCmd
	ZRevRange(key string, start, stop int64) *redis.StringSliceCmd
	ZRevRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	ZRevRangeByScore(key string, opt ZRangeBy) *redis.StringSliceCmd
	ZRevRangeByLex(key string, opt ZRangeBy) *redis.StringSliceCmd
	ZRevRangeByScoreWithScores(key string, opt ZRangeBy) *redis.ZSliceCmd
	ZRevRank(key, member string) *redis.IntCmd
	ZScore(key, member string) *redis.FloatCmd
	ZUnionStore(dest string, store ZStore, keys ...string) *redis.IntCmd
	PFAdd(key string, els ...interface{}) *redis.IntCmd
	PFCount(keys ...string) *redis.IntCmd
	PFMerge(dest string, keys ...string) *redis.StatusCmd
	BgRewriteAOF() *redis.StatusCmd
	BgSave() *redis.StatusCmd
	ClientKill(ipPort string) *redis.StatusCmd
	ClientList() *redis.StringCmd
	ClientPause(dur time.Duration) *redis.BoolCmd
	ConfigGet(parameter string) *redis.SliceCmd
	ConfigResetStat() *redis.StatusCmd
	ConfigSet(parameter, value string) *redis.StatusCmd
	DbSize() *redis.IntCmd
	FlushAll() *redis.StatusCmd
	FlushAllAsync() *redis.StatusCmd
	FlushDB() *redis.StatusCmd
	FlushDBAsync() *redis.StatusCmd
	Info(section ...string) *redis.StringCmd
	LastSave() *redis.IntCmd
	Save() *redis.StatusCmd
	Shutdown() *redis.StatusCmd
	ShutdownSave() *redis.StatusCmd
	ShutdownNoSave() *redis.StatusCmd
	SlaveOf(host, port string) *redis.StatusCmd
	Time() *redis.TimeCmd
	Eval(script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(scripts ...string) *redis.BoolSliceCmd
	ScriptFlush() *redis.StatusCmd
	ScriptKill() *redis.StatusCmd
	ScriptLoad(script string) *redis.StringCmd
	DebugObject(key string) *redis.StringCmd
	PubSubChannels(pattern string) *redis.StringSliceCmd
	PubSubNumSub(channels ...string) *redis.StringIntMapCmd
	PubSubNumPat() *redis.IntCmd
	Publish(channel, message string) *redis.IntCmd
	ClusterSlots() *redis.ClusterSlotsCmd
	ClusterNodes() *redis.StringCmd
	ClusterMeet(host, port string) *redis.StatusCmd
	ClusterForget(nodeID string) *redis.StatusCmd
	ClusterReplicate(nodeID string) *redis.StatusCmd
	ClusterResetSoft() *redis.StatusCmd
	ClusterResetHard() *redis.StatusCmd
	ClusterInfo() *redis.StringCmd
	ClusterKeySlot(key string) *redis.IntCmd
	ClusterCountFailureReports(nodeID string) *redis.IntCmd
	ClusterCountKeysInSlot(slot int) *redis.IntCmd
	ClusterDelSlots(slots ...int) *redis.StatusCmd
	ClusterDelSlotsRange(min, max int) *redis.StatusCmd
	ClusterSaveConfig() *redis.StatusCmd
	ClusterSlaves(nodeID string) *redis.StringSliceCmd
	ClusterFailover() *redis.StatusCmd
	ClusterAddSlots(slots ...int) *redis.StatusCmd
	ClusterAddSlotsRange(min, max int) *redis.StatusCmd
	GeoAdd(key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd
	GeoPos(key string, members ...string) *redis.GeoPosCmd
	GeoRadius(key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd
	GeoRadiusByMember(key, member string, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd
	GeoDist(key string, member1, member2, unit string) *redis.FloatCmd
	GeoHash(key string, members ...string) *redis.StringSliceCmd
	Command() *redis.CommandsInfoCmd
	TxPipeline() redis.Pipeliner
	TxPipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
*/
