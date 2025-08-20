package cache

import "C"
import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Item struct {
	Object     []byte
	Expiration int64
	invalid    bool  // 是否有效
	err        error // 错误信息
}

type kvStore func(path string) (Store, error)

const (
	NoExpiration time.Duration = 0 // 不过期
)

type cache struct {
	DirtyTotal      uint            // 脏数据总数
	expiration      time.Duration   // 默认过期时间
	items           map[string]Item // 缓存数据
	mu              sync.RWMutex
	store           Store                 // 缓存存储
	saveInterval    time.Duration         // 缓存文件保存间隔
	writeInterval   uint                  // 缓存文件写入间隔
	cleanupInterval time.Duration         // 清理过期项目的间隔时间
	janitor         *janitor              // 清理过期项目的后台 goroutine
	DirtyKey        map[DirtyOpt][]string // 脏数据键
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
		if c.saveInterval >= 1*time.Second {
			c.startSaving()
		}
		// 添加退出信号处理
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go func() {
			<-sigChan
			log.Println("收到退出信号，正在将缓存保存到文件...")
			log.Printf("缓存数据总数：%d", len(c.items))
			log.Printf("脏数据总数：%d\n", c.DirtyTotal)
			if c.DirtyTotal > 0 {
				if err = c.Sync(); err != nil {
					log.Println("缓存保存失败", err)
				}
			}
			log.Println("缓存保存完成")
			os.Exit(1)
		}()
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
func WithSave(f Store, interval time.Duration, writeNum uint) Options {
	if f == nil {
		return func(c *cache) {}
	}
	return func(c *cache) {
		c.store = f
		c.saveInterval = interval
		c.writeInterval = writeNum
		c.DirtyKey = make(map[DirtyOpt][]string)
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
