package DB

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func (mapper *Mapper) Find(params any, args ...any) *Mapper {
	v := reflect.ValueOf(params)

	if v.Kind() == reflect.String {
		return mapper.where(params.(string), args...)
	}

	if v.Kind() == reflect.Map {
		return mapper.whereMap(params.(map[string]any))
	}

	if v.Kind() != reflect.Pointer {
		return mapper
	}

	elem := v.Elem()
	if elem.Kind() == reflect.Struct {
		return mapper.whereStruct(elem)
	}
	return mapper
}

/*
Where 自定义where条件

	@where	string;--条件查询
	@args	[]any;--条件值
*/
func (mapper *Mapper) where(where string, args ...any) *Mapper {
	if where == "" {
		return mapper
	}
	if mapper.Debris.where == "" {
		mapper.Debris.where = "where "
	} else {
		mapper.Debris.where += " and "
	}
	mapper.Debris.where += fmt.Sprintf("(%s) ", where)

	mapper.Complete.Args = append(mapper.Complete.Args, args...)
	return mapper
}

/*
WhereIn in查询

	field string; --字段名称
	params Slice; --切片
*/
// func (mapper *Mapper) WhereIn(field string, params any) *Mapper {
// 	if reflect.TypeOf(params).Kind() != reflect.Slice {
// 		return mapper
// 	}
// 	s := reflect.ValueOf(params)
// 	l := s.Len()
// 	if l == 0 {
// 		return mapper
// 	}
// 	for i := 0; i < s.Len(); i++ {
// 		mapper.Complete.Args = append(mapper.Complete.Args, s.Index(i).Interface())
// 	}
// 	query := fmt.Sprintf(" %s in (%s)", field, placeholders(l))
// 	return mapper.where(query)
// }

func (mapper *Mapper) whereStruct(elem reflect.Value) *Mapper {
	var fields string
	var args []any
	var next string
	for i := range elem.NumField() {
		find := elem.Type().Field(i).Tag.Get("find")
		exclude := elem.Type().Field(i).Tag.Get("ignore")
		if exclude == fmt.Sprintf("%v", elem.Field(i).Interface()) {
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
		fields += fmt.Sprintf(" %s %s", field, next)
		args = append(args, arg...)
	}
	fields = strings.TrimSuffix(fields, next)
	return mapper.where(fields, args...)
}

/*
WhereMap map条件查询

	@params map[string]any；--map条件
	map[column:where]v
*/
func (mapper *Mapper) whereMap(params map[string]any) *Mapper {
	var fields string
	var args []any
	for k, v := range params {
		column := strings.Split(k, ":")
		var find string
		if len(column) > 1 {
			find = column[1]
		}
		field, arg := getfind(CamelCaseToUdnderscore(column[0]), find, v)
		fields += fmt.Sprintf(" %s", field)

		args = append(args, arg...)
	}
	return mapper.where(fields, args...)
}

func getfind(column, find string, val any) (where string, args []any) {
	find = strings.ToLower(find)
	switch {
	case regexp.MustCompile(`eq|lt|gt|=|<|>|!`).MatchString(find):
		args = append(args, val)
		where = fmt.Sprintf("%s %s ? ", column, find)
		return
	case strings.Contains(find, "null"):
		where = fmt.Sprintf("%s %s", column, find)
		return
	case strings.Contains(find, "in"):
		if reflect.TypeOf(val).Kind() != reflect.Slice {
			return
		}
		s := reflect.ValueOf(val)
		l := s.Len()
		if l == 0 {
			return
		}
		for i := range l {
			args = append(args, s.Index(i).Interface())
		}
		where = fmt.Sprintf("%s %s (%s) ", column, find, Placeholders(l))
		return
	case strings.Contains(find, "between"):
		s := reflect.ValueOf(val)
		if s.Kind() != reflect.Slice {
			return
		}
		l := s.Len()
		if l != 2 {
			return
		}
		for i := range l {
			args = append(args, s.Index(i).Interface())
		}
		where = fmt.Sprintf("%s between  ? and ?  ", column)
		return
	case strings.Contains(find, "like"):
		if val == "%" || val == "" || val == "%%" {
			return
		}
		args = append(args, val)
		where = fmt.Sprintf("%s like ? ", column)
		return
	default:
		args = append(args, val)
		where = column + " = ? "
		return
	}
}
