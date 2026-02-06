/*
 * @Author: Chris
 * @Date: 2023-06-14 22:27:14
 * @LastEditors: Strong
 * @LastEditTime: 2025-03-22 17:26:58
 * @Description: 请填写简介
 */
package qiao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/chris-liu-zh/qiao/tools"
)

type qiaoError struct {
	Msg      string `json:"msg,omitempty"`
	Err      error  `json:"err,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Id       string `json:"id,omitempty"`
	FuncName string `json:"funcName,omitempty"`
	Other    any    `json:"other,omitempty"`
	level    slog.Level
	printLog bool
}

type options func(*qiaoError)

func SetLevel(level slog.Level) options {
	return func(qe *qiaoError) {
		qe.level = level
	}
}

func SetOther(other any) options {
	return func(qe *qiaoError) {
		qe.Other = other
	}
}

func SetPrintLog(printLog bool) options {
	return func(qe *qiaoError) {
		qe.printLog = printLog
	}
}

func AsErr(err error) *qiaoError {
	if err == nil {
		return nil
	}
	var qe *qiaoError
	if ok := errors.As(err, &qe); ok {
		return qe
	}
	return nil
}

func Err(msg string, err error, opt ...options) error {
	if err == nil {
		return err
	}
	qe := &qiaoError{
		level:    slog.LevelError,
		printLog: true,
	}
	if ok := errors.As(err, &qe); ok {
		return err
	}

	for _, o := range opt {
		o(qe)
	}

	if funcName, file, line, ok := runtime.Caller(1); ok {
		errId := tools.UUIDV7().String()
		if qe.printLog {
			go slog.Log(
				context.Background(), qe.level, msg,
				slog.String("id", errId),
				slog.String("file", fmt.Sprintf("%s:%d", file, line)),
				slog.String("err", err.Error()),
				slog.Any("other", qe.Other),
			)
		}

		qe.Err = err
		qe.Msg = msg
		qe.File = file
		qe.Line = line
		qe.Id = errId
		qe.FuncName = runtime.FuncForPC(funcName).Name()
		return qe
	}
	return qe
}

func (qe *qiaoError) Error() string {
	return qe.Err.Error()
}

func (qe *qiaoError) Unwrap() error {
	return qe.Err
}
