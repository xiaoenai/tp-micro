package micro

import tp "github.com/henrylee2cn/teleport"

// NOTE: error code range [-1,999]
var (
	// RerrClientClosed: client is closed.
	RerrClientClosed = tp.NewStatus(100, "client is closed", "")
	// RerrInvalidParameter: Invalid Parameter
	RerrInvalidParameter = tp.NewStatus(tp.CodeBadMessage, "Invalid Parameter", "")
	// RerrInternalServerError: System is busy, please try again later
	RerrInternalServerError = tp.NewStatus(tp.CodeInternalServerError, "System is busy, please try again later", "")
	// RerrNotFound: Not Found
	RerrNotFound = tp.NewStatus(tp.CodeNotFound, "Not Found", "")
	// RerrNotOnline: User is not online
	RerrNotOnline = tp.NewStatus(tp.CodeNotFound, "Not Found", "User is not online")
	// RerrRenderFailed: Template Rendering Failed
	RerrRenderFailed = tp.NewStatus(tp.CodeInternalServerError, "Template Rendering Failed", "")
)
