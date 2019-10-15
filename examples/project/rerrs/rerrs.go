package rerrs

import (
	tp "github.com/henrylee2cn/teleport"
)

var (
	// RerrInvalidParameter status
	RerrInvalidParameter = tp.NewStatus(100001, "Invalid Parameter", "Contains invalid request parameters")
)
