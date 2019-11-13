package mongo

import (
	"fmt"
	"time"

	"github.com/xiaoenai/tp-micro/v6/model/redis"
	"gopkg.in/mgo.v2"
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

// Init initialize *DB.
func (p *PreDB) Init2(dbConfig *Config, redisClient *redis.Client) (err error) {
	// connect to mongodb
	db, err := mgo.DialWithInfo(dbConfig.Source())
	if err != nil {
		return err
	}
	p.DB.Session = db
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
func (p *PreDB) RegCacheableDB(ormStructPtr Cacheable, cacheExpiration time.Duration) (*CacheableDB, error) {
	if p.inited {
		return p.DB.RegCacheableDB(ormStructPtr, cacheExpiration)
	}

	tableName := ormStructPtr.TableName()
	if _, ok := p.preFuncs[ormStructPtr.TableName()]; ok {
		return nil, fmt.Errorf("re-register cacheable table: %s", tableName)
	}
	var cacheableDB = new(CacheableDB)

	var preFunc = func() error {
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
