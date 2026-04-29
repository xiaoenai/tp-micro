module github.com/xiaoenai/tp-micro/v2

go 1.23.0

replace (
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de
	google.golang.org/genproto/googleapis/api => google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de
	google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de
	google.golang.org/grpc => google.golang.org/grpc v1.67.1
	google.golang.org/protobuf => google.golang.org/protobuf v1.35.2
)

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.9.3
	github.com/gogo/protobuf v1.3.2
	github.com/howeyc/fsnotify v0.9.0
	github.com/lib/pq v1.12.3
	github.com/mattn/go-sqlite3 v1.14.42
	github.com/swxctx/teleport v1.0.0
	github.com/urfave/cli v1.22.17
	github.com/valyala/fasthttp v1.59.0
	go.etcd.io/etcd/client/v3 v3.5.17
	golang.org/x/crypto v0.36.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.34.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.17 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.17 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
