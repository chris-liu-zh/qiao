package qiao

import (
	"time"

	"github.com/go-redis/redis"
)

var Redis *redis.Client

func Redisinit() {
	Redis = redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379", // Redis地址
		Password:    "",               // Redis账号
		DB:          0,                // Redis库
		PoolSize:    5,                // Redis连接池大小
		MaxRetries:  3,                // 最大重试次数
		IdleTimeout: 10 * time.Second, // 空闲链接超时时间
	})
	_, err := Redis.Ping().Result()
	if err != nil {
		panic(err)
	}
}
