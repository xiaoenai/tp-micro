package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/goutil/errors"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/model/redis"
	"github.com/xiaoenai/tp-micro/model/sqlx"
	"github.com/xiaoenai/tp-micro/model/sqlx/reflectx"
)

// DB is a wrapper around sqlx.DB and redis.Client.
type DB struct {
	*sqlx.DB
	Cache        *redis.Client
	dbConfig     *Config
	redisConfig  *redis.Config
	cacheableDBs map[string]*CacheableDB
}

// Connect to a database and verify with a ping.
func Connect(dbConfig *Config, redisConfig *redis.Config) (*DB, error) {
	var cache *redis.Client
	if !dbConfig.NoCache && redisConfig != nil {
		var err error
		cache, err = redis.NewClient(redisConfig)
		if err != nil {
			return nil, err
		}
	}

	// this Pings the database trying to connect, panics on error
	// use sqlx.Open() for sql.Open() semantics
	db, err := sqlx.Connect("mysql", dbConfig.Source())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)
	// db.MapperFunc(goutil.SnakeString)
	db.Mapper = reflectx.NewMapperFunc("json", goutil.SnakeString)

	return &DB{
		DB:           db,
		dbConfig:     dbConfig,
		Cache:        cache,
		redisConfig:  redisConfig,
		cacheableDBs: make(map[string]*CacheableDB),
	}, nil
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
	cols              []string
	priCols           []string
	cacheExpiration   time.Duration
	typeName          string
	priFieldsIndex    []int          // primary column index in struct
	fieldsIndexMap    map[string]int // key:colName, value:field index in struct
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

	type Col struct {
		ColumnName string `json:"COLUMN_NAME" db:"COLUMN_NAME"`
		ColumnKey  string `json:"COLUMN_KEY" db:"COLUMN_KEY"`
	}

	var colsResult []*Col
	err := d.DB.Select(&colsResult, `SELECT COLUMN_NAME, COLUMN_KEY FROM information_schema.columns WHERE table_schema = ? AND table_name = ?;`, d.dbConfig.Database, tableName)
	if err != nil {
		return nil, fmt.Errorf("RegCacheableDB(): %s", err.Error())
	}

	priCols := make([]string, 0, 1)
	cols := make([]string, 0, len(colsResult))
	for _, col := range colsResult {
		cols = append(cols, col.ColumnName)
		if col.ColumnKey == "PRI" {
			priCols = append(priCols, col.ColumnName)
		}
	}

	if len(priCols) == 0 {
		return nil, fmt.Errorf("RegCacheableDB(): table '%s.%s' has no primary key", d.dbConfig.Database, tableName)
	}
	sort.Strings(priCols)

	t := reflect.TypeOf(ormStructPtr)
	var typeName = t.String()
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("ormStructPtr must be *struct type: %s", typeName)
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ormStructPtr must be *struct type: %s", typeName)
	}
	structMap := d.Mapper.TypeMap(t)
	var fieldsIndexMap = make(map[string]int, len(structMap.Index))
	for i, idx := range structMap.Index {
		for _, col := range cols {
			if col == idx.Name {
				fieldsIndexMap[col] = i
				break
			}
		}
	}
	// tp.Debugf("fields:%#v", fieldsIndexMap)

	var priFieldsIndex = make([]int, len(priCols))
	for i, col := range priCols {
		priFieldsIndex[i] = fieldsIndexMap[col]
	}

	module := redis.NewModule(d.dbConfig.Database + ":" + tableName)
	c := &CacheableDB{
		DB:                d,
		tableName:         tableName,
		cachePriKeyPrefix: module.Key(strings.Join(priCols, "&")),
		cols:              cols,
		priCols:           priCols,
		cacheExpiration:   cacheExpiration,
		typeName:          typeName,
		priFieldsIndex:    priFieldsIndex,
		fieldsIndexMap:    fieldsIndexMap,
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

var emptyValue = reflect.Value{}

// CreateCacheKey creates cache key and fields' values.
// NOTE:
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) CreateCacheKey(structPtr Cacheable, fields ...string) (CacheKey, reflect.Value, error) {
	var t = reflect.TypeOf(structPtr)
	var typeName = t.String()
	if c.typeName != typeName {
		return emptyCacheKey, emptyValue, fmt.Errorf("CreateCacheKey(): unmatch Cacheable: want %s, have %s", c.typeName, typeName)
	}
	var (
		v        = reflect.ValueOf(structPtr).Elem()
		values   = make([]interface{}, 0, 2)
		cacheKey string
		isPriKey bool
	)
	if len(fields) == 0 {
		// use primary fields
		for _, idx := range c.priFieldsIndex {
			vv := v.Field(idx)
			values = append(values, vv.Interface())
		}
		bs, err := json.Marshal(values)
		if err != nil {
			return emptyCacheKey, emptyValue, errors.New("CreateCacheKey(): " + err.Error())
		}
		isPriKey = true
		cacheKey = c.cachePriKeyPrefix + goutil.BytesToString(bs)
	} else {
		for i, field := range fields {
			fields[i] = goutil.SnakeString(field)
			vv := v.FieldByName(goutil.CamelString(field))
			values = append(values, vv.Interface())
		}
		var err error
		cacheKey, err = c.CreateCacheKeyByFields(fields, values)
		if err != nil {
			return emptyCacheKey, emptyValue, err
		}
		if strings.HasPrefix(cacheKey, c.cachePriKeyPrefix+"[") {
			isPriKey = true
		}
	}
	return CacheKey{
		Key:         cacheKey,
		FieldValues: values,
		isPriKey:    isPriKey,
	}, v, nil
}

func (c *CacheableDB) createPrikey(structElemValue reflect.Value) (string, error) {
	var (
		values = make([]interface{}, 0, 2)
	)
	for _, idx := range c.priFieldsIndex {
		values = append(values, structElemValue.Field(idx).Interface())
	}
	bs, err := json.Marshal(values)
	if err != nil {
		return "", errors.New("*CacheableDB.createPrikey(): " + err.Error())
	}
	return c.cachePriKeyPrefix + goutil.BytesToString(bs), nil
}

// CreateGetQuery creates query string of selecting one row data.
// NOTE:
//  If whereFields is empty, auto-use primary fields.
func (c *CacheableDB) CreateGetQuery(whereFields ...string) string {
	if len(whereFields) == 0 {
		whereFields = c.priCols
	}
	var queryAll = "SELECT"
	for _, col := range c.cols {
		queryAll += " `" + col + "`,"
	}
	queryAll = queryAll[:len(queryAll)-1] + " FROM `" + c.tableName + "` WHERE"
	for _, col := range whereFields {
		queryAll += " `" + col + "`=? AND"
	}
	queryAll = queryAll[:len(queryAll)-4] + " LIMIT 1;"
	return queryAll
}

// CacheGet selects one row by primary key.
// Priority from the read cache.
// NOTE:
//  If the cache does not exist, then write the cache;
//  destStructPtr must be a *struct type;
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) CacheGet(destStructPtr Cacheable, fields ...string) error {
	var cacheKey, structElemValue, err = c.CreateCacheKey(destStructPtr, fields...)
	if err != nil {
		return err
	}

	if c.DB.dbConfig.NoCache {
		// read db
		return c.DB.Get(destStructPtr, c.CreateGetQuery(fields...), cacheKey.FieldValues...)
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
		// clean
		c.cleanDestCacheable(structElemValue)

		exist, err = c.getFirstCache(key, destStructPtr)
		if err != nil {
			return err
		}
		if exist {
			// check secondary cache
			if !cacheKey.isPriKey && !c.checkSecondCache(structElemValue, fields, cacheKey.FieldValues) {
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

		// read db
		err = c.DB.Get(destStructPtr, c.CreateGetQuery(fields...), cacheKey.FieldValues...)
		if err != nil {
			return
		}
		key, err = c.createPrikey(structElemValue)
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

func (c *CacheableDB) createCacheKeyByWhere(structPtr Cacheable, whereNamedCond string) (CacheKey, string, error) {
	whereCond, values, err := c.BindNamed(whereNamedCond, structPtr)
	if err != nil {
		return emptyCacheKey, whereCond, err
	}
	bs, err := json.Marshal(values)
	if err != nil {
		return emptyCacheKey, whereCond, errors.New("CreateCacheKeyByFields(): " + err.Error())
	}
	return CacheKey{
		Key:         c.module.Key(whereCond + goutil.BytesToString(bs)),
		FieldValues: values,
		isPriKey:    false,
	}, whereCond, nil
}

func (c *CacheableDB) createGetQueryByWhere(whereCond string) string {
	var queryAll = "SELECT"
	for _, col := range c.cols {
		queryAll += " `" + col + "`,"
	}
	return queryAll[:len(queryAll)-1] + " FROM `" + c.tableName + "` WHERE " + whereCond + " LIMIT 1;"
}

// CacheGetByWhere selects one row by the whereNamedCond.
// Priority from the read cache.
// NOTE:
//  If the cache does not exist, then write the cache;
//  destStructPtr must be a *struct type;
//  whereNamedCond e.g. 'id=:id AND created_at>1520000000'.
func (c *CacheableDB) CacheGetByWhere(destStructPtr Cacheable, whereNamedCond string) error {
	cacheKey, whereCond, err := c.createCacheKeyByWhere(destStructPtr, whereNamedCond)
	if err != nil {
		return err
	}
	structElemValue := reflect.ValueOf(destStructPtr).Elem()

	if c.DB.dbConfig.NoCache {
		// read db
		return c.DB.Get(destStructPtr, c.createGetQueryByWhere(whereCond), cacheKey.FieldValues...)
	}

	var (
		key                 = cacheKey.Key
		gettedFirstCacheKey bool
		b                   []byte
		exist               bool
	)

	// read secondary cache
	b, err = c.Cache.Get(key).Bytes()
	if err == nil {
		key = goutil.BytesToString(b)
		gettedFirstCacheKey = true
	} else if !redis.IsRedisNil(err) {
		return err
	}

	// get first cache
	if gettedFirstCacheKey {
		// clean
		c.cleanDestCacheable(structElemValue)

		exist, err = c.getFirstCache(key, destStructPtr)
		if err != nil {
			return err
		}
		if exist {
			// check secondary cache
			cacheKey2, _, _ := c.createCacheKeyByWhere(destStructPtr, whereNamedCond)
			if cacheKey2.Key != cacheKey.Key {
				c.Cache.Del(cacheKey.Key)
			} else {
				return nil
			}
		}
	}

	// to lock or get first cache
	c.Cache.LockCallback("lock_"+key, func() {
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

		// read db
		err = c.DB.Get(destStructPtr, c.createGetQueryByWhere(whereCond), cacheKey.FieldValues...)
		if err != nil {
			return
		}
		key, err = c.createPrikey(structElemValue)
		if err != nil {
			tp.Errorf("CacheGetByWhere(): createPrikey: %s", err.Error())
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
			tp.Errorf("CacheGetByWhere(): %s", err.Error())
			err = nil
		}
	})

	return err
}

func (c *CacheableDB) cleanDestCacheable(destStructElemValue reflect.Value) {
	for _, i := range c.fieldsIndexMap {
		fv := destStructElemValue.Field(i)
		fv.Set(reflect.New(fv.Type()).Elem())
	}
}

func (c *CacheableDB) checkSecondCache(destStructElemValue reflect.Value, fields []string, values []interface{}) bool {
	for i, field := range fields {
		vv := destStructElemValue.FieldByName(goutil.CamelString(field))
		if vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
			if vv.Kind() == reflect.Invalid {
				if values[i] != nil {
					return false
				}
			} else if vv.Interface() != reflect.ValueOf(values[i]).Elem().Interface() {
				return false
			}
		} else if values[i] != vv.Interface() {
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
// NOTE:
//  destStructPtr must be a *struct type;
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) PutCache(srcStructPtr Cacheable, fields ...string) error {
	if c.DB.dbConfig.NoCache {
		return nil
	}
	cacheKey, structElemValue, err := c.CreateCacheKey(srcStructPtr, fields...)
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
	key, err = c.createPrikey(structElemValue)
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
// NOTE:
//  destStructPtr must be a *struct type;
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) DeleteCache(srcStructPtr Cacheable, fields ...string) error {
	if c.DB.dbConfig.NoCache {
		return nil
	}
	cacheKey, _, err := c.CreateCacheKey(srcStructPtr, fields...)
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

// Callback non-transactional operations.
func (d *DB) Callback(fn func(sqlx.DbOrTx) error, tx ...*sqlx.Tx) error {
	if fn == nil {
		return nil
	}
	if len(tx) > 0 && tx[0] != nil {
		return fn(tx[0])
	}
	return fn(d)
}

// TransactCallback transactional operations.
// nOTE: if an error is returned, the rollback method should be invoked outside the function.
func (d *DB) TransactCallback(fn func(*sqlx.Tx) error, tx ...*sqlx.Tx) (err error) {
	if fn == nil {
		return
	}
	var _tx *sqlx.Tx
	if len(tx) > 0 {
		_tx = tx[0]
	}
	if _tx == nil {
		_tx, err = d.Beginx()
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				_tx.Rollback()
			} else {
				_tx.Commit()
			}
		}()
	}
	err = fn(_tx)
	return err
}

// CallbackInSession non-transactional operations in one session.
func (d *DB) CallbackInSession(fn func(context.Context, *sqlx.Conn) error, ctx ...context.Context) error {
	if fn == nil {
		return nil
	}
	var _ctx = context.Background()
	if len(ctx) > 0 {
		_ctx = ctx[0]
	}
	conn, err := d.Conn(_ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(_ctx, conn)
}

// TransactCallbackInSession transactional operations in one session.
// nOTE: if an error is returned, the rollback method should be invoked outside the function.
func (d *DB) TransactCallbackInSession(fn func(context.Context, *sqlx.Tx) error, ctx ...context.Context) (err error) {
	if fn == nil {
		return
	}
	var _ctx = context.Background()
	if len(ctx) > 0 {
		_ctx = ctx[0]
	}
	conn, err := d.Conn(_ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	_tx, err := conn.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_tx.Rollback()
		} else {
			_tx.Commit()
		}
	}()
	err = fn(_ctx, _tx)
	return err
}

// ErrNoRows is returned by Scan when QueryRow doesn't return a
// row. In such a case, QueryRow returns a placeholder *Row value that
// defers this error until a Scan.
var ErrNoRows = sql.ErrNoRows

// IsNoRows is the data exist or not.
func IsNoRows(err error) bool {
	return ErrNoRows == err
}
