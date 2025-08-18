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

func (c *cache) flushDirty() {
	c.DirtyKey = make(map[DirtyOpt][]string)
	c.DirtyTotal = 0
}

func (c *cache) setDirtyKey(key string, opt DirtyOpt) {
	if c.store == nil {
		return
	}
	var err error
	defer func() {
		if err != nil {
			slog.Error("setDirtyKey recover", "err", err)
		}
	}()
	// 实时保存
	if c.saveInterval < 1*time.Second && c.writeInterval == 0 {
		if opt {
			err = c.store.put(key, c.items[key].Object, c.items[key].Expiration)
			return
		}
		err = c.store.delete(key)
		return
	}
	c.DirtyKey[opt] = append(c.DirtyKey[opt], key)
	c.DirtyTotal++
	if c.DirtyTotal > c.writeInterval && c.saveInterval < 1*time.Second {
		go c.Sync()
	}
}
