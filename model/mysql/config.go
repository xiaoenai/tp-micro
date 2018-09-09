package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/henrylee2cn/cfgo"
)

// Config db config
type Config struct {
	Database string
	Username string
	Password string
	Host     string
	Port     int
	// the maximum number of connections in the idle connection pool.
	//
	// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns
	// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit
	//
	// If n <= 0, no idle connections are retained.
	MaxIdleConns int `yaml:"max_idle_conns"`
	// the maximum number of open connections to the database.
	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
	// MaxIdleConns, then MaxIdleConns will be reduced to match the new
	// MaxOpenConns limit
	//
	// If n <= 0, then there is no limit on the number of open connections.
	// The default is 0 (unlimited).
	MaxOpenConns int `yaml:"max_open_conns"`
	// maximum amount of second a connection may be reused.
	// If d <= 0, connections are reused forever.
	ConnMaxLifetime int64 `yaml:"conn_max_lifetime"`

	// NoCache whether to disable cache
	NoCache bool `yaml:"no_cache"`

	init bool
}

var configs = make(map[string]*Config)

// ReadConfig gets a mysql db config form yaml.
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
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Database: "test",
	}
}

// Source returns the mysql connection string.
func (cfg *Config) Source() string {
	pwd := cfg.Password
	if pwd != "" {
		pwd = ":" + pwd
	}
	port := cfg.Port
	if port == 0 {
		port = 3306
	}
	return fmt.Sprintf("%s%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&interpolateParams=true", cfg.Username, pwd, cfg.Host, port, cfg.Database)
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
	if len(cfg.Host) == 0 {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port <= 0 {
		cfg.Port = 3306
	}
	if len(cfg.Username) == 0 {
		cfg.Username = "root"
	}
	if len(cfg.Database) == 0 {
		cfg.Database = "test"
	}
	cfg.init = true
	return nil
}
