// Temporary file used to ensure successful compilation!
// It will be overwritten when `micro gen` is run!
// When the project is completed, it should be removed!

package handler

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/examples/project/sdk"
	// "github.com/xiaoenai/tp-micro/examples/project/internal/model"
)

// Stat handler
func Stat(ctx tp.PushCtx, arg *sdk.StatArg) *tp.Rerror {
	return nil
}
