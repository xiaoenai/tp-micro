package create

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/ants/cmd/ant/info"
)

const (
	// API_PULL_ROUTER name of the interface used to register the pull route in the template
	API_PULL_ROUTER = "__API__PULL__"
	// API_PUSH_ROUTER name of the interface used to register the push route in the template
	API_PUSH_ROUTER = "__API__PUSH__"
	// MODEL name of the struct used to create model
	MODEL = "__MODEL__"
)

type (
	// Project project Information
	Project struct {
		fileSet      *token.FileSet
		astFile      *ast.File
		Name         string
		ImprotPrefix string
		Types        []*TypeStructGroup
		CtrlStructs  map[string]*CtrlStruct
		PullApis     []*Handler
		PushApis     []*Handler
		Models       map[string]*Model
	}
	// TypeStructGroup a group of defined types
	TypeStructGroup struct {
		Doc     string
		Structs []*TypeStruct
	}
	// TypeStruct defined struct in types directory
	TypeStruct struct {
		Doc  string
		Name string
		Body string
		expr ast.Expr
	}
	// CtrlStruct defined controller struct
	CtrlStruct struct {
		Doc     string
		Name    string
		Methods []*Handler
	}
	// Handler defined router handler
	Handler struct {
		IsCtrl bool
		Name   string
		Doc    string
		Param  string
		Result string
	}
	Model struct {
		Name             string
		SnakeName        string
		LowerFirstName   string
		LowerFirstLetter string
		StructDefinition string
		NameSql          string
		QuerySql         [2]string
		UpdateSql        string
		code             string
	}
)

// NewProject new project.
func NewProject(src []byte) *Project {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		tp.Fatalf("[ant] %v", err)
	}

	var proj Project
	proj.CtrlStructs = make(map[string]*CtrlStruct)
	proj.Name = info.ProjName()
	proj.ImprotPrefix = info.ProjPath()
	proj.fileSet = fset
	proj.astFile = file
	proj.Models = make(map[string]*Model)

	return &proj
}

// Prepare prepares project.
func (p *Project) Prepare() {
	for k := range codeFiles {
		p.fillFile(k)
	}

	for _, decl := range p.astFile.Decls {
		genDecl := decl.(*ast.GenDecl)

		var typeStructGroup = &TypeStructGroup{
			Doc: groupComment(genDecl.Doc.Text()),
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			name := typeSpec.Name.Name
			var doc = typeStructGroup.Doc
			if len(genDecl.Specs) > 1 {
				doc = comment(name, typeSpec.Doc.Text(), typeSpec.Comment.Text())
			}

			switch t := typeSpec.Type.(type) {

			case *ast.StructType, *ast.Ident:

				// model
				if name == MODEL {
					mod, ok := t.(*ast.StructType)
					if !ok {
						tp.Fatalf("[ant] the type of %s must be struct", MODEL)
					}
					p.parseToModels(mod)
					continue
				}

				// types
				if !goutil.IsExportedName(name) {
					tp.Fatalf("[ant] Unexported types: %s", name)
				}

				dst := bytes.NewBuffer(nil)
				format.Node(dst, p.fileSet, typeSpec.Type)
				body := dst.String()

				typeStructGroup.Structs = append(typeStructGroup.Structs, &TypeStruct{
					Name: name,
					Doc:  doc,
					Body: body,
					expr: t,
				})

			case *ast.InterfaceType:

				var handlers []*Handler

				for _, method := range t.Methods.List {

					switch tt := method.Type.(type) {
					case *ast.FuncType:
						// function
						subName := method.Names[0].String()
						if tt.Params.NumFields() != 1 {
							tp.Fatalf("[ant] Invalid method: %s", subName)
							continue
						}
						dst := bytes.NewBuffer(nil)
						format.Node(dst, p.fileSet, tt.Params.List[0].Type)
						param := dst.String()

						var result string
						if n := tt.Results.NumFields(); n > 1 {
							tp.Fatalf("[ant] Invalid method: %s", subName)
						} else if n == 1 {
							dst.Reset()
							format.Node(dst, p.fileSet, tt.Results.List[0].Type)
							result = dst.String()
						}

						r := &Handler{
							IsCtrl: false,
							Name:   subName,
							Doc:    comment(subName, method.Doc.Text(), method.Comment.Text()),
							Param:  param,
							Result: result,
						}
						handlers = append(handlers, r)

					case *ast.Ident:
						// interface
						subName := tt.String()
						r := &Handler{
							IsCtrl: true,
							Name:   subName,
							Doc:    comment(subName, method.Doc.Text(), method.Comment.Text()),
						}
						handlers = append(handlers, r)
					}
				}

				switch name {
				case API_PULL_ROUTER:
					p.PullApis = append(p.PullApis, handlers...)

				case API_PUSH_ROUTER:
					p.PushApis = append(p.PushApis, handlers...)

				default:
					// controller
					p.CtrlStructs[name] = &CtrlStruct{
						Name:    name,
						Doc:     doc,
						Methods: handlers,
					}
				}
			}
		}

		if len(typeStructGroup.Structs) > 0 {
			p.Types = append(p.Types, typeStructGroup)
		}
	}
}

func (p *Project) fillFile(k string) {
	v, ok := codeFiles[k]
	if !ok {
		return
	}
	v = strings.Replace(v, "${import_prefix}", p.ImprotPrefix, -1)
	switch k {
	case "main.go", "config.go", "logic/model/init.go":
		codeFiles[k] = v
	case "logic/tmp_code.gen.go":
		codeFiles[k] = "// Code generated by 'ant gen' command.\n" +
			"// The temporary code used to ensure successful compilation!\n" +
			"// When the project is completed, it should be removed!\n\n" + v
	default:
		codeFiles[k] = "// Code generated by 'ant gen' command.\n// DO NOT EDIT!\n\n" + v
	}
}

func (p *Project) parseToModels(mod *ast.StructType) {
	for _, field := range mod.Fields.List {
		t, ok := field.Type.(*ast.Ident)
		if !ok {
			starExpr, ok := field.Type.(*ast.StarExpr)
			if ok {
				if t, ok = starExpr.X.(*ast.Ident); ok {
					p.Models[t.Name] = newModel(t.Name)
					continue
				}
			}
			tp.Fatalf("[ant] the field %s of %s must be struct", t.Name, MODEL)
		}
		p.Models[t.Name] = newModel(t.Name)
	}
	// fmt.Printf("%v", p.Models)
}

func newModel(name string) *Model {
	model := &Model{
		Name:      name,
		SnakeName: goutil.SnakeString(name),
	}
	model.LowerFirstLetter = strings.ToLower(model.Name[:1])
	model.LowerFirstName = model.LowerFirstLetter + model.Name[1:]
	return model
}

var codeFiles = map[string]string{
	"main.go": `package main

import (
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"

	"${import_prefix}/api"
)

func main() {
	srv := micro.NewServer(
		cfg.Srv,
		discovery.ServicePlugin(cfg.Srv.InnerIpPort(), cfg.Etcd),
	)
	api.Route("/${service_api_prefix}", srv.Router())
	srv.ListenAndServe()
}
`,
	"config.go": `package main

import (
	"time"
	
	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	"github.com/xiaoenai/ants/model"
	"github.com/xiaoenai/ants/model/redis"

	mod "${import_prefix}/logic/model"
)

type config struct {
	Srv      micro.SrvConfig ` + "`yaml:\"srv\"`" + `
	Etcd     etcd.EasyConfig ` + "`yaml:\"etcd\"`" + `
	DB       model.Config    ` + "`yaml:\"db\"`" + `
	Redis    redis.Config    ` + "`yaml:\"redis\"`" + `
	LogLevel string          ` + "`yaml:\"log_level\"`" + `
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err != nil {
		return err
	}
	if len(c.LogLevel) == 0 {
		c.LogLevel = "TRACE"
	}
	tp.SetLoggerLevel(c.LogLevel)
	err = mod.Init(c.DB, c.Redis)
	if err != nil {
		tp.Errorf("%v", err)
	}
	return nil
}

var cfg = &config{
	Srv: micro.SrvConfig{
		ListenAddress:     ":9090",
		EnableHeartbeat:   true,
		PrintDetail:       true,
		CountTime:         true,
		SlowCometDuration: time.Millisecond * 500,
	},
	Etcd: etcd.EasyConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	},
	DB: model.Config{
		Port: 3306,
	},
	Redis:    *redis.NewConfig(),
	LogLevel: "TRACE",
}

func init() {
	goutil.WritePidFile()
	cfgo.MustReg("${service_api_prefix}", cfg)
}
`,
	"types/types.gen.go": `package types
import (
	${has_model} "${import_prefix}/logic/model"
)
${type_define_list}
`,

	"logic/tmp_code.gen.go": `package logic
import (
	tp "github.com/henrylee2cn/teleport"

	"${import_prefix}/types"
	// "${import_prefix}/logic/model"
	// "${import_prefix}/rerrs"
)
${logic_api_define}
`,

	"logic/model/init.go": `package model

import (
	"github.com/xiaoenai/ants/model"
	"github.com/xiaoenai/ants/model/redis"
)

// dbHandler preset DB handler
var dbHandler = model.NewPreDB()

// Init initializes the model packet.
func Init(dbConfig model.Config, redisConfig redis.Config) error {
	return dbHandler.Init(&dbConfig, &redisConfig)
}

// GetDB returns the DB handler.
func GetDB() *model.DB {
	return dbHandler.DB
}

// GetRedis returns the redis client.
func GetRedis() *redis.Client {
	return dbHandler.DB.Cache
}`,

	"api/handler.gen.go": `package api
import (
    tp "github.com/henrylee2cn/teleport"

    "${import_prefix}/logic"
    "${import_prefix}/types"
)
${handler_api_define}
`,

	"api/router.gen.go": `
package api
import (
    tp "github.com/henrylee2cn/teleport"
)
// Route registers handlers to router.
func Route(_root string, _router *tp.Router) {
    // root router group
    _group := _router.SubRoute(_root)
    
    // custom router
    customRoute(_group.ToRouter())
   
    // automatically generated router
    ${register_router_list}}
`,

	"sdk/rpc.gen.go": `package sdk
import (
	micro "github.com/henrylee2cn/tp-micro"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
    "github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"

	"${import_prefix}/types"
)
var client *micro.Client
// Init initializes client with configs.
func Init(cliConfig micro.CliConfig, etcdConfing etcd.EasyConfig) {
	client = micro.NewClient(
		cliConfig,
		discovery.NewLinker(etcdConfing),
	)
}
// InitWithClient initializes client with specified object.
func InitWithClient(cli *micro.Client) {
	client = cli
}
${rpc_call_define}
`,

	"sdk/rpc.gen_test.go": `package sdk
import (
	"testing"

	micro "github.com/henrylee2cn/tp-micro"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"

	"${import_prefix}/types"
)

// TestSdk test SDK.
func TestSdk(t *testing.T) {
	Init(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	)
	${rpc_call_test_define}}
`}

func mustMkdirAll(dir string) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		tp.Fatalf("[ant] %v", err)
	}
}

// Generator generates code files.
func (p *Project) Generator() {
	// generate all codes
	p.genMainFile()
	p.genTypeAndModelFiles()
	p.genRouterFile()
	p.genHandlerAndLogicAndSdkFiles()
	// make all directorys
	mustMkdirAll("types")
	mustMkdirAll("api")
	mustMkdirAll("logic/model")
	mustMkdirAll("sdk")
	// write files
	notFirst := goutil.FileExists("main.go")
	for k, v := range codeFiles {
		if notFirst && (k == "main.go" || k == "config.go" || k == "logic/model/init.go") {
			continue
		}
		f, err := os.OpenFile(k, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if err != nil {
			tp.Fatalf("[ant] Create files error: %v", err)
		}
		b := formatSource(goutil.StringToBytes(v))
		f.Write(b)
		f.Close()
		fmt.Printf("generate %s\n", info.ProjPath()+"/"+k)
	}
}

func (p *Project) genMainFile() {
	replace("main.go", "${service_api_prefix}", goutil.SnakeString(p.Name))
	replace("config.go", "${service_api_prefix}", goutil.SnakeString(p.Name))
}

func (p *Project) genTypeAndModelFiles() {
	var s string
	for _, t := range p.Types {
		s += t.createTypesAndModels(p.Models)
	}
	if len(p.Models) > 0 {
		replace("types/types.gen.go", "${has_model}", "")
	} else {
		replace("types/types.gen.go", "${has_model}", "//")
	}
	replaceWithLine("types/types.gen.go", "${type_define_list}", s)
	for k, v := range p.Models {
		fileName := "logic/model/" + goutil.SnakeString(k) + ".gen.go"
		codeFiles[fileName] = v.code
		p.fillFile(fileName)
	}
}

func (p *Project) genRouterFile() {
	var s string
	if len(p.PullApis) > 0 {
		s += "\n// PULL APIs...\n"
		for _, r := range p.PullApis {
			if r.IsCtrl {
				s += fmt.Sprintf("_group.RoutePull(new(%s))\n", r.Name)
			} else {
				s += fmt.Sprintf("_group.RoutePullFunc(%s)\n", r.Name)
			}
		}
	}
	if len(p.PushApis) > 0 {
		s += "\n// PUSH APIs...\n"
		for _, r := range p.PushApis {
			if r.IsCtrl {
				s += fmt.Sprintf("_group.RoutePush(new(%s))\n", r.Name)
			} else {
				s += fmt.Sprintf("_group.RoutePushFunc(%s)\n", r.Name)
			}
		}
	}
	replaceWithLine("api/router.gen.go", "${register_router_list}", s)
}

func (p *Project) genHandlerAndLogicAndSdkFiles() {
	var handler, logic, sdk, sdkTest = p.createHandlerAndLogicAndSdk(p.PullApis, true)
	var a, b, c, d = p.createHandlerAndLogicAndSdk(p.PushApis, false)
	handler += a
	logic += b
	sdk += c
	sdkTest += d
	replaceWithLine("api/handler.gen.go", "${handler_api_define}", handler)
	replaceWithLine("logic/tmp_code.gen.go", "${logic_api_define}", logic)
	replaceWithLine("sdk/rpc.gen.go", "${rpc_call_define}", sdk)
	replaceWithLine("sdk/rpc.gen_test.go", "${rpc_call_test_define}", sdkTest)
}

func (p *Project) createHandlerAndLogicAndSdk(routers []*Handler, isPull bool) (handler, logic, sdk, sdkTest string) {
	for _, r := range routers {
		if !r.IsCtrl {
			a, b, c, d := p.createFunc(r, isPull)
			handler += a
			logic += b
			sdk += c
			sdkTest += d
		} else {
			c, ok := p.CtrlStructs[r.Name]
			if ok {
				if c.Doc == "" {
					c.Doc = r.Doc
				}
				a, b, c, d := p.createCtrlStruct(c, isPull)
				handler += a
				logic += b
				sdk += c
				sdkTest += d
			}
		}
	}
	return
}

func (p *Project) createFunc(r *Handler, isPull bool) (handler, logic, sdk, sdkTest string) {
	paramAndresult := p.checkHandler(r, isPull)
	camelName := goutil.CamelString(r.Name)
	camelDoc := strings.Replace(r.Doc, "// "+r.Name, "// "+camelName, 1)
	if isPull {
		return fmt.Sprintf(
				"%sfunc %s(ctx tp.PullCtx,args %s)(%s,*tp.Rerror){\nreturn logic.%s(ctx,args)\n}\n",
				r.Doc, r.Name, paramAndresult[0], paramAndresult[1], camelName,
			), fmt.Sprintf(
				"%sfunc %s(ctx tp.PullCtx,args %s)(%s,*tp.Rerror){\nreturn new(%s),nil\n}\n",
				camelDoc, camelName, paramAndresult[0], paramAndresult[1], paramAndresult[1][1:],
			), fmt.Sprintf(
				"%sfunc %s(args %s, setting ...socket.PacketSetting)(%s,*tp.Rerror){\n"+
					"reply := new(%s)\n"+
					"rerr := client.Pull(\"%s\", args, reply, setting...).Rerror()\n"+
					"return reply, rerr\n}\n",
				camelDoc, camelName, paramAndresult[0], paramAndresult[1],
				paramAndresult[1][1:],
				path.Join("/", goutil.SnakeString(p.Name), tp.ToUriPath(r.Name)),
			), fmt.Sprintf(
				"{\n"+
					"reply, rerr :=%s(new(%s))\n"+
					"if rerr != nil {\ntp.Errorf(\"%s: rerr: %%v\", rerr)\n} else {\ntp.Infof(\"%s: reply: %%#v\", reply)\n}\n"+
					"}\n",
				camelName, paramAndresult[0][1:], camelName, camelName,
			)

	} else {
		return fmt.Sprintf(
				"%sfunc %s(ctx tp.PushCtx,args %s)*tp.Rerror{\nreturn logic.%s(ctx,args)\n}\n",
				r.Doc, r.Name, paramAndresult[0], camelName,
			), fmt.Sprintf(
				"%sfunc %s(ctx tp.PushCtx,args %s)*tp.Rerror{\nreturn nil\n}\n",
				camelDoc, camelName, paramAndresult[0],
			), fmt.Sprintf(
				"%sfunc %s(args %s, setting ...socket.PacketSetting)*tp.Rerror{\n"+
					"return client.Push(\"%s\", args, setting...)\n"+
					"}\n",
				camelDoc, camelName, paramAndresult[0],
				path.Join("/", goutil.SnakeString(p.Name), tp.ToUriPath(r.Name)),
			), fmt.Sprintf(
				"{\n"+
					"rerr :=%s(new(%s))\n"+
					"tp.Infof(\"%s: rerr: %%v\", rerr)\n"+
					"}\n",
				camelName, paramAndresult[0][1:], camelName,
			)
	}
}

func (p *Project) createCtrlStruct(c *CtrlStruct, isPull bool) (handler, logic, sdk, sdkTest string) {
	var ctx string
	if isPull {
		ctx = "tp.PullCtx"
	} else {
		ctx = "tp.PushCtx"
	}
	handler += fmt.Sprintf("%stype %s struct{\n%s\n}\n\n", c.Doc, c.Name, ctx)
	for _, r := range c.Methods {
		a, b, c, d := p.createMethod(c.Name, r, isPull)
		handler += a
		logic += b
		sdk += c
		sdkTest += d
	}
	return
}

func (p *Project) createMethod(ctrl string, r *Handler, isPull bool) (handler, logic, sdk, sdkTest string) {
	paramAndresult := p.checkHandler(r, isPull)
	first := strings.ToLower(ctrl[:1])
	fullName := goutil.CamelString(ctrl + "_" + r.Name)
	fullDoc := strings.Replace(r.Doc, "// "+r.Name, "// "+fullName, 1)
	if isPull {
		return fmt.Sprintf(
				"%sfunc(%s *%s) %s(args %s)(%s,*tp.Rerror){\nreturn logic.%s(%s.PullCtx,args)\n}\n",
				r.Doc, first, ctrl, r.Name, paramAndresult[0], paramAndresult[1], fullName, first,
			), fmt.Sprintf(
				"%sfunc %s(ctx tp.PullCtx,args %s)(%s,*tp.Rerror){\nreturn new(%s),nil\n}\n",
				fullDoc, fullName, paramAndresult[0], paramAndresult[1], paramAndresult[1][1:],
			), fmt.Sprintf(
				"%sfunc %s(args %s, setting ...socket.PacketSetting)(%s,*tp.Rerror){\n"+
					"reply := new(%s)\n"+
					"rerr := client.Pull(\"%s\", args, reply, setting...).Rerror()\n"+
					"return reply, rerr\n}\n",
				fullDoc, fullName, paramAndresult[0], paramAndresult[1],
				paramAndresult[1][1:],
				path.Join("/", goutil.SnakeString(p.Name), tp.ToUriPath(ctrl), tp.ToUriPath(r.Name)),
			), fmt.Sprintf(
				"{\n"+
					"reply, rerr :=%s(new(%s))\n"+
					"if rerr != nil {\ntp.Errorf(\"%s: rerr: %%v\", rerr)\n} else {\ntp.Infof(\"%s: reply: %%#v\", reply)\n}\n"+
					"}\n",
				fullName, paramAndresult[0][1:], fullName, fullName,
			)

	} else {
		return fmt.Sprintf(
				"%sfunc(%s *%s) %s(args %s)*tp.Rerror{\nreturn logic.%s(%s.PushCtx,args)\n}\n",
				r.Doc, first, ctrl, r.Name, paramAndresult[0], fullName, first,
			), fmt.Sprintf(
				"%sfunc %s(ctx tp.PushCtx,args %s)*tp.Rerror{\nreturn nil\n}\n",
				fullDoc, fullName, paramAndresult[0],
			), fmt.Sprintf(
				"%sfunc %s(args %s, setting ...socket.PacketSetting)*tp.Rerror{\n"+
					"return client.Push(\"%s\", args, setting...)\n"+
					"}\n",
				fullDoc, fullName, paramAndresult[0],
				path.Join("/", goutil.SnakeString(p.Name), tp.ToUriPath(ctrl), tp.ToUriPath(r.Name)),
			), fmt.Sprintf(
				"{\n"+
					"rerr :=%s(new(%s))\n"+
					"tp.Infof(\"%s: rerr: %%v\", rerr)\n"+
					"}\n",
				fullName, paramAndresult[0][1:], fullName,
			)
	}
}

func (t *TypeStructGroup) createTypesAndModels(models map[string]*Model) (s string) {
	defer func() {
		b, _ := format.Source([]byte(s))
		s = string(b)
	}()
	switch a := len(t.Structs); {
	case a == 0:
		return ""
	case a == 1:
		tt := t.Structs[0]
		doc := t.Doc
		if doc == "" {
			doc = tt.Doc
		}
		tt.Doc = doc
		if mod, ok := models[tt.Name]; ok {
			mod.createModel(tt)
			return fmt.Sprintf("\n%stype %s = model.%s\n", tt.Doc, tt.Name, tt.Name)
		}
		return fmt.Sprintf("\n%stype %s %s\n", tt.Doc, tt.Name, addTag(tt.Body))
	default:
		var body string
		for _, tt := range t.Structs {
			if mod, ok := models[tt.Name]; ok {
				mod.createModel(tt)
				body += fmt.Sprintf("%s%s = model.%s\n", tt.Doc, tt.Name, tt.Name)
			} else {
				body += fmt.Sprintf("%s%s %s\n\n", tt.Doc, tt.Name, addTag(tt.Body))
			}
		}
		return fmt.Sprintf("\n%stype(\n%s)\n", t.Doc, body)
	}
}

var jsonRegexp = regexp.MustCompile("^[^/]*[`\\s\"]*json:\"")

func addTag(body string) string {
	body = strings.Replace(body, "\n\n", "\n", -1)
	body = strings.Replace(body, "\r\n\r\n", "\n", -1)
	body = strings.Replace(body, "\r\n", "\n", -1)
	a := strings.Split(body, "\n")
	for i, s := range a {
		if i == 0 || i == len(a)-1 {
			continue
		}
		s = strings.TrimSpace(s)
		ss := strings.TrimPrefix(s, "*")
		if len(ss) == 0 || !unicode.IsUpper(rune(ss[0])) || jsonRegexp.MatchString(ss) {
			continue
		}
		var lastIsSpace bool
		var cnt int
		var col [4]string
		for _, r := range s {
			if unicode.IsSpace(r) {
				if !lastIsSpace {
					if cnt < 3 {
						cnt++
					} else {
						col[cnt] += string(r)
					}
				}
				lastIsSpace = true
			} else {
				lastIsSpace = false
				col[cnt] += string(r)
			}
		}
		jsTag := fmt.Sprintf("json:\"%s\"", goutil.SnakeString(col[0]))
		if col[1] == "" {
			col[2] = "`" + jsTag + "`"
		} else if col[1][0] == '/' {
			col[1] = "`" + jsTag + "`" + col[1]
		} else if col[1][0] == '`' && !strings.Contains(col[1], "json:\"") {
			col[1] = col[1][:1] + jsTag + " " + col[1][1:]
		} else if col[2] == "" {
			col[2] = "`" + jsTag + "`"
		} else if col[2][0] == '/' {
			col[2] = "`" + jsTag + "`" + col[2]
		} else if !strings.Contains(col[2], "json:\"") {
			col[2] = col[2][:1] + jsTag + " " + col[2][1:]
		}
		a[i] = strings.Join(col[:], " ")
	}
	return strings.Join(a, "\n")
}

func groupComment(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return ""
	}
	return "// " + strings.Replace(s, "\n", "\n// ", -1) + "\n"
}

func comment(name, s1, s2 string) string {
	s := s1
	if s == "" {
		s = s2
	}
	s = groupComment(s)
	if len(s) == 0 {
		return "// " + name + " comment...\n"
	}
	return s
}

func (p *Project) checkHandler(r *Handler, isPull bool) (paramAndResult [2]string) {
	var tt = []string{r.Param}
	if isPull {
		if r.Result == "" {
			panic("Pull Handler mush have result: " + r.Name)
		}
		tt = append(tt, r.Result)
	} else {
		if r.Result != "" {
			panic("Push Handler must not have result: " + r.Name)
		}
	}
	for i, t := range tt {
		if len(t) < 2 || t[0] != '*' {
			panic("Arguments and results must be exised struct pointers: " + r.Name)
		}
		if t == "*struct{}" {
			paramAndResult[i] = t
		} else {
			name := t[1:]

			paramAndResult[i] = t[:1] + "types." + name
			// Q:
			// 	for _, ty := range p.Types {
			// 		for _, s := range ty.Structs {
			// 			if s.Name == name {
			// 				paramAndResult[i] = t[:1] + "types." + name
			// 				break Q
			// 			}
			// 		}
			// 	}
			// 	if paramAndResult[i] == "" {
			// 		panic("Arguments and results must be exised struct pointers: " + r.Name)
			// 	}
		}
	}
	return
}

func replace(key, placeholder, value string) string {
	a := strings.Replace(codeFiles[key], placeholder, value, -1)
	codeFiles[key] = a
	return a
}

func replaceWithLine(key, placeholder, value string) string {
	return replace(key, placeholder, "\n"+value)
}

func formatSource(src []byte) []byte {
	b, err := format.Source(src)
	if err != nil {
		panic(err.Error())
	}
	return b
}
