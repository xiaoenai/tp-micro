package create

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/micro/create/tpl"
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
		tp.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	tpl.RestoreAsset("./", MicroGenLock)

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
		b = []byte(strings.Replace(__tpl__, "__PROJ_NAME__", info.ProjName(), -1))
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

	tp.Infof("Completed README.md generation by api!")

	// gen err code info
	appendErrorInfo()

	tp.Infof("Completed README.md generation!")
}

// appendErrorInfo append err code to README.md
func appendErrorInfo() {
	ctn := `
## Error List

|Code|Message(输出时 Msg 将会被转为 JSON string)|
|------|------|
`
	var appendRow = func(code string, msg string) {
		ctn += fmt.Sprintf("|%s|%s|\n", code, msg)
	}

	b, err := ioutil.ReadFile("rerrs/rerrs.go")
	if err != nil {
		tp.Errorf("[micro] Append error list error: %v", err)
	}

	re := regexp.MustCompile(`\(([1-9][0-9]*), rerrors\.(Msg\{.*\})`)
	a := re.FindAllStringSubmatch(string(b), -1)
	for _, row := range a {
		if len(row) != 3 {
			continue
		}
		appendRow(row[1], row[2])
	}

	f, err := os.OpenFile("README.md", os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		tp.Errorf("[micro] Append error list error: %v", err)
	}
	defer f.Close()
	f.WriteString(ctn)
	tp.Infof("Appended error list to README.md!")
}
