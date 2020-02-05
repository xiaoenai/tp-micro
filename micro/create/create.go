package create

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/goutil"
	"github.com/xiaoenai/tp-micro/v6/micro/create/tpl"
	"github.com/xiaoenai/tp-micro/v6/micro/info"
)

// MicroTpl template file name
const MicroTpl = "__tp-micro__tpl__.go"

// MicroGenLock the file is used to markup generated project
const MicroGenLock = "__tp-micro__gen__.lock"

// CreateProject creates a project.
func CreateProject(force, newdoc bool) {
	erpc.Infof("Generating project: %s", info.ProjPath())

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		erpc.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	force = force || !goutil.FileExists(MicroGenLock)

	// creates base files
	if force {
		tpl.Create()
	}

	// read temptale file
	b, err := ioutil.ReadFile(MicroTpl)
	if err != nil {
		b = []byte(strings.Replace(__tpl__, "__PROJ_NAME__", info.ProjName(), -1))
	}

	// new project code
	proj := NewProject(b)
	proj.Generator(force, force || newdoc)

	// write template file
	f, err := os.OpenFile(MicroTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		erpc.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	tpl.RestoreAsset("./", MicroGenLock)

	erpc.Infof("Completed code generation!")
}

// CreateDoc creates a project doc.
func CreateDoc() {
	erpc.Infof("Generating README.md: %s", info.ProjPath())

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		erpc.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	// read temptale file
	b, err := ioutil.ReadFile(MicroTpl)
	if err != nil {
		b = []byte(strings.Replace(__tpl__, "__PROJ_NAME__", info.ProjName(), -1))
	}

	// new project code
	proj := NewProject(b)
	proj.gen()
	proj.genAndWriteReadmeFile()

	// write template file
	f, err := os.OpenFile(MicroTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		erpc.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	erpc.Infof("Completed README.md generation!")
}
