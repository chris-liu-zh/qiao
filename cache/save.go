package cache

import (
	"log/slog"
	"time"
)

// 定期保存缓存
func (c *Cache) startSaving() {
	ticker := time.NewTicker(c.saveInterval)
	go func() {
		for range ticker.C {
			if c.DirtyTotal > 0 && c.DirtyTotal >= c.writeInterval {
				if err := c.Sync(); err != nil {
					slog.Error("failed to sync cache", "err", err)
				}
			}
		}
	}()
}

func (c *Cache) Sync() error {
	if c.store == nil {
		return nil
	}
	slog.Info("start sync cache")
	startT := time.Now()
	delKeys := c.DirtyKey[DirtyOpDel]
	if err := c.store.sync(c, delSql, delKeys); err != nil {
		return err
	}
	putKeys := c.DirtyKey[DirtyOpPut]
	if err := c.store.sync(c, putSql, putKeys); err != nil {
		return err
	}
	tc := time.Since(startT) // 计算耗时
	slog.Info("sync cache ", "cost", tc)
	c.flushDirty()
	return nil
}
