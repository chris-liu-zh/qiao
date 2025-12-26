package redisCache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrRedisCacheOffline = errors.New("redis cache offline")
var ErrRedisCacheNotInit = errors.New("redis cache not init")

type RedisCache struct {
	Client            redis.Cmdable
	ctx               context.Context
	retryIng          bool
	sign              string
	online            atomic.Bool
	ReconnectNum      int           //重连次数
	ReconnectInterval time.Duration //重连间隔时间
}

func ping(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Ping(ctx).Err()
}

// NewStandaloneClient 创建单机Redis客户端
func NewStandaloneClient(options *RedisOptions) (*RedisCache, error) {
	client := redis.NewClient(options.Opt)
	if err := ping(client); err != nil {
		return nil, err
	}
	cache := &RedisCache{
		Client:            client,
		ctx:               context.Background(),
		sign:              options.Base.Sign + ":",
		ReconnectNum:      options.Base.ReconnectNum,
		ReconnectInterval: options.Base.ReconnectInterval,
	}
	cache.online.Store(true)
	return cache, nil
}

// NewFailoverClient 创建主从Redis客户端
func NewFailoverClient(options *FailoverOptions) (*RedisCache, error) {
	client := redis.NewFailoverClient(options.Opt)
	if err := ping(client); err != nil {
		return nil, err
	}
	cache := &RedisCache{
		Client:            client,
		ctx:               context.Background(),
		sign:              options.Base.Sign + ":",
		ReconnectNum:      options.Base.ReconnectNum,
		ReconnectInterval: options.Base.ReconnectInterval,
	}
	cache.online.Store(true)
	return cache, nil
}

func (cache *RedisCache) getClient() *redis.Client {
	if cache.Client == nil {
		slog.Error("Redis server not started", "error", ErrRedisCacheNotInit)
		return nil
	}
	if client, ok := cache.Client.(*redis.Client); ok {
		return client
	}
	return nil
}

func (cache *RedisCache) ShowPoolStats() {
	if client := cache.getClient(); client != nil {
		stats := client.PoolStats()
		// 打印连接池详细信息
		fmt.Println("\n===== Redis连接池统计信息 =====")
		fmt.Printf("总连接数（已创建的所有连接）: %d\n", stats.TotalConns)
		fmt.Printf("活跃连接数（正在使用的连接）: %d\n", stats.PubSubStats.Active)
		fmt.Printf("空闲连接数（池中空闲的连接）: %d\n", stats.IdleConns)
		fmt.Printf("等待连接的请求数: %d\n", stats.WaitCount)
		fmt.Printf("等待连接的总时长(纳秒): %d\n", stats.WaitDurationNs)
		fmt.Printf("连接超时次数: %d\n", stats.Timeouts)
		fmt.Printf("命中空闲连接的次数: %d\n", stats.Hits)
		fmt.Printf("未命中空闲连接的次数: %d\n", stats.Misses)
		return
	}
	slog.Error("Redis server not started", "error", ErrRedisCacheNotInit)
}

func (cache *RedisCache) ShowSentinelInfo() {
	if client := cache.getClient(); client != nil {
		sentinelInfo := client.Info(cache.ctx, "replication").String()
		fmt.Println("===== Redis哨兵信息 =====")
		fmt.Println(sentinelInfo)
		return
	}
	slog.Error("Redis server not started", "error", ErrRedisCacheNotInit)
}

func (cache *RedisCache) CheckOpError(err error) error {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		cache.online.Store(false)
		slog.Error("Redis连接异常", "error", err)
		go cache.Retry()
		return ErrRedisCacheOffline
	}
	return err
}

func (cache *RedisCache) Retry() {
	if cache.retryIng {
		return
	}
	cache.retryIng = true
	defer func() { cache.retryIng = false }()
	for range cache.ReconnectNum {
		if cache.online.Load() == true {
			return
		}
		if err := cache.Reconnect(); err != nil {
			slog.Error("Redis重连失败", "error", err)
			time.Sleep(cache.ReconnectInterval)
			continue
		}

		slog.Info("Redis重连成功", "sign", cache.sign)
		cache.online.Store(true)
		return
	}
}

// HealthCheck 检查Redis连接是否正常
func (cache *RedisCache) Reconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return cache.Client.Ping(ctx).Err()
}
