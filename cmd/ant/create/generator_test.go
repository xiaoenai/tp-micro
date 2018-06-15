package create

import (
	"strings"
	"testing"

	"github.com/xiaoenai/ants/cmd/ant/info"
)

func TestGenerator(t *testing.T) {
	info.Init("test")
	proj := NewProject([]byte(src))
	proj.Prepare()
	proj.genMainFile()
	proj.genTypeAndModelFiles()
	proj.genRouterFile()
	proj.genHandlerAndLogicAndSdkFiles()
	t.Logf("main.go:\n%s", codeFiles["main.go"])
	t.Logf("config.go:\n%s", codeFiles["config.go"])
	t.Logf("args/type.gen.go:\n%s", codeFiles["args/type.gen.go"])
	t.Logf("logic/tmp_code.gen.go:\n%s", codeFiles["logic/tmp_code.gen.go"])
	t.Logf("logic/model/init.go:\n%s", codeFiles["logic/model/init.go"])
	t.Logf("api/handler.gen.go:\n%s", codeFiles["api/handler.gen.go"])
	t.Logf("api/router.gen.go:\n%s", codeFiles["api/router.gen.go"])
	t.Logf("sdk/rpc.gen.go:\n%s", codeFiles["sdk/rpc.gen.go"])
	t.Logf("sdk/rpc_test.gen.go:\n%s", codeFiles["sdk/rpc.gen_test.go"])
	for k, v := range codeFiles {
		if strings.HasPrefix(k, "logic/model") {
			t.Logf("%s:\n%s", k, v)
		}
	}
}

const src = `// package __ANT__TPL__ is the project template
package __ANT__TPL__

// __API__PULL__ register PULL router:
//  /home
//  /math/divide
type __API__PULL__ interface {
	Home(*struct{}) *HomeResult
	Math
}

// __API__PUSH__ register PUSH router:
//  /stat
type __API__PUSH__ interface {
	Stat(*StatArg)
}

// MODEL create model
type __MODEL__ struct {
	DivideArg
	User
}

// Math controller
type Math interface {
	// Divide handler
	Divide(*DivideArg) *DivideResult
}

// HomeResult home result
type HomeResult struct {
	Content string // text
}

type (
	// DivideArg divide api arg
	DivideArg struct {
		// dividend
		A float64
		// divisor
		B float64 ` + "`param:\"<range: 0.01:100000>\"`" + `
	}
	// DivideResult divide api result
	DivideResult struct {
		// quotient
		C float64
	}
)

// StatArg stat handler arg
type StatArg struct {
	Ts int64 // timestamps
}

// User user info
type User struct {
	Id   int64
	Name string
	Age  int32
}
`
