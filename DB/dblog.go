/*
 * @Author: Chris
 * @Date: 2024-05-16 22:38:04
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 12:09:41
 * @Description: 请填写简介
 */
package DB

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"

	"github.com/chris-liu-zh/qiao"
)

type sqlLog struct {
	Title   string `json:"title"`
	Message string `json:"msg"`
	Sqlstr  string `json:"sql"`
	Args    []any  `json:"args"`
}

var (
	loggerPre  = make(map[string]*slog.Logger)
	loggerList = []string{"DEBUG", "INFO", "WARNING", "ERROR", "CUSTOM"}
)

// LogEntry 定义 JSON 日志结构
type LogEntry struct {
	Time    string         `json:"time"`            // 时间
	Level   string         `json:"level"`           // 日志级别
	Message string         `json:"message"`         // 日志消息
	File    string         `json:"file,omitempty"`  // 文件名
	Line    int            `json:"line,omitempty"`  // 行号
	Attrs   map[string]any `json:"attrs,omitempty"` // 动态属性
}

// CustomHandler 自定义 slog.Handler
type CustomHandler struct {
	output io.Writer
	level  slog.Level
	IsJson bool
}

func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	// 检查日志级别是否满足条件
	return level >= h.level
}

func (h *CustomHandler) Handle(_ context.Context, r slog.Record) (err error) {
	// 获取调用栈信息
	_, file, line, ok := runtime.Caller(4) // 调整调用栈深度
	if !ok {
		file = "unknown"
		line = 0
	}
	var logbyte []byte
	if h.IsJson {
		logmsg := LogEntry{
			Message: r.Message,
			Level:   r.Level.String(),
			Time:    r.Time.Format("2006-01-02 15:04:05"),
			File:    file,
			Line:    line,
			Attrs:   make(map[string]any),
		}
		// 添加属性（key 和 value）
		r.Attrs(func(attr slog.Attr) bool {
			logmsg.Attrs[attr.Key] = attr.Value.Any()
			return true // 继续遍历
		})
		if logbyte, err = json.Marshal(logmsg); err != nil {
			return err
		}
		_, err = h.output.Write(append(logbyte, '\n'))
		return
	}

	// 自定义日志格式
	msg := r.Message
	level := r.Level.String()
	timeStr := r.Time.Format("2006-01-02 15:04:05")
	var logString string
	r.Attrs(func(attr slog.Attr) bool {
		logString += fmt.Sprintf(` %s="%v"`, attr.Key, attr.Value)
		return true // 继续遍历
	})
	logbyte = fmt.Appendf(nil, `[%s] time="%s" Source="%s:%d" msg="%s" %s`, level, timeStr, file, line, msg, logString)
	_, err = h.output.Write(append(logbyte, '\n'))
	return
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// 返回一个新的 Handler，包含附加的属性
	// 这里简单返回自身，实际可以根据需要实现
	return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	// 返回一个新的 Handler，用于分组日志属性
	// 这里简单返回自身，实际可以根据需要实现
	return h
}

func getLog(filename string, maxSize int, maxBackups int, maxAge int, compress bool, Level slog.Level, isJson bool) (*slog.Logger, error) {
	loggerRotate, err := qiao.NewLoggerRotate(filename, maxSize, maxBackups, maxAge, compress)
	if err != nil {
		return nil, err
	}
	multiWriter := io.MultiWriter(os.Stdout, loggerRotate)
	customHandler := &CustomHandler{
		output: multiWriter,
		level:  Level,
		IsJson: isJson,
	}
	return slog.New(customHandler), nil
}

// SetDBLog 设置日志记录器
// LevelDebug  = -4
// LevelInfo   = 0
// LevelWarn   = 4
// LevelError  = 8
func SetDBLog(path string, maxSize int, maxBackups int, maxAge int, compress bool, level slog.Level) error {
	for _, logType := range loggerList {
		filename := fmt.Sprintf("%s/%s.log", path, logType)
		if logType == "CUSTOM" {
			level = slog.LevelDebug
		}
		newSlog, err := getLog(filename, maxSize, maxBackups, maxAge, compress, level, false)
		if err != nil {
			return err
		}
		loggerPre[logType] = newSlog
	}
	return nil
}

func (info *sqlLog) logDEBUG() {
	if loggerPre["DEBUG"] == nil {
		info.formatLog("DEBUG")
		return
	}
	loggerPre["DEBUG"].Debug(info.Message, "sql", info.Sqlstr, "args", info.Args, "DBTitle", info.Title)
}

func (info *sqlLog) logINFO() {
	if loggerPre["INFO"] == nil {
		info.formatLog("INFO")
		return
	}
	loggerPre["INFO"].Info(info.Message, "sql", info.Sqlstr, "args", info.Args, "DBTitle", info.Title)
}

func (info *sqlLog) logWARNING() {
	if loggerPre["WARNING"] == nil {
		info.formatLog("WARNING")
		return
	}
	loggerPre["WARNING"].Warn(info.Message, "sql", info.Sqlstr, "args", info.Args, "DBTitle", info.Title)
}

func (info *sqlLog) logERROR() {
	if loggerPre["ERROR"] == nil {
		info.formatLog("ERROR")
		return
	}
	loggerPre["ERROR"].Error(info.Message, "sql", info.Sqlstr, "args", info.Args, "DBTitle", info.Title)
}

func (info *sqlLog) formatLog(types string) {
	logf := log.New(os.Stdout, "[DB]["+types+"]", log.Ldate|log.Ltime)
	logf.Printf("Messag=%s; sql=%s; args=%v; DBTitle=%s \n", info.Message, info.Sqlstr, info.Args, info.Title)
}

func (mapper *Mapper) debug(msg string) {
	if mapper.Complete.Debug {
		if loggerPre["CUSTOM"] == nil {
			logf := log.New(os.Stdout, "[DB][CUSTOM]", log.Ldate|log.Ltime)
			logf.Printf("Messag=%s; sql=%s; args=%v;\n", msg, mapper.Complete.Sql, mapper.Complete.Args)
			return
		}
		loggerPre["CUSTOM"].Debug(msg, "Sqlstr", mapper.Complete.Sql, "Args", mapper.Complete.Args)
	}
}
