package cache

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"time"
)

// Expired 如果项目已过期，则返回 true。
func (item Item) Expired() bool {
	if item.Expiration <= 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint | uintptr | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

// Set 将项目添加到缓存，替换现有值，使用指定的到期时间, 表示不过期
func (c *cache) Set(k string, x any, exps ...time.Duration) {
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
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
	c.writeTotal++
}

// Get 获取缓存中的项目。如果项目不存在或已过期，则返回 nil。
func (c *cache) Get(k string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expired() {
		return nil, false
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
func Increment[T Numeric](c *cache, k string, n T) (T, error) {
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
	c.writeTotal++
	return v.Object.(T), nil
}

// Decrement 将缓存中存储的数字减少 n。如果键不存在或值不是数字，则返回错误。
func Decrement[T Numeric](c *cache, k string, n T) (T, error) {
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
	c.writeTotal++
	return v.Object.(T), nil
}

// Del 删除缓存中的项目
func (c *cache) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, evicted := c.delete(k)
	if evicted {
		c.onEvicted(k, v)
	}
}

// delete 删除缓存中的项目。如果键不在缓存中，则不执行任何操作。
func (c *cache) delete(k string) (any, bool) {
	if v, found := c.items[k]; found {
		delete(c.items, k)
		c.writeTotal++
		return v.Object, true
	}
	return nil, false
}

// DeleteExpired 从缓存中删除所有过期的项目。
func (c *cache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v.Expired() {
			c.delete(k)
		}
	}
}

// List 将缓存中所有未过期的项目复制到新映射中并返回
func (c *cache) List() []Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var m []Item
	for _, v := range c.items {
		if v.Expired() {
			continue
		}
		m = append(m, v)
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
	if c.dataFile != nil {
		if err := os.Truncate(c.dataFile.Name(), 0); err != nil {
			return fmt.Errorf("error clearing cache file: %v", err)
		}
	}
	return nil
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
