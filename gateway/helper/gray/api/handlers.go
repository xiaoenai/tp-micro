package api

import (
	tp "github.com/henrylee2cn/teleport/v6"

	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/logic"
	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/types"
)

// IsGray check whether the service should use grayscale based on the uid.
func IsGray(ctx tp.CallCtx, args *types.IsGrayArgs) (*types.IsGrayResult, *tp.Status) {
	return logic.IsGray(args)
}

// Get get the rule of gray.
func Get(ctx tp.CallCtx, args *types.GetArgs) (*types.GrayMatch, *tp.Status) {
	return logic.Get(args)
}

// Delete delete the rule of gray.
func Delete(ctx tp.CallCtx, args *types.DeleteArgs) (*struct{}, *tp.Status) {
	return logic.Delete(args)
}

// Set insert or update the regular expression for matching the URI.
func Set(ctx tp.CallCtx, args *types.SetArgs) (*struct{}, *tp.Status) {
	return logic.Set(args)
}
