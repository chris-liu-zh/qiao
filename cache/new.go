package cache

import "C"
import (
	"bytes"
	"encoding/gob"
	"log"
	"runtime"
	"sync"
	"time"
)

type Item struct {
	Object     []byte
	Expiration int64
	invalid    bool  // 是否有效
	err        error // 错误信息
}

const (
	NoExpiration time.Duration = 0 // 不过期
)

type cache struct {
	writeTotal      int             // 写入总数
	expiration      time.Duration   // 默认过期时间
	items           map[string]Item // 缓存数据
	mu              sync.RWMutex
	janitor         *janitor      // 清理过期项目的后台 goroutine
	store           Store         // 缓存文件路径
	saveInterval    time.Duration // 缓存文件保存间隔
	writeInterval   int           // 缓存文件写入间隔
	cleanupInterval time.Duration // 清理过期项目的间隔时间
}

// newCacheWithJanitor 创建一个新的缓存实例，同时运行一个清理器 goroutine
func (c *cache) newCacheWithJanitor() (err error) {
	if c.cleanupInterval > 0 {
		c.runJanitor()                                // 运行清理器 goroutine
		runtime.SetFinalizer(c, (*cache).stopJanitor) // 设置清理器 goroutine 的最终izer
	}
	if c.store != nil {
		if c.items, err = c.store.load(); err != nil {
			return
		}
		if c.saveInterval > 0 {
			c.startSaving()
		}
		// 添加退出信号处理
		// sigChan := make(chan os.Signal, 1)
		// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		// go func() {
		// 	<-sigChan
		// 	log.Println("Received exit signal, saving cache to file...")
		// 	os.Exit(0)
		// }()
	}
	return
}

type Options func(*cache)

// WithDefaultExpiration 设置缓存项的默认过期时间
func WithDefaultExpiration(d time.Duration) Options {
	return func(c *cache) {
		if d <= 0 {
			d = NoExpiration
		}
		c.expiration = d
	}
}

// WithSave 设置缓存保存间隔
func WithSave(path string, interval time.Duration, writeNum int) Options {
	kv, err := NewKVStore(path)
	if err != nil {
		log.Fatalf("Error opening cache file: %v\n", err)
		return func(c *cache) {}
	}
	return func(c *cache) {
		c.store = kv
		c.saveInterval = interval
		c.writeInterval = writeNum
	}
}

// WithDatas 设置缓存数据
func WithDatas(items map[string]Item) Options {
	return func(c *cache) {
		c.items = items
	}
}

// WithCleanupInterval 设置缓存清理间隔
func WithCleanupInterval(interval time.Duration) Options {
	return func(c *cache) {
		c.cleanupInterval = interval
	}
}

// New 创建一个新的缓存实例
func New(opts ...Options) (*cache, error) {
	c := &cache{
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
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
