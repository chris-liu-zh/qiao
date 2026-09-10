package Http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/chris-liu-zh/qiao/tools"
)

func BodyTOStruct(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// GetQuery 解析 URL 查询参数到结构体
// 支持标签 `query:"name"` 指定参数名，默认使用蛇形命名（如 UserName -> user_name）
// 忽略标签 `query:"~"` 的字段
// 支持多同名参数数组：is_sync[]=2&is_sync[]=0，支持 *[]int 指针切片
func GetQuery(r *http.Request, v any) error {
	queryVals := r.URL.Query()
	return parseRequestValues(v, "query", func(key string) []string {
		return queryVals[key]
	})
}

// PathValue 解析 URL 路径参数到结构体
// 支持标签 `path:"name"` 指定参数名，默认使用蛇形命名（如 UserID -> user_id）
// 忽略标签 `path:"~"` 的字段
func PathValue(r *http.Request, v any) error {
	return parseRequestValues(v, "path", func(key string) []string {
		val := r.PathValue(key)
		if val == "" {
			return nil
		}
		return []string{val}
	})
}

// parseRequestValues v必须是非空结构体指针
// getter: 根据key返回字符串切片，适配单值/多值参数
func parseRequestValues(v any, tagKey string, getter func(key string) []string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("v must be a non-nil struct pointer")
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("v must point to a struct")
	}

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := elem.Type().Field(i)

		tag := fieldType.Tag.Get(tagKey)
		if tag == "" {
			tag = tools.CamelCaseToUdnderscore(fieldType.Name)
		}
		if tag == "~" {
			continue
		}
		if !field.CanSet() {
			continue
		}

		vals := getter(tag)
		if err := setFieldValue(field, vals); err != nil {
			return fmt.Errorf("field %q: %w", fieldType.Name, err)
		}
	}
	return nil
}

// setFieldValue 将字符串切片转换为对应字段类型并设置
// vals: 所有同名参数；单值类型取vals[0]，slice类型遍历全部vals
// 支持普通类型、指针类型、*[]int指针切片
func setFieldValue(field reflect.Value, vals []string) error {
	// 空输入
	if len(vals) == 0 {
		switch field.Kind() {
		case reflect.String:
			field.SetString("")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(0)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			field.SetUint(0)
		case reflect.Bool:
			field.SetBool(false)
		case reflect.Float32, reflect.Float64:
			field.SetFloat(0)
		case reflect.Complex64, reflect.Complex128:
			field.SetComplex(0)
		case reflect.Slice:
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		case reflect.Pointer:
			// 指针类型无参数，置nil
			field.Set(reflect.Zero(field.Type()))
		}
		return nil
	}

	// 取第一个值用于单值类型解析
	firstVal := vals[0]

	switch field.Kind() {
	case reflect.String:
		field.SetString(firstVal)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv, err := strconv.ParseInt(firstVal, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", firstVal, err)
		}
		field.SetInt(iv)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		uv, err := strconv.ParseUint(firstVal, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid uint value %q: %w", firstVal, err)
		}
		field.SetUint(uv)

	case reflect.Bool:
		bv, err := strconv.ParseBool(firstVal)
		if err != nil {
			return fmt.Errorf("invalid bool value %q: %w", firstVal, err)
		}
		field.SetBool(bv)

	case reflect.Float32, reflect.Float64:
		fv, err := strconv.ParseFloat(firstVal, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", firstVal, err)
		}
		field.SetFloat(fv)

	case reflect.Complex64, reflect.Complex128:
		cv, err := strconv.ParseComplex(firstVal, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid complex value %q: %w", firstVal, err)
		}
		field.SetComplex(cv)

	case reflect.Slice:
		sliceType := field.Type()
		elemType := sliceType.Elem()
		newSlice := reflect.MakeSlice(sliceType, 0, len(vals))

		for _, s := range vals {
			elemVal := reflect.New(elemType).Elem()
			switch elemType.Kind() {
			case reflect.String:
				elemVal.SetString(s)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				iv, err := strconv.ParseInt(s, 10, elemType.Bits())
				if err != nil {
					return fmt.Errorf("parse slice int %q: %w", s, err)
				}
				elemVal.SetInt(iv)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				uv, err := strconv.ParseUint(s, 10, elemType.Bits())
				if err != nil {
					return fmt.Errorf("parse slice uint %q: %w", s, err)
				}
				elemVal.SetUint(uv)
			case reflect.Bool:
				bv, err := strconv.ParseBool(s)
				if err != nil {
					return fmt.Errorf("parse slice bool %q: %w", s, err)
				}
				elemVal.SetBool(bv)
			case reflect.Float32, reflect.Float64:
				fv, err := strconv.ParseFloat(s, elemType.Bits())
				if err != nil {
					return fmt.Errorf("parse slice float %q: %w", s, err)
				}
				elemVal.SetFloat(fv)
			default:
				return fmt.Errorf("unsupported slice element type %s", elemType.Kind())
			}
			newSlice = reflect.Append(newSlice, elemVal)
		}
		field.Set(newSlice)

	case reflect.Pointer:
		elemType := field.Type().Elem()
		newPtr := reflect.New(elemType)
		elemField := newPtr.Elem()
		if err := setFieldValue(elemField, vals); err != nil {
			return fmt.Errorf("pointer field: %w", err)
		}
		field.Set(newPtr)

	default:
		return fmt.Errorf("unsupported field type %s", field.Kind())
	}
	return nil
}
