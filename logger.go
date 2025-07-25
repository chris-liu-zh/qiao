/*
 * @Author: Chris
 * @Date: 2024-05-16 22:38:04
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-10 11:47:24
 * @Description: 请填写简介
 */
package qiao

import (
	"io"
	"log/slog"
	"os"
)

/*
*SetLog 设置日志记录器
LevelDebug  = -4
LevelInfo   = 0
LevelWarn   = 4
LevelError  = 8
*/

type LogOption struct {
	path       string     // 日志文件路径
	maxSize    int        // 每个日志文件的最大大小（MB）
	maxBackups int        // 保留的最大备份文件数
	maxAge     int        // 保留的最大天数
	compress   bool       // 是否压缩日志文件
	level      slog.Level // 日志级别
	outJson    bool       // 是否输出为 JSON 格式
	viewSource bool       // 是否显示调用者信息
	viewOut    bool       // 是否显示输出
}

func NewLog() *LogOption {
	return &LogOption{
		path:       "./log/run.log",
		maxSize:    10,
		maxBackups: 180,
		maxAge:     180,
		compress:   false,
		level:      slog.LevelInfo,
		outJson:    true,
		viewSource: false,
		viewOut:    false,
	}
}

func (opt *LogOption) SetFilePath(dirPath string) *LogOption {
	opt.path = dirPath + "/run.log"
	return opt
}
func (opt *LogOption) SetMaxSize(maxSize int) *LogOption {
	opt.maxSize = maxSize
	return opt
}

func (opt *LogOption) SetMaxBackups(maxBackups int) *LogOption {
	opt.maxBackups = maxBackups
	return opt
}
func (opt *LogOption) SetMaxAge(maxAge int) *LogOption {
	opt.maxAge = maxAge
	return opt
}
func (opt *LogOption) SetCompress(compress bool) *LogOption {
	opt.compress = compress
	return opt
}
func (opt *LogOption) SetLevel(level slog.Level) *LogOption {
	opt.level = level
	return opt
}
func (opt *LogOption) SetOutJson(outJson bool) *LogOption {
	opt.outJson = outJson
	return opt
}
func (opt *LogOption) SetViewSource(viewSource bool) *LogOption {
	opt.viewSource = viewSource
	return opt
}

func (opt *LogOption) SetViewOut(viewOut bool) *LogOption {
	opt.viewOut = viewOut
	return opt
}

func (opt *LogOption) SetDefault() error {
	var output io.Writer
	// 创建 NewLoggerRotate 实例
	output, err := NewLoggerRotate(opt.path, opt.maxSize, opt.maxBackups, opt.maxAge, opt.compress)
	if err != nil {
		return err
	}
	if opt.viewOut {
		// 将 slog 的输出重定向到 NewLoggerRotate 和控制台
		output = io.MultiWriter(os.Stdout, output)
	}

	logOpt := &slog.HandlerOptions{AddSource: opt.viewSource, Level: opt.level}
	var handler slog.Handler
	if opt.outJson {
		handler = slog.NewJSONHandler(output, logOpt)
	} else {
		handler = slog.NewTextHandler(output, logOpt)
	}
	// 设置默认日志记录器
	slog.SetDefault(slog.New(handler))
	return nil
}
