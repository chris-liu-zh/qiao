package Http

import (
	"fmt"
	"testing"
)

type RolePage struct {
	Page        uint64   `json:"current" query:"current" find:"~"`
	Size        uint64   `json:"size" query:"size" find:"~"`
	Name        string   `json:"roleName" query:"roleName" find:"like" ignore:""`
	Code        string   `json:"roleCode" query:"roleCode" find:"like" ignore:""`
	Description string   `json:"description" query:"description" find:"like" ignore:""`
	Enabled     string   `json:"enabled" query:"enabled" ignore:""`
	CreateTime  []string `json:"daterange" query:"startTime,endTime" find:"between" ignore:""`
}

func TestTools(t *testing.T) {
	var rp RolePage
	parseRequestValues(&rp, "query", func(key string) string {
		return query(key)
	})
	fmt.Println(rp)
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
