package model

import (
	"github.com/xiaoenai/ants/model/mysql"
	"github.com/xiaoenai/ants/model/redis"
)

// dbHandler preset DB handler
var dbHandler = mysql.NewPreDB()

// Init initializes the model packet.
func Init(dbConfig mysql.Config, redisConfig redis.Config) error {
	return dbHandler.Init(&dbConfig, &redisConfig)
}

// GetDB returns the DB handler.
func GetDB() *mysql.DB {
	return dbHandler.DB
}

// GetRedis returns the redis client.
func GetRedis() *redis.Client {
	return dbHandler.DB.Cache
}
