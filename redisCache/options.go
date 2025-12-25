package redisCache

import (
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptions 实现Options接口
type RedisOptions struct {
	Opt  *redis.Options `json:"Opt"`
	Base BaseOptions
}
type FailoverOptions struct {
	Opt  *redis.FailoverOptions `json:"Opt"`
	Base BaseOptions
}

type BaseOptions struct {
	Sign              string `json:"Sign"` //签名
	Online            atomic.Bool
	ReconnectNum      int           `json:"ReconnectNum"`      //重连次数
	ReconnectInterval time.Duration `json:"ReconnectInterval"` //重连间隔时间
}
