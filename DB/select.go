/*
 * @Author: Chris
 * @Date: 2023-07-20 15:30:06
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:50:58
 * @Description: 请填写简介
 */
package DB

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/chris-liu-zh/qiao"
)

const Select = "select ${field} from ${table} ${where} ${order} ${group}"

/*
查询最大值

	@field string; --字段名称
	@params slice; --切片
*/
func (mapper *Mapper) Max(_struct any, field string) (max int, err error) {
	ReflectV := reflect.ValueOf(_struct)
	elem := ReflectV.Elem()
	if elem.Kind() != reflect.Struct {
		return 0, ErrNotStruct
	}
	if mapper.Debris.table == "" {
		mapper.Debris.table = qiao.CamelCaseToUdnderscore(elem.Type().Name())
	}
	mapper.Debris.field = fmt.Sprintf("max(%s) as max", field)
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	mapper.debug("Max")
	if max, err = mapper.Read().Count(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}

// 获取单行map数据
func (mapper *Mapper) GetRowMap() (data map[string]any, err error) {
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return nil, err
	}
	mapper.debug("GetRowMap")
	if mapper.sqlRows, err = mapper.Read().Query(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return nil, err
	}
	if data, err = mapper.ScanRowMap(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	return
}

// 获取多行map数据
func (mapper *Mapper) GetListMap() (list []map[string]any, err error) {
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	mapper.debug("GetListMap")
	if mapper.sqlRows, err = mapper.Read().Query(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	if list, err = mapper.ScanListMap(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	return
}

// 获取单行struct数据
func (mapper *Mapper) Get(_struct any) (err error) {
	ReflectV := reflect.ValueOf(_struct)
	if ReflectV.Kind() != reflect.Pointer {
		return ErrNotPtr
	}
	elem := ReflectV.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	if mapper, err = mapper.getMapper(elem); err != nil {
		return
	}
	mapper.debug("Get")
	if mapper.sqlRows, err = mapper.Read().Query(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	if err = mapper.ScanRowStruct(_struct); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	return
}

// 获取多行struct数据
func (mapper *Mapper) GetList(_struct any) (err error) {

	reflectT := reflect.TypeOf(_struct)
	if reflectT.Kind() != reflect.Pointer {
		return ErrNotPtr
	}
	sliceType := reflectT.Elem()
	if reflect.Slice != sliceType.Kind() {
		return ErrNotSlice
	}
	elem := sliceType.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	if mapper.Debris.table == "" {
		mapper.Debris.table = qiao.CamelCaseToUdnderscore(elem.Name())
	}
	if mapper.Debris.field == "" {
		for i := range elem.NumField() {
			columns := strings.Split(elem.Field(i).Tag.Get("db"), ";")
			if ReadOnlyField(columns) {
				if c := getColumn(columns); c != "" {
					mapper.Debris.field += c + `,`
				} else {
					mapper.Debris.field += qiao.CamelCaseToUdnderscore(elem.Field(i).Name) + `,`
				}
			}
		}
	}
	mapper.Debris.field = strings.TrimRight(mapper.Debris.field, ",")
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	mapper.debug("GetList")
	if mapper.sqlRows, err = mapper.Read().Query(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	if err = mapper.ScanListStruct(_struct); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}

	return
}

func (mapper *Mapper) Count(_struct any, index string) (count int, err error) {
	ReflectV := reflect.ValueOf(_struct)
	if ReflectV.Kind() != reflect.Pointer {
		return 0, ErrNotPtr
	}
	elem := ReflectV.Elem()
	if elem.Kind() != reflect.Struct {
		return 0, ErrNotStruct
	}
	if mapper, err = mapper.getMapper(elem); err != nil {
		return 0, err
	}
	if index == "" {
		index = "*"
	}
	mapper.Complete.Sql = fmt.Sprintf("select count(%s) from(%s) a", index, mapper.Complete.Sql)
	mapper.debug("Count")
	if count, err = mapper.Read().Count(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}

func (mapper *Mapper) getMapper(elem reflect.Value) (*Mapper, error) {
	if mapper.Debris.table == "" {
		mapper.Debris.table = qiao.CamelCaseToUdnderscore(elem.Type().Name())
	}
	if mapper.Debris.field == "" {
		for i := range elem.NumField() {
			fields := strings.Split(elem.Type().Field(i).Tag.Get("db"), ";")
			if ReadOnlyField(fields) {
				if c := getColumn(fields); c != "" {
					mapper.Debris.field += c + `,`
				} else {
					mapper.Debris.field += qiao.CamelCaseToUdnderscore(elem.Type().Field(i).Name) + `,`
				}
			}
		}
	}
	mapper.Debris.field = strings.TrimRight(mapper.Debris.field, ",")
	var err error
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return mapper, err
	}
	return mapper, nil
}
