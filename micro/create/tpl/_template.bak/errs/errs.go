package errs

import (
	tp "github.com/henrylee2cn/teleport/v6"
)

var (
	// InvalidParameter status
	InvalidParameter = tp.NewStatus(100001, "Invalid Parameter", "Contains invalid request parameters")
)
