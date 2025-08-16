package cache

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
	ErrKeyInvalid  = errors.New("key invalid")
)

type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint | uintptr | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

// Set 将项目添加到缓存，替换现有值，使用指定的到期时间, 表示不过期
func (c *cache) Set(k string, v any, exps ...time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := gobEncode(v)
	if err != nil {
		return err
	}
	e := time.Now().Add(c.expiration).UnixNano()
	for _, exp := range exps {
		if exp > 0 {
			e = time.Now().Add(exp).UnixNano()
			break
		}
		e = -1
	}
	c.items[k] = Item{
		Object:     data,
		Expiration: e,
	}
	dirtyPut(k, data, e, DirtyOpPut)
	c.writeTotal++
	return nil
}

// Get 获取缓存中的项目。如果项目不存在或已过期，则返回 nil。
func (c *cache) Get(k string) (item Item) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return item.setInvalid(ErrKeyNotFound)
	}
	if item.Expired() {
		return item.setInvalid(ErrKeyExpired)
	}
	return item
}

// anyToNumber 将 any 类型转换为 Numeric 类型
func anyToNumber[T Numeric](k string, value T) (T, error) {
	rv := reflect.ValueOf(value)
	if rv.IsValid() && rv.Type().ConvertibleTo(reflect.TypeOf(*new(T))) {
		return T(rv.Convert(reflect.TypeOf(*new(T))).Interface().(T)), nil
	}
	return 0, fmt.Errorf("the value for %v does not have type Numeric", k)
}

// Increment 将缓存中存储的数字增加 n。如果键不存在或值不是数字，则返回错误。
func Increment[T Numeric](c *cache, k string, n T) (T, error) {
	item := c.Get(k)
	var val T
	if err := item.Scan(&val); err != nil {
		return 0, err
	}
	num, err := anyToNumber(k, val)
	if err != nil {
		return 0, err
	}
	result := num + n
	if item.Object, err = gobEncode(result); err != nil {
		return 0, err
	}
	c.items[k] = item
	dirtyPut(k, item.Object, item.Expiration, DirtyOpPut)
	c.writeTotal++
	return result, nil
}

// Decrement 将缓存中存储的数字减少 n。如果键不存在或值不是数字，则返回错误。
func Decrement[T Numeric](c *cache, k string, n T) (T, error) {
	item := c.Get(k)
	c.mu.Lock()
	defer c.mu.Unlock()
	var val T
	if err := item.Scan(&val); err != nil {
		return 0, err
	}
	num, err := anyToNumber(k, val)
	if err != nil {
		return 0, err
	}
	result := num - n
	if item.Object, err = gobEncode(result); err != nil {
		return 0, err
	}
	c.items[k] = item
	dirtyPut(k, item.Object, item.Expiration, DirtyOpPut)
	c.writeTotal++
	return result, nil
}

// Del 删除缓存中的项目
func (c *cache) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delete(k)
	dirtyPut(k, nil, 0, DirtyOpDel)
	c.writeTotal++
}

// delete 删除缓存中的项目。如果键不在缓存中，则不执行任何操作。
func (c *cache) delete(k string) {
	delete(c.items, k)
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
	if c.store != nil {
		return c.store.flush()
	}
	FlushDirtyItems()
	c.items = map[string]Item{}
	c.writeTotal = 0
	return nil
}
