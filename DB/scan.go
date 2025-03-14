package DB

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"sync"
)

var (
	columns      []string
	pointer      []any
	rlock        sync.RWMutex
	ErrNotPtr    = errors.New("type is not reflect.Ptr")
	ErrNotStruct = errors.New("type is not reflect.Struct")
	ErrNotSlice  = errors.New("type is not reflect.Slice")
)

func ScanRowMap(Rows *sql.Rows) (row map[string]any, err error) {
	rlock.Lock()
	defer rlock.Unlock()
	defer Rows.Close()
	if columns, err = Rows.Columns(); err != nil {
		return
	}
	length := len(columns)
	pointer = make([]any, length)
	for i := range length {
		pointer[i] = new(any)
	}
	if !Rows.Next() {
		if err = Rows.Err(); err != nil {
			return
		}
		return nil, sql.ErrNoRows
	}

	if err = Rows.Scan(pointer...); err != nil {
		return
	}
	row = make(map[string]any)
	for i := range length {
		row[columns[i]] = *pointer[i].(*any)
	}

	return
}

func ScanRowStruct(Rows *sql.Rows, _struct any) (err error) {
	rlock.Lock()
	defer rlock.Unlock()
	defer Rows.Close()
	if columns, err = Rows.Columns(); err != nil {
		return
	}
	length := len(columns)
	pointer = make([]any, length)

	ReflectV := reflect.ValueOf(_struct)
	if ReflectV.Kind() != reflect.Ptr {
		return ErrNotPtr
	}
	elem := ReflectV.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	reflectT := reflect.TypeOf(_struct)
	if reflectT.Kind() != reflect.Ptr {
		return ErrNotPtr
	}
	for i := range length {
		pointer[i] = elem.FieldByName(reflectT.Elem().Field(i).Name).Addr().Interface()
	}

	if !Rows.Next() {
		if err = Rows.Err(); err != nil {
			return
		}
		return sql.ErrNoRows
	}
	if err = Rows.Scan(pointer...); err != nil {
		return
	}
	return
}

//------GetList-----------------------------------------------------------------------------------------------//

func ScanListMap(Rows *sql.Rows) (list []map[string]any, err error) {
	rlock.Lock()
	defer rlock.Unlock()
	defer Rows.Close()
	if columns, err = Rows.Columns(); err != nil {
		return
	}
	length := len(columns)
	pointer = make([]any, length)
	for i := range length {
		var val any
		pointer[i] = &val
	}
	for Rows.Next() {
		row := make(map[string]any)
		if err = Rows.Scan(pointer...); err == nil {
			for i := range length {
				row[columns[i]] = *pointer[i].(*any)
			}
			list = append(list, row)
		}
	}
	return
}

func ScanListStruct(Rows *sql.Rows, _struct any) (err error) {
	rlock.Lock()
	defer rlock.Unlock()
	defer Rows.Close()

	reflectT := reflect.TypeOf(_struct)
	if reflectT.Kind() != reflect.Ptr {
		return ErrNotPtr
	}
	sliceVal := reflect.Indirect(reflect.ValueOf(_struct))
	sliceItem := reflect.New(reflectT.Elem().Elem()).Elem()
	columns, err := Rows.Columns()
	if err != nil {
		return
	}
	length := len(columns)
	for Rows.Next() {
		pointer = make([]any, 0, length)
		for i := range reflectT.Elem().Elem().NumField() {
			columns := strings.Split(reflectT.Elem().Elem().Field(i).Tag.Get("db"), ";")
			if ReadOnlyField(columns) {
				fieldVal := sliceItem.FieldByName(reflectT.Elem().Elem().Field(i).Name)
				pointer = append(pointer, fieldVal.Addr().Interface())
			}
		}

		if err = Rows.Scan(pointer...); err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, sliceItem))
	}
	return Rows.Err()
}
