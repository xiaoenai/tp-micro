// Temporary file used to ensure successful compilation!
// It will be overwritten when `micro gen` is run!
// When the project is completed, it should be removed!

package handler

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/examples/project/sdk"
	// "github.com/xiaoenai/tp-micro/examples/project/internal/model"
)

// Home handler
func Home(ctx tp.CallCtx, arg *sdk.EmptyStruct) (*sdk.HomeResult, *tp.Rerror) {
	return new(sdk.HomeResult), nil
}

// Math controller
type Math struct {
	tp.CallCtx
}

// Divide handler
func (m *Math) Divide(arg *sdk.DivideArg) (*sdk.DivideResult, *tp.Rerror) {
	return new(sdk.DivideResult), nil
}
