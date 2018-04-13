// Package model is a mysql ORM(Object Role Modeling) package with redis cache.
package model

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
	"github.com/xiaoenai/ants/model/redis"
	"github.com/xiaoenai/ants/model/sqlx"
	"github.com/xiaoenai/ants/model/sqlx/reflectx"
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
// Note: if fields is empty, auto-use primary fields.
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
// Note:
//  If whereFields is empty, auto-use primary fields.
func (c *CacheableDB) CreateGetQuery(whereFields ...string) string {
	if len(whereFields) == 0 {
		whereFields = c.priCols
	}
	var queryAll = "SELECT"
	for _, col := range c.cols {
		queryAll += " `" + col + "`,"
	}
	queryAll = queryAll[:len(queryAll)-1] + " FROM " + c.tableName + " WHERE"
	for _, col := range whereFields {
		queryAll += " `" + col + "`=? AND"
	}
	queryAll = queryAll[:len(queryAll)-4] + " LIMIT 1;"
	return queryAll
}

// CacheGet selects one row by primary key.
// Priority from the read cache.
// Note:
//  If the cache does not exist, then write the cache;
//  @destStructPtr must be a *struct type;
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
			// check
			if !cacheKey.isPriKey && !c.checkSecondCache(structElemValue, fields, cacheKey.FieldValues) {
				c.Cache.Del(cacheKey.Key)
				return ErrNoRows
			}
			return nil
		}
	}

	// to lock or get first cache
	c.Cache.LockCallback("lock_"+key, func() {
		var b []byte
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
// Note:
//  @destStructPtr must be a *struct type;
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
// Note:
//  @destStructPtr must be a *struct type;
//  If fields is empty, auto-use primary fields.
func (c *CacheableDB) DeleteCache(srcStructPtr Cacheable, fields ...string) error {
	if c.DB.dbConfig.NoCache {
		return nil
	}
	cacheKey, structElemValue, err := c.CreateCacheKey(srcStructPtr, fields...)
	if err != nil {
		return err
	}

	key := cacheKey.Key
	if cacheKey.isPriKey {
		return c.Cache.Del(key).Err()
	}

	// secondary cache
	key, err = c.createPrikey(structElemValue)
	if err != nil {
		return err
	}
	return c.Cache.Del(key, cacheKey.Key).Err()
}

// DbOrTx contains all the exportable methods of *sqlx.DB
type DbOrTx interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	// NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	// PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	// PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Callback non-transactional operations.
func (d *DB) Callback(fn func(DbOrTx) error, tx ...*sqlx.Tx) error {
	if fn == nil {
		return nil
	}
	if len(tx) > 0 && tx[0] != nil {
		return fn(tx[0])
	}
	return fn(d)
}

// TransactCallback transactional operations.
// note: if an error is returned, the rollback method should be invoked outside the function.
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
	return
}

// ErrNoRows is returned by Scan when QueryRow doesn't return a
// row. In such a case, QueryRow returns a placeholder *Row value that
// defers this error until a Scan.
var ErrNoRows = sql.ErrNoRows

// IsNoRows is the data exist or not.
func IsNoRows(err error) bool {
	return ErrNoRows == err
}
