package middlewares

import (
	"context"
	"io"
	"net/http"
	"time"

	"zonst/qipai-golang-libs/httputil"

	"github.com/Sirupsen/logrus"
)

type loggerContextKey string

const (
	loggerContextName loggerContextKey = "_logger_"
)

// GetLogger 获取上下文的日志器
func GetLogger(ctx context.Context) *logrus.Logger {
	return ctx.Value(loggerContextName).(*logrus.Logger)
}

// WithLogger 获取含日志的上下文
func WithLogger(parent context.Context, level int, output io.Writer) context.Context {
	logger := logrus.New()
	logger.Level = logrus.Level(level)
	logger.Out = output
	return context.WithValue(parent, loggerContextName, logger)
}

// LoggerHandler 请求日志中间件
func LoggerHandler(level int, output io.Writer) http.HandlerFunc {
	logger := logrus.New()
	logger.Level = logrus.Level(level)
	logger.Out = output
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		httputil.WithValue(ctx, loggerContextName, logger)

		// 记录请求时间间隔
		start := time.Now()
		httputil.Next(ctx)
		logger.Infof("[%s] [%s] from [%s] used %.3fsecs",
			r.Method, r.URL.Path, r.RemoteAddr, time.Now().Sub(start).Seconds())
	}
}
