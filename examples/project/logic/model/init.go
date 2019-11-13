package model

import (
	"strings"
	"time"

	"github.com/xiaoenai/tp-micro/v6/model/mongo"
	"github.com/xiaoenai/tp-micro/v6/model/mysql"
	"github.com/xiaoenai/tp-micro/v6/model/redis"
)

// mysqlHandler preset mysql DB handler
var mysqlHandler = mysql.NewPreDB()

// mongoHandler preset mongo DB handler
var mongoHandler = mongo.NewPreDB()

var (
	redisClient *redis.Client
	cacheExpire time.Duration
)

// Init initializes the model packet.
func Init(mysqlConfig *mysql.Config, mongoConfig *mongo.Config, redisConfig *redis.Config, _cacheExpire time.Duration) error {
	cacheExpire = _cacheExpire
	var err error
	if redisConfig != nil {
		redisClient, err = redis.NewClient(redisConfig)
		if err != nil {
			return err
		}
	}
	if mysqlConfig != nil {
		if err = mysqlHandler.Init2(mysqlConfig, redisClient); err != nil {
			return err
		}
	}
	if mongoConfig != nil {
		if err = mongoHandler.Init2(mongoConfig, redisClient); err != nil {
			return err
		}
	}
	return nil
}

// GetMysqlDB returns the mysql DB handler.
func GetMysqlDB() *mysql.DB {
	return mysqlHandler.DB
}

// GetMongoDB returns the mongo DB handler.
func GetMongoDB() *mongo.DB {
	return mongoHandler.DB
}

// GetRedis returns the redis client.
func GetRedis() *redis.Client {
	return redisClient
}

func index(s string, sub ...string) int {
	var i, ii = -1, -1
	for _, ss := range sub {
		ii = strings.Index(s, ss)
		if ii != -1 && (ii < i || i == -1) {
			i = ii
		}
	}
	return i
}

func insertZeroDeletedTsField(whereCond string) string {
	whereCond = strings.TrimSpace(whereCond)
	whereCond = strings.TrimRight(whereCond, ";")
	i := index(
		whereCond,
		"`deleted_ts`",
		" deleted_ts",
	)
	if i != -1 {
		return whereCond
	}
	i = index(
		whereCond,
		"ORDER BY", "order by",
		"GROUP BY", "group by",
		"OFFSET", "offset",
		"LIMIT", "limit",
	)
	if i == -1 {
		return whereCond + " AND `deleted_ts`=0"
	}
	return whereCond[:i] + " AND `deleted_ts`=0 " + whereCond[i:]
}
