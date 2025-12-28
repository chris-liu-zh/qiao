package DB

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
)

var (
	// rlock        sync.RWMutex
	ErrNotPtr    = errors.New("type is not reflect.Pointer")
	ErrNotStruct = errors.New("type is not reflect.Struct")
	ErrNotSlice  = errors.New("type is not reflect.Slice")
)

func (mapper *Mapper) ScanRowMap() (row map[string]any, err error) {
	// rlock.Lock()
	// defer rlock.Unlock()
	defer mapper.sqlRows.Close()
	columns, err := mapper.sqlRows.Columns()
	if err != nil {
		return
	}
	length := len(columns)
	pointer := make([]any, length)
	for i := range length {
		pointer[i] = new(any)
	}
	if !mapper.sqlRows.Next() {
		if err = mapper.sqlRows.Err(); err != nil {
			return
		}
		return nil, sql.ErrNoRows
	}

	if err = mapper.sqlRows.Scan(pointer...); err != nil {
		return
	}
	row = make(map[string]any)
	for i := range length {
		row[columns[i]] = *pointer[i].(*any)
	}
	return
}

func (mapper *Mapper) ScanRowStruct(_struct any) (err error) {
	// rlock.Lock()
	// defer rlock.Unlock()
	defer mapper.sqlRows.Close()
	columns, err := mapper.sqlRows.Columns()
	if err != nil {
		return
	}
	length := len(columns)
	pointer := make([]any, length)

	ReflectV := reflect.ValueOf(_struct)
	if ReflectV.Kind() != reflect.Pointer {
		return ErrNotPtr
	}
	elem := ReflectV.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	reflectT := reflect.TypeOf(_struct)
	if reflectT.Kind() != reflect.Pointer {
		return ErrNotPtr
	}
	for i := range length {
		pointer[i] = elem.FieldByName(reflectT.Elem().Field(i).Name).Addr().Interface()
	}

	if !mapper.sqlRows.Next() {
		if err = mapper.sqlRows.Err(); err != nil {
			return
		}
		return sql.ErrNoRows
	}
	if err = mapper.sqlRows.Scan(pointer...); err != nil {
		return
	}
	return
}

//------GetList-----------------------------------------------------------------------------------------------//

func (mapper *Mapper) ScanListMap() (list []map[string]any, err error) {
	// rlock.Lock()
	// defer rlock.Unlock()
	defer mapper.sqlRows.Close()
	columns, err := mapper.sqlRows.Columns()
	if err != nil {
		return
	}
	length := len(columns)
	pointer := make([]any, length)
	for i := range length {
		pointer[i] = new(any)
	}
	for mapper.sqlRows.Next() {
		row := make(map[string]any)
		if err = mapper.sqlRows.Scan(pointer...); err == nil {
			for i := range length {
				row[columns[i]] = *pointer[i].(*any)
			}
			list = append(list, row)
		}
	}
	return
}

func (mapper *Mapper) ScanListStruct(_struct any) (err error) {
	// rlock.Lock()
	// defer rlock.Unlock()
	defer mapper.sqlRows.Close()

	reflectT := reflect.TypeOf(_struct)
	if reflectT.Kind() != reflect.Pointer {
		return ErrNotPtr
	}
	sliceVal := reflect.Indirect(reflect.ValueOf(_struct))
	sliceItem := reflect.New(reflectT.Elem().Elem()).Elem()
	columns, err := mapper.sqlRows.Columns()
	if err != nil {
		return
	}
	length := len(columns)
	for mapper.sqlRows.Next() {
		pointer := make([]any, 0, length)
		for i := range reflectT.Elem().Elem().NumField() {
			columns := strings.Split(reflectT.Elem().Elem().Field(i).Tag.Get("db"), ";")
			if ReadOnlyField(columns) {
				fieldVal := sliceItem.FieldByName(reflectT.Elem().Elem().Field(i).Name)
				pointer = append(pointer, fieldVal.Addr().Interface())
			}
		}

		if err = mapper.sqlRows.Scan(pointer...); err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, sliceItem))
	}
	return mapper.sqlRows.Err()
}
