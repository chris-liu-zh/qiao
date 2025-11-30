package Http

import (
	"fmt"
	"testing"
)

type RolePage struct {
	Page        *uint64   `json:"current" query:"current" find:"~"`
	Size        *uint64   `json:"size" query:"size" find:"~"`
	Name        *string   `json:"roleName" query:"roleName" find:"like" ignore:""`
	Code        *string   `json:"roleCode" query:"roleCode" find:"like" ignore:""`
	Description string    `json:"description" query:"description" find:"like" ignore:""`
	Enabled     *bool     `json:"enabled" query:"enabled" ignore:""`
	CreateTime  *[]string `json:"daterange" query:"startTime,endTime" find:"between" ignore:""`
}

// 测试结构体，包含各种类型
type TestStruct struct {
	StringField  string   `query:"string_field"`
	IntField     int      `query:"int_field"`
	Int64Field   int64    `query:"int64_field"`
	UintField    uint     `query:"uint_field"`
	Uint64Field  uint64   `query:"uint64_field"`
	BoolField    bool     `query:"bool_field"`
	FloatField   float64  `query:"float_field"`
	StringPtr    *string  `query:"string_ptr"`
	IntPtr       *int     `query:"int_ptr"`
	BoolPtr      *bool    `query:"bool_ptr"`
	SliceField   []string `query:"slice_field"`
	IgnoredField string   `query:"~"`
	DefaultTag   string   // 使用默认标签
}

func TestTools(t *testing.T) {
	t.Run("RolePage测试", func(t *testing.T) {
		var rp RolePage
		if err := parseRequestValues(&rp, "query", func(key string) string {
			return query(key)
		}); err != nil {
			t.Fatal(err)
		}

		// 验证指针字段
		if rp.Page == nil || *rp.Page != 1 {
			t.Errorf("Page expected 1, got %v", rp.Page)
		}
		if rp.Size == nil || *rp.Size != 10 {
			t.Errorf("Size expected 10, got %v", rp.Size)
		}
		if rp.Name == nil || *rp.Name != "admin" {
			t.Errorf("Name expected 'admin', got %v", rp.Name)
		}
		if rp.Enabled == nil || !*rp.Enabled {
			t.Errorf("Enabled expected true, got %v", rp.Enabled)
		}

		fmt.Printf("RolePage测试通过: %+v\n", rp)
	})

	t.Run("全面类型测试", func(t *testing.T) {
		var ts TestStruct
		if err := parseRequestValues(&ts, "query", func(key string) string {
			return testQuery(key)
		}); err != nil {
			t.Fatal(err)
		}

		// 验证基本类型
		if ts.StringField != "test" {
			t.Errorf("StringField expected 'test', got %s", ts.StringField)
		}
		if ts.IntField != 42 {
			t.Errorf("IntField expected 42, got %d", ts.IntField)
		}
		if ts.BoolField != true {
			t.Errorf("BoolField expected true, got %v", ts.BoolField)
		}

		// 验证指针类型
		if ts.StringPtr == nil || *ts.StringPtr != "pointer" {
			t.Errorf("StringPtr expected 'pointer', got %v", ts.StringPtr)
		}
		if ts.IntPtr == nil || *ts.IntPtr != 100 {
			t.Errorf("IntPtr expected 100, got %v", ts.IntPtr)
		}

		// 验证切片
		if len(ts.SliceField) != 2 || ts.SliceField[0] != "value1" || ts.SliceField[1] != "value2" {
			t.Errorf("SliceField expected [value1 value2], got %v", ts.SliceField)
		}

		// 验证默认标签（使用蛇形命名）
		if ts.DefaultTag != "default_value" {
			t.Errorf("DefaultTag expected 'default_value', got %s", ts.DefaultTag)
		}

		fmt.Printf("全面类型测试通过: %+v\n", ts)
	})

	t.Run("空值处理测试", func(t *testing.T) {
		var ts TestStruct
		if err := parseRequestValues(&ts, "query", func(key string) string {
			return "" // 返回空值
		}); err != nil {
			t.Fatal(err)
		}

		// 验证指针字段为nil
		if ts.StringPtr != nil {
			t.Errorf("StringPtr expected nil for empty value, got %v", ts.StringPtr)
		}
		if ts.IntPtr != nil {
			t.Errorf("IntPtr expected nil for empty value, got %v", ts.IntPtr)
		}

		fmt.Println("空值处理测试通过")
	})

	t.Run("错误处理测试", func(t *testing.T) {
		// 测试非指针参数
		var ts TestStruct
		err := parseRequestValues(ts, "query", func(key string) string {
			return "test"
		})
		if err == nil {
			t.Error("Expected error for non-pointer parameter")
		}

		// 测试nil指针
		err = parseRequestValues(nil, "query", func(key string) string {
			return "test"
		})
		if err == nil {
			t.Error("Expected error for nil pointer")
		}

		fmt.Println("错误处理测试通过")
	})
}

func query(key string) string {
	switch key {
	case "current":
		return "1"
	case "size":
		return "10"
	case "roleName":
		return "admin"
	case "roleCode":
		return "admin"
	case "description":
		return "system administrator"
	case "enabled":
		return "true"
	case "startTime":
		return "2023-01-01"
	case "endTime":
		return "2023-12-31"
	}
	return key
}

func testQuery(key string) string {
	switch key {
	case "string_field":
		return "test"
	case "int_field":
		return "42"
	case "int64_field":
		return "64"
	case "uint_field":
		return "32"
	case "uint64_field":
		return "128"
	case "bool_field":
		return "true"
	case "float_field":
		return "3.14"
	case "string_ptr":
		return "pointer"
	case "int_ptr":
		return "100"
	case "bool_ptr":
		return "true"
	case "slice_field":
		return "value1,value2"
	case "default_tag":
		return "default_value"
	}
	return ""
}
