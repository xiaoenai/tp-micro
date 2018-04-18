package sdk

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	types "github.com/xiaoenai/ants/gateway/helper/gray/types"
	gwLogic "github.com/xiaoenai/ants/gateway/logic"
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
func IsGray(args *types.IsGrayArgs, setting ...socket.PacketSetting) (*types.IsGrayResult, *tp.Rerror) {
	reply := new(types.IsGrayResult)
	rerr := _formalClient.Pull("/gw/"+gwLogic.ApiVersion()+"/gray/is_gray", args, reply, setting...).Rerror()
	return reply, rerr
}

// Get get the rule of gray.
func Get(args *types.GetArgs, setting ...socket.PacketSetting) (*types.GrayMatch, *tp.Rerror) {
	reply := new(types.GrayMatch)
	rerr := _formalClient.Pull("/gw/"+gwLogic.ApiVersion()+"/gray/get", args, reply, setting...).Rerror()
	return reply, rerr
}

// Delete delete the rule of gray.
func Delete(args *types.DeleteArgs, setting ...socket.PacketSetting) (*struct{}, *tp.Rerror) {
	reply := new(struct{})
	rerr := _formalClient.Pull("/gw/"+gwLogic.ApiVersion()+"/gray/delete", args, reply, setting...).Rerror()
	return reply, rerr
}

// Set insert or update the regular expression for matching the URI.
func Set(args *types.SetArgs, setting ...socket.PacketSetting) (*struct{}, *tp.Rerror) {
	reply := new(struct{})
	rerr := _formalClient.Pull("/gw/"+gwLogic.ApiVersion()+"/gray/set", args, reply, setting...).Rerror()
	return reply, rerr
}
