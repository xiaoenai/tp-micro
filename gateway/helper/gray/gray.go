package gray

import (
	"strings"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	"github.com/xiaoenai/ants/gateway/client"
	"github.com/xiaoenai/ants/gateway/helper/gray/api"
	"github.com/xiaoenai/ants/gateway/helper/gray/logic"
	mod "github.com/xiaoenai/ants/gateway/helper/gray/logic/model"
	types "github.com/xiaoenai/ants/gateway/helper/gray/types"
	gwLogic "github.com/xiaoenai/ants/gateway/logic"
	gwTypes "github.com/xiaoenai/ants/gateway/types"
	"github.com/xiaoenai/ants/model"
	"github.com/xiaoenai/ants/model/redis"
)

// SetGray sets gray model to *gwTypes.Business.
// Note: the result grayClient may be used externally.
func SetGray(
	biz *gwTypes.Business,
	grayClientConfig micro.CliConfig,
	grayEtcdConfig etcd.EasyConfig,
	mysqlConfig model.Config,
	redisConfig redis.Config,
	protoFunc socket.ProtoFunc,
) (grayClient *micro.Client, err error) {
	err = mod.Init(mysqlConfig, redisConfig)
	if err != nil {
		return nil, err
	}
	if protoFunc == nil {
		protoFunc = socket.NewFastProtoFunc
	}
	grayEtcdClient, err := etcd.EasyNew(grayEtcdConfig)
	if err != nil {
		return nil, err
	}
	grayClient = micro.NewClient(
		grayClientConfig,
		discovery.NewLinkerFromEtcd(grayEtcdClient),
	)
	grayClient.SetProtoFunc(protoFunc)

	biz.InnerServerPlugins = append(biz.InnerServerPlugins, new(innerServerPlugin))
	biz.ProxySelector = func(label *plugin.ProxyLabel) plugin.Caller {
		idx := strings.Index(label.Uri, "?")
		var uri string
		if idx != -1 {
			uri = label.Uri[:idx]
		} else {
			uri = label.Uri
		}
		r, rerr := logic.IsGray(&types.IsGrayArgs{
			Uri: uri,
			Uid: label.SessionId,
		})
		if rerr != nil {
			tp.Errorf("%s", rerr.String())
			return client.DynamicClient()
		}
		if !r.Gray {
			return client.DynamicClient()
		}
		return grayClient
	}
	return grayClient, nil
}

type innerServerPlugin struct{}

func (*innerServerPlugin) Name() string {
	return "route_gray"
}

var _ tp.PostNewPeerPlugin = (*innerServerPlugin)(nil)

func (*innerServerPlugin) PostNewPeer(peer tp.EarlyPeer) error {
	api.Route("/gw/"+gwLogic.ApiVersion()+"/gray", peer.Router())
	return nil
}
