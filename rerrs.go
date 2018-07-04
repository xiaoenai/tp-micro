package micro

import tp "github.com/henrylee2cn/teleport"

// NOTE: error code range [-1,999]
var (
	// RerrClientClosed: client is closed.
	RerrClientClosed = tp.NewRerror(100, "client is closed", "")
	// RerrInvalidParameter: Invalid Parameter
	RerrInvalidParameter = tp.NewRerror(tp.CodeBadPacket, "Invalid Parameter", "")
	// RerrInternalServerError: System is busy, please try again later
	RerrInternalServerError = tp.NewRerror(tp.CodeInternalServerError, "System is busy, please try again later", "")
	// RerrNotFound: Not Found
	RerrNotFound = tp.NewRerror(tp.CodeNotFound, "Not Found", "")
	// RerrNotOnline: User is not online
	RerrNotOnline = tp.NewRerror(tp.CodeNotFound, "Not Found", "User is not online")
	// RerrRenderFailed: Template Rendering Failed
	RerrRenderFailed = tp.NewRerror(tp.CodeInternalServerError, "Template Rendering Failed", "")
)
