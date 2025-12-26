package qiao

import (
	"fmt"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/redisCache"
	"github.com/redis/go-redis/v9"
)

func GetRedisConfig() *redisCache.FailoverOptions {
	return &redisCache.FailoverOptions{
		Opt: &redis.FailoverOptions{
			MasterName:      "mymaster", // Redis地址
			SentinelAddrs:   []string{}, // 哨兵地址
			Password:        "",         // Redis密码
			DB:              0,          // Redis库
			PoolSize:        10,         // Redis连接池大小
			MinIdleConns:    5,          // 最小空闲连接数
			MaxRetries:      3,          // 最大重试次数（重连后重试）
			MinRetryBackoff: 100 * time.Millisecond,
			DialTimeout:     5 * time.Second, // 连接超时时间
			ReadTimeout:     3 * time.Second, // 读超时
			WriteTimeout:    3 * time.Second, // 写超时
		},
		Base: redisCache.BaseOptions{
			Sign:              "memory",
			ReconnectNum:      10,
			ReconnectInterval: 5 * time.Second,
		},
	}
}

func TestRedisClient(t *testing.T) {
	cache := redisCache.NewFailoverClient(GetRedisConfig())
	if err := cache.Ping(); err != nil {
		t.Errorf("Failed to connect to Redis: %v", err)
		return
	}
	cache.ShowSentinelInfo()
	cache.ShowPoolStats()
	if cmd := cache.Set("test_key", 123456, 1*time.Hour); cmd.Err() != nil {
		t.Errorf("Failed to set key: %v", cmd.Err())
	}
	var data int
	if err := cache.Get("test_key").Scan(&data); err != nil {
		t.Errorf("Failed to get key: %v", err)
	}
	fmt.Printf("\nRetrieved from Redis: %v\n", data)
}
