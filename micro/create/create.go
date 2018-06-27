package create

import (
	"io/ioutil"
	"os"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/micro/create/test"
	"github.com/xiaoenai/tp-micro/micro/create/tpl"
	"github.com/xiaoenai/tp-micro/micro/info"
)

// MicroTpl template file name
const MicroTpl = "__tp-micro__tpl__.go"

const microGenLock = "__tp-micro__gen__.lock"

// CreateProject creates a project.
func CreateProject() {
	tp.Infof("Generating project: %s", info.ProjPath())

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		tp.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	// creates base files
	if !goutil.FileExists(microGenLock) {
		tpl.Create()
	}

	// read temptale file
	b, err := ioutil.ReadFile(MicroTpl)
	if err != nil {
		b = test.MustAsset(MicroTpl)
	}

	// new project code
	proj := NewProject(b)
	proj.Generator()

	// write template file
	f, err := os.OpenFile(MicroTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		tp.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	tp.Infof("Completed code generation!")
}
