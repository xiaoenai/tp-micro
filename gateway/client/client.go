package client

import (
	"github.com/henrylee2cn/ant"
)

var cli *ant.Client

// Init initializes a common inner ant client.
func Init(cliCfg ant.CliConfig) {

}

// Client returns the common inner ant client.
func Client() *ant.Client {
	return cli
}
