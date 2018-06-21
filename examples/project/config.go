package main

import (
	"time"

	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/model/etcd"
	"github.com/xiaoenai/tp-micro/model/mysql"
	"github.com/xiaoenai/tp-micro/model/redis"

	"github.com/xiaoenai/tp-micro/examples/project/logic/model"
)

type config struct {
	Srv      micro.SrvConfig `yaml:"srv"`
	Etcd     etcd.EasyConfig `yaml:"etcd"`
	DB       mysql.Config    `yaml:"db"`
	Redis    redis.Config    `yaml:"redis"`
	LogLevel string          `yaml:"log_level"`
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err != nil {
		return err
	}
	if len(c.LogLevel) == 0 {
		c.LogLevel = "TRACE"
	}
	tp.SetLoggerLevel(c.LogLevel)
	err = model.Init(c.DB, c.Redis)
	if err != nil {
		tp.Errorf("%v", err)
	}
	return nil
}

var cfg = &config{
	Srv: micro.SrvConfig{
		ListenAddress:     ":9090",
		EnableHeartbeat:   true,
		PrintDetail:       true,
		CountTime:         true,
		SlowCometDuration: time.Millisecond * 500,
	},
	Etcd: etcd.EasyConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	},
	DB: mysql.Config{
		Port: 3306,
	},
	Redis:    *redis.NewConfig(),
	LogLevel: "TRACE",
}

func init() {
	goutil.WritePidFile()
	cfgo.MustReg("project", cfg)
}
