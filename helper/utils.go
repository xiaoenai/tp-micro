package helper

import (
	"net/http"

	tp "github.com/henrylee2cn/teleport/v6"
)

// Redirect implements request redirection of HTTP gateway.
func Redirect(ctx tp.CallCtx, code int32, targetUrl string) *tp.Status {
	ctx.Output().Meta().Set("Location", targetUrl)
	return tp.NewStatus(code, http.StatusText(int(code)), "")
}
