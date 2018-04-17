package rerrs

import (
	tp "github.com/henrylee2cn/teleport"
)

var (
	// RerrInvalidParameter error
	RerrInvalidParameter = tp.NewRerror(400, "Invalid Parameter", "")
	// RerrServerError error
	RerrServerError = tp.NewRerror(500, "System is busy, please try again later", "")
	// RerrNotFound error
	RerrNotFound = tp.NewRerror(404, "Not Found", "")
)
