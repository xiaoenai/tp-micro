package create

import (
	"encoding/json"
	"testing"
)

var tInfo = newTplInfo([]byte(`
package create
import (
	"testing"
)
import f "fmt"

type __API__PULL__ interface {
	// Stat doc
	Stat(X) *struct{}// Stat comment
	// Math doc
	Math // Math comment
}

type Math interface {
	// Divide handler
	Divide(*DivideArg) *DivideResult
	Sub
}

type Sub interface {
	// Divide handler
	Divide(*DivideArg) *DivideResult
}

// the comment can not be caught!
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

type __API__PUSH__ interface {
	Stat(*StatArg)
}

// StatArg stat handler arg
type StatArg struct {
	Ts int64 // timestamps
}

type __MYSQL__MODEL__ struct {
	DivideResult
	StatArg
}

type __MONGO__MODEL__ struct {
	A
}

// A comment ...
type A struct{
	// X doc ...
	X string // X comment ...
	// Y doc ...
	Y int // Y comment ...
}

// X A alias
type X = A

type Y X  
`))

func TestParse(t *testing.T) {
	tInfo.Parse()
	t.Logf("TypeImportString: %s", tInfo.TypeImportString())
	t.Logf("TypesString:\n%v", tInfo.TypesString())
	t.Logf("pull apis:\n%v", tInfo.PullHandlerString(nil))
	t.Logf("push apis:\n%v", tInfo.PushHandlerString(nil))
	t.Logf("router:\n%v", tInfo.RouterString("_group"))
	for _, m := range tInfo.models.mysql {
		t.Logf("mysql:\n%v", m)
	}
	for _, m := range tInfo.models.mongo {
		t.Logf("mongo:\n%v", m)
	}
}

func jsonString(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
