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
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/chris-liu-zh/qiao"
)

var accessLog *log.Logger
var errorLog *log.Logger

func NewLog(logPath string, maxSize int, maxBackups int, maxAge int, compress bool) error {
	accessFile := fmt.Sprintf("%s/access_log", logPath)
	errorFile := fmt.Sprintf("%s/error_log", logPath)
	accessIo, err := qiao.NewLoggerRotate(accessFile, maxSize, maxBackups, maxAge, compress)
	if err != nil {
		return err
	}
	errorIo, err := qiao.NewLoggerRotate(errorFile, maxSize, maxBackups, maxAge, compress)
	if err != nil {
		return err
	}
	accessLog = log.New(accessIo, "[ACCESS] ", log.Ldate|log.Ltime)
	errorLog = log.New(errorIo, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

// LogAccess 记录访问日志
func LogAccess(r *http.Request, status int, bytesWritten int) {
	if accessLog == nil {
		accessLog = log.New(os.Stdout, "[ACCESS] ", log.Ldate|log.Ltime)
	}
	accessLog.Printf("%s - - \"%s %s %s\" %d %d \"%s\" \"%s\"",
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
		errorLog = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
	}
	errorLog.Printf("%s - - \"%s %s %s\" %d %d \"%s\" \"%s\" %s",
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
			log.Printf("%s - - \"%s\" %d %s", msg, file, line, runtime.FuncForPC(funcName).Name())
		}
	}
}
