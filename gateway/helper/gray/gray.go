package gray

import (
	"strings"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin/proxy"
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/gateway/client"
	"github.com/xiaoenai/tp-micro/gateway/helper/gray/api"
	"github.com/xiaoenai/tp-micro/gateway/helper/gray/logic"
	mod "github.com/xiaoenai/tp-micro/gateway/helper/gray/logic/model"
	types "github.com/xiaoenai/tp-micro/gateway/helper/gray/types"
	gwLogic "github.com/xiaoenai/tp-micro/gateway/logic"
	gwTypes "github.com/xiaoenai/tp-micro/gateway/types"
	"github.com/xiaoenai/tp-micro/model/etcd"
	"github.com/xiaoenai/tp-micro/model/mysql"
	"github.com/xiaoenai/tp-micro/model/redis"
)

// SetGray sets gray model to *gwTypes.Business.
// Note: the result grayClient may be used externally.
func SetGray(
	biz *gwTypes.Business,
	grayClientConfig micro.CliConfig,
	grayEtcdConfig etcd.EasyConfig,
	mysqlConfig mysql.Config,
	redisConfig redis.Config,
	protoFunc socket.ProtoFunc,
) (grayClient *micro.Client, err error) {
	err = mod.Init(mysqlConfig, redisConfig)
	if err != nil {
		return nil, err
	}
	if protoFunc == nil {
		protoFunc = socket.NewRawProtoFunc
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
	biz.ProxySelector = func(label *proxy.ProxyLabel) proxy.Caller {
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
