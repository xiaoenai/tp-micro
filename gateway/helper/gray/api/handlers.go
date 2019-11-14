package api

import (
	"github.com/henrylee2cn/erpc/v6"

	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/logic"
	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/types"
)

// IsGray check whether the service should use grayscale based on the uid.
func IsGray(ctx erpc.CallCtx, args *types.IsGrayArgs) (*types.IsGrayResult, *erpc.Status) {
	return logic.IsGray(args)
}

// Get get the rule of gray.
func Get(ctx erpc.CallCtx, args *types.GetArgs) (*types.GrayMatch, *erpc.Status) {
	return logic.Get(args)
}

// Delete delete the rule of gray.
func Delete(ctx erpc.CallCtx, args *types.DeleteArgs) (*struct{}, *erpc.Status) {
	return logic.Delete(args)
}

// Set insert or update the regular expression for matching the URI.
func Set(ctx erpc.CallCtx, args *types.SetArgs) (*struct{}, *erpc.Status) {
	return logic.Set(args)
}
