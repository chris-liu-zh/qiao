package ignore

import "C"
import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Item struct {
	Object     any
	Expiration int64
}

const (
	opInsert = iota + 1
	opUpdate
	opDelete
)

const (
	NoExpiration time.Duration = -1 // 不过期
)

type cache struct {
	writeTotal      int             // 写入总数
	expiration      time.Duration   // 默认过期时间
	datas           map[string]Item // 缓存数据
	mu              sync.RWMutex
	onEvicted       func(string, any) // 项目被删除时调用的回调函数
	janitor         *janitor          // 清理过期项目的后台 goroutine
	dataFile        *os.File          // 缓存文件路径
	saveInterval    time.Duration     // 缓存文件保存间隔
	writeInterval   int               // 缓存文件写入间隔
	cleanupInterval time.Duration     // 清理过期项目的间隔时间
}

// newCacheWithJanitor 创建一个新的缓存实例，同时运行一个清理器 goroutine
func (c *cache) newCacheWithJanitor() (err error) {
	if c.cleanupInterval > 0 {
		c.runJanitor()                                // 运行清理器 goroutine
		runtime.SetFinalizer(c, (*cache).stopJanitor) // 设置清理器 goroutine 的最终izer
	}
	if c.dataFile != nil {
		if err = c.LoadFile(); err != nil {
			return
		}
		c.startSaving()
		// 添加退出信号处理
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go func() {
			<-sigChan
			log.Println("Received exit signal, saving cache to file...")
			os.Exit(0)
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
func WithSave(path string, interval time.Duration, writeNum int) Options {
	dataFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening cache file: %v\n", err)
	}
	return func(c *cache) {
		c.dataFile = dataFile
		c.saveInterval = interval
		c.writeInterval = writeNum
	}
}

// WithDatas 设置缓存数据
func WithDatas(items map[string]Item) Options {
	return func(c *cache) {
		c.datas = items
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
		expiration:      3600,
		cleanupInterval: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.datas == nil {
		c.datas = make(map[string]Item)
	}
	if err := c.newCacheWithJanitor(); err != nil {
		return nil, err
	}
	return c, nil
}
