package cache

import (
	"log/slog"
	"slices"
	"time"
)

// 定期保存缓存
func (c *cache) startSaving() {
	slog.Info("start save cache")
	ticker := time.NewTicker(c.saveInterval)
	go func() {
		for range ticker.C {
			if c.writeTotal >= c.writeInterval {
				slog.Info("start sync cache")
				if err := c.Sync(); err != nil {
					slog.Error("failed to sync cache", "err", err)
				}
				c.writeTotal = 0
			}
		}
	}()
}

func (c *cache) Sync() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	slices.Sort(sortKeys)
	for _, id := range sortKeys {
		item := DirtyItems[id]
		if item.Opt == DirtyOpPut {
			if err := c.store.insert(item.Key, item.Value, item.Expire); err != nil {
				slog.Error("failed to insert cache", "err", err)
			}
			continue
		}
		if item.Opt == DirtyOpDel {
			if err := c.store.delete(item.Key); err != nil {
				slog.Error("failed to delete cache", "err", err)
			}
		}
	}
	return c.store.deleteExpire()
}
