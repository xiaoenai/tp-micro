package model

import (
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
