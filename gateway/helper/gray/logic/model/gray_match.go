package model

import (
	"database/sql"
	"time"

	"github.com/henrylee2cn/goutil/coarsetime"
	tp "github.com/henrylee2cn/teleport/v6"
	"github.com/xiaoenai/tp-micro/v6/model/mysql"
	"github.com/xiaoenai/tp-micro/v6/model/sqlx"
)

// GrayMatch
type GrayMatch struct {
	Uri       string `protobuf:"bytes,1,opt,name=Uri,proto3" json:"uri"`
	Regexp    string `protobuf:"bytes,2,opt,name=Regexp,proto3" json:"regexp"`
	CreatedAt int64  `protobuf:"varint,3,opt,name=CreatedAt,proto3" json:"created_at"`
	UpdatedAt int64  `protobuf:"varint,4,opt,name=UpdatedAt,proto3" json:"updated_at"`
}

// TableName implements 'github.com/xiaoenai/tp-micro/model'.Cacheable
func (*GrayMatch) TableName() string {
	return "gray_match"
}

var grayMatchDB, _ = dbHandler.RegCacheableDB(
	new(GrayMatch),
	time.Hour*24,
	`CREATE TABLE IF NOT EXISTS gray_match (
	uri VARCHAR(190) COMMENT 'URI',
	`+"`regexp`"+` LONGTEXT COMMENT 'regular expression to match UID',
	updated_at INT(11) NOT NULL COMMENT 'updated time',
	created_at INT(11) NOT NULL COMMENT 'created time',
	PRIMARY KEY(uri)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='gray rule';
`)

// GetGrayMatchDB returns the GrayMatch DB handler.
func GetGrayMatchDB() *mysql.CacheableDB {
	return grayMatchDB
}

// UpsertGrayMatch insert a GrayMatch data into database.
func UpsertGrayMatch(_g *GrayMatch, tx ...*sqlx.Tx) error {
	_g.UpdatedAt = coarsetime.FloorTimeNow().Unix()
	if _g.CreatedAt == 0 {
		_g.CreatedAt = _g.UpdatedAt
	}
	return grayMatchDB.TransactCallback(func(tx *sqlx.Tx) error {
		query := "INSERT INTO `gray_match` (`uri`,`regexp`,`updated_at`,`created_at`)VALUES(:uri,:regexp,:updated_at,:created_at)\n" +
			"ON DUPLICATE KEY UPDATE `regexp`=VALUES(`regexp`),`updated_at`=VALUES(`updated_at`);"
		_, err := tx.NamedExec(query, _g)
		if err != nil {
			return err
		}
		return grayMatchDB.PutCache(_g)
	}, tx...)
}

// UpdateGrayMatchByUri update the GrayMatch data in database by id.
func UpdateGrayMatchByUri(_g *GrayMatch, tx ...*sqlx.Tx) error {
	return grayMatchDB.TransactCallback(func(tx *sqlx.Tx) error {
		_g.UpdatedAt = coarsetime.FloorTimeNow().Unix()
		_, err := tx.NamedExec("UPDATE `gray_match` SET `uri`=:uri,`regexp`=:regexp,`created_at`=:created_at,`updated_at`=:updated_at WHERE uri=:uri LIMIT 1;", _g)
		if err != nil {
			return err
		}
		return grayMatchDB.PutCache(_g)
	}, tx...)
}

// DeleteGrayMatchByUri delete a GrayMatch data in database by id.
func DeleteGrayMatchByUri(uri string, tx ...*sqlx.Tx) error {
	return grayMatchDB.TransactCallback(func(tx *sqlx.Tx) error {
		_, err := tx.Exec("DELETE FROM `gray_match` WHERE uri=?;", uri)
		if err != nil {
			return err
		}
		return grayMatchDB.PutCache(&GrayMatch{
			Uri: uri,
		})
	}, tx...)
}

// GetGrayMatchByUri query a GrayMatch data from database by id.
// If @reply bool=false error=nil, means the data is not exist.
func GetGrayMatchByUri(uri string) (*GrayMatch, bool, error) {
	var _g = &GrayMatch{
		Uri: uri,
	}
	err := grayMatchDB.CacheGet(_g)
	switch err {
	case nil:
		if _g.CreatedAt == 0 {
			return nil, false, nil
		}
		return _g, true, nil
	case sql.ErrNoRows:
		err2 := grayMatchDB.PutCache(_g)
		if err2 != nil {
			tp.Errorf("%s", err2.Error())
		}
		return nil, false, nil
	default:
		return nil, false, err
	}
}

// GetGrayMatchByWhere query a GrayMatch data from database by WHERE condition.
// If @reply bool=false error=nil, means the data is not exist.
func GetGrayMatchByWhere(whereCond string, args ...interface{}) (*GrayMatch, bool, error) {
	var _g = new(GrayMatch)
	err := grayMatchDB.Get(_g, "SELECT `uri`,`regexp`,`created_at`,`updated_at` FROM `gray_match` WHERE "+whereCond+" LIMIT 1;", args...)
	switch err {
	case nil:
		return _g, true, nil
	case sql.ErrNoRows:
		return nil, false, nil
	default:
		return nil, false, err
	}
}

// SelectGrayMatchByWhere query some GrayMatch data from database by WHERE condition.
func SelectGrayMatchByWhere(whereCond string, args ...interface{}) ([]*GrayMatch, error) {
	var objs = new([]*GrayMatch)
	err := grayMatchDB.Select(objs, "SELECT `uri`,`regexp`,`created_at`,`updated_at` FROM `gray_match` WHERE "+whereCond, args...)
	return *objs, err
}

// CountGrayMatchByWhere count GrayMatch data number from database by WHERE condition.
func CountGrayMatchByWhere(whereCond string, args ...interface{}) (int64, error) {
	var count int64
	err := grayMatchDB.Get(&count, "SELECT count(1) FROM `gray_match` WHERE "+whereCond, args...)
	return count, err
}
