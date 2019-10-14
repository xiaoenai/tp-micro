package status

import (
	tp "github.com/henrylee2cn/teleport"
)

var (
	// RerrInvalidParameter error
	RerrInvalidParameter = tp.NewStatus(100001, "Invalid Parameter", "Contains invalid request parameters")
)
