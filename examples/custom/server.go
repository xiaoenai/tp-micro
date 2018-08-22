package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/gateway/helper/agent"
	"github.com/xiaoenai/tp-micro/helper"
	html "github.com/xiaoenai/tp-micro/helper/mod-html"
	"github.com/xiaoenai/tp-micro/model/etcd"
	"github.com/xiaoenai/tp-micro/model/redis"
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
func Home(ctx tp.CallCtx, args *struct{}) ([]byte, *tp.Rerror) {
	return html.Render(ctx, "home", "Home Page Test!")
}

// Home2 HTML home page
func Home2(ctx tp.CallCtx, args *struct{}) ([]byte, *tp.Rerror) {
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
func (m *Math) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

type config struct {
	Srv   micro.SrvConfig
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
		Etcd: etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
		Redis: *redis.NewConfig(),
	}
	cfgo.AllowAppsShare(true)
	cfgo.MustReg("server", &cfg)

	redisClient, err := redis.NewClient(&cfg.Redis)
	agent.Init(redisClient, redisClient)
	if err != nil {
		tp.Fatalf("%v", err)
	}
	tp.Go(func() {
		agentNewsChan := agent.Subscribe()
		for news := range agentNewsChan {
			tp.Infof("agent news: sessionId:%s, event:%s",
				news.SessionId, news.Event,
			)
		}
	})
	srv := micro.NewServer(cfg.Srv, discovery.ServicePlugin(
		cfg.Srv.InnerIpPort(),
		cfg.Etcd,
	))
	srv.RouteCallFunc(Home)
	srv.RouteCallFunc(Home2)
	srv.RouteCall(new(Math))
	srv.ListenAndServe()
}
