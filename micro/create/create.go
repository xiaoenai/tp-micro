package create

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/micro/info"
)

// MicroTpl template file name
const MicroTpl = "__tp-micro__tpl__.go"

// MicroGenLock the file is used to markup generated project
const MicroGenLock = "__tp-micro__gen__.lock"

// CreateProject creates a project.
func CreateProject(force, newdoc bool) {
	tp.Infof("Generating project: %s", info.ProjPath())

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		tp.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	force = force || !goutil.FileExists(MicroGenLock)
	if force {
		mustMkdirAll("internal/handler")
		mustMkdirAll("internal/model")
		mustMkdirAll("sdk")
		ioutil.WriteFile(
			".gitignore",
			[]byte(strings.Replace(__gitignore__, "${PROJ_NAME}", info.ProjName(), -1)),
			os.FileMode(0665),
		)
		ioutil.WriteFile("sdk/rerr.go", []byte(__rerr__), os.FileMode(0665))
	}

	// read temptale file
	b, err := ioutil.ReadFile(MicroTpl)
	if err != nil {
		b = []byte(strings.Replace(__tpl__, "${PROJ_NAME}", info.ProjName(), -1))
	}

	// new project code
	proj := NewProject(b)
	proj.Generator(force, force || newdoc)

	// write template file
	f, err := os.OpenFile(MicroTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		tp.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	ioutil.WriteFile(MicroGenLock, []byte(__genlock__), os.FileMode(0665))

	tp.Infof("Completed code generation!")
}

// CreateDoc creates a project doc.
func CreateDoc() {
	tp.Infof("Generating README.md: %s", info.ProjPath())

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		tp.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	// read temptale file
	b, err := ioutil.ReadFile(MicroTpl)
	if err != nil {
		b = []byte(strings.Replace(__tpl__, "${PROJ_NAME}", info.ProjName(), -1))
	}

	// new project code
	proj := NewProject(b)
	proj.gen()
	proj.genAndWriteReadmeFile()

	// write template file
	f, err := os.OpenFile(MicroTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		tp.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	tp.Infof("Completed README.md generation!")
}
