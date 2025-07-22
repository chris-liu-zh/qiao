/*
 * @Author: Chris
 * @Date: 2023-06-14 22:27:14
 * @LastEditors: Strong
 * @LastEditTime: 2025-03-22 17:26:58
 * @Description: 请填写简介
 */
package qiao

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

type qiaoError struct {
	Msg      string `json:"msg"`
	Err      error  `json:"err"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Id       int64  `json:"id"`
	FuncName string `json:"funcName"`
	Other    any    `json:"other"`
}

var errId int64 = 1

func Err(msg string, err error, other ...any) error {
	var qe *qiaoError
	if ok := errors.As(err, &qe); ok && qe.Id == errId {
		return err
	}
	if funcName, file, line, ok := runtime.Caller(1); ok {
		errId = newID()
		return &qiaoError{
			Msg:      msg,
			Err:      err,
			File:     file,
			Line:     line,
			Other:    other,
			Id:       errId,
			FuncName: runtime.FuncForPC(funcName).Name(),
		}
	}
	return nil
}

func newID() int64 {
	timestamp := time.Now().UnixMilli()
	rand.New(rand.NewSource(timestamp))
	randomNum := rand.Intn(9000) + 1000
	id, _ := strconv.ParseInt(fmt.Sprintf("%d%d", timestamp, randomNum), 10, 64)
	return id
}

func (e *qiaoError) Error() string {
	jsonErr, _ := json.Marshal(e)
	return string(jsonErr)
}

func (e *qiaoError) Unwrap() error {
	return e.Err
}
