package DB

import (
	"fmt"
	"reflect"
	"strings"
)

type Order string

const (
	Desc Order = "desc"
	Asc  Order = "asc"
)

type Mapper struct {
	SqlTpl   string
	Debris   SqlDebris
	Complete SqlComplete
}

type SqlComplete struct {
	Sql   string
	Args  []any
	Debug bool //开启调试
}

type SqlDebris struct {
	table     string `tpl:"table"`
	field     string `tpl:"field"`
	joinField string `tpl:"join_field"`
	where     string `tpl:"where"`
	set       string `tpl:"set"`
	group     string `tpl:"group"`
	order     string `tpl:"order"`
	sign      string `tpl:"sign"`
	join      string `tpl:"join"`
}

func QiaoDB(openDebug ...bool) *Mapper {
	debug := false
	for _, debug = range openDebug {
	}
	return &Mapper{SqlTpl: Select, Complete: SqlComplete{Debug: debug}}
}

func (mapper *Mapper) Begin() *Begin {
	tx := Write().Begin()
	tx.Mapper = mapper
	return tx
}

/*
Set 设置

	@set string map struct；--设置updata set参数
*/
func (mapper *Mapper) Set(set any, args ...any) *Mapper {
	v := reflect.ValueOf(set)
	if v.Kind() == reflect.String {
		mapper.Debris.set += set.(string)
		mapper.Complete.Args = append(mapper.Complete.Args, args...)
		return mapper
	}
	if v.Kind() == reflect.Map {
		for k, v := range set.(map[string]any) {
			mapper.Debris.set += fmt.Sprintf(`%s = ?,`, CamelCaseToUdnderscore(k))
			mapper.Complete.Args = append(mapper.Complete.Args, v)
		}
		mapper.Debris.set = strings.TrimRight(mapper.Debris.set, ",")
		return mapper
	}

	if v.Kind() == reflect.Ptr {
		elem := v.Elem()
		if mapper.Debris.table == "" {
			mapper.Debris.table = CamelCaseToUdnderscore(elem.Type().Name())
		}
		l := elem.NumField()
		var column string
		if elem.Kind() == reflect.Struct {
			for i := range l {
				fields := strings.Split(elem.Type().Field(i).Tag.Get("db"), ";")
				if WritableField(fields) {
					if c := getColumn(fields); c != "" {
						column += c + "=?"
					} else {
						column += CamelCaseToUdnderscore(elem.Type().Field(i).Name) + `=?`
					}
					if i != l-1 {
						column += ","
					}
					mapper.Complete.Args = append(mapper.Complete.Args, elem.Field(i).Interface())
				}
			}
		}
		mapper.Debris.set = column
		return mapper
	}

	return mapper
}

// Group 设置分组
func (mapper *Mapper) Group(group string) *Mapper {
	mapper.Debris.group = "group by " + group
	return mapper
}

// Table 设置表
func (mapper *Mapper) Table(tableName string) *Mapper {
	mapper.Debris.table = tableName
	return mapper
}

/*
Order	设置排序
*/
func (mapper *Mapper) Order(order string) *Mapper {
	if order == "" {
		return mapper
	}
	mapper.Debris.order = "order by " + order
	return mapper
}

// Field 设置字段
func (mapper *Mapper) Field(field string) *Mapper {
	mapper.Debris.field = field
	return mapper
}

/*
Limit	设置分页

	@size int;-- 页面大小
	@page int;-- 当前页
*/
func (mapper *Mapper) Limit(size, page int) *Mapper {
	return Read().DBFunc.Page(mapper, size, page)
}
