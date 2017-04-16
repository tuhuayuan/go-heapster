package middlewares

import (
	"context"
	"net/http"
	"time"
	"zonst/qipai/logagent/utils"

	"zonst/qipai-golang-libs/httputil"

	"github.com/garyburd/redigo/redis"
)

// 消除工具警告
type redisContext string

const (
	redisContextLabel redisContext = "_redis_pool_"
)

// GetRedisConn 获取Redis连接，需要自己释放
func GetRedisConn(ctx context.Context) redis.Conn {
	pool := ctx.Value(redisContextLabel).(*redis.Pool)
	return pool.Get()
}

// WithRedisConn 获取带连接的上下文
func WithRedisConn(parent context.Context, host string, passwd string, db int) context.Context {
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

	ctx := context.WithValue(parent, redisContextLabel, pool)
	return ctx
}

// RedisConnHandler Redis连接中间件, 全局插件不要单独使用
func RedisConnHandler(host string, passwd string, db int) http.HandlerFunc {
	redisCtx := WithRedisConn(context.Background(), host, passwd, db)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		httputil.WithValue(ctx, redisContextLabel, redisCtx.Value(redisContextLabel))
		httputil.Next(ctx)
	}
}
