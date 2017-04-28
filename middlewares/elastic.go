package middlewares

import (
	"context"
	"net/http"

	"zonst/qipai-golang-libs/httputil"

	elastic "gopkg.in/olivere/elastic.v5"
)

type elasticContextKey string

// 上下文key
const (
	elasticContextConn elasticContextKey = "_elastic_"
)

// GetElasticConn 获取ES连接
func GetElasticConn(ctx context.Context) *elastic.Client {
	return ctx.Value(elasticContextConn).(*elastic.Client)
}

// WithElasticConn 获取带连接的上下文
func WithElasticConn(parent context.Context, hosts []string, user string, password string) context.Context {
	// 使用简单Client，暂时忽略错误
	conn, _ := elastic.NewSimpleClient(
		elastic.SetURL(hosts...),
		elastic.SetBasicAuth(user, password),
	)
	return context.WithValue(parent, elasticContextConn, conn)
}

// ElasticConnHandler 中间件：管理elasticsearch连接
func ElasticConnHandler(hosts []string, user string, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		conn, err := elastic.NewSimpleClient(
			elastic.SetURL(hosts...),
			elastic.SetBasicAuth(user, password),
		)
		if err != nil {
			panic(err)
		}
		httputil.WithValue(ctx, elasticContextConn, conn)
		httputil.Next(ctx)
	}
}
