/*
 * @Author: Chris
 * @Date: 2023-06-08 10:04:58
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:48:56
 * @Description: 请填写简介
 */
package DB

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/chris-liu-zh/qiao/tools"
)

var Insert = "INSERT INTO ${table}(${field})VALUES (${sign})"

// 以struct添加数据
func (mapper *Mapper) Add(data any) (r sql.Result, err error) {
	mapper = mapper.getInsert(data)
	return mapper.Exec()
}

func (mapper *Mapper) LastAddId(data any) (int64, error) {
	mapper = mapper.getInsert(data)
	return mapper.lastInsertId()
}

func (mapper *Mapper) getInsert(data any) *Mapper {
	v := reflect.ValueOf(data)
	var field strings.Builder
	var l int
	if v.Kind() == reflect.Pointer {
		elem := v.Elem()
		if elem.Kind() != reflect.Struct {
			return mapper
		}
		if mapper.Debris.table == "" {
			mapper.Debris.table = tools.CamelCaseToUdnderscore(elem.Type().Name())
		}
		for i := range elem.NumField() {
			if !elem.Field(i).IsValid() || (elem.Field(i).Kind() == reflect.Pointer && elem.Field(i).IsNil()) || !elem.Field(i).CanInterface() {
				continue
			}
			fields := strings.Split(elem.Type().Field(i).Tag.Get("db"), ";")
			if WritableField(fields) {
				if c := getColumn(fields); c != "" {
					field.WriteString(c + `,`)
				} else {
					field.WriteString(tools.CamelCaseToUdnderscore(elem.Type().Field(i).Name) + `,`)
				}

				mapper.Complete.Args = append(mapper.Complete.Args, elem.Field(i).Interface())
				l++
			}
		}
	}

	if v.Kind() == reflect.Map {
		for k, v := range data.(map[string]any) {
			field.WriteString(tools.CamelCaseToUdnderscore(k) + `,`)
			mapper.Complete.Args = append(mapper.Complete.Args, v)
		}
	}

	mapper.Debris.sign = Placeholders(l)
	mapper.Debris.field = strings.TrimRight(field.String(), ",")
	mapper.SqlTpl = Insert
	return mapper
}

func (mapper *Mapper) lastInsertId() (insertId int64, err error) {
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log("get sql error").logERROR(err)
		return
	}
	mapper.debug("lastInsertId")
	if insertId, err = mapper.Write().DBFunc.AddReturnId(mapper); err != nil {
		return
	}
	return
}
