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
	RateControl(keys []string, every time.Duration, times int) bool
}

// 依赖Redis的TTL实现
type redisRateLimiter struct {
	ctx  context.Context
	pool *redis.Pool
}

// NewRedisRateLimiter 新建一个Ratelimiter
func NewRedisRateLimiter(ctx context.Context, host string, passwd string, db int) RateLimiter {
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
		ctx:  ctx,
		pool: pool,
	}
	return r
}

func (limiter *redisRateLimiter) RateControl(keys []string, every time.Duration, times int) bool {
	conn := limiter.pool.Get()
	defer conn.Close()
	for _, k := range keys {
		limitKey := fmt.Sprintf("_ratelimit_key_%s_%d_%d", k, int(every.Seconds()), times)
		count, err := redis.Int(conn.Do("GET", limitKey))
		if err != nil && err != redis.ErrNil {
			log.Printf("rate limiter bypassed because driver error %s", err)
		}
		if err == redis.ErrNil {
			conn.Do("SET", limitKey, 1, "EX", every.Seconds())
		} else if count < times {
			conn.Do("INCR", limitKey)
		} else {
			// Forbidden
			return true
		}
	}
	return false
}

// RateLimitEvery 每X时间Y次
func RateLimitEvery(every time.Duration, times int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := httputil.GetContext(r)
		limiter := ctx.Value(RateLimiterContextKey).(RateLimiter)
		rawKeys := ctx.Value("_ratelimit_key").(map[string]interface{})
		keys := make([]string, 0, len(rawKeys))

		for k := range rawKeys {
			keys = append(keys, k)
		}
		if limiter.RateControl(keys, every, times) {
			w.WriteHeader(403)
		} else {
			ctx.Next()
		}
	}
}

// RateLimitKey 限制的Key
func RateLimitKey(keys ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			keyMap = make(map[string]interface{})
		)
		ctx := httputil.GetContext(r)
		for _, k := range keys {
			keyMap[k] = nil
		}
		ctx.Set("_ratelimit_key", keyMap)
		ctx.Next()
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
	limiter := NewRedisRateLimiter(context.Background(), host, passwd, db)
	return context.WithValue(parent, RateLimiterContextKey, limiter)
}

// GetRateLimiter 拆箱
func GetRateLimiter(ctx context.Context) (RateLimiter, error) {
	limiter, ok := ctx.Value(RateLimiterContextKey).(RateLimiter)
	if ok {
		return limiter, nil
	}
	return nil, fmt.Errorf("no ratelimiter in context")
}

// RateLimiterHandler 中间件
func RateLimiterHandler(host string, passwd string, db int) http.HandlerFunc {
	limiter := NewRedisRateLimiter(context.Background(), host, passwd, db)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := httputil.GetContext(r)
		ctx.Set(RateLimiterContextKey, limiter)
		ctx.Next()
	}
}
