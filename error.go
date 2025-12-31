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
	"encoding/json"
	"errors"
	"log/slog"
	"runtime"
)

type qiaoError struct {
	Msg      string `json:"msg,omitempty"`
	Err      error  `json:"err,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Id       string `json:"id,omitempty"`
	FuncName string `json:"funcName,omitempty"`
	Other    any    `json:"other,omitempty"`
}

var errId string

func Err(msg string, err error, level slog.Level, other ...any) error {
	var qe *qiaoError
	if ok := errors.As(err, &qe); ok && qe.Id == errId {
		return err
	}

	if funcName, file, line, ok := runtime.Caller(1); ok {
		errId = UUIDV7().String()
		if err != nil {
			slog.Log(context.Background(), slog.LevelError, msg, slog.String("file", file), slog.Int("line", line), slog.String("err", err.Error()), slog.Any("other", other))
		}

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

func (e *qiaoError) Error() string {
	jsonErr, _ := json.Marshal(e)
	return string(jsonErr)
}

func (e *qiaoError) Unwrap() error {
	return e.Err
}
