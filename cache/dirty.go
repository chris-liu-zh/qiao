package cache

import (
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

func (c *cache) flushDirty() {
	c.DirtyKey = make(map[DirtyOpt][]string)
	c.DirtyTotal = 0
}

func (c *cache) setPutKey(key string, data []byte, expiration int64) error {
	if c.store == nil {
		return nil
	}
	if c.saveInterval < 1*time.Second && c.DirtyTotal > c.writeInterval {
		go c.Sync()
	}
	if c.saveInterval > 1*time.Second || c.writeInterval > 0 {
		c.DirtyKey[DirtyOpPut] = append(c.DirtyKey[DirtyOpPut], key)
		c.DirtyTotal++
		return nil
	}
	return c.store.put(key, data, expiration)
}

func (c *cache) setDelKey(key string) error {
	if c.store == nil {
		return nil
	}
	if c.saveInterval < 1*time.Second && c.DirtyTotal > c.writeInterval {
		go c.Sync()
	}
	if c.saveInterval > 1*time.Second || c.writeInterval > 0 {
		c.DirtyKey[DirtyOpDel] = append(c.DirtyKey[DirtyOpDel], key)
		c.DirtyTotal++
		return nil
	}
	return c.store.delete(key)
}
