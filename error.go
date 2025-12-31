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
	"log/slog"
	"runtime"
)

type qiaoError struct {
	Msg      string     `json:"msg,omitempty"`
	Err      error      `json:"err,omitempty"`
	File     string     `json:"file,omitempty"`
	Line     int        `json:"line,omitempty"`
	Id       string     `json:"id,omitempty"`
	Level    slog.Level `json:"level,omitempty"`
	FuncName string     `json:"funcName,omitempty"`
	Other    any        `json:"other,omitempty"`
}

var errId string

type options func(*qiaoError)

func SetLevel(level slog.Level) options {
	return func(qe *qiaoError) {
		qe.Level = level
	}
}

func SetOther(other any) options {
	return func(qe *qiaoError) {
		qe.Other = other
	}
}

func Err(msg string, err error, opt ...options) error {
	qe := &qiaoError{
		Level: slog.LevelError,
	}
	if ok := errors.As(err, &qe); ok && qe.Id == errId {
		return err
	}

	for _, o := range opt {
		o(qe)
	}

	if funcName, file, line, ok := runtime.Caller(1); ok {
		errId = UUIDV7().String()

		if err != nil {
			slog.Log(context.Background(), qe.Level, msg, slog.String("file", file), slog.Int("line", line), slog.String("err", err.Error()), slog.Any("other", qe.Other))
		}
		qe.Err = err
		qe.Msg = msg
		qe.File = file
		qe.Line = line
		qe.Id = errId
		qe.FuncName = runtime.FuncForPC(funcName).Name()
		return qe
	}
	return nil
}

func (qe *qiaoError) Error() string {
	return qe.Err.Error()
}

func (qe *qiaoError) Unwrap() error {
	return qe.Err
}
