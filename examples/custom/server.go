package main

import (
	"encoding/json"
	"time"

	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport/v6"
	"github.com/henrylee2cn/teleport/v6/codec"
	micro "github.com/xiaoenai/tp-micro/v6"
	"github.com/xiaoenai/tp-micro/v6/discovery"
	sagent "github.com/xiaoenai/tp-micro/v6/gateway/helper/agent/socket"
	gwSdk "github.com/xiaoenai/tp-micro/v6/gateway/sdk"
	gwTypes "github.com/xiaoenai/tp-micro/v6/gateway/types"
	"github.com/xiaoenai/tp-micro/v6/helper"
	html "github.com/xiaoenai/tp-micro/v6/helper/mod-html"
	"github.com/xiaoenai/tp-micro/v6/model/etcd"
	"github.com/xiaoenai/tp-micro/v6/model/redis"
)

func init() {
	// html.ParseGlob("*.tpl")

	html.Parse("home", `<!DOCTYPE html>
<html>
	<head>
	    <title>Home</title>
	</head>
	<body>
		<br/>
		<h2><center>{{.}}</center></h2>
	</body>
</html>`)
}

// Home HTML home page
func Home(ctx tp.CallCtx, args *struct{}) ([]byte, *tp.Status) {
	return html.Render(ctx, "home", "Home Page Test!")
}

// Home2 HTML home page
func Home2(ctx tp.CallCtx, args *struct{}) ([]byte, *tp.Status) {
	return nil, helper.Redirect(ctx, 302, "http://localhost:5000/home")
}

// Args args
type Args struct {
	A int
	B int `param:"<range:1:>"`
}

// Math handler
type Math struct {
	tp.CallCtx
}

// Divide divide API
func (m *Math) Divide(args *Args) (int, *tp.Status) {
	return args.A / args.B, nil
}

type config struct {
	Srv   micro.SrvConfig
	Cli   micro.CliConfig
	Etcd  etcd.EasyConfig
	Redis redis.Config
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	err := bind()
	return err
}

func main() {
	cfg := config{
		Srv: micro.SrvConfig{
			ListenAddress:   ":0",
			EnableHeartbeat: true,
			PrintDetail:     true,
		},
		Cli: micro.CliConfig{},
		Etcd: etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
		Redis: *redis.NewConfig(),
	}
	cfgo.AllowAppsShare(true)
	cfgo.MustReg("server", &cfg)

	redisClient, err := redis.NewClient(&cfg.Redis)
	sagent.Init(redisClient, redisClient)
	if err != nil {
		tp.Fatalf("%v", err)
	}

	tp.Go(func() {
		agentNewsChan := sagent.Subscribe()
		for news := range agentNewsChan {
			tp.Infof("agent news: sessionId:%s, event:%s",
				news.SessionId, news.Event,
			)
		}
	})

	p := discovery.ServicePlugin(
		cfg.Srv.InnerIpPort(),
		cfg.Etcd,
	)
	srv := micro.NewServer(cfg.Srv, p)
	srv.RouteCallFunc(Home)
	srv.RouteCallFunc(Home2)
	srv.RouteCall(new(Math))
	go srv.ListenAndServe()

	// test pushing, when the client progress is existed.
	gwSdk.Init("v1", nil)
	for i := 0; ; i++ {
		tp.Infof("test pushing msg after 10s")
		time.Sleep(time.Second * 10)
		push([]string{"client-auth-info-12345"}, &Args{A: i, B: i})
	}
	select {}
}

func push(uids []string, args *Args) {
	agts, stat := sagent.QueryAgent(uids)
	if stat != nil {
		tp.Errorf("push fail: %v", stat)
	}
	for _, agt := range agts.Agents {
		if agt.IsOffline {
			continue
		}
		addr := agt.GetInnerGw()
		target := []*gwTypes.MpushTarget{
			{
				SessionId:       agt.GetSessionId(),
				AdditionalQuery: "testpush=ok",
			},
		}
		targetUri := "/push"
		msgBytes, _ := json.Marshal(args)
		_, stat := gwSdk.SocketMpush(
			addr,
			&gwTypes.SocketMpushArgs{
				Target:    target,
				Uri:       targetUri,
				Body:      msgBytes,
				BodyCodec: codec.ID_JSON,
			},
			tp.WithBodyCodec(codec.ID_JSON),
		)
		if stat != nil {
			tp.Errorf("push fail: %v", stat)
		} else {
			tp.Infof("push ok")
		}
	}
}
