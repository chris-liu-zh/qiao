package cache

import (
	"log/slog"
	"time"
)

// 定期保存缓存
func (c *cache) startSaving() {
	ticker := time.NewTicker(c.saveInterval)
	go func() {
		for range ticker.C {
			if c.DirtyTotal >= c.writeInterval {
				if err := c.Sync(); err != nil {
					slog.Error("failed to sync cache", "err", err)
				}
			}
		}
	}()
}

func (c *cache) Sync() error {
	if c.store == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	slog.Info("start sync cache")
	startT := time.Now()
	delKeys := c.DirtyKey[DirtyOpDel]
	if err := c.store.sync(c, delKeys, delSql); err != nil {
		return err
	}
	putKeys := c.DirtyKey[DirtyOpPut]
	if err := c.store.sync(c, putKeys, putSql); err != nil {
		return err
	}
	tc := time.Since(startT) // 计算耗时
	slog.Info("sync cache ", "cost", tc)
	c.flushDirty()
	return nil
}
