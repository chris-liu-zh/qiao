/*
 * @Author: Chris
 * @Date: 2025-03-05 13:24:30
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 12:08:59
 * @Description: 请填写简介
 */
package qiao

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/chris-liu-zh/qiao"
	"github.com/chris-liu-zh/qiao/DB"
)

func initdb() error {
	db := &DB.DBNew{
		Title: "test",
		Part:  "master",
		Type:  "mssql",
		Dsn:   "sqlserver://api:CF18.COM@192.168.1.217:1433?database=cf2024a&encrypt=disable&parseTime=true",
	}
	if err := db.NewDB(); err != nil {
		return qiao.Err(err)
	}
	return nil
}

func initlog() (err error) {
	logger, err := qiao.SetLog("./log/run.log", 10, 10, 180, true, -4, true, true)
	if err != nil {
		return qiao.Err(err)
	}
	qiao.SetDefaultSlog(logger)
	return
}

func initialize() (err error) {
	if err = initlog(); err != nil {
		return qiao.Err(err)
	}
	if err = initdb(); err != nil {
		return qiao.Err(err)
	}

	return
}

type Ptype struct {
	Typeid   string `json:"typeid" db:"typeid;~>"`
	UserCode string `json:"usercode" db:"usercode;~>"`
	FullName string `json:"fullname" db:"fullname;~>"`
}

func Test_Get(t *testing.T) {
	if err := initialize(); err != nil {
		slog.Error(err.Error())
		t.Fatalf("%v", err)
	}
	ptype := Ptype{}
	if err := DB.QiaoDB().Find("typeid=?", "00000").Get(&ptype); err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Println(ptype)
}

func Test_List(t *testing.T) {
	if err := initialize(); err != nil {
		t.Fatalf("%v", err)
	}
	ptypes := Ptype{}
	if err := DB.QiaoDB().Find("typeid=?", "00000").GetList(&ptypes); err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Println(ptypes)
}

func Test_Add(t *testing.T) {
	if err := initialize(); err != nil {
		t.Fatalf("%v", err)
	}
	ptype := Ptype{
		UserCode: "001",
		FullName: "张三",
	}
	if _, err := DB.QiaoDB().Add(&ptype); err != nil {
		t.Fatalf("%v", err)
	}
	Test_List(t)
}

func Test_Update(t *testing.T) {
	if err := initialize(); err != nil {
		t.Fatalf("%v", err)
	}
	ptype := Ptype{
		UserCode: "001",
		FullName: "张三",
	}
	if _, err := DB.QiaoDB().Find("typeid=?", "00000").Update(&ptype); err != nil {
		t.Fatalf("%v", err)
	}
	Test_List(t)
}
