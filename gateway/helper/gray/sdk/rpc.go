package sdk

import (
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	types "github.com/xiaoenai/tp-micro/gateway/helper/gray/types"
	gwLogic "github.com/xiaoenai/tp-micro/gateway/logic"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

var _formalClient *micro.Client

// Init initializes formal client with configs.
func Init(formalClientConfig micro.CliConfig, formalEtcdConfing etcd.EasyConfig) {
	_formalClient = micro.NewClient(
		formalClientConfig,
		discovery.NewLinker(formalEtcdConfing),
	)
}

// InitWithClient initializes formal client with specified object.
func InitWithClient(formalClient *micro.Client) {
	_formalClient = formalClient
}

// IsGray check whether the service should use grayscale based on the uid.
func IsGray(args *types.IsGrayArgs, setting ...tp.MessageSetting) (*types.IsGrayResult, *tp.Status) {
	reply := new(types.IsGrayResult)
	rerr := _formalClient.Call("/gw/"+gwLogic.ApiVersion()+"/gray/is_gray", args, reply, setting...).Rerror()
	return reply, rerr
}

// Get get the rule of gray.
func Get(args *types.GetArgs, setting ...tp.MessageSetting) (*types.GrayMatch, *tp.Status) {
	reply := new(types.GrayMatch)
	rerr := _formalClient.Call("/gw/"+gwLogic.ApiVersion()+"/gray/get", args, reply, setting...).Rerror()
	return reply, rerr
}

// Delete delete the rule of gray.
func Delete(args *types.DeleteArgs, setting ...tp.MessageSetting) (*struct{}, *tp.Status) {
	reply := new(struct{})
	rerr := _formalClient.Call("/gw/"+gwLogic.ApiVersion()+"/gray/delete", args, reply, setting...).Rerror()
	return reply, rerr
}

// Set insert or update the regular expression for matching the URI.
func Set(args *types.SetArgs, setting ...tp.MessageSetting) (*struct{}, *tp.Status) {
	reply := new(struct{})
	rerr := _formalClient.Call("/gw/"+gwLogic.ApiVersion()+"/gray/set", args, reply, setting...).Rerror()
	return reply, rerr
}
