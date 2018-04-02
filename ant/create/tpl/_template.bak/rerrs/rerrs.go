package rerrs

import (
	tp "github.com/henrylee2cn/teleport"
)

var (
	// RerrInvalidParameter error
	RerrInvalidParameter = tp.NewRerror(100001, "Invalid Parameter", "Contains invalid request parameters")
)
