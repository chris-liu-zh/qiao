package cache

import (
	"log/slog"
	"time"
)

type DirtyOpt bool

const (
	DirtyOpPut DirtyOpt = true
	DirtyOpDel DirtyOpt = false
)

type DirtyKey struct {
	Key   []string
	Total uint
}

func (c *Cache) flushDirty() {
	c.DirtyKey = make(map[DirtyOpt][]string)
	c.DirtyTotal = 0
}

func (c *Cache) setPutKey(key string, data []byte, expiration int64) error {
	if c.store == nil {
		return nil
	}
	if c.saveInterval < 1*time.Second && c.DirtyTotal > c.writeInterval {
		c.Sync()
	}
	if c.saveInterval > 1*time.Second || c.writeInterval > 0 {
		c.DirtyKey[DirtyOpPut] = append(c.DirtyKey[DirtyOpPut], key)
		c.DirtyTotal++
		return nil
	}
	return c.store.put(key, data, expiration)
}

func (c *Cache) setDelKey(key string) error {
	if c.store == nil {
		return nil
	}
	if c.saveInterval < 1*time.Second && c.DirtyTotal > c.writeInterval {
		if err := c.Sync(); err != nil {
			slog.Error("failed to sync cache", "err", err)
		}
	}
	if c.saveInterval > 1*time.Second || c.writeInterval > 0 {
		c.DirtyKey[DirtyOpDel] = append(c.DirtyKey[DirtyOpDel], key)
		c.DirtyTotal++
		return nil
	}
	return c.store.delete(key)
}
