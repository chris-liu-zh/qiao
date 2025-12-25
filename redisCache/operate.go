package redisCache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Set 设置键值对
func (cache *RedisCache) Set(key string, value any, ttl time.Duration) *redis.StatusCmd {
	cmd := &redis.StatusCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	if !cache.Online.Load() {
		cmd.SetErr(ErrRedisCacheOffline)
		return cmd
	}
	cmd = cache.client.Set(cache.ctx, cache.sign+key, value, ttl)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Get 获取键值对
func (cache *RedisCache) Get(key string) *redis.StringCmd {
	cmd := &redis.StringCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	if !cache.Online.Load() {
		cmd.SetErr(ErrRedisCacheOffline)
		return cmd
	}
	cmd = cache.client.Get(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Delete 删除键值对
func (cache *RedisCache) Delete(key string) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.Del(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Exists 检查键是否存在
func (cache *RedisCache) Exists(key string) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.Exists(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Expire 设置键的过期时间
func (cache *RedisCache) Expire(key string, ttl time.Duration) *redis.BoolCmd {
	cmd := &redis.BoolCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.Expire(cache.ctx, cache.sign+key, ttl)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// TTL 获取键的剩余生存时间
func (cache *RedisCache) TTL(key string) *redis.DurationCmd {
	cmd := &redis.DurationCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.TTL(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// HSet 设置哈希字段的值
func (cache *RedisCache) HSet(key, field string, value ...any) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.HSet(cache.ctx, cache.sign+key, field, value)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

func (cache *RedisCache) HExpire(key string, ttl time.Duration, fields ...string) *redis.IntSliceCmd {
	cmd := &redis.IntSliceCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.HExpire(cache.ctx, cache.sign+key, ttl, fields...)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// HGet 获取哈希字段的值
func (cache *RedisCache) HGet(key, field string) *redis.StringCmd {
	cmd := &redis.StringCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.HGet(cache.ctx, cache.sign+key, field)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// HDel 删除哈希字段
func (cache *RedisCache) HDel(key, field string) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.HDel(cache.ctx, cache.sign+key, field)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// HGetAll 获取哈希中所有字段和值
func (cache *RedisCache) HGetAll(key string) *redis.MapStringStringCmd {
	cmd := &redis.MapStringStringCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.HGetAll(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Incr 对键的值进行自增
func (cache *RedisCache) Incr(key string) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.Incr(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// IncrBy 对键的值进行指定步长的自增
func (cache *RedisCache) IncrBy(key string, increment int64) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.IncrBy(cache.ctx, cache.sign+key, increment)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Decr 对键的值进行自减
func (cache *RedisCache) Decr(key string) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.Decr(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// DecrBy 对键的值进行指定步长的自减
func (cache *RedisCache) DecrBy(key string, decrement int64) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.DecrBy(cache.ctx, cache.sign+key, decrement)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// LPush 将一个或多个值插入到列表头部
func (cache *RedisCache) LPush(key string, values ...any) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.LPush(cache.ctx, cache.sign+key, values...)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// RPush 将一个或多个值插入到列表尾部
func (cache *RedisCache) RPush(key string, values ...any) *redis.IntCmd {
	cmd := &redis.IntCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.RPush(cache.ctx, cache.sign+key, values...)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// LPop 移出并获取列表的第一个元素
func (cache *RedisCache) LPop(key string) *redis.StringCmd {
	cmd := &redis.StringCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.LPop(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// RPop 移出并获取列表的最后一个元素
func (cache *RedisCache) RPop(key string) *redis.StringCmd {
	cmd := &redis.StringCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.RPop(cache.ctx, cache.sign+key)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Keys 获取所有匹配模式的键
func (cache *RedisCache) Keys(pattern string) *redis.StringSliceCmd {
	cmd := &redis.StringSliceCmd{}
	if cache.client == nil {
		cmd.SetErr(ErrRedisCacheNotInit)
		return cmd
	}
	cmd = cache.client.Keys(cache.ctx, cache.sign+pattern)
	if err := cmd.Err(); err != nil {
		cmd.SetErr(cache.CheckOpError(err))
	}
	return cmd
}

// Flush 清空当前数据库中以cache.sign开头的key
func (cache *RedisCache) Flush() error {
	if cache.client == nil {
		return ErrRedisCacheNotInit
	}

	// 使用SCAN命令查找所有以cache.sign开头的key
	pattern := cache.sign + "*"
	cursor := uint64(0)
	var err error
	for {
		var keys []string
		var nextCursor uint64
		keys, nextCursor, err = cache.client.Scan(cache.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return cache.CheckOpError(err)
		}

		if len(keys) > 0 {
			// 删除找到的keys
			delCmd := cache.client.Del(cache.ctx, keys...)
			if delErr := delCmd.Err(); delErr != nil {
				return cache.CheckOpError(delErr)
			}

			// 如果游标回到0，表示遍历完成
			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}
	}
	return nil
}
