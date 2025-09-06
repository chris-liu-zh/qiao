package cache

import (
	"encoding/gob"
	"log/slog"
	"os"
	"time"
)

// 定期保存缓存
func (c *Cache) startSaving() {
	ticker := time.NewTicker(c.saveInterval)
	go func() {
		for range ticker.C {
			if c.DirtyTotal > 0 && c.DirtyTotal >= c.writeInterval {
				c.mu.RLock()
				if err := c.Sync(); err != nil {
					slog.Error("failed to sync cache", "err", err)
				}
				c.mu.RUnlock()
			}
		}
	}()
}

func (c *Cache) Sync() error {
	slog.Info("start sync cache")
	startT := time.Now()
	file, err := os.Create(c.filename) // 创建或清空目标文件
	if err != nil {
		return err
	}
	defer file.Close()

	err = gob.NewEncoder(file).Encode(c.items)
	if err != nil {
		return err
	}
	tc := time.Since(startT) // 计算耗时
	slog.Info("sync cache ", "cost", tc)
	c.clearDirty()
	return nil
}
