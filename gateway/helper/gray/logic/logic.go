package logic

import (
	"regexp"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"

	"github.com/xiaoenai/ants/gateway/helper/gray/logic/model"
	"github.com/xiaoenai/ants/gateway/helper/gray/rerrs"
	"github.com/xiaoenai/ants/gateway/helper/gray/types"
)

var (
	regexpCache   = goutil.AtomicMap()
	notGrayResult = new(types.IsGrayResult)
	isGrayResult  = &types.IsGrayResult{
		Gray: true,
	}
)

// IsGray check whether the service should use grayscale based on the uid.
func IsGray(args *types.IsGrayArgs) (*types.IsGrayResult, *tp.Rerror) {
	g, exist, err := model.GetGrayMatchByUri(args.Uri)
	if err != nil {
		return nil, rerrs.RerrServerError.Copy().SetDetail(err.Error())
	}
	if !exist {
		return notGrayResult, nil
	}
	var re *regexp.Regexp
	reIface, ok := regexpCache.Load(g.Regexp)
	if !ok {
		re, err = regexp.Compile(g.Regexp)
		if err != nil {
			return nil, rerrs.RerrServerError.Copy().SetDetail(err.Error())
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
func Get(args *types.GetArgs) (*types.GrayMatch, *tp.Rerror) {
	g, exist, err := model.GetGrayMatchByUri(args.Uri)
	if err != nil {
		return nil, rerrs.RerrServerError.Copy().SetDetail(err.Error())
	}
	if !exist {
		return nil, rerrs.RerrNotFound
	}
	return g, nil
}

// Delete delete the rule of gray.
func Delete(args *types.DeleteArgs) (*struct{}, *tp.Rerror) {
	err := model.DeleteGrayMatchByUri(args.Uri)
	if err != nil {
		return nil, rerrs.RerrServerError.Copy().SetDetail(err.Error())
	}
	return new(struct{}), nil
}

// Set insert or update the regular expression for matching the URI.
func Set(args *types.SetArgs) (*struct{}, *tp.Rerror) {
	_, ok := regexpCache.Load(args.Regexp)
	if !ok {
		re, err := regexp.Compile(args.Regexp)
		if err != nil {
			return nil, rerrs.RerrInvalidParameter.Copy().SetDetail(err.Error())
		}
		regexpCache.Store(args.Regexp, re)
	}
	// put
	err := model.UpsertGrayMatch(&model.GrayMatch{
		Uri:    args.Uri,
		Regexp: args.Regexp,
	})
	if err != nil {
		return nil, rerrs.RerrServerError.Copy().SetDetail(err.Error())
	}
	return new(struct{}), nil
}
