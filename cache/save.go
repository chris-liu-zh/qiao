package cache

import (
	"log/slog"
	"slices"
	"time"
)

// 定期保存缓存
func (c *cache) startSaving() {
	ticker := time.NewTicker(c.saveInterval)
	go func() {
		for range ticker.C {
			if c.DirtyTotal >= c.writeInterval {
				slog.Info("start sync cache")
				if err := c.Sync(); err != nil {
					slog.Error("failed to sync cache", "err", err)
				}
				c.flushDirty()
			}
		}
	}()
}

func (c *cache) Sync() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	slices.Sort(c.sortDirtyKeys)
	for _, id := range c.sortDirtyKeys {
		item := c.dirtyItems[id]
		if item.Opt == DirtyOpPut {
			if err := c.store.insert(item.Key, item.Value, item.Expire); err != nil {
				return err
			}
			continue
		}
		if item.Opt == DirtyOpDel {
			if err := c.store.delete(item.Key); err != nil {
				return err
			}
		}
	}
	return c.store.deleteExpire()
}

func (c *cache) SyncAll() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if err := c.store.flush(); err != nil {
		return err
	}
	c.flushDirty()
	return c.store.batchSet(c.List())
}
