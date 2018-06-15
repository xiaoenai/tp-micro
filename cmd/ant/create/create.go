package create

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/ants/cmd/ant/create/test"
	"github.com/xiaoenai/ants/cmd/ant/create/tpl"
	"github.com/xiaoenai/ants/cmd/ant/info"
)

const (
	defAntTpl = "__ant__tpl__.go"
)

// CreateProject creates a project.
func CreateProject(tplFile string) {
	tp.Infof("Generating project: %s", info.ProjPath())

	// read temptale file
	var noTplFile = len(tplFile) == 0
	if noTplFile {
		tplFile = defAntTpl
	}

	absTplFile, err := filepath.Abs(tplFile)
	if err != nil {
		tp.Fatalf("[ant] Invalid template file: %s", tplFile)
	}

	b, err := ioutil.ReadFile(absTplFile)
	if err != nil {
		if !noTplFile {
			tp.Fatalf("[ant] Write project files failed: %v", err)
		} else {
			b = test.MustAsset(defAntTpl)
		}
	}

	// creates project

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err = os.Chdir(info.AbsPath())
	if err != nil {
		tp.Fatalf("[ant] Jump working directory failed: %v", err)
	}

	// creates base files
	if !goutil.FileExists(".ant_gen_lock") {
		tpl.Create()
	}

	// new project code
	proj := NewProject(b)
	proj.Prepare()
	proj.Generator()

	// write template file
	f, err := os.OpenFile(defAntTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		tp.Fatalf("[ant] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	tp.Infof("Completed code generation!")
}
