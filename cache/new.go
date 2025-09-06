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
	expiration      time.Duration   // 默认过期时间
	items           map[string]Item // 缓存数据
	mu              sync.RWMutex
	cleanupInterval time.Duration // 清理过期项目的间隔时间
	janitor         *janitor      // 清理过期项目的后台 goroutine
}

// newCacheWithJanitor 创建一个新的缓存实例，同时运行一个清理器 goroutine
func (c *Cache) newCacheWithJanitor() (err error) {
	if c.cleanupInterval > 0 {
		c.runJanitor()                                // 运行清理器 goroutine
		runtime.SetFinalizer(c, (*Cache).stopJanitor) // 设置清理器 goroutine 的最终izer
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
