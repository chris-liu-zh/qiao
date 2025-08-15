package ignore

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"time"
)

// Expired 如果项目已过期，则返回 true。
func (item cache.Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint | uintptr | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

// Set 将项目添加到缓存，替换现有值，使用指定的到期时间, 表示不过期
func (c *cache.cache) Set(k string, x any, exps ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := time.Now().Add(c.expiration).UnixNano()
	for _, exp := range exps {
		if exp > 0 {
			e = time.Now().Add(exp).UnixNano()
			break
		}
		e = -1
	}
	c.datas[k] = cache.Item{
		Object:     x,
		Expiration: e,
	}
	c.writeTotal++
}

// Get 获取缓存中的项目。如果项目不存在或已过期，则返回 nil。
func (c *cache.cache) Get(k string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.datas[k]
	if !found {
		return nil, false
	}
	if item.Expired() {
		return nil, false
	}
	return item.Object, true
}

// GetWithExpiration 获取数据和过期时间
func (c *cache.cache) GetWithExpiration(k string) (any, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.datas[k]
	if !found {
		return nil, time.Time{}, false
	}

	if item.Expiration > 0 {
		if item.Expired() {
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
func Increment[T Numeric](c *cache.cache, k string, n T) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, found := c.datas[k]
	if !found || v.Expired() {
		return n, fmt.Errorf("item %s not found", k)
	}
	num, err := anyToNumber[T](k, v.Object)
	if err != nil {
		return 0, err
	}
	result := num + n
	v.Object = result
	c.datas[k] = v
	c.writeTotal++
	return v.Object.(T), nil
}

// Decrement 将缓存中存储的数字减少 n。如果键不存在或值不是数字，则返回错误。
func Decrement[T Numeric](c *cache.cache, k string, n T) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, found := c.datas[k]
	if !found || v.Expired() {
		return 0, fmt.Errorf("item %s not found", k)
	}
	num, err := anyToNumber[T](k, v.Object)
	if err != nil {
		return 0, err
	}
	result := num - n
	v.Object = result
	c.datas[k] = v
	c.writeTotal++
	return v.Object.(T), nil
}

// Del 从缓存中删除一个项目。如果键不在缓存中，则不执行任何操作。
func (c *cache.cache) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, evicted := c.delete(k)
	if evicted {
		c.onEvicted(k, v)
	}
}

func (c *cache.cache) delete(k string) (any, bool) {
	defer func() {
		c.writeTotal++
	}()
	if c.onEvicted != nil {
		if v, found := c.datas[k]; found {
			delete(c.datas, k)
			return v.Object, true
		}
	}
	delete(c.datas, k)
	return nil, false
}

// DeleteExpired 从缓存中删除所有过期的项目。
func (c *cache.cache) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.datas {
		if v.Expiration > 0 && now > v.Expiration {
			if ov, evicted := c.delete(k); evicted {
				c.onEvicted(k, ov)
			}
		}
	}
}

// OnEvicted 设置一个回调函数，当缓存项被逐出时执行
func (c *cache.cache) OnEvicted(f func(string, any)) {
	c.mu.Lock()
	c.onEvicted = f
	c.mu.Unlock()
}

// List 将缓存中所有未过期的项目复制到新映射中并返回
func (c *cache.cache) List() map[string]cache.Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string]cache.Item, len(c.datas))
	now := time.Now().UnixNano()
	for k, v := range c.datas {
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
func (c *cache.cache) Count() int {
	c.mu.RLock()
	count := len(c.datas)
	c.mu.RUnlock()
	return count
}

// PrintCache 打印缓存中的所有项目
func PrintCache(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var data map[string]cache.Item
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
