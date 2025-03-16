/*
 * @Author: Chris
 * @Date: 2025-03-17 00:14:34
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-17 00:58:40
 * @Description: 请填写简介
 */
package Http

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

func ReadAll(r *http.Request, v any) error {
	// 检查 v 是否为指针
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, v); err != nil {
		return err
	}
	return nil
}
