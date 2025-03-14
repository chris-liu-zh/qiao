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

func SetLog(filename string, maxSize int, maxBackups int, maxAge int, compress bool, Level slog.Level, OutJson, viewSource bool) (*slog.Logger, error) {
	// 创建 NewLoggerRotate 实例
	logger, err := NewLoggerRotate(filename, maxSize, maxBackups, maxAge, compress)
	if err != nil {
		return nil, err
	}

	// 将 slog 的输出重定向到 NewLoggerRotate 和控制台
	multiWriter := io.MultiWriter(os.Stdout, logger)
	opt := &slog.HandlerOptions{AddSource: viewSource, Level: Level}
	var handler slog.Handler
	if OutJson {
		handler = slog.NewJSONHandler(multiWriter, opt)
	} else {
		handler = slog.NewTextHandler(multiWriter, opt)
	}
	// 设置默认日志记录器
	return slog.New(handler), nil
}

func SetDefaultSlog(s *slog.Logger) {
	slog.SetDefault(s)
}
