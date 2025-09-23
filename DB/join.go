/*
 * @Author: Chris
 * @Date: 2024-04-30 13:46:58
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-04 11:55:06
 * @Description: 请填写简介
 */
package DB

import (
	"fmt"
	"reflect"
	"strings"
)

// SELECT
// 	bi_t_item_info.item_no,
// 	bi_t_item_info.item_subno,
// 	goods_details.recommend,
// 	goods_details.hot,
// 	goods_details.special_offer,
// 	goods_details.image_list,
// 	goods_details.details,
// 	goods_details.item_no
// FROM
// 	dbo.bi_t_item_info
// 	LEFT JOIN
// 	dbo.goods_details
// 	ON
// 		bi_t_item_info.item_no = goods_details.item_no
// WHERE
// 	goods_details.item_no IS NOT NULL

// join

const SelectJoin = "select ${field} from (select ${joinfield} from ${table} ${join} ${where} ${order} ${group}) temp"

func (mapper *Mapper) Join(join, joinField, on string) *Mapper {
	mapper.Debris.joinField = joinField
	mapper.Debris.join = fmt.Sprintf("%s on %s", join, on)
	mapper.SqlTpl = SelectJoin
	return mapper
}

func (mapper *Mapper) JoinFind(params any, args ...any) *Mapper {
	v := reflect.ValueOf(params)

	if v.Kind() == reflect.String {
		return mapper.where(params.(string), args...)
	}

	if v.Kind() != reflect.Pointer {
		return mapper
	}

	elem := v.Elem()
	if elem.Kind() == reflect.Struct {
		table := CamelCaseToUdnderscore(elem.Type().Name())
		var fields string
		var args []any
		var next string
		for i := range elem.NumField() {
			find := elem.Type().Field(i).Tag.Get("find")
			exclude := elem.Type().Field(i).Tag.Get("exclude")
			if exclude == elem.Field(i).Interface() {
				continue
			}
			column := CamelCaseToUdnderscore(elem.Type().Field(i).Name)
			field, arg := getfind(column, find, elem.Field(i).Interface())
			if field == "" {
				continue
			}
			if next = elem.Type().Field(i).Tag.Get("next"); next == "" {
				next = "and"
			}
			next = strings.TrimSpace(next)
			fields += fmt.Sprintf(" %s.%s %s", table, field, next)
			args = append(args, arg...)
		}
		fields = strings.TrimSuffix(fields, next)
		return mapper.where(fields, args...)
	}
	return mapper
}
