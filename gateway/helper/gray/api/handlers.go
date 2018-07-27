package api

import (
	tp "github.com/henrylee2cn/teleport"

	"github.com/xiaoenai/tp-micro/gateway/helper/gray/logic"
	"github.com/xiaoenai/tp-micro/gateway/helper/gray/types"
)

// IsGray check whether the service should use grayscale based on the uid.
func IsGray(ctx tp.CallCtx, args *types.IsGrayArgs) (*types.IsGrayResult, *tp.Rerror) {
	return logic.IsGray(args)
}

// Get get the rule of gray.
func Get(ctx tp.CallCtx, args *types.GetArgs) (*types.GrayMatch, *tp.Rerror) {
	return logic.Get(args)
}

// Delete delete the rule of gray.
func Delete(ctx tp.CallCtx, args *types.DeleteArgs) (*struct{}, *tp.Rerror) {
	return logic.Delete(args)
}

// Set insert or update the regular expression for matching the URI.
func Set(ctx tp.CallCtx, args *types.SetArgs) (*struct{}, *tp.Rerror) {
	return logic.Set(args)
}
