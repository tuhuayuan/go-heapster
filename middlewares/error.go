package middlewares

import (
	"encoding/json"
	"net/http"
	"zonst/qipai-golang-libs/httputil"
)

// APIReponseError 通用错误结构
type APIReponseError struct {
	Status   int    `json:"-"`
	ErrorNo  int    `json:"errno"`
	ErrorMsg string `json:"errmsg"`
}

// SetJsonError 快速写入错误
func SetJsonError(r *http.Request, status int, errno int, err error) {
	ctx := httputil.GetContext(r)
	apiErr := ctx.Value("error").(*APIReponseError)
	apiErr.Status = status
	if err != nil {
		apiErr.ErrorMsg = err.Error()
	}
	apiErr.ErrorNo = errno
}

// SetJsonOk 返回API ok
func SetJsonOk(w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write([]byte(`{"errno":0,"errmsg":"ok"}`))
}

// ErrorHandler API错误处理插件
func ErrorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiErr := &APIReponseError{}
		ctx := httputil.GetContext(r)
		ctx.Set("error", apiErr)
		ctx.Next()
		if apiErr.ErrorNo != 0 {
			raw, err := json.Marshal(apiErr)
			if err != nil {
				w.WriteHeader(500)
			} else {
				w.WriteHeader(apiErr.Status)
				w.Write(raw)
			}
			ctx.Set("errno", apiErr.ErrorNo)
		}
	}
}
