package middlewares

import (
	"net/http"
	"time"
	"zonst/qipai-golang-libs/httputil"

	"github.com/Sirupsen/logrus"
)

// Logger 请求日志中间件
func Logger(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			errno int
			ok    bool
		)
		ctx := httputil.GetContext(r)
		start := time.Now()
		ctx.Next()
		errno, ok = ctx.Value("errno").(int)
		if ok && errno != 0 {
			logger.Errorf("[%s] [%s] from [%s] used %.3fsecs Content-Type [%s] ErrNo[%d] ",
				r.Method, r.URL.Path, r.RemoteAddr, time.Now().Sub(start).Seconds(), r.Header.Get("Content-Type"), errno)
		} else {
			logger.Infof("[%s] [%s] from [%s] used %.3fsecs",
				r.Method, r.URL.Path, r.RemoteAddr, time.Now().Sub(start).Seconds())
		}
	}
}
