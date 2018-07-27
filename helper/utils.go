package helper

import (
	"net/http"

	tp "github.com/henrylee2cn/teleport"
)

// Redirect implements request redirection of HTTP gateway.
func Redirect(ctx tp.CallCtx, code int32, targetUrl string) *tp.Rerror {
	ctx.Output().Meta().Set("Location", targetUrl)
	return tp.NewRerror(code, http.StatusText(int(code)), "")
}
