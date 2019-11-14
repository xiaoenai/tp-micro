package main

import (
	"time"

	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/goutil"
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/model/etcd"
	"github.com/xiaoenai/tp-micro/v6/model/mongo"
	"github.com/xiaoenai/tp-micro/v6/model/mysql"
	"github.com/xiaoenai/tp-micro/v6/model/redis"

	"github.com/xiaoenai/tp-micro/v6/examples/project/logic/model"
)

type config struct {
	Srv         micro.SrvConfig `yaml:"srv"`
	Cli         micro.CliConfig `yaml:"cli"`
	Etcd        etcd.EasyConfig `yaml:"etcd"`
	Mysql       mysql.Config    `yaml:"mysql"`
	Mongo       mongo.Config    `yaml:"mongo"`
	Redis       redis.Config    `yaml:"redis"`
	CacheExpire time.Duration   `yaml:"cache_expire"`
	LogLevel    string          `yaml:"log_level"`
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err != nil {
		return err
	}
	if c.CacheExpire == 0 {
		c.CacheExpire = time.Hour * 24
	}
	if len(c.LogLevel) == 0 {
		c.LogLevel = "TRACE"
	}
	erpc.SetLoggerLevel(c.LogLevel)
	var (
		mysqlConfig *mysql.Config
		mongoConfig *mongo.Config
		redisConfig = &c.Redis
	)
	if len(c.Mysql.Host) > 0 {
		mysqlConfig = &c.Mysql
	}
	if len(c.Mongo.Addrs) > 0 {
		mongoConfig = &c.Mongo
	}
	err = model.Init(mysqlConfig, mongoConfig, redisConfig, c.CacheExpire)
	if err != nil {
		erpc.Errorf("%v", err)
	}
	return nil
}

var cfg = &config{
	Srv: micro.SrvConfig{
		ListenAddress:     ":0",
		EnableHeartbeat:   true,
		PrintDetail:       true,
		CountTime:         true,
		SlowCometDuration: time.Millisecond * 500,
	},
	Etcd: etcd.EasyConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	},
	Redis:       *redis.NewConfig(),
	CacheExpire: time.Hour * 24,
	LogLevel:    "TRACE",
}

func init() {
	goutil.WritePidFile()
	cfgo.MustReg("project", cfg)
}
