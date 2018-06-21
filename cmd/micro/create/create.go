package create

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/cmd/micro/create/test"
	"github.com/xiaoenai/tp-micro/cmd/micro/create/tpl"
	"github.com/xiaoenai/tp-micro/cmd/micro/info"
)

const (
	defMicroTpl     = "__tp-micro__tpl__.go"
	defMicroGenLock = "__tp-micro__gen__.lock"
)

// CreateProject creates a project.
func CreateProject(tplFile string) {
	tp.Infof("Generating project: %s", info.ProjPath())

	// read temptale file
	var noTplFile = len(tplFile) == 0
	if noTplFile {
		tplFile = defMicroTpl
	}

	absTplFile, err := filepath.Abs(tplFile)
	if err != nil {
		tp.Fatalf("[micro] Invalid template file: %s", tplFile)
	}

	b, err := ioutil.ReadFile(absTplFile)
	if err != nil {
		if !noTplFile {
			tp.Fatalf("[micro] Write project files failed: %v", err)
		} else {
			b = test.MustAsset(defMicroTpl)
		}
	}

	// creates project

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err = os.Chdir(info.AbsPath())
	if err != nil {
		tp.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	// creates base files
	if !goutil.FileExists(defMicroGenLock) {
		tpl.Create()
	}

	// new project code
	proj := NewProject(b)
	proj.Prepare()
	proj.Generator()

	// write template file
	f, err := os.OpenFile(defMicroTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		tp.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	tp.Infof("Completed code generation!")
}
