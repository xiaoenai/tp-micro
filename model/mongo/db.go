package mongo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/goutil/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/model/redis"
)

// GO 1.9 alias
type M = bson.M
type Collection = mgo.Collection
type ObjectId = bson.ObjectId

var (
	ErrNotFound = mgo.ErrNotFound
	NewObjectId = bson.NewObjectId
)

// DB is a wrapper around sqlx.DB and redis.Client.
type DB struct {
	*mgo.Session
	Cache        *redis.Client
	dbConfig     *Config
	redisConfig  *redis.Config
	cacheableDBs map[string]*CacheableDB
}

func (c *CacheableDB) getSession() (*mgo.Session, error) {
	if c.DB.Session.Ping() != nil {
		// Creating a maintenance socket pool for session
		// connect to mongodb
		session, err := mgo.DialWithInfo(c.DB.dbConfig.Source())
		if err != nil {
			return nil, err
		}
		c.DB.Session = session
	}
	return c.DB.Session.Clone(), nil
}

// Cacheable the interface that can use cache.
// It must be orm-struct.
type Cacheable interface {
	TableName() string
}

// CacheableDB cacheable DB handle
type CacheableDB struct {
	*DB
	tableName         string
	cachePriKeyPrefix string
	cacheExpiration   time.Duration
	typeName          string
	module            *redis.Module
}

// ErrCacheNil error: *DB.Cache (redis) is nil
var ErrCacheNil = errors.New("*DB.Cache (redis) is nil")

// RegCacheableDB registers a cacheable table.
func (d *DB) RegCacheableDB(ormStructPtr Cacheable, cacheExpiration time.Duration) (*CacheableDB, error) {
	tableName := ormStructPtr.TableName()
	if _, ok := d.cacheableDBs[tableName]; ok {
		return nil, fmt.Errorf("re-register cacheable table: %s", tableName)
	}
	if !d.dbConfig.NoCache && d.Cache == nil {
		return nil, ErrCacheNil
	}

	module := redis.NewModule(d.dbConfig.Database + ":" + tableName)
	t := reflect.TypeOf(ormStructPtr)
	c := &CacheableDB{
		DB:        d,
		tableName: tableName,
		// 采用固定前缀的二级缓存key
		cachePriKeyPrefix: module.Key("_id"),
		cacheExpiration:   cacheExpiration,
		typeName:          t.String(),
		module:            module,
	}
	d.cacheableDBs[tableName] = c
	return c, nil
}

// GetCacheableDB returns the specified *CacheableDB
func (d *DB) GetCacheableDB(tableName string) (*CacheableDB, error) {
	c, ok := d.cacheableDBs[tableName]
	if !ok {
		return nil, fmt.Errorf("has not called *DB.RegCacheableDB() to register: %s", tableName)
	}
	return c, nil
}

// CacheKey cache key and values corresponding to primary keys
type CacheKey struct {
	Key         string
	FieldValues []interface{}
	isPriKey    bool
}

var emptyCacheKey = CacheKey{}

// CreateCacheKeyByFields creates the cache key string by specified fields and values.
func (c *CacheableDB) CreateCacheKeyByFields(fields []string, values []interface{}) (string, error) {
	if len(fields) != len(values) {
		return "", errors.New("CreateCacheKeyByFields(): len(fields) != len(values)")
	}
	bs, err := json.Marshal(values)
	if err != nil {
		return "", errors.New("CreateCacheKeyByFields(): " + err.Error())
	}
	return c.module.Key(strings.Join(fields, "&") + goutil.BytesToString(bs)), nil
}

// CreateCacheKey creates cache key and fields' values.
// Note: if fields is empty, auto-use primary fields.
func (c *CacheableDB) CreateCacheKey(structPtr Cacheable, fields ...string) (CacheKey, error) {
	var t = reflect.TypeOf(structPtr)
	var typeName = t.String()
	if c.typeName != typeName {
		return emptyCacheKey, fmt.Errorf("CreateCacheKey(): unmatch Cacheable: want %s, have %s", c.typeName, typeName)
	}
	var v = reflect.ValueOf(structPtr).Elem()
	var values = make([]interface{}, 0, 2)
	var cacheKey string
	if len(fields) == 0 {
		// TODO 这里必须要传fields
		return emptyCacheKey, errors.New("CreateCacheKey(): must transfer fields")
	} else {
		for i, field := range fields {
			vv := v.FieldByName(goutil.CamelString(field))
			if vv.Kind() == reflect.Ptr {
				vv = vv.Elem()
			}
			values = append(values, vv.Interface())
			// values = append(values, v.FieldByName(goutil.CamelString(field)).Interface())
			fields[i] = goutil.SnakeString(field)
		}
		var err error
		cacheKey, err = c.CreateCacheKeyByFields(fields, values)
		if err != nil {
			return emptyCacheKey, err
		}
	}
	return CacheKey{
		Key:         cacheKey,
		FieldValues: values,
	}, nil
}

// CacheGet selects one row by primary key.
// Priority from the read cache.
// Note:
//  If the cache does not exist, then write the cache;
//  @destStructPtr must be a *struct type;
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) CacheGet(destStructPtr Cacheable, fields ...string) error {
	var cacheKey, err = c.CreateCacheKey(destStructPtr, fields...)
	if err != nil {
		return err
	}

	if c.DB.dbConfig.NoCache {
		// read db
		return c.WitchCollection(func(collect *Collection) error {
			return collect.Find(c.CreateGetQuery(cacheKey.FieldValues, fields...)).One(destStructPtr)
		})
	}

	var (
		key                 = cacheKey.Key
		gettedFirstCacheKey = cacheKey.isPriKey
	)

	// read secondary cache
	if !gettedFirstCacheKey {
		var b []byte
		b, err = c.Cache.Get(key).Bytes()
		if err == nil {
			key = goutil.BytesToString(b)
			gettedFirstCacheKey = true
		} else if !redis.IsRedisNil(err) {
			return err
		}
	}

	var exist bool

	// get first cache
	if gettedFirstCacheKey {
		exist, err = c.getFirstCache(key, destStructPtr)
		if err != nil {
			return err
		}
		if exist {
			// check
			if !cacheKey.isPriKey && !c.checkSecondCache(destStructPtr, fields, cacheKey.FieldValues) {
				c.Cache.Del(cacheKey.Key)
			} else {
				return nil
			}
		}
	}

	// to lock or get first cache
	c.Cache.LockCallback("lock_"+key, func() {
		var b []byte
		if !exist {
		FIRST:
			if gettedFirstCacheKey {
				exist, err = c.getFirstCache(key, destStructPtr)
				if exist {
					err = nil
					return
				}
				if err != nil {
					return
				}
			} else {
				b, err = c.Cache.Get(key).Bytes()
				if err == nil {
					key = goutil.BytesToString(b)
					gettedFirstCacheKey = true
					goto FIRST
				} else if !redis.IsRedisNil(err) {
					return
				}
			}
		}

		err = c.WitchCollection(func(collect *Collection) error {
			return collect.Find(c.CreateGetQuery(cacheKey.FieldValues, fields...)).One(destStructPtr)
		})
		if err != nil {
			return
		}

		key, err = c.createPrikey(destStructPtr)
		if err != nil {
			tp.Errorf("CacheGet(): createPrikey: %s", err.Error())
			err = nil
			return
		}

		// write cache
		data, _ := json.Marshal(destStructPtr)
		err = c.Cache.Set(key, data, c.cacheExpiration).Err()
		if err == nil && !cacheKey.isPriKey {
			err = c.Cache.Set(cacheKey.Key, key, c.cacheExpiration).Err()
		}
		if err != nil {
			tp.Errorf("CacheGet(): %s", err.Error())
			err = nil
		}
	})

	return err
}

func (c *CacheableDB) checkSecondCache(destStructPtr Cacheable, fields []string, values []interface{}) bool {
	v := reflect.ValueOf(destStructPtr).Elem()
	for i, field := range fields {
		vv := v.FieldByName(goutil.CamelString(field))
		if vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}
		if values[i] != vv.Interface() {
			return false
		}
	}
	return true
}

// get first cache
func (c *CacheableDB) getFirstCache(key string, destStructPtr Cacheable) (bool, error) {
	data, err := c.Cache.Get(key).Bytes()
	if err == nil {
		err = json.Unmarshal(data, destStructPtr)
		if err == nil {
			return true, nil
		}
		tp.Errorf("CacheGet(): %s", err.Error())

	} else if !redis.IsRedisNil(err) {
		return false, err
	}
	return false, nil
}

// PutCache caches one row by primary key.
// Note:
//  @destStructPtr must be a *struct type;
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) PutCache(srcStructPtr Cacheable, fields ...string) error {
	if c.DB.dbConfig.NoCache {
		return nil
	}
	cacheKey, err := c.CreateCacheKey(srcStructPtr, fields...)
	if err != nil {
		return err
	}
	data, err := json.Marshal(srcStructPtr)
	if err != nil {
		return err
	}

	key := cacheKey.Key

	if cacheKey.isPriKey {
		return c.Cache.Set(key, data, c.cacheExpiration).Err()
	}

	// secondary cache
	key, err = c.createPrikey(srcStructPtr)
	if err != nil {
		return err
	}
	err = c.Cache.Set(key, data, c.cacheExpiration).Err()
	if err != nil {
		return err
	}
	return c.Cache.Set(cacheKey.Key, key, c.cacheExpiration).Err()
}

// DeleteCache deletes one row form cache by primary key.
// Note:
//  @destStructPtr must be a *struct type;
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) DeleteCache(srcStructPtr Cacheable, fields ...string) error {
	if c.DB.dbConfig.NoCache {
		return nil
	}
	cacheKey, err := c.CreateCacheKey(srcStructPtr, fields...)
	if err != nil {
		return err
	}
	var keys = []string{cacheKey.Key}
	// secondary cache
	if !cacheKey.isPriKey {
		// get first cache key
		firstKey, err := c.Cache.Get(cacheKey.Key).Result()
		if err == nil {
			keys = append(keys, firstKey)
		}
	}
	return c.Cache.Del(keys...).Err()
}

func (c *CacheableDB) createPrikey(structPtr Cacheable) (string, error) {
	var v = reflect.ValueOf(structPtr).Elem()
	// Parse objectIdHex's hex
	objectIdHex := ObjectId(v.FieldByName("Id").String())
	values := []interface{}{
		objectIdHex.Hex(),
	}
	bs, err := json.Marshal(values)
	if err != nil {
		return "", errors.New("*CacheableDB.createPrikey(): " + err.Error())
	}
	return c.cachePriKeyPrefix + goutil.BytesToString(bs), nil
}

func (c *CacheableDB) CreateGetQuery(values []interface{}, whereFields ...string) M {
	m := M{}
	if len(whereFields) == 0 {
		return m
	}
	for index, col := range whereFields {
		m[col] = values[index]
	}
	return m
}

// Common method
func (c *CacheableDB) WitchCollection(s func(*Collection) error) error {
	session, err := c.getSession()
	if err != nil {
		return fmt.Errorf("Mongodb connection error:%s", err)
	}
	defer func() {
		session.Close()
		if err := recover(); err != nil {
			tp.Errorf("Mongodb close session err:%s", err)
		}
	}()
	collection := c.DB.DB(c.DB.dbConfig.Database).C(c.tableName)
	return s(collection)
}
