package errs

import (
	"github.com/henrylee2cn/erpc/v6"
)

var (
	// InvalidParameter status
	InvalidParameter = erpc.NewStatus(100001, "Invalid Parameter", "Contains invalid request parameters")
)
