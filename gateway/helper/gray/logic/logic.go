package logic

import (
	"regexp"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"

	"github.com/xiaoenai/tp-micro/gateway/helper/gray/logic/model"
	"github.com/xiaoenai/tp-micro/gateway/helper/gray/types"
)

var (
	regexpCache   = goutil.AtomicMap()
	notGrayResult = new(types.IsGrayResult)
	isGrayResult  = &types.IsGrayResult{
		Gray: true,
	}
)

// IsGray check whether the service should use grayscale based on the uid.
func IsGray(args *types.IsGrayArgs) (*types.IsGrayResult, *tp.Status) {
	g, exist, err := model.GetGrayMatchByUri(args.Uri)
	if err != nil {
		return nil, micro.RerrInternalServerError.Copy().SetReason(err.Error())
	}
	if !exist {
		return notGrayResult, nil
	}
	var re *regexp.Regexp
	reIface, ok := regexpCache.Load(g.Regexp)
	if !ok {
		re, err = regexp.Compile(g.Regexp)
		if err != nil {
			return nil, micro.RerrInternalServerError.Copy().SetReason(err.Error())
		}
		regexpCache.Store(g.Regexp, re)
	} else {
		re = reIface.(*regexp.Regexp)
	}
	if re.MatchString(args.Uid) {
		return isGrayResult, nil
	}
	return notGrayResult, nil
}

// Get get the rule of gray.
func Get(args *types.GetArgs) (*types.GrayMatch, *tp.Status) {
	g, exist, err := model.GetGrayMatchByUri(args.Uri)
	if err != nil {
		return nil, micro.RerrInternalServerError.Copy().SetReason(err.Error())
	}
	if !exist {
		return nil, micro.RerrNotFound
	}
	return g, nil
}

// Delete delete the rule of gray.
func Delete(args *types.DeleteArgs) (*struct{}, *tp.Status) {
	err := model.DeleteGrayMatchByUri(args.Uri)
	if err != nil {
		return nil, micro.RerrInternalServerError.Copy().SetReason(err.Error())
	}
	return new(struct{}), nil
}

// Set insert or update the regular expression for matching the URI.
func Set(args *types.SetArgs) (*struct{}, *tp.Status) {
	_, ok := regexpCache.Load(args.Regexp)
	if !ok {
		re, err := regexp.Compile(args.Regexp)
		if err != nil {
			return nil, micro.RerrInvalidParameter.Copy().SetReason(err.Error())
		}
		regexpCache.Store(args.Regexp, re)
	}
	// put
	err := model.UpsertGrayMatch(&model.GrayMatch{
		Uri:    args.Uri,
		Regexp: args.Regexp,
	})
	if err != nil {
		return nil, micro.RerrInternalServerError.Copy().SetReason(err.Error())
	}
	return new(struct{}), nil
}
