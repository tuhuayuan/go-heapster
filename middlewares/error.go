package middlewares

import (
	"encoding/json"
	"net/http"
)

// APIReponseError API通用错误结构
type APIReponseError struct {
	ErrorNo  int    `json:"errno"`
	ErrorMsg string `json:"errmsg"`
}

// ErrorWrite 写入HTTPAPI错误
func ErrorWrite(w http.ResponseWriter, status int, errno int, err error) {
	resp := &APIReponseError{
		ErrorNo:  errno,
		ErrorMsg: err.Error(),
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(status)
	w.Write(raw)
}

// ErrorWriteOK 写入HTTPAPI成功
func ErrorWriteOK(w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write([]byte(`{"errno":0,"errmsg":"ok"}`))
}
