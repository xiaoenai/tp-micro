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
	"github.com/henrylee2cn/erpc/v6"
	"github.com/xiaoenai/tp-micro/v6/micro/info"
	"golang.org/x/crypto/ssh"
)

// ConnConfig connection config
type ConnConfig struct {
	MysqlConfig
	Tables  []string
	SshHost string
	SshPort string
	SshUser string
}

// MysqlConfig config
type MysqlConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Db       string
}

// ConnString returns the connection string.
func (c MysqlConfig) ConnString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.User, c.Password, c.Host, c.Port, c.Db)
}

// AddTableStructToTpl adds struct to tpl
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
		db, err = newMysqlDbInSSH(cfg.SshHost+":"+cfg.SshPort, cfg.SshUser, &cfg.MysqlConfig)
	} else {
		db, err = sql.Open("mysql", cfg.MysqlConfig.ConnString())
	}
	if err != nil {
		erpc.Fatalf("[micro] op driver: %s", err.Error())
	}

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err = os.Chdir(info.AbsPath())
	if err != nil {
		erpc.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	// read temptale file
	b, err := ioutil.ReadFile(MicroTpl)
	if err != nil {
		b = []byte(strings.Replace(__tpl__, "__PROJ_NAME__", info.ProjName(), -1))
	}
	b = formatSource(b)
	b = bytes.Replace(b, []byte{'\r', '\n'}, []byte{'\n'}, -1)
	p := []byte("\ntype __MYSQL_MODEL__ struct {\n")
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
		tb, hasTimeImport, err := parseTable(db, tabName)
		if err != nil {
			erpc.Fatalf("[micro] parse table: %s", err.Error())
		}
		if m[tb.name] {
			continue
		}
		tb.initModel()
		fieldLine := "\t" + tb.name + "\n"
		if !strings.Contains(sub, fieldLine) {
			text = text[:j] + fieldLine + text[j:]
			j += len(fieldLine)
		}
		if !strings.Contains(text, "type "+tb.name+" struct {") {
			structString += "\n" + tb.String() + "\n"
			appendTimePackage = appendTimePackage || hasTimeImport
		}
		m[tb.name] = true
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
		erpc.Fatalf("[micro] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	erpc.Infof("Added mysql model struct code to project template!")
}

func parseTable(db *sql.DB, tableName string) (tb *structType, hasTimeImport bool, err error) {
	row, err := db.Query("desc `" + tableName + "`")
	if err != nil {
		return
	}
	defer row.Close()
	tb = new(structType)
	tb.modelStyle = "mysql"
	tb.name = goutil.CamelString(tableName)
	var updatedAt, createdAt, deletedTs bool
	for row.Next() {
		var (
			f         = new(field)
			modelType string
			key       string
			discard   interface{}
		)
		if err = row.Scan(&f.ModelName, &modelType, &discard, &key, &discard, &discard); err != nil {
			return
		}
		if containsAny(modelType, "bigint", "timestamp") {
			f.Typ = "int64"
		} else if containsAny(modelType, "tinyint(1)") {
			f.Typ = "bool"
		} else if containsAny(modelType, "int") {
			f.Typ = "int32"
		} else if containsAny(modelType, "float", "double") {
			f.Typ = "float64"
		} else if containsAny(modelType, "char", "text", "decimal") {
			f.Typ = "string"
		} else if containsAny(modelType, "time", "date", "year") {
			f.Typ = "time.Time"
			hasTimeImport = true
		} else {
			f.Typ = "[]byte"
		}
		switch k := strings.ToLower(key); k {
		case "pri", "uni":
			f.tag = fmt.Sprintf("`json:\"%s\" key:\"%s\"`", f.ModelName, key)
		default:
			f.tag = fmt.Sprintf("`json:\"%s\"`", f.ModelName)
		}
		f.Name = goutil.CamelString(f.ModelName)
		tb.fields = append(tb.fields, f)

		updatedAt = updatedAt || isDefaultField(f, "UpdatedAt", "int64")
		createdAt = createdAt || isDefaultField(f, "CreatedAt", "int64")
		deletedTs = deletedTs || isDefaultField(f, "DeletedTs", "int64")
	}
	if !(updatedAt && createdAt && deletedTs) {
		erpc.Warnf("Generated struct fields is not in conformity with the table fields: %s", tableName)
	}
	newTabName := goutil.SnakeString(tb.name)
	if newTabName != tableName {
		erpc.Warnf("Generated table name is not in conformity with the table name: new: %s, raw: %s", newTabName, tableName)
	}
	return
}

func isDefaultField(f *field, fieldName string, fieldType string) bool {
	return f.Name == fieldName && f.Typ == fieldType
}

func containsAny(s string, substr ...string) bool {
	for _, sub := range substr {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func publicKeyFile(file string) ssh.AuthMethod {
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

func sshConfig(user string) *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			publicKeyFile(os.Getenv("HOME") + "/.ssh/id_rsa"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return sshConfig
}

type viaSSHDialer struct {
	client *ssh.Client
}

func (self *viaSSHDialer) Dial(addr string) (net.Conn, error) {
	return self.client.Dial("tcp", addr)
}

func newMysqlDbInSSH(host, user string, cfg *MysqlConfig) (*sql.DB, error) {
	sshcon, err := ssh.Dial("tcp", host, sshConfig(user))
	if err != nil {
		return nil, err
	}
	mysql.RegisterDial("tcp", (&viaSSHDialer{sshcon}).Dial)

	db, err := sql.Open("mysql", cfg.ConnString())
	if err != nil {
		return nil, err
	}
	return db, nil
}
