/*
 * @Author: Chris
 * @Date: 2024-06-12 17:53:09
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 12:09:32
 * @Description: 请填写简介
 */
package Http

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/chris-liu-zh/qiao"
)

var accessLog *log.Logger
var errorLog *log.Logger

func NewLog(dirPath string, maxSize int, maxBackups int, maxAge int, compress, viewOut bool) (err error) {
	accessFile := fmt.Sprintf("%s/access_log", dirPath)
	errorFile := fmt.Sprintf("%s/error_log", dirPath)
	var accIo, errIo io.Writer
	if accIo, err = qiao.NewLoggerRotate(accessFile, maxSize, maxBackups, maxAge, compress); err != nil {
		return err
	}
	if errIo, err = qiao.NewLoggerRotate(errorFile, maxSize, maxBackups, maxAge, compress); err != nil {
		return err
	}
	if viewOut {
		accIo = io.MultiWriter(os.Stdout, accIo)
		errIo = io.MultiWriter(os.Stdout, errIo)
	}
	accessLog = log.New(accIo, "[ACCESS] ", log.Ldate|log.Ltime)
	errorLog = log.New(errIo, "[ERROR] ", log.Ldate|log.Ltime)
	return nil
}

// LogAccess 记录访问日志
func LogAccess(r *http.Request, status int, bytesWritten int) {
	if accessLog == nil {
		return
	}
	accessLog.Printf("%s -- \"%s %s %s\" %d %d \"%s\" \"%s\"",
		r.RemoteAddr,
		r.Method,
		r.URL.Path,
		r.Proto,
		status,
		bytesWritten,
		r.Referer(),
		r.UserAgent(),
	)
}

// LogError 记录错误日志
func LogError(r *http.Request, status int, bytesWritten int, msg string) {
	if errorLog == nil {
		return
	}
	errorLog.Printf("%s -- \"%s %s %s\" %d %d \"%s\" \"%s\" %s",
		r.RemoteAddr,
		r.Method,
		r.URL.Path,
		r.Proto,
		status,
		bytesWritten,
		r.Referer(),
		r.UserAgent(),
		msg,
	)
}

func LogDebug(msg string, skip int) {
	if msg != "" {
		if funcName, file, line, ok := runtime.Caller(skip); ok {
			filepath := slog.String("file", file)
			lint := slog.Int("line", line)
			fnName := slog.String("func", runtime.FuncForPC(funcName).Name())
			slog.Debug(msg, filepath, lint, fnName)
		}
	}
}
