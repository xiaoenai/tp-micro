package mysql

import (
	"database/sql/driver"
	"encoding/json"
)

var (
	emptyObjectValue = []byte("{}")
	emptyArrayValue  = []byte("[]")
)

// Int32s db column value type.
type Int32s []int32

// Scan implements the sql.Scanner interface.
func (i *Int32s) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(data, i)
}

// Value implements the driver.Valuer interface.
func (i Int32s) Value() (driver.Value, error) {
	if i == nil {
		return emptyArrayValue, nil
	}
	return json.Marshal(i)
}

// Strings db column value type.
type Strings []string

// Scan implements the sql.Scanner interface.
func (s *Strings) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(data, s)
}

// Value implements the driver.Valuer interface.
func (s Strings) Value() (driver.Value, error) {
	if s == nil {
		return emptyArrayValue, nil
	}
	return json.Marshal(s)
}

// Int8IFaceMap db column value type.
type Int8IFaceMap map[int8]interface{}

// Scan implements the sql.Scanner interface.
func (i *Int8IFaceMap) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(data, i)
}

// Value implements the driver.Valuer interface.
func (i Int8IFaceMap) Value() (driver.Value, error) {
	if i == nil {
		return emptyObjectValue, nil
	}
	return json.Marshal(i)
}

// IntStringMap db column value type.
type IntStringMap map[int]string

// Scan implements the sql.Scanner interface.
func (i *IntStringMap) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(data, i)
}

// Value implements the driver.Valuer interface.
func (i IntStringMap) Value() (driver.Value, error) {
	if i == nil {
		return emptyObjectValue, nil
	}
	return json.Marshal(i)
}
