package mongo

import (
	"time"

	"github.com/henrylee2cn/cfgo"
	"gopkg.in/mgo.v2"
)

// Config mongodb setting
type Config struct {
	Addrs    []string      `yaml:"addrs"`
	Timeout  time.Duration `yaml:"timeout"`
	Database string        `yaml:"database"`
	// Username and Password inform the credentials for the initial authentication
	// done on the database defined by the Source field. See Session.Login.
	Username string `yaml:"username"`
	Password string `yaml:"passward"`
	// PoolLimit defines the per-server socket pool limit. Defaults to 4096.
	// See Session.SetPoolLimit for details.
	PoolLimit int `yaml:"pool_limit"`
	// NoCache whether to disable cache
	NoCache bool `yaml:"no_cache"`

	init bool
}

var configs = make(map[string]*Config)

// ReadConfig gets a mongodb db config form yaml.
func ReadConfig(configSection string) (*Config, error) {
	conf, ok := configs[configSection]
	if ok {
		return conf, nil
	}
	conf = NewConfig()
	err := cfgo.Reg(configSection, conf)
	if err == nil {
		configs[configSection] = conf
		return conf, nil
	} else {
		return nil, err
	}
}

// NewConfig creates a default config.
func NewConfig() *Config {
	return &Config{
		Addrs:     []string{"127.0.0.1:27017"},
		Timeout:   10,
		PoolLimit: 256,
		Username:  "root",
		Password:  "",
		Database:  "test",
	}
}

// Reload sync automatically config from config file.
func (cfg *Config) Reload(bind cfgo.BindFunc) error {
	if cfg.init {
		return nil
	}
	err := bind()
	if err != nil {
		return err
	}
	cfg.init = true
	return nil
}

// Set config
func (mgoConfig *Config) Source() *mgo.DialInfo {
	dialInfo := &mgo.DialInfo{
		Addrs:     mgoConfig.Addrs,
		Username:  mgoConfig.Username,
		Password:  mgoConfig.Password,
		Database:  mgoConfig.Database,
		Timeout:   mgoConfig.Timeout,
		PoolLimit: mgoConfig.PoolLimit,
	}

	return dialInfo
}
