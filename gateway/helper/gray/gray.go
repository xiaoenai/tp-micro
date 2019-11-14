package gray

import (
	"strings"

	"github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/erpc/v6/plugin/proxy"
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/clientele"
	"github.com/xiaoenai/tp-micro/v6/discovery"
	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/api"
	"github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/logic"
	mod "github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/logic/model"
	types "github.com/xiaoenai/tp-micro/v6/gateway/helper/gray/types"
	gwLogic "github.com/xiaoenai/tp-micro/v6/gateway/logic"
	gwTypes "github.com/xiaoenai/tp-micro/v6/gateway/types"
	"github.com/xiaoenai/tp-micro/v6/model/etcd"
	"github.com/xiaoenai/tp-micro/v6/model/mysql"
	"github.com/xiaoenai/tp-micro/v6/model/redis"
)

// SetGray sets gray model to *gwTypes.Business.
// Note: the result grayClient may be used externally.
func SetGray(
	biz *gwTypes.Business,
	grayClientConfig micro.CliConfig,
	grayEtcdConfig etcd.EasyConfig,
	mysqlConfig mysql.Config,
	redisConfig redis.Config,
	protoFunc erpc.ProtoFunc,
) (grayClient *micro.Client, err error) {
	err = mod.Init(mysqlConfig, redisConfig)
	if err != nil {
		return nil, err
	}
	// if protoFunc == nil {
	// 	protoFunc = socket.NewRawProtoFunc
	// }
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
	biz.ProxySelector = func(label *proxy.Label) proxy.Forwarder {
		idx := strings.Index(label.ServiceMethod, "?")
		var uri string
		if idx != -1 {
			uri = label.ServiceMethod[:idx]
		} else {
			uri = label.ServiceMethod
		}
		r, stat := logic.IsGray(&types.IsGrayArgs{
			Uri: uri,
			Uid: label.SessionID,
		})
		if stat != nil {
			erpc.Errorf("%s", stat.String())
			return clientele.GetDynamicClient()
		}
		if !r.Gray {
			return clientele.GetDynamicClient()
		}
		return grayClient
	}
	return grayClient, nil
}

type innerServerPlugin struct{}

func (*innerServerPlugin) Name() string {
	return "route_gray"
}

var _ erpc.PostNewPeerPlugin = (*innerServerPlugin)(nil)

func (*innerServerPlugin) PostNewPeer(peer erpc.EarlyPeer) error {
	api.Route("/gw/"+gwLogic.ApiVersion()+"/gray", peer.Router())
	return nil
}
