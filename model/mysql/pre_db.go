package mysql

import (
	"fmt"
	"time"

	"github.com/henrylee2cn/goutil"
	"github.com/xiaoenai/tp-micro/model/redis"
	"github.com/xiaoenai/tp-micro/model/sqlx"
	"github.com/xiaoenai/tp-micro/model/sqlx/reflectx"
)

// PreDB preset *DB
type PreDB struct {
	*DB
	preFuncs map[string]func() error
	inited   bool
}

// NewPreDB creates a unconnected *DB
func NewPreDB() *PreDB {
	return &PreDB{
		DB: &DB{
			cacheableDBs: make(map[string]*CacheableDB),
		},
		preFuncs: make(map[string]func() error),
	}
}

// Init initialize *DB.
func (p *PreDB) Init(dbConfig *Config, redisConfig *redis.Config) (err error) {
	var cache *redis.Client
	if !dbConfig.NoCache && redisConfig != nil {
		cache, err = redis.NewClient(redisConfig)
		if err != nil {
			return err
		}
	}
	return p.Init2(dbConfig, cache)
}

// Init2 initialize *DB.
func (p *PreDB) Init2(dbConfig *Config, redisClient *redis.Client) (err error) {
	p.DB.DB, err = sqlx.Connect("mysql", dbConfig.Source())
	if err != nil {
		return err
	}
	p.DB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	p.DB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	p.DB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)
	// p.DB.MapperFunc(goutil.SnakeString)
	p.DB.Mapper = reflectx.NewMapperFunc("json", goutil.SnakeString)
	p.DB.dbConfig = dbConfig
	if !dbConfig.NoCache && redisClient != nil {
		p.DB.Cache = redisClient
		p.DB.redisConfig = redisClient.Config()
	}

	for _, preFunc := range p.preFuncs {
		if err = preFunc(); err != nil {
			return err
		}
	}
	p.inited = true
	return nil
}

// RegCacheableDB registers a cacheable table.
func (p *PreDB) RegCacheableDB(ormStructPtr Cacheable, cacheExpiration time.Duration, initQuery string, args ...interface{}) (*CacheableDB, error) {
	if p.inited {
		if len(initQuery) > 0 {
			_, err := p.DB.Exec(initQuery, args...)
			if err != nil {
				return nil, err
			}
		}
		return p.DB.RegCacheableDB(ormStructPtr, cacheExpiration)
	}

	tableName := ormStructPtr.TableName()
	if _, ok := p.preFuncs[ormStructPtr.TableName()]; ok {
		return nil, fmt.Errorf("re-register cacheable table: %s", tableName)
	}
	var cacheableDB = new(CacheableDB)
	var preFunc = func() error {
		if len(initQuery) > 0 {
			_, err := p.DB.Exec(initQuery, args...)
			if err != nil {
				return err
			}
		}
		_cacheableDB, err := p.DB.RegCacheableDB(ormStructPtr, cacheExpiration)
		if err == nil {
			*cacheableDB = *_cacheableDB
			p.DB.cacheableDBs[tableName] = cacheableDB
		}
		return err
	}
	p.preFuncs[tableName] = preFunc
	return cacheableDB, nil
}
