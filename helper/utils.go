package helper

import (
	"net/http"

	"github.com/henrylee2cn/erpc/v6"
)

// Redirect implements request redirection of HTTP gateway.
func Redirect(ctx erpc.CallCtx, code int32, targetUrl string) *erpc.Status {
	ctx.Output().Meta().Set("Location", targetUrl)
	return erpc.NewStatus(code, http.StatusText(int(code)), "")
}
