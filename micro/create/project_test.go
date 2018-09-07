package create

import (
	"strings"
	"testing"

	"github.com/xiaoenai/tp-micro/micro/info"
)

func TestGenerator(t *testing.T) {
	info.Init("test")
	proj := NewProject([]byte(__tpl__))
	proj.gen()
	t.Logf("main.go:\n%s", proj.codeFiles["main.go"])
	t.Logf("config.go:\n%s", proj.codeFiles["config.go"])
	t.Logf("args/const.gen.go:\n%s", proj.codeFiles["args/const.gen.go"])
	t.Logf("args/type.gen.go:\n%s", proj.codeFiles["args/type.gen.go"])
	t.Logf("api/router.gen.go:\n%s", proj.codeFiles["api/router.gen.go"])
	t.Logf("sdk/rpc.gen.go:\n%s", proj.codeFiles["sdk/rpc.gen.go"])
	t.Logf("sdk/rpc_test.gen.go:\n%s", proj.codeFiles["sdk/rpc.gen_test.go"])
	for k, v := range proj.codeFiles {
		if strings.HasPrefix(k, "internal/") {
			t.Logf("%s:\n%s", k, v)
		}
	}
	t.Logf("README.md:\n%s", proj.genReadme())
}
