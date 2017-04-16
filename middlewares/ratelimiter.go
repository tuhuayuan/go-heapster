package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"

	"zonst/qipai-golang-libs/httputil"
	"zonst/qipai/logagent/utils"
)

type rateLimiterContextKey string

// 上下文key
const (
	RateLimiterContextKey rateLimiterContextKey = "_ratelimiter_"
)

// RateLimiter 频率控制接口
type RateLimiter interface {
	TryAccept(keys []string, every time.Duration, times int) bool
	Accept(keys []string, every time.Duration, times int)
}

// 依赖Redis的TTL实现
type redisRateLimiter struct {
	pool *redis.Pool
}

// NewRedisRateLimiter 新建一个Ratelimiter
func NewRedisRateLimiter(host string, passwd string, db int) RateLimiter {
	pool := &redis.Pool{
		MaxIdle:     16,
		MaxActive:   16,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			ops := []redis.DialOption{
				redis.DialConnectTimeout(time.Second * 5),
				redis.DialDatabase(db),
			}
			if passwd != "" {
				ops = append(ops, redis.DialPassword(passwd))
			}
			conn, err := redis.Dial("tcp", host, ops...)
			if err != nil {
				utils.Logger.Warnf("Redis output dial redis error %q", err)
				return nil, err
			}
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	r := &redisRateLimiter{
		pool: pool,
	}
	return r
}

func (limiter *redisRateLimiter) TryAccept(keys []string, every time.Duration, times int) bool {
	conn := limiter.pool.Get()
	defer conn.Close()

	milliseconds := every.Nanoseconds() / 1000000

	for _, k := range keys {
		limitKey := fmt.Sprintf("_ratelimit_key_%s_%d_%d", k, milliseconds, times)
		count, err := redis.Int(conn.Do("GET", limitKey))
		if err != nil && err != redis.ErrNil {
			log.Printf("rate limiter bypassed because driver error %s", err)
		}
		if err == redis.ErrNil {
			conn.Do("SET", limitKey, 1, "PX", milliseconds)
		} else if count < times {
			conn.Do("INCR", limitKey)
		} else {
			// Forbidden
			return true
		}
	}
	return false
}

func (limiter *redisRateLimiter) Accept(keys []string, every time.Duration, times int) {
	conn := limiter.pool.Get()
	defer conn.Close()

	milliseconds := every.Nanoseconds() / 1000000

	for _, k := range keys {
		limitKey := fmt.Sprintf("_ratelimit_key_%s_%d_%d", k, milliseconds, times)
		count, err := redis.Int(conn.Do("GET", limitKey))
		if err != nil && err != redis.ErrNil {
			log.Printf("rate limiter bypassed because driver error %s", err)
		}
		if err == redis.ErrNil {
			conn.Do("SET", limitKey, 1, "PX", milliseconds)
		}
		if count < times {
			conn.Do("INCR", limitKey)
		} else {
			for {
				ttl, err := redis.Int64(conn.Do("PTTL", limitKey))
				if err != nil || ttl <= 0 {
					break
				}
				time.Sleep(time.Duration(ttl) * time.Millisecond)
			}
			// 等待结束算访问一次
			conn.Do("SET", limitKey, 1, "PX", milliseconds)
		}
	}
}

// RateLimitEvery 每X时间Y次
func RateLimitEvery(every time.Duration, times int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		limiter := ctx.Value(RateLimiterContextKey).(RateLimiter)
		rawKeys := ctx.Value("_ratelimit_key").(map[string]interface{})
		keys := make([]string, 0, len(rawKeys))

		for k := range rawKeys {
			keys = append(keys, k)
		}
		if limiter.TryAccept(keys, every, times) {
			w.WriteHeader(403)
		} else {
			httputil.Next(ctx)
		}
	}
}

// RateLimitKey 限制的Key
func RateLimitKey(keys ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			keyMap = make(map[string]interface{})
		)
		ctx := r.Context()
		for _, k := range keys {
			keyMap[k] = nil
		}
		httputil.WithValue(ctx, "_ratelimit_key", keyMap)
		httputil.Next(ctx)
	}
}

// RateLimitByIP 用客户端IP最为Key
func RateLimitByIP() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		RateLimitKey(r.RemoteAddr)(w, r)
	}
}

// WithRateLimiter 装箱
func WithRateLimiter(parent context.Context, host string, passwd string, db int) context.Context {
	limiter := NewRedisRateLimiter(host, passwd, db)
	return context.WithValue(parent, RateLimiterContextKey, limiter)
}

// GetRateLimiter 拆箱
func GetRateLimiter(ctx context.Context) RateLimiter {
	return ctx.Value(RateLimiterContextKey).(RateLimiter)
}

// RateLimiterHandler 中间件
func RateLimiterHandler(host string, passwd string, db int) http.HandlerFunc {
	limiter := NewRedisRateLimiter(host, passwd, db)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		httputil.WithValue(ctx, RateLimiterContextKey, limiter)
		httputil.Next(ctx)
	}
}
