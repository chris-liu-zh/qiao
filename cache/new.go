package cache

import "C"
import (
	"bytes"
	"encoding/gob"
	"runtime"
	"sync"
	"time"
)

type Item struct {
	Object     []byte
	Expiration int64
}

const (
	NoExpiration time.Duration = 0 // 不过期
)

type Cache struct {
	DirtyTotal      uint            // 脏数据总数
	expiration      time.Duration   // 默认过期时间
	items           map[string]Item // 缓存数据
	mu              sync.RWMutex
	filename        string        // 缓存文件名
	saveInterval    time.Duration // 缓存文件保存间隔
	writeInterval   uint          // 缓存文件写入间隔
	cleanupInterval time.Duration // 清理过期项目的间隔时间
	janitor         *janitor      // 清理过期项目的后台 goroutine
}

// newCacheWithJanitor 创建一个新的缓存实例，同时运行一个清理器 goroutine
func (c *Cache) newCacheWithJanitor() (err error) {
	if c.cleanupInterval > 0 {
		c.runJanitor()                                // 运行清理器 goroutine
		runtime.SetFinalizer(c, (*Cache).stopJanitor) // 设置清理器 goroutine 的最终izer
	}
	if c.saveInterval >= 1*time.Second {
		c.startSaving()
	}
	return
}

type Options func(*Cache)

// WithDefaultExpiration 设置缓存项的默认过期时间
func WithDefaultExpiration(d time.Duration) Options {
	return func(c *Cache) {
		if d <= 0 {
			d = NoExpiration
		}
		c.expiration = d
	}
}

// WithSave 设置缓存保存间隔,interval单位为秒
func WithSave(cache string, interval uint64, writeNum uint) Options {
	if cache == "" {
		cache = "cache.dat"
	}

	return func(c *Cache) {
		c.filename = cache
		c.saveInterval = time.Duration(interval) * time.Second
		c.writeInterval = writeNum
	}
}

// WithDatas 设置缓存数据
func WithDatas(items map[string]Item) Options {
	return func(c *Cache) {
		c.items = items
	}
}

// WithCleanupInterval 设置过期缓存清理间隔,interval单位为秒
func WithCleanupInterval(interval int64) Options {
	return func(c *Cache) {
		c.cleanupInterval = time.Duration(interval) * time.Second
	}
}

// New 创建一个新的缓存实例
func New(opts ...Options) (*Cache, error) {
	c := &Cache{
		expiration:      5 * time.Minute,
		cleanupInterval: 5 * time.Minute,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.items == nil {
		c.items = make(map[string]Item)
	}

	if err := c.newCacheWithJanitor(); err != nil {
		return nil, err
	}
	return c, nil
}

func gobDecode(data []byte, valType any) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(valType)
}

func gobEncode(data any) ([]byte, error) {
	var buf bytes.Buffer
	return buf.Bytes(), gob.NewEncoder(&buf).Encode(data)
}
