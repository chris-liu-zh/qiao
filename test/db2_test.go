package qiao

import (
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/DB"
)

type Project struct {
	Id          *int64  `db:"id;~>" json:"id"`
	Cid         *int64  `db:"cid" json:"cid"`
	Typeid      *int64  `db:"typeid" json:"typeId"`
	Name        *string `db:"name" json:"name"`
	Passwd      *string `db:"passwd" json:"passwd"`
	StartTime   *string `db:"startTime" json:"startTime"`
	EndTime     *string `db:"endTime" json:"endTime"`
	StopTime    *string `db:"stopTime" json:"stopTime"`
	CountTime   *uint64 `db:"countTime" json:"countTime"`
	Description *string `db:"description" json:"description"`
	Enabled     *bool   `db:"enabled" json:"enabled"`
	Points      *uint64 `db:"points" json:"points"`
}

func Test_db(t *testing.T) {
	if err := initdb(); err != nil {
		t.Fatalf("%v", err)
	}

	for range 100 {
		go func(t *testing.T) {
			query(t)
		}(t)

	}

	time.Sleep(10 * time.Second)
}

func query(t *testing.T) {
	var p []Project
	if err := DB.QiaoDB().Order("startTime").GetList(&p); err != nil {
		t.Fatalf("%v", err)
		return
	}
}
