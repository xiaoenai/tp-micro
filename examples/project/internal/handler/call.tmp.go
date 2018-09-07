package handler

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/examples/project/sdk"
	// "${import_prefix}/internal/model"
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
