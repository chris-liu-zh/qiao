package Http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/chris-liu-zh/qiao"
)

func BodyTOStruct(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func parseRequestValues(v any, tagKey string, getter func(key string) string) error {
	// 检查 v 必须是非空结构体指针
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("v must be a non-nil struct pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("v must point to a struct")
	}

	// 遍历结构体字段
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := elem.Type().Field(i)

		// 获取字段对应的标签
		tag := fieldType.Tag.Get(tagKey)
		if tag == "" {
			tag = qiao.CamelCaseToUdnderscore(fieldType.Name)
		}
		// 跳过忽略标签（如 `query:"~"`）
		if tag == "~" {
			continue
		}

		// 2. 从请求中获取值
		val := getter(tag)
		if val == "" {
			continue // 空值不覆盖结构体默认值
		}

		// 3. 检查字段是否可设置
		if !field.CanSet() {
			continue // 跳过未导出字段或不可设置的字段
		}

		// 4. 类型转换并设置字段值
		if err := setFieldValue(field, val); err != nil {
			return fmt.Errorf("field %q: %w", fieldType.Name, err)
		}
	}

	return nil
}

// GetQuery 解析 URL 查询参数到结构体
// 支持标签 `query:"name"` 指定参数名，默认使用蛇形命名（如 UserName -> user_name）
// 忽略标签 `query:"~"` 的字段
func GetQuery(r *http.Request, v any) error {
	return parseRequestValues(v, "query", func(key string) string {
		return r.URL.Query().Get(key)
	})
}

// PathValue 解析 URL 路径参数到结构体
// 支持标签 `path:"name"` 指定参数名，默认使用蛇形命名（如 UserID -> user_id）
// 忽略标签 `path:"~"` 的字段
func PathValue(r *http.Request, v any) error {
	return parseRequestValues(v, "path", func(key string) string {
		return r.PathValue(key)
	})
}

// setFieldValue 将字符串值转换为对应字段类型并设置
func setFieldValue(field reflect.Value, val string) error {
	switch field.Kind() {
	// 字符串类型
	case reflect.String:
		field.SetString(val)

	// 有符号整数类型
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv, err := strconv.ParseInt(val, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", val, err)
		}
		field.SetInt(iv)

	// 无符号整数类型
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		uv, err := strconv.ParseUint(val, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid uint value %q: %w", val, err)
		}
		field.SetUint(uv)

	// 布尔类型
	case reflect.Bool:
		bv, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid bool value %q: %w", val, err)
		}
		field.SetBool(bv)

	// 浮点类型
	case reflect.Float32, reflect.Float64:
		fv, err := strconv.ParseFloat(val, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", val, err)
		}
		field.SetFloat(fv)

	// 复数类型
	case reflect.Complex64, reflect.Complex128:
		cv, err := strconv.ParseComplex(val, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid complex value %q: %w", val, err)
		}
		field.SetComplex(cv)
	default:
		return fmt.Errorf("unsupported field type %s", field.Kind())
	}
	return nil
}
