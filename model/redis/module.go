package redis

import (
	"fmt"
)

// Module defines a module for the redis key prefix.
type Module struct {
	keyFormat string
	module    string
	keyPrefix string
}

// NewModule creates a module for the redis key prefix.
func NewModule(module string) *Module {
	return &Module{
		keyFormat: "%s:%s", // module:key
		module:    module,
		keyPrefix: fmt.Sprintf("%s:", module),
	}
}

// Key completes the internal short key to the true key in redis.
func (m *Module) Key(shortKey string) string {
	return fmt.Sprintf(m.keyFormat, m.module, shortKey)
}

// Prefix returns the key prefix of current module.
func (m *Module) Prefix() string {
	return m.keyPrefix
}

// String returns module description.
func (m *Module) String() string {
	return fmt.Sprintf("module: %s", m.module)
}
