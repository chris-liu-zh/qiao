package qiao

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type QiaoLogger struct {
	filename   string     // 日志文件路径
	maxSize    int        // 单个日志文件的最大大小（MB）
	maxBackups int        // 保留的旧日志文件数量，-1代表不限制
	maxAge     int        // 保留旧日志文件的最大天数
	compress   bool       // 是否压缩旧日志文件
	file       *os.File   // 当前日志文件
	mu         sync.Mutex // 互斥锁，确保并发安全
}

// NewLoggerRotate 创建一个新的 QiaoLogger 实例
func NewLoggerRotate(filename string, maxSize, maxBackups, maxAge int, compress bool) (*QiaoLogger, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &QiaoLogger{
		filename:   filename,
		maxSize:    maxSize,
		maxBackups: maxBackups,
		maxAge:     maxAge,
		compress:   compress,
	}, nil
}

// Write 实现 io.Writer 接口，用于写入日志
func (l *QiaoLogger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 检查当前日志文件大小
	fileInfo, err := os.Stat(l.filename)
	if err == nil && fileInfo.Size() > int64(l.maxSize*1024*1024) {
		// 触发日志切割
		if err = l.rotate(); err != nil {
			return 0, err
		}
	}

	// 打开日志文件（如果未打开）
	if l.file == nil {
		if err := l.openFile(); err != nil {
			return 0, err
		}
	}

	// 写入日志
	return l.file.Write(p)
}

// openFile 打开日志文件
func (l *QiaoLogger) openFile() error {
	file, err := os.OpenFile(l.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	l.file = file
	return nil
}

// rotate 实现日志切割和轮转
func (l *QiaoLogger) rotate() error {
	// 关闭当前日志文件
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return err
		}
		l.file = nil
	}

	// 重命名当前日志文件
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupFilename := fmt.Sprintf("%s.%s", l.filename, timestamp)
	if err := os.Rename(l.filename, backupFilename); err != nil {
		return err
	}

	// 如果需要压缩，压缩旧日志文件
	if l.compress {
		if err := compressFile(backupFilename); err != nil {
			return err
		}
		backupFilename += ".gz" // 压缩后的文件名
	}

	// 打开新的日志文件
	if err := l.openFile(); err != nil {
		return err
	}

	// 清理旧日志文件
	l.cleanupOldLogs()
	return nil
}

// 清理过期的旧日志文件
func (l *QiaoLogger) cleanupOldLogs() {
	files, err := os.ReadDir(filepath.Dir(l.filename))
	if err != nil {
		return
	}

	// 过滤出符合条件的旧日志文件（仅 .gz 文件）
	var backups []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// 匹配文件名格式：<filename>.<timestamp>.gz
		if matched, _ := filepath.Match(l.filename+".*.gz", file.Name()); matched {
			backups = append(backups, file.Name())
		}
	}

	// 使用 slices.Sort 进行排序
	slices.Sort(backups)

	// 使用 range 遍历并删除超出数量的旧日志文件
	if len(backups) > l.maxBackups {
		for _, file := range backups[:len(backups)-l.maxBackups] {
			os.Remove(file)
		}
	}
}

// compressFile 压缩日志文件
func compressFile(filename string) error {
	// 打开原始日志文件
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建压缩文件
	gzFile, err := os.Create(filename + ".gz")
	if err != nil {
		return err
	}
	defer gzFile.Close()

	// 使用 gzip 压缩
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	// 将原始文件内容写入压缩文件
	if _, err := io.Copy(gzWriter, file); err != nil {
		return err
	}

	// 删除原始日志文件
	return os.Remove(filename)
}
