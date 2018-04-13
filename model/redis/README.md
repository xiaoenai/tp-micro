# redis

Redis (cluster) client package.
Based on `github.com/go-redis/redis` v6.5.2

## Example

```
package main

import (
	"log"
	"time"

	"github.com/xiaoenai/redis"
)

func main() {
	cfg, err := redis.ReadConfig("test_redis")
	if err != nil {
		log.Fatalf("redis.ReadConfig(\"test_redis\"): %v", err)
	}
	c, err := redis.NewClient(cfg)
	if err != nil {
		log.Fatalf("redis.NewClient(\"test_redis\"): %v", err)
	}

	m := redis.NewModule("test", "v1.0")

	s, err := c.Set(m.Key("a_key"), "a_value", time.Second).Result()
	if err != nil {
		log.Fatalf("c.Set().Result() error: %v", err)
	}
	log.Printf("c.Set().Result() result: %s", s)

	s, err = c.Get(m.Key("a_key")).Result()
	if err != nil {
		log.Fatalf("c.Get().Result() error: %v", err)
	}
	log.Printf("c.Get().Result() result: %s", s)
	time.Sleep(2 * time.Second)

	s, err = c.Get(m.Key("a_key")).Result()
	if err == nil {
		log.Fatalf("[after 2s] c.Get().Result() result: %s", s)
	}
	log.Printf("[after 2s] c.Get().Result() error: %s", err)
}
```

## API doc

[http://godoc.org/gopkg.in/go-redis/redis.v6](http://godoc.org/gopkg.in/go-redis/redis.v6)


List of `*Client` methods, inherited from `github.com/go-redis/redis.Cmdable` interface:

```
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
```
