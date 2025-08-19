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
	exp := time.Now().Add(c.expiration).UnixNano()
	for _, e := range exps {
		if e == 0 {
			exp = 0
			break
		}
		exp = time.Now().Add(e).UnixNano()
	}
	if err := c.setPutKey(k, data, exp); err != nil {
		return err
	}
	c.items[k] = Item{
		Object:     data,
		Expiration: exp,
	}
	return nil
}

// Get 获取缓存中的项目。如果项目不存在或已过期，则返回 nil。
func (c *cache) Get(k string) (item Item) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return item.SetInvalid(ErrKeyNotFound)
	}
	if item.Expired() {
		return item.SetInvalid(ErrKeyExpired)
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

// getNewNum 获取缓存中的数字类型值
// plus 表示是否增加，false 表示减少
func getNewNum[T Numeric](c *cache, k string, plus bool, n T) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, found := c.items[k]
	if !found {
		return 0, ErrKeyNotFound
	}
	if item.Expired() {
		return 0, ErrKeyExpired
	}
	var val T
	if err := item.Scan(&val); err != nil {
		return 0, err
	}
	num, err := anyToNumber(k, val)
	if err != nil {
		return 0, err
	}
	result := num - n
	if plus {
		result = num + n
	}
	if item.Object, err = gobEncode(result); err != nil {
		return 0, err
	}
	if err := c.setPutKey(k, item.Object, item.Expiration); err != nil {
		return 0, err
	}
	c.items[k] = item
	return result, nil
}

// Increment 将缓存中存储的数字增加 n。如果键不存在或值不是数字，则返回错误。
func Increment[T Numeric](c *cache, k string, n T) (number T, err error) {
	return getNewNum(c, k, true, n)
}

// Decrement 将缓存中存储的数字减少 n。如果键不存在或值不是数字，则返回错误。
func Decrement[T Numeric](c *cache, k string, n T) (T, error) {
	return getNewNum(c, k, false, n)
}

// Del 删除缓存中的项目
func (c *cache) Del(k string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.setDelKey(k); err != nil {
		return err
	}
	c.delete(k)
	return nil
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
func (c *cache) List() map[string]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var m = make(map[string]Item, 0)
	for k, v := range c.items {
		if v.Expired() {
			continue
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
	if c.store != nil {
		return c.store.flush()
	}
	c.flushDirty()
	c.items = map[string]Item{}
	return nil
}
