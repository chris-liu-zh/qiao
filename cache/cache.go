package cache

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Item struct {
	Object     any
	Expiration int64
}

// 如果项目已过期，则返回 true。
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Cache struct {
	*cache
}

type cache struct {
	writeSum          int
	cachefile         string
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
	onEvicted         func(string, any)
	janitor           *janitor
}

// 清理过期项目的后台 goroutine
type janitor struct {
	interval time.Duration
	stop     chan bool
}

type save struct {
	cachefile     string
	saveInterval  time.Duration // 缓存文件保存间隔
	writeInterval int           // 缓存文件写入间隔
}

type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint | uintptr | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func (c *cache) Set(k string, x any, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
	c.writeSum++
}

// SetDefault 将项目添加到缓存，替换现有值，使用默到期时间。
func (c *cache) SetDefault(k string, x any) {
	c.Set(k, x, DefaultExpiration)
}

func (c *cache) Get(k string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Object, true
}

// GetWithExpiration 获取数据和过期时间
func (c *cache) GetWithExpiration(k string) (any, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, time.Time{}, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, time.Time{}, false
		}
		return item.Object, time.Unix(0, item.Expiration), true
	}
	return item.Object, time.Time{}, true
}

// anyToNumber 将 any 类型转换为 Numeric 类型
func anyToNumber[T Numeric](k string, value any) (T, error) {
	rv := reflect.ValueOf(value)
	if rv.IsValid() && rv.Type().ConvertibleTo(reflect.TypeOf(*new(T))) {
		return T(rv.Convert(reflect.TypeOf(*new(T))).Interface().(T)), nil
	}
	return 0, fmt.Errorf("the value for %v does not have type Numeric", k)
}

// Increment 将缓存中存储的数字增加 n。如果键不存在或值不是数字，则返回错误。
func Increment[T Numeric](c *Cache, k string, n T) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, found := c.items[k]
	if !found || v.Expired() {
		return n, fmt.Errorf("item %s not found", k)
	}
	num, err := anyToNumber[T](k, v.Object)
	if err != nil {
		return 0, err
	}
	result := num + n
	v.Object = result
	c.items[k] = v
	c.writeSum++
	return v.Object.(T), nil
}

// Decrement 将缓存中存储的数字减少 n。如果键不存在或值不是数字，则返回错误。
func Decrement[T Numeric](c *Cache, k string, n T) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, found := c.items[k]
	if !found || v.Expired() {
		return 0, fmt.Errorf("item %s not found", k)
	}
	num, err := anyToNumber[T](k, v.Object)
	if err != nil {
		return 0, err
	}
	result := num - n
	v.Object = result
	c.items[k] = v
	c.writeSum++
	return v.Object.(T), nil
}

// Del 从缓存中删除一个项目。如果键不在缓存中，则不执行任何操作。
func (c *cache) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, evicted := c.delete(k)
	if evicted {
		c.onEvicted(k, v)
	}
}

func (c *cache) delete(k string) (any, bool) {
	defer func() {
		c.writeSum++
	}()
	if c.onEvicted != nil {
		if v, found := c.items[k]; found {
			delete(c.items, k)
			return v.Object, true
		}
	}
	delete(c.items, k)
	return nil, false
}

// DeleteExpired 从缓存中删除所有过期的项目。
func (c *cache) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			if ov, evicted := c.delete(k); evicted {
				c.onEvicted(k, ov)
			}
		}
	}
}

// OnEvicted 设置一个回调函数，当缓存项被逐出时执行
func (c *cache) OnEvicted(f func(string, any)) {
	c.mu.Lock()
	c.onEvicted = f
	c.mu.Unlock()
}

func (c *cache) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	c.mu.RLock()
	defer c.mu.RUnlock()
	return enc.Encode(&c.items)
}

func (c *cache) SaveFile(fname string) error {
	tempFile, err := os.CreateTemp("", "cache-*.tmp")
	if err != nil {
		return err
	}
	tempFileName := tempFile.Name()
	fmt.Println(tempFileName)
	defer func() {
		tempFile.Close()
		os.Remove(tempFileName)
	}()

	if err = c.Save(tempFile); err != nil {
		return err
	}

	// 将临时文件重命名为目标文件
	if err = os.Rename(tempFileName, fname); err != nil {
		return err
	}
	return nil
}

// 定期保存缓存
func (c *cache) startSaving(s *save) {
	if s.saveInterval > 0 {
		ticker := time.NewTicker(s.saveInterval)
		go func() {
			for range ticker.C {
				if c.writeSum >= s.writeInterval {
					if err := c.SaveFile(s.cachefile); err != nil {
						log.Printf("Error saving cache to file: %v\n", err)
						continue
					}
					c.writeSum = 0
				}
			}
		}()
	}
}

func (c *cache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	if err := dec.Decode(&items); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range items {
		ov, found := c.items[k]
		if !found || ov.Expired() {
			c.items[k] = v
		}
	}
	return nil
}

func (c *cache) LoadFile(fname string) error {
	fp, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	if err = c.Load(fp); err != nil {
		return err
	}
	return nil
}

// Items 将缓存中所有未过期的项目复制到新映射中并返回
func (c *cache) Items() map[string]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string]Item, len(c.items))
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.Expiration > 0 {
			if now > v.Expiration {
				continue
			}
		}
		m[k] = v
	}
	return m
}

// Count 返回缓存中的项目数量。这个数量可能包括已经过期但尚未被清理的项目
func (c *cache) Count() int {
	c.mu.RLock()
	count := len(c.items)
	c.mu.RUnlock()
	return count
}

// Flush 清除缓存中的所有项目
func (c *cache) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
	if c.cachefile != "" {
		if err := os.Truncate(c.cachefile, 0); err != nil {
			return fmt.Errorf("error clearing cache file: %v", err)
		}
	}
	return nil
}

func (j *janitor) Run(c *cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

// stopJanitor 停止清理器 goroutine
func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}

// runJanitor 运行一个清理器 goroutine，定期删除过期的缓存项
func runJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.Run(c)
}

// newCache 创建一个新的缓存实例
func newCache(de time.Duration, m map[string]Item) *cache {
	if de == 0 {
		de = -1
	}
	c := &cache{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

// newCacheWithJanitor 创建一个新的缓存实例，同时运行一个清理器 goroutine
func newCacheWithJanitor(de, ci time.Duration, m map[string]Item, s *save) *Cache {
	c := newCache(de, m)
	C := &Cache{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}
	if s != nil {
		c.cachefile = s.cachefile
		c.LoadFile(s.cachefile)
		c.startSaving(s)
		// 添加退出信号处理
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go func() {
			<-sigChan
			log.Println("Received exit signal, saving cache to file...")
			if err := c.SaveFile(s.cachefile); err != nil {
				log.Printf("Error saving cache on exit: %v\n", err)
			} else {
				log.Println("Cache saved successfully on exit")
			}
			os.Exit(0)
		}()
	}
	return C
}

func DefaultSave() *save {
	return &save{
		cachefile:     "cache",
		saveInterval:  60 * time.Second,
		writeInterval: 10,
	}
}

// CustomSave 创建一个自定义的保存配置
func CustomSave(cachefile string, saveInterval time.Duration, writeInterval int) *save {
	return &save{
		cachefile:     cachefile,
		saveInterval:  saveInterval,
		writeInterval: writeInterval,
	}
}

// NewFromFile 创建一个新的缓存实例，从指定的文件加载缓存数据
func NewFromFile(defaultExpiration, cleanupInterval time.Duration, s *save) *Cache {
	if s == nil {
		return New(defaultExpiration, cleanupInterval)
	}
	items := make(map[string]Item)
	c := newCacheWithJanitor(defaultExpiration, cleanupInterval, items, s)
	return c
}

// New 创建一个新的缓存实例
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]Item)
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items, nil)
}

// NewFrom 创建一个新的缓存实例，从指定的映射加载缓存数据
func NewFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) *Cache {
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items, nil)
}

// PrintCache 打印缓存中的所有项目
func PrintCache(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var data map[string]Item
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Error decoding file:", err)
		return
	}

	for k, v := range data {
		fmt.Println(k, v)
	}
}
