package DB

import (
	"database/sql"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

var ErrNoConn = errors.New("no available database connection")

/*
struct转map

	@any struct
*/
func Struct2Map(_struct any) (map[string]string, error) {
	v := reflect.ValueOf(_struct)
	if v.Kind() != reflect.Pointer {
		return nil, ErrNotPtr
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}
	data := make(map[string]string)
	for i := range elem.NumField() {
		data[elem.Type().Field(i).Tag.Get("tpl")] = elem.Field(i).String()
	}
	return data, nil
}

/*
sql模板拼接

	@sqlTpl string;	--sql模板
	@str struct;	--sqlstr
*/
func (mapper *Mapper) getSql() (sql string, err error) {
	if mapper.Debris.field == "" {
		mapper.Debris.field = "*"
	}
	if mapper.Debris.joinField == "" {
		mapper.Debris.joinField = mapper.Debris.field
	}
	sqlMap, err := Struct2Map(&mapper.Debris)
	if err != nil {
		return
	}
	return os.Expand(mapper.SqlTpl, func(k string) string { return sqlMap[k] }), nil
}

/*
调正sql占位符

	@s string;	--sql语句
	@old string;	--sql原占位符
	@new string;	--sql新占位符
*/
func Replace(s, old, new string) string {
	if old == new {
		return s
	}

	m := strings.Count(s, old)
	if m == 0 {
		return s
	}

	var b strings.Builder
	b.Grow(len(s) + m*(len(new)-len(old)))
	start := 0
	for i := range m {
		j := start
		if len(old) == 0 {
			if i > 0 {
				_, wid := utf8.DecodeRuneInString(s[start:])
				j += wid
			}
		} else {
			j += strings.Index(s[start:], old)
		}
		b.WriteString(s[start:j])
		b.WriteString(new + strconv.Itoa(i+1))
		start = j + len(old)
	}
	b.WriteString(s[start:])
	return b.String()
}

func Placeholders(n int) string {
	var b strings.Builder
	for range n - 1 {
		b.WriteString("?,")
	}
	if n > 0 {
		b.WriteString("?")
	}
	return b.String()
}

func handleNull(args ...any) (anydata []any) {
	var val any
	for _, value := range args {
		switch value.(type) {
		case sql.NullBool:
			if b, ok := (value).(sql.NullBool); ok {
				val = b.Bool
			}
			anydata = append(anydata, val)
		case sql.NullString:
			if b, ok := (value).(sql.NullString); ok {
				val = b.String
			}
			anydata = append(anydata, val)
		case sql.NullByte:
			if b, ok := (value).(sql.NullByte); ok {
				val = b.Byte
			}
			anydata = append(anydata, val)
		case sql.NullFloat64:
			if b, ok := (value).(sql.NullFloat64); ok {
				val = b.Float64
			}
			anydata = append(anydata, val)
		case sql.NullInt64:
			if b, ok := (value).(sql.NullInt64); ok {
				val = b.Int64
			}
			anydata = append(anydata, val)
		case sql.NullInt32:
			if b, ok := (value).(sql.NullInt32); ok {
				val = b.Int32
			}
			anydata = append(anydata, val)
		case sql.NullInt16:
			if b, ok := (value).(sql.NullInt16); ok {
				val = b.Int16
			}
			anydata = append(anydata, val)
		case sql.NullTime:
			if b, ok := (value).(sql.NullTime); ok {
				val = b.Time
			}
			anydata = append(anydata, val)
		default:
			anydata = append(anydata, value)
		}
	}
	return
}

func (mapper *Mapper) log(msg string) *sqlLog {
	return &sqlLog{Message: msg, Sqlstr: mapper.Complete.Sql, Args: mapper.Complete.Args}
}

func (db *ConnDB) log(msg, sqlstr string, args ...any) *sqlLog {
	return &sqlLog{Message: msg, Title: db.Conf.Title, Sqlstr: sqlstr, Args: args}
}

func (mapper *Mapper) GetSql() (sql SqlComplete, err error) {
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		return mapper.Complete, err
	}
	return mapper.Complete, nil
}
