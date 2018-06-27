package create

var codeFiles = map[string]string{
	"main.go": `package main

import (
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/discovery"

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
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/model/etcd"
	"github.com/xiaoenai/tp-micro/model/mongo"
	"github.com/xiaoenai/tp-micro/model/mysql"
	"github.com/xiaoenai/tp-micro/model/redis"

	"${import_prefix}/logic/model"
)

type config struct {
	Srv      micro.SrvConfig ` + "`yaml:\"srv\"`" + `
	Etcd     etcd.EasyConfig ` + "`yaml:\"etcd\"`" + `
	Mysql    mysql.Config    ` + "`yaml:\"mysql\"`" + `
	Mongo    mongo.Config    ` + "`yaml:\"mongo\"`" + `
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
	var (
		mysqlConfig *mysql.Config
		mongoConfig *mongo.Config
		redisConfig = &c.Redis
	)
	if len(c.Mysql.Host)>0{
		mysqlConfig=&c.Mysql
	}
	if len(c.Mongo.Addrs)>0{
		mongoConfig=&c.Mongo
	}
	err = model.Init(mysqlConfig, mongoConfig, redisConfig)
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
	Redis:    *redis.NewConfig(),
	LogLevel: "TRACE",
}

func init() {
	goutil.WritePidFile()
	cfgo.MustReg("${service_api_prefix}", cfg)
}
`,

	"logic/model/init.go": `package model

import (
	"strings"

	"github.com/xiaoenai/tp-micro/model/mongo"
	"github.com/xiaoenai/tp-micro/model/mysql"
	"github.com/xiaoenai/tp-micro/model/redis"
)

// mysqlHandler preset mysql DB handler
var mysqlHandler = mysql.NewPreDB()

// mongoHandler preset mongo DB handler
var mongoHandler = mongo.NewPreDB()

var redisClient *redis.Client

// Init initializes the model packet.
func Init(mysqlConfig *mysql.Config, mongoConfig *mongo.Config, redisConfig *redis.Config) error {
	var err error
	if redisConfig != nil {
		redisClient, err = redis.NewClient(redisConfig)
		if err != nil {
			return err
		}
	}
	if mysqlConfig!=nil{
		if err = mysqlHandler.Init2(mysqlConfig, redisClient);err!=nil{
			return err
		}
	}
	if mongoConfig!=nil{
		if err = mongoHandler.Init2(mongoConfig, redisClient);err!=nil{
			return err
		}
	}
	return nil
}

// GetMysqlDB returns the mysql DB handler.
func GetMysqlDB() *mysql.DB {
	return mysqlHandler.DB
}

// GetMongoDB returns the mongo DB handler.
func GetMongoDB() *mongo.DB {
	return mongoHandler.DB
}

// GetRedis returns the redis client.
func GetRedis() *redis.Client {
	return redisClient
}

func insertZeroDeletedTsField(whereCond string)string{
	whereCond=strings.TrimSpace(whereCond)
	whereCond=strings.TrimRight(whereCond,";")
	i:=strings.Index(whereCond,"OFFSET")
	if i==-1{
		i=strings.Index(whereCond,"offset")
		if i==-1{
			i=strings.Index(whereCond,"Offset")
		}
	}
	i2:=strings.Index(whereCond,"LIMIT")
	if i2==-1{
		i2=strings.Index(whereCond,"limit")
		if i2==-1{
			i2=strings.Index(whereCond,"Limit")
		}
	}
	if i==-1 || (i>i2 && i2!=-1) {
		i=i2
	}
	if i==-1 {
		return	whereCond+" ` + "AND `deleted_ts`" + `=0"
	}
	return whereCond[:i]+" ` + "AND `deleted_ts`" + `=0 "+whereCond[i:]
}
`,

	"args/type.gen.go": `package args
import (${import_list}
)
${type_define_list}
`,

	"logic/tmp_code.gen.go": `package logic
import (
	tp "github.com/henrylee2cn/teleport"

	"${import_prefix}/args"
	// "${import_prefix}/logic/model"
	// "${import_prefix}/rerrs"
)
${logic_api_define}
`,

	"api/pull_handler.gen.go": `package api
import (
    tp "github.com/henrylee2cn/teleport"

    "${import_prefix}/logic"
    "${import_prefix}/args"
)
${handler_api_define}
`,

	"api/push_handler.gen.go": `package api
import (
    tp "github.com/henrylee2cn/teleport"

    "${import_prefix}/logic"
    "${import_prefix}/args"
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
	micro "github.com/xiaoenai/tp-micro"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
    "github.com/xiaoenai/tp-micro/discovery"
	"github.com/xiaoenai/tp-micro/model/etcd"

	"${import_prefix}/args"
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

	micro "github.com/xiaoenai/tp-micro"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/model/etcd"

	"${import_prefix}/args"
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

const mysqlModelTpl = `package model

import (
	"database/sql"
	"time"
	"unsafe"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/goutil/coarsetime"
	"github.com/xiaoenai/tp-micro/model/mysql"
	"github.com/xiaoenai/tp-micro/model/sqlx"

	"${import_prefix}/args"
)

{{.Doc}}type {{.Name}} args.{{.Name}}

// To{{.Name}} converts to *{{.Name}} type.
func To{{.Name}}(_{{.LowerFirstLetter}} *args.{{.Name}}) *{{.Name}} {
	return (*{{.Name}})(unsafe.Pointer(_{{.LowerFirstLetter}}))
}

// ToArgs{{.Name}} converts to *args.{{.Name}} type.
func ToArgs{{.Name}}(_{{.LowerFirstLetter}} *{{.Name}}) *args.{{.Name}} {
	return (*args.{{.Name}})(unsafe.Pointer(_{{.LowerFirstLetter}}))
}

// TableName implements 'github.com/xiaoenai/tp-micro/model'.Cacheable
func (*{{.Name}}) TableName() string {
	return "{{.SnakeName}}"
}

func (_{{.LowerFirstLetter}} *{{.Name}}) isZeroPrimaryKey() bool {
	{{range .PrimaryFields}}var _{{.ModelName}} {{.Typ}}
	if _{{$.LowerFirstLetter}}.{{.Name}}!=_{{.ModelName}}{
		return false
	}
	{{end}}return true
}

var {{.LowerFirstName}}DB, _ = mysqlHandler.RegCacheableDB(new({{.Name}}), time.Hour*24, ` + "``" + `)

// Get{{.Name}}DB returns the {{.Name}} DB handler.
func Get{{.Name}}DB() *mysql.CacheableDB {
	return {{.LowerFirstName}}DB
}

// Insert{{.Name}} insert a {{.Name}} data into database.
// NOTE:
//  Primary key:{{range .PrimaryFields}} '{{.ModelName}}'{{end}};
//  Without cache layer.
func Insert{{.Name}}(_{{.LowerFirstLetter}} *{{.Name}}, tx ...*sqlx.Tx) ({{if .IsDefaultPrimary}}int64,{{end}}error) {
	_{{.LowerFirstLetter}}.UpdatedAt = coarsetime.FloorTimeNow().Unix()
	if _{{.LowerFirstLetter}}.CreatedAt == 0 {
		_{{.LowerFirstLetter}}.CreatedAt = _{{.LowerFirstLetter}}.UpdatedAt
	}
	return {{if .IsDefaultPrimary}}_{{.LowerFirstLetter}}.Id,{{end}}{{.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
		var (
			query string
			isZeroPrimaryKey=_{{.LowerFirstLetter}}.isZeroPrimaryKey()
		)
		if isZeroPrimaryKey {
			query = "INSERT INTO {{.NameSql}} ({{index .QuerySql 0}})VALUES({{index .QuerySql 1}});"
		} else {
			query = "INSERT INTO {{.NameSql}} ({{range .PrimaryFields}}` + "`{{.ModelName}}`," + `{{end}}{{index .QuerySql 0}})VALUES({{range .PrimaryFields}}:{{.ModelName}},{{end}}{{index .QuerySql 1}});"
		}
		{{if .IsDefaultPrimary}}r, err := tx.NamedExec(query, _{{.LowerFirstLetter}})
		if isZeroPrimaryKey {
			_{{.LowerFirstLetter}}.Id, err = r.LastInsertId()
		}
		{{else}}_, err := tx.NamedExec(query, _{{.LowerFirstLetter}})
		{{end}}return err
	}, tx...)
}

// Upsert{{.Name}} insert or update the {{.Name}} data by primary key.
// NOTE:
//  Primary key:{{range .PrimaryFields}} '{{.ModelName}}'{{end}};
//  With cache layer;
//  Insert data if the primary key is specified;
//  Update data based on _updateFields if no primary key is specified;
//  _updateFields' members must be db field style (snake format);
//  Automatic update 'updated_at' field;
//  Don't update the primary keys, 'created_at' key and 'deleted_ts' key;
//  Update all fields except the primary keys, 'created_at' key and 'deleted_ts' key, if _updateFields is empty.
func Upsert{{.Name}}(_{{.LowerFirstLetter}} *{{.Name}}, _updateFields []string, tx ...*sqlx.Tx) ({{if .IsDefaultPrimary}}int64,{{end}}error) {
	if _{{.LowerFirstLetter}}.UpdatedAt == 0 {
		_{{.LowerFirstLetter}}.UpdatedAt = coarsetime.FloorTimeNow().Unix()
	}
	if _{{.LowerFirstLetter}}.CreatedAt == 0 {
		_{{.LowerFirstLetter}}.CreatedAt = _{{.LowerFirstLetter}}.UpdatedAt
	}
	return {{if .IsDefaultPrimary}}_{{.LowerFirstLetter}}.Id,{{end}}{{.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
		var (
			query string
			isZeroPrimaryKey=_{{.LowerFirstLetter}}.isZeroPrimaryKey()
		)
		if isZeroPrimaryKey {
			query = "INSERT INTO {{.NameSql}} ({{index .QuerySql 0}})VALUES({{index .QuerySql 1}})"
		} else {
			query = "INSERT INTO {{.NameSql}} ({{range .PrimaryFields}}` + "`{{.ModelName}}`," + `{{end}}{{index .QuerySql 0}})VALUES({{range .PrimaryFields}}:{{.ModelName}},{{end}}{{index .QuerySql 1}})"
		}
		query +=" ON DUPLICATE KEY UPDATE "
		if len(_updateFields) == 0 {
			query += "{{.UpsertSqlSuffix}}"
		} else {
			for _, s := range _updateFields {
				if s == "updated_at" || s == "created_at" || s == "deleted_ts"{{range .PrimaryFields}} || s == "{{.ModelName}}"{{end}} {
					continue
				}
				query += ` + "\"`\" + s + \"`=VALUES(`\" + s + \"`),\"" + `
			}
			if query[len(query)-1] != ',' {
				return nil
			}
			query += ` + "\"`updated_at`=VALUES(`updated_at`),`deleted_ts`=0;\"" + `
		}
		{{if .IsDefaultPrimary}}r, err := tx.NamedExec(query, _{{.LowerFirstLetter}})
		if isZeroPrimaryKey {
			rowsAffected, err := r.RowsAffected()
			if err == nil && rowsAffected == 1 {
				_{{.LowerFirstLetter}}.Id, err = r.LastInsertId()
			}
		}
		{{else}}_, err := tx.NamedExec(query, _{{.LowerFirstLetter}})
		{{end}}return err
	}, tx...)
}

// Update{{.Name}}ByPrimary update the {{.Name}} data in database by primary key.
// NOTE:
//  Primary key:{{range .PrimaryFields}} '{{.ModelName}}'{{end}};
//  With cache layer;
//  _updateFields' members must be db field style (snake format);
//  Automatic update 'updated_at' field;
//  Don't update the primary keys, 'created_at' key and 'deleted_ts' key;
//  Update all fields except the primary keys, 'created_at' key and 'deleted_ts' key, if _updateFields is empty.
func Update{{.Name}}ByPrimary(_{{.LowerFirstLetter}} *{{.Name}}, _updateFields []string, tx ...*sqlx.Tx) error {
	_{{.LowerFirstLetter}}.UpdatedAt = coarsetime.FloorTimeNow().Unix()
	err := {{.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
		query := "UPDATE {{.NameSql}} SET "
		if len(_updateFields) == 0 {
			query += "{{.UpdateSql}} WHERE ` + "{{range $.PrimaryFields}}`{{.ModelName}}`=:{{.ModelName}} AND {{end}}`deleted_ts`=0" + ` LIMIT 1;"
		} else {
			for _, s := range _updateFields {
				if s == "updated_at" || s == "created_at" || s == "deleted_ts"{{range .PrimaryFields}} || s == "{{.ModelName}}"{{end}} {
					continue
				}
				query += ` + "\"`\" + s + \"`=:\" + s + \",\"" + `
			}
			if query[len(query)-1] != ',' {
				return nil
			}
			query += ` + "\"`updated_at`=:updated_at WHERE {{range .PrimaryFields}}`{{.ModelName}}`:={{.ModelName}} AND {{end}}`deleted_ts`=0 LIMIT 1;\"" + `
		}
		_, err := tx.NamedExec(query, _{{.LowerFirstLetter}})
		return err
	}, tx...)
	if err != nil {
		return err
	}
	err = {{.LowerFirstName}}DB.DeleteCache(_{{.LowerFirstLetter}})
	if err != nil {
		tp.Errorf("%s", err.Error())
	}
	return nil
}

{{range .UniqueFields}}
// Update{{$.Name}}By{{.Name}} update the {{$.Name}} data in database by '{{.ModelName}}' unique key.
// NOTE:
//  With cache layer;
//  _updateFields' members must be db field style (snake format);
//  Automatic update 'updated_at' field;
//  Don't update the primary keys, 'created_at' key and 'deleted_ts' key;
//  Update all fields except the primary keys, '{{.ModelName}}' unique key, 'created_at' key and 'deleted_ts' key, if _updateFields is empty.
func Update{{$.Name}}By{{.Name}}(_{{$.LowerFirstLetter}} *{{$.Name}}, _updateFields []string, tx ...*sqlx.Tx) error {
	_{{$.LowerFirstLetter}}.UpdatedAt = coarsetime.FloorTimeNow().Unix()
	err := {{$.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
		query := "UPDATE {{$.NameSql}} SET "
		if len(_updateFields) == 0 {
			query += "{{$.UpdateSql}} WHERE ` + "`{{.ModelName}}`=:{{.ModelName}} AND `deleted_ts`=0" + ` LIMIT 1;"
		} else {
			for _, s := range _updateFields {
				if s == "updated_at" || s == "created_at" || s == "deleted_ts" || s == "{{.ModelName}}"{{range $.PrimaryFields}} || s == "{{.ModelName}}"{{end}} {
					continue
				}
				query += ` + "\"`\" + s + \"`=:\" + s + \",\"" + `
			}
			if query[len(query)-1] != ',' {
				return nil
			}
			query += ` + "\"`updated_at`=:updated_at WHERE `{{.ModelName}}`:={{.ModelName}} AND `deleted_ts`=0 LIMIT 1;\"" + `
		}
		_, err := tx.NamedExec(query, _{{$.LowerFirstLetter}})
		return err
	}, tx...)
	if err != nil {
		return err
	}
	err = {{$.LowerFirstName}}DB.DeleteCache(_{{$.LowerFirstLetter}},"{{.ModelName}}")
	if err != nil {
		tp.Errorf("%s", err.Error())
	}
	return nil
}
{{end}}

// Delete{{.Name}}ByPrimary delete a {{.Name}} data in database by primary key.
// NOTE:
//  Primary key:{{range .PrimaryFields}} '{{.ModelName}}'{{end}};
//  With cache layer.
func Delete{{.Name}}ByPrimary({{range .PrimaryFields}}_{{.ModelName}} {{.Typ}}, {{end}}deleteHard bool, tx ...*sqlx.Tx) error {
	var err error
	if deleteHard {
		// Immediately delete from the hard disk.
		err = {{.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
				_, err := tx.Exec("DELETE FROM {{.NameSql}} WHERE {{range .PrimaryFields}}` + "`{{.ModelName}}`=? AND {{end}}`deleted_ts`=0;" + `", {{range .PrimaryFields}}_{{.ModelName}}, {{end}})
				return err
			}, tx...)

	}else {
		// Delay delete from the hard disk.
		ts := coarsetime.FloorTimeNow().Unix()
		err = {{.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
			_, err := tx.Exec("UPDATE {{.NameSql}} SET ` + "`updated_at`=?, `deleted_ts`=?" + ` WHERE {{range .PrimaryFields}}` + "`{{.ModelName}}`=? AND {{end}}`deleted_ts`=0;" + `", ts, ts, {{range .PrimaryFields}}_{{.ModelName}}, {{end}})
			return err
		}, tx...)
	}
	
	if err != nil {
		return err
	}
	err = {{.LowerFirstName}}DB.DeleteCache(&{{.Name}}{
		{{range .PrimaryFields}}{{.Name}}:_{{.ModelName}},
		{{end}} })
	if err != nil {
		tp.Errorf("%s", err.Error())
	}
	return nil
}

{{range .UniqueFields}}
// Delete{{$.Name}}By{{.Name}} delete a {{$.Name}} data in database by '{{.ModelName}}' unique key.
// NOTE:
//  With cache layer.
func Delete{{$.Name}}By{{.Name}}(_{{.ModelName}} {{.Typ}}, deleteHard bool, tx ...*sqlx.Tx) error {
	var err error
	if deleteHard {
		// Immediately delete from the hard disk.
		err = {{$.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
				_, err := tx.Exec("DELETE FROM {{$.NameSql}} WHERE ` + "`{{.ModelName}}`=? AND `deleted_ts`=0;" + `", _{{.ModelName}})
				return err
			}, tx...)

	}else {
		// Delay delete from the hard disk.
		ts := coarsetime.FloorTimeNow().Unix()
		err = {{$.LowerFirstName}}DB.Callback(func(tx sqlx.DbOrTx) error {
			_, err := tx.Exec("UPDATE {{$.NameSql}} SET ` + "`updated_at`=?, `deleted_ts`=?" + ` WHERE ` + "`{{.ModelName}}`=? AND `deleted_ts`=0;" + `", ts, ts, _{{.ModelName}})
			return err
		}, tx...)
	}
	
	if err != nil {
		return err
	}
	err = {{$.LowerFirstName}}DB.DeleteCache(&{{$.Name}}{
		{{.Name}}:_{{.ModelName}},
		},"{{.ModelName}}")
	if err != nil {
		tp.Errorf("%s", err.Error())
	}
	return nil
}
{{end}}

// Get{{.Name}}ByPrimary query a {{.Name}} data from database by primary key.
// NOTE:
//  Primary key:{{range .PrimaryFields}} '{{.ModelName}}'{{end}};
//  With cache layer;
//  If @return bool=false error=nil, means the data is not exist.
func Get{{.Name}}ByPrimary({{range .PrimaryFields}}_{{.ModelName}} {{.Typ}}, {{end}}) (*{{.Name}}, bool, error) {
	var _{{.LowerFirstLetter}} = &{{.Name}}{
		{{range .PrimaryFields}}{{.Name}}:_{{.ModelName}},
		{{end}} }
	err := {{.LowerFirstName}}DB.CacheGet(_{{.LowerFirstLetter}})
	switch err {
	case nil:
		if _{{.LowerFirstLetter}}.CreatedAt == 0 {
			return nil, false, nil
		}
		return _{{.LowerFirstLetter}}, true, nil
	case sql.ErrNoRows:
		err2 := {{.LowerFirstName}}DB.PutCache(_{{.LowerFirstLetter}})
		if err2 != nil {
			tp.Errorf("%s", err2.Error())
		}
		return nil, false, nil
	default:
		return nil, false, err
	}
}

{{range .UniqueFields}}
// Get{{$.Name}}By{{.Name}} query a {{$.Name}} data from database by '{{.ModelName}}' unique key.
// NOTE:
//  With cache layer;
//  If @return bool=false error=nil, means the data is not exist.
func Get{{$.Name}}By{{.Name}}(_{{.ModelName}} {{.Typ}}) (*{{$.Name}}, bool, error) {
	var _{{$.LowerFirstLetter}} = &{{$.Name}}{
		{{.Name}}:_{{.ModelName}},
		}
	err := {{$.LowerFirstName}}DB.CacheGet(_{{$.LowerFirstLetter}},"{{.ModelName}}")
	switch err {
	case nil:
		if _{{$.LowerFirstLetter}}.CreatedAt == 0 {
			return nil, false, nil
		}
		return _{{$.LowerFirstLetter}}, true, nil
	case sql.ErrNoRows:
		err2 := {{$.LowerFirstName}}DB.PutCache(_{{$.LowerFirstLetter}},"{{.ModelName}}")
		if err2 != nil {
			tp.Errorf("%s", err2.Error())
		}
		return nil, false, nil
	default:
		return nil, false, err
	}
}
{{end}}

// Get{{.Name}}ByWhere query a {{.Name}} data from database by WHERE condition.
// NOTE:
//  Without cache layer;
//  If @return bool=false error=nil, means the data is not exist.
func Get{{.Name}}ByWhere(whereCond string, arg ...interface{}) (*{{.Name}}, bool, error) {
	var _{{.LowerFirstLetter}} = new({{.Name}})
	err := {{.LowerFirstName}}DB.Get(_{{.LowerFirstLetter}}, "SELECT {{range .PrimaryFields}}` + "`{{.ModelName}}`," + `{{end}}{{index .QuerySql 0}} FROM {{.NameSql}} WHERE "+insertZeroDeletedTsField(whereCond)+ " LIMIT 1;", arg...)
	switch err {
	case nil:
		return _{{.LowerFirstLetter}}, true, nil
	case sql.ErrNoRows:
		return nil, false, nil
	default:
		return nil, false, err
	}
}

// Select{{.Name}}ByWhere query some {{.Name}} data from database by WHERE condition.
// NOTE:
//  Without cache layer.
func Select{{.Name}}ByWhere(whereCond string, arg ...interface{}) ([]*{{.Name}}, error) {
	var objs = new([]*{{.Name}})
	err := {{.LowerFirstName}}DB.Select(objs, "SELECT {{range .PrimaryFields}}` + "`{{.ModelName}}`," + `{{end}}{{index .QuerySql 0}} FROM {{.NameSql}} WHERE "+insertZeroDeletedTsField(whereCond), arg...)
	return *objs, err
}

// Count{{.Name}}ByWhere count {{.Name}} data number from database by WHERE condition.
// NOTE:
//  Without cache layer.
func Count{{.Name}}ByWhere(whereCond string, arg ...interface{}) (int64, error) {
	var count int64
	err := {{.LowerFirstName}}DB.Get(&count, "SELECT count(1) FROM {{.NameSql}} WHERE "+insertZeroDeletedTsField(whereCond), arg...)
	return count, err
}
`

const mongoModelTpl = `package model
// TODO
`
