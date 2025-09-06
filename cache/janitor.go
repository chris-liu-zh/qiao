package cache

import (
	"time"
)

type janitor struct {
	interval time.Duration
	stop     chan bool
}

// Run 运行清理器 goroutine，定期删除过期的缓存项
func (j *janitor) run(c *Cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.Clear()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

// stopJanitor 停止清理器 goroutine
func (c *Cache) stopJanitor() {
	c.janitor.stop <- true
}

// runJanitor 运行一个清理器 goroutine，定期删除过期的缓存项
func (c *Cache) runJanitor() {
	j := &janitor{
		interval: c.cleanupInterval,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.run(c)
}
