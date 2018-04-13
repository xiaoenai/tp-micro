package create

import (
	"bytes"
	"fmt"
	"go/ast"
	"strings"
	"text/template"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

func (mod *Model) createModel(t *TypeStruct) {
	st, ok := t.expr.(*ast.StructType)
	if !ok {
		tp.Fatalf("[ant] the type of model must be struct: %s", t.Name)
	}

	mod.NameSql = fmt.Sprintf("`%s`", mod.SnakeName)
	mod.QuerySql = [2]string{}
	mod.UpdateSql = ""

	var (
		fields                            []string
		querySql1, querySql2              string
		hasId, hasCreatedAt, hasUpdatedAt bool
	)
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			tp.Fatalf("[ant] the type of model can't have anonymous field")
		}
		name := field.Names[0].Name
		if !goutil.IsExportedName(name) {
			continue
		}
		name = goutil.SnakeString(name)
		switch name {
		case "id":
			hasId = true
		case "created_at":
			hasCreatedAt = true
		case "updated_at":
			hasUpdatedAt = true
		}
		fields = append(fields, name)
	}

	if !hasId {
		t.appendHeadField(`Id int64`)
		fields = append([]string{"id"}, fields...)
	}
	if !hasCreatedAt {
		t.appendTailField(`CreatedAt int64`)
		fields = append(fields, "created_at")
	}
	if !hasUpdatedAt {
		t.appendTailField(`UpdatedAt int64`)
		fields = append(fields, "updated_at")
	}
	mod.StructDefinition = fmt.Sprintf("\n%stype %s %s\n", t.Doc, t.Name, addTag(t.Body))

	for _, field := range fields {
		if field == "id" {
			continue
		}
		querySql1 += fmt.Sprintf("`%s`,", field)
		querySql2 += fmt.Sprintf(":%s,", field)
		mod.UpdateSql += fmt.Sprintf("`%s`=:%s,", field, field)
	}
	mod.QuerySql = [2]string{querySql1[:len(querySql1)-1], querySql2[:len(querySql2)-1]}
	mod.UpdateSql = mod.UpdateSql[:len(mod.UpdateSql)-1]

	m, err := template.New("").Parse(modelTpl)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(nil)
	err = m.Execute(buf, mod)
	if err != nil {
		panic(err)
	}
	mod.code = buf.String()
}

func (t *TypeStruct) appendHeadField(fieldLine string) {
	idx := strings.Index(t.Body, "{") + 1
	t.Body = t.Body[:idx] + "\n" + fieldLine + "\n" + t.Body[idx:]
}

func (t *TypeStruct) appendTailField(fieldLine string) {
	idx := strings.LastIndex(t.Body, "}")
	t.Body = t.Body[:idx] + "\n" + fieldLine + "\n" + t.Body[idx:]
}

const modelTpl = `package model

import (
	"database/sql"
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/goutil/coarsetime"
	"github.com/xiaoenai/ants/model"
	"github.com/xiaoenai/ants/model/sqlx"
)

{{.StructDefinition}}

// TableName implements 'github.com/xiaoenai/ants/model'.Cacheable
func (*{{.Name}}) TableName() string {
	return "{{.SnakeName}}"
}

var {{.LowerFirstName}}DB, _ = dbHandler.RegCacheableDB(new({{.Name}}), time.Hour*24, ` + "``" + `)

// Get{{.Name}}DB returns the {{.Name}} DB handler.
func Get{{.Name}}DB() *model.CacheableDB {
	return {{.LowerFirstName}}DB
}

// Insert{{.Name}} insert a {{.Name}} data into database.
func Insert{{.Name}}(_{{.LowerFirstLetter}} *{{.Name}}, tx ...*sqlx.Tx) (int64, error) {
	_{{.LowerFirstLetter}}.UpdatedAt = coarsetime.FloorTimeNow().Unix()
	if _{{.LowerFirstLetter}}.CreatedAt == 0 {
		_{{.LowerFirstLetter}}.CreatedAt = _{{.LowerFirstLetter}}.UpdatedAt
	}
	return _{{.LowerFirstLetter}}.Id, {{.LowerFirstName}}DB.TransactCallback(func(tx *sqlx.Tx) error {
		var query string
		if _{{.LowerFirstLetter}}.Id > 0 {
			query = "INSERT INTO {{.NameSql}} (id,{{index .QuerySql 0}})VALUES(:id,{{index .QuerySql 1}});"
		} else {
			query = "INSERT INTO {{.NameSql}} ({{index .QuerySql 0}})VALUES({{index .QuerySql 1}});"
		}
		r, err := tx.NamedExec(query, _{{.LowerFirstLetter}})
		if err != nil {
			return err
		}
		id, err := r.LastInsertId()
		if err != nil {
			return err
		}
		_{{.LowerFirstLetter}}.Id = id
		return {{.LowerFirstName}}DB.PutCache(_{{.LowerFirstLetter}})
	}, tx...)
}

// Update{{.Name}}ById update the {{.Name}} data in database by id.
func Update{{.Name}}ById(_{{.LowerFirstLetter}} *{{.Name}}, tx ...*sqlx.Tx) error {
	return {{.LowerFirstName}}DB.TransactCallback(func(tx *sqlx.Tx) error {
		_{{.LowerFirstLetter}}.UpdatedAt = coarsetime.FloorTimeNow().Unix()
		_, err := tx.NamedExec("UPDATE {{.NameSql}} SET {{.UpdateSql}} WHERE id=:id LIMIT 1;", _{{.LowerFirstLetter}})
		if err != nil {
			return err
		}
		return {{.LowerFirstName}}DB.PutCache(_{{.LowerFirstLetter}})
	}, tx...)
}

// Delete{{.Name}}ById delete a {{.Name}} data in database by id.
func Delete{{.Name}}ById(id int64, tx ...*sqlx.Tx) error {
	return {{.LowerFirstName}}DB.TransactCallback(func(tx *sqlx.Tx) error {
		_, err := tx.Exec("DELETE FROM {{.NameSql}} WHERE id=?;", id)
		if err != nil {
			return err
		}
		return {{.LowerFirstName}}DB.PutCache(&{{.Name}}{
			Id: id,
		})
	}, tx...)
}

// Get{{.Name}}ById query a {{.Name}} data from database by id.
// If @reply bool=false error=nil, means the data is not exist.
func Get{{.Name}}ById(id int64) (*{{.Name}}, bool, error) {
	var _{{.LowerFirstLetter}} = &{{.Name}}{
		Id: id,
	}
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

// Get{{.Name}}ByWhere query a {{.Name}} data from database by WHERE condition.
// If @reply bool=false error=nil, means the data is not exist.
func Get{{.Name}}ByWhere(whereCond string, args ...interface{}) (*{{.Name}}, bool, error) {
	var _{{.LowerFirstLetter}} = new({{.Name}})
	err := {{.LowerFirstName}}DB.Get(_{{.LowerFirstLetter}}, "SELECT id,{{index .QuerySql 0}} FROM {{.NameSql}} WHERE "+whereCond+" LIMIT 1;", args...)
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
func Select{{.Name}}ByWhere(whereCond string, args ...interface{}) ([]*{{.Name}}, error) {
	var objs = new([]*{{.Name}})
	err := {{.LowerFirstName}}DB.Select(objs, "SELECT id,{{index .QuerySql 0}} FROM {{.NameSql}} WHERE "+whereCond, args...)
	return *objs, err
}

// Count{{.Name}}ByWhere count {{.Name}} data number from database by WHERE condition.
func Count{{.Name}}ByWhere(whereCond string, args ...interface{}) (int64, error) {
	var count int64
	err := {{.LowerFirstName}}DB.Get(&count, "SELECT count(1) FROM {{.NameSql}} WHERE "+whereCond, args...)
	return count, err
}`
