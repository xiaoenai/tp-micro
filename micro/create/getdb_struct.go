package create

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/micro/create/test"
	"github.com/xiaoenai/tp-micro/micro/info"
	"golang.org/x/crypto/ssh"
)

type ConnConfig struct {
	MysqlConfig
	Tables  []string
	SshHost string
	SshPort string
	SshUser string
}

func AddTableStructToTpl(cfg ConnConfig) {
	if len(cfg.Tables) == 0 {
		return
	}
	var tabs []string
	for _, tb := range cfg.Tables {
		tabs = append(tabs, strings.Split(tb, ",")...)
	}
	cfg.Tables = tabs

	var db *sql.DB
	var err error
	if cfg.SshHost != "" {
		db, err = NewMysqlDbInSSH(cfg.SshHost+":"+cfg.SshPort, cfg.SshUser, &cfg.MysqlConfig)
	} else {
		db, err = sql.Open("mysql", cfg.MysqlConfig.ConnString())
	}
	if err != nil {
		tp.Fatalf("[micro] op driver: %s", err.Error())
	}

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err = os.Chdir(info.AbsPath())
	if err != nil {
		tp.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	// read temptale file
	b, err := ioutil.ReadFile(MicroTpl)
	if err != nil {
		b = test.MustAsset(MicroTpl)
	}
	b = formatSource(b)
	b = bytes.Replace(b, []byte{'\r', '\n'}, []byte{'\n'}, -1)
	p := []byte("\ntype __MYSQL__MODEL__ struct {\n")
	var i, j int
	i = bytes.Index(b, p)
	if i == -1 {
		b = append(b, '\n')
		b = append(b, p...)
		i = len(b)
		j = i
		b = append(b, "\n}\n\n"...)
	} else {
		i += len(p)
		var v byte
		for j, v = range b[i:] {
			if v == '}' {
				j += i
				break
			}
		}
	}
	var text = string(b)
	var sub = string(b[i:j])
	var structString string
	var m = make(map[string]bool)
	var appendTimePackage bool

	for _, tabName := range cfg.Tables {
		table, err := ParseTable(db, &TableConfig{
			TableName: tabName,
			fieldsMap: make(map[string]string, 0),
		})
		if err != nil {
			tp.Fatalf("[micro] parse table: %s", err.Error())
		}
		name := goutil.CamelString(table.name)
		if m[name] {
			continue
		}
		m[name] = true
		field := "\t" + name + "\n"
		if !strings.Contains(sub, field) {
			text = text[:j] + field + text[j:]
			j += len(field)
		}

		if !strings.Contains(text, "type "+name+" struct {") {
			structString += "\n" + table.String() + "\n"
			appendTimePackage = appendTimePackage || table.timeImport
		}
	}
	if len(structString) > 0 {
		fmt.Printf(
			"Added mysql model struct code:\n%s",
			formatSource([]byte(structString)),
		)
		if appendTimePackage {
			tInfo := newTplInfo([]byte(text))
			tInfo.Parse()
			for _, v := range tInfo.typeImports {
				if v == `"time"` {
					appendTimePackage = false
					break
				}
			}
			if appendTimePackage {
				text = strings.Replace(text, "package __TPL__\n", "package __TPL__\nimport \"time\"\n", 1)
			}
		}
	}
	b = []byte(text + "\n\n" + structString)

	// write template file
	f, err := os.OpenFile(MicroTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		tp.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	tp.Infof("Added mysql model struct code to project template!")
}

//-----------------------------------------------table parser-------------------------------------------------//
type Table struct {
	name       string
	Fields     []string
	Types      []string
	Flags      []string
	timeImport bool
}

func (t *Table) String() string {
	buf := strings.Builder{}
	buf.WriteString("type ")
	buf.WriteString(goutil.CamelString(t.name) + " ")
	buf.WriteString("struct { \n")
	for i := range t.Fields {
		buf.WriteString(goutil.CamelString(t.Fields[i]))
		buf.WriteByte(' ')
		buf.WriteString(t.Types[i])
		buf.WriteByte(' ')
		buf.WriteString(t.Flags[i])
		buf.WriteByte('\n')
	}
	buf.WriteString("}")
	return buf.String()
}

// func (t *Table) InsertAllStr() string {
// 	buf := strings.Builder{}
// 	buf.WriteString("INSERT INTO ")
// 	buf.WriteString(t.name + " (`")
// 	buf.WriteString(strings.Join(t.Fields, "`, `"))
// 	buf.WriteString("`) VALUES (:")
// 	buf.WriteString(strings.Join(t.Fields, ", :"))
// 	buf.WriteString(");")
// 	return buf.String()
// }

// func (t *Table) SelectAllByWhereStr() string {
// 	buf := strings.Builder{}
// 	buf.WriteString("SELECT `")
// 	buf.WriteString(strings.Join(t.Fields, "`, `"))
// 	buf.WriteString("` FROM " + t.name)
// 	buf.WriteString(" WHERE ")
// 	return buf.String()

// }
// func (t *Table) UpdateAllStr() string {
// 	buf := strings.Builder{}
// 	buf.WriteString("UPDATE " + t.name + " SET ")
// 	for i, v := range t.Fields {
// 		buf.WriteString(" `" + v + "`=:" + v)
// 		if i != len(t.Fields)-1 {
// 			buf.WriteString(", ")
// 		}
// 	}
// 	buf.WriteString(" WHERE ")
// 	return buf.String()
// }

type TableConfig struct {
	TableName string
	fieldsMap map[string]string
}

func (c *TableConfig) AddFieldsMap(k, v string) *TableConfig {
	c.fieldsMap[k] = c.fieldsMap[v]
	return c
}

func (c *TableConfig) MergeFieldsMap(m map[string]string) {
	for k, v := range m {
		c.fieldsMap[k] = v
	}
}

func (c *TableConfig) FieldsContains(v string) bool {
	if _, ok := c.fieldsMap[v]; ok {
		return true
	}
	return false
}

func (c *TableConfig) Fields(v string) string {
	if k, ok := c.fieldsMap[v]; ok {
		return k
	}
	return "interface{}"
}

func NewTableConfig(table string) *TableConfig {
	return &TableConfig{TableName: table}
}

func ParseTable(db *sql.DB, cfg *TableConfig) (*Table, error) {
	tb := new(Table)
	row, err := db.Query("desc `" + cfg.TableName + "`")
	if err != nil {
		return nil, err
	}
	tb.name = cfg.TableName
	for row.Next() {
		var Field, Type string
		var tmp interface{}
		if err := row.Scan(&Field, &Type, &tmp, &tmp, &tmp, &tmp); err != nil {
			return nil, err
		}
		f := Field
		tb.Fields = append(tb.Fields, f)
		var t string
		if cfg.FieldsContains(Type) {
			t = cfg.Fields(Type)
		} else if strings.Contains(Type, "bigint") {
			t = "int64"
		} else if strings.Contains(Type, "tinyint(1)") {
			t = "bool"
		} else if strings.Contains(Type, "int") {
			t = "int32"
		} else if strings.Contains(Type, "char") {
			t = "string"
		} else if strings.Contains(Type, "datetime") {
			t = "time.Time"
			tb.timeImport = true
		} else {
			t = "interface{}"
		}
		tb.Types = append(tb.Types, t)
		g := " `json:\"" + Field + "\"`"
		tb.Flags = append(tb.Flags, g)
	}
	return tb, nil
}

//-----------------------------------------------ssh driver-------------------------------------------------//
type MysqlConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Db       string
}

func (c MysqlConfig) ConnString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.User, c.Password, c.Host, c.Port, c.Db)
}

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func SshConfig(user string) *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			PublicKeyFile(os.Getenv("HOME") + "/.ssh/id_rsa"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return sshConfig
}

type ViaSSHDialer struct {
	client *ssh.Client
}

func (self *ViaSSHDialer) Dial(addr string) (net.Conn, error) {
	return self.client.Dial("tcp", addr)
}

func NewMysqlDbInSSH(host, user string, cfg *MysqlConfig) (*sql.DB, error) {
	sshcon, err := ssh.Dial("tcp", host, SshConfig(user))
	if err != nil {
		return nil, err
	}
	mysql.RegisterDial("tcp", (&ViaSSHDialer{sshcon}).Dial)

	db, err := sql.Open("mysql", cfg.ConnString())
	if err != nil {
		return nil, err
	}
	return db, nil
}
