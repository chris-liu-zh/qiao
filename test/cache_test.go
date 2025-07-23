/*
 * @Author: Chris
 * @Date: 2025-03-07 12:03:36
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 15:25:36
 * @Description: 请填写简介
 */
package qiao

import (
	"fmt"
	"testing"
	"time"

	"qiao/cache"
)

func Test_CacheSet(t *testing.T) {
	c := cache.New(3600*time.Second, 2*time.Second)
	c.OnEvicted(func(key string, value any) {
		fmt.Printf("Item %s with value %v has been evicted from the cache.\n", key, value)
	})
	c.Set("test1", 1, 1*time.Second)

	time.Sleep(5 * time.Second)
	c.DeleteExpired()
}

func Test_IandD(t *testing.T) {
	c := cache.New(3600*time.Second, 10*time.Second)
	c.Set("test", 1, cache.DefaultExpiration)

	aa, err := cache.Increment(c, "test", 1.1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("v:%v t: %T\n", aa, aa)
	aa, err = cache.Decrement(c, "test", 1.1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("v:%v t: %T\n", aa, aa)
}

func Test_CacheGet(t *testing.T) {
	s := cache.CustomSave("cache", 1*time.Second, 0)
	c := cache.NewFromFile(3600*time.Second, 10*time.Second, s)

	for i := range 10 {
		if v, ok := c.Get(fmt.Sprintf("test%d", i)); ok {
			fmt.Printf("test%d:%v t: %T\n", i, v, v)
		}
	}
}

func Test_cacheFromfile(t *testing.T) {
	s := cache.CustomSave("cache", 1*time.Second, 0)
	c := cache.NewFromFile(3600*time.Second, 10*time.Second, s)
	for i := range 10 {
		c.Set(fmt.Sprintf("test%d", i), i, cache.DefaultExpiration)
	}

	time.Sleep(3 * time.Second)
}

func Test_ViewCache(t *testing.T) {
	cache.PrintCache("cache")
}
