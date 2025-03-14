/*
 * @Author: Chris
 * @Date: 2023-06-14 22:27:14
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-09 00:23:28
 * @Description: 请填写简介
 */
package qiao

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"
)

type qiaoError struct {
	Err      error  `json:"err"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	FuncName string `json:"funcName"`
	Other    any    `json:"other"`
}

var errid string

func Err(err error, other ...any) error {
	if id := extractID(err.Error()); id != "" && id == errid {
		return err
	}
	if err != nil {
		if funcName, file, line, ok := runtime.Caller(1); ok {
			return &qiaoError{
				Err:      err,
				File:     file,
				Line:     line,
				Other:    other,
				FuncName: runtime.FuncForPC(funcName).Name(),
			}
		}
	}
	return nil
}

func extractID(input string) string {
	startIndex := strings.Index(input, "id=")
	if startIndex == -1 {
		return ""
	}
	startIndex += len("id=")
	endIndex := strings.Index(input[startIndex:], "  ")
	if endIndex == -1 {
		return input[startIndex:]
	}
	endIndex += startIndex

	return input[startIndex:endIndex]
}

func generateID() string {
	timestamp := time.Now().UnixMilli()
	rand.New(rand.NewSource(timestamp))
	randomNum := rand.Intn(9000) + 1000
	id := fmt.Sprintf("%d%d", timestamp, randomNum)
	return id
}

func (e *qiaoError) Error() string {
	errid = generateID()
	return fmt.Sprintf(`file=%s:%d func=%s err=%s id=%s`, e.File, e.Line, e.FuncName, e.Err, errid)
}
