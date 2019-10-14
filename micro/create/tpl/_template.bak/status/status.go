package status

import (
	tp "github.com/henrylee2cn/teleport"
)

var (
	// StatusInvalidParameter status
	StatusInvalidParameter = tp.NewStatus(100001, "Invalid Parameter", "Contains invalid request parameters")
)
