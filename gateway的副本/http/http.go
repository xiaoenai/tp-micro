package http

// import (
// 	"github.com/valyala/fasthttp"
// )

// OuterHttpSrvConfig config of HTTP server
type OuterHttpSrvConfig struct {
	ListenAddress string `yaml:"listen_address"`
	TlsCertFile   string `yaml:"tls_cert_file"`
	TlsKeyFile    string `yaml:"tls_key_file"`
}

// Serve starts HTTP gateway service.
func Serve(srvCfg OuterHttpSrvConfig) {

}
