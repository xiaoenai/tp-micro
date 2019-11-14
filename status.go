package micro

import "github.com/henrylee2cn/erpc/v6"

// NOTE: error code range [-1,999]
var (
	// RerrClientClosed: client is closed.
	RerrClientClosed = erpc.NewStatus(100, "client is closed", "")
	// RerrInvalidParameter: Invalid Parameter
	RerrInvalidParameter = erpc.NewStatus(erpc.CodeBadMessage, "Invalid Parameter", "")
	// RerrInternalServerError: System is busy, please try again later
	RerrInternalServerError = erpc.NewStatus(erpc.CodeInternalServerError, "System is busy, please try again later", "")
	// RerrNotFound: Not Found
	RerrNotFound = erpc.NewStatus(erpc.CodeNotFound, "Not Found", "")
	// RerrNotOnline: User is not online
	RerrNotOnline = erpc.NewStatus(erpc.CodeNotFound, "Not Found", "User is not online")
	// RerrRenderFailed: Template Rendering Failed
	RerrRenderFailed = erpc.NewStatus(erpc.CodeInternalServerError, "Template Rendering Failed", "")
)
