package middlewares

import (
	"context"
	"net/http"

	influxdb "github.com/influxdata/influxdb/client/v2"

	"zonst/qipai-golang-libs/httputil"
)

type influxdbContextKey string

// 上下文key
const (
	influxdbContextName   influxdbContextKey = "_influxdb_"
	influxdbContextDBName influxdbContextKey = "_influxdb_db"
)

// GetInfluxDB 获取上下文连接
func GetInfluxDB(ctx context.Context) influxdb.Client {
	return ctx.Value(influxdbContextName).(influxdb.Client)
}

// GetInfluxDBName 获取配置的DB
func GetInfluxDBName(ctx context.Context) string {
	return ctx.Value(influxdbContextDBName).(string)
}

// WithInfluxDB 返回带db连接的上下文
func WithInfluxDB(parent context.Context, host string, user string, passwd string, db string) (context.Context, error) {
	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     host,
		Username: user,
		Password: passwd,
	})
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(parent, influxdbContextName, client)
	ctx = context.WithValue(ctx, influxdbContextDBName, db)
	return ctx, nil
}

// InfluxDBHandler http中间件
func InfluxDBHandler(host string, user string, passwd string, db string) http.HandlerFunc {
	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     host,
		Username: user,
		Password: passwd,
	})
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx = context.WithValue(ctx, influxdbContextDBName, db)
		ctx = httputil.WithValue(ctx, influxdbContextName, client)
		httputil.Next(ctx)
	}
}
