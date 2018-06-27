package model

import (
	"strings"

	"github.com/xiaoenai/tp-micro/model/mongo"
	"github.com/xiaoenai/tp-micro/model/mysql"
	"github.com/xiaoenai/tp-micro/model/redis"
)

// mysqlHandler preset mysql DB handler
var mysqlHandler = mysql.NewPreDB()

// mongoHandler preset mongo DB handler
var mongoHandler = mongo.NewPreDB()

var redisClient *redis.Client

// Init initializes the model packet.
func Init(mysqlConfig *mysql.Config, mongoConfig *mongo.Config, redisConfig *redis.Config) error {
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

func insertZeroDeletedTsField(whereCond string) string {
	whereCond = strings.TrimSpace(whereCond)
	whereCond = strings.TrimRight(whereCond, ";")
	i := strings.Index(whereCond, "OFFSET")
	if i == -1 {
		i = strings.Index(whereCond, "offset")
		if i == -1 {
			i = strings.Index(whereCond, "Offset")
		}
	}
	i2 := strings.Index(whereCond, "LIMIT")
	if i2 == -1 {
		i2 = strings.Index(whereCond, "limit")
		if i2 == -1 {
			i2 = strings.Index(whereCond, "Limit")
		}
	}
	if i > i2 && i2 != -1 {
		i = i2
	}
	if i == -1 {
		return whereCond + " AND `deleted_ts`=0"
	}
	return whereCond[:i] + " AND `deleted_ts`=0 " + whereCond[i:]
}
