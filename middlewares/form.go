package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"zonst/qipai-golang-libs/httputil"
)

type bindBodyContextKey string

const (
	bindBodyContextInstance bindBodyContextKey = "_bindbody_instance"
	bindBodyContextError    bindBodyContextKey = "_bindbody_error"
)

// GetBindBody 获取绑定
func GetBindBody(ctx context.Context) (interface{}, error) {
	err, ok := ctx.Value(bindBodyContextError).(error)
	if ok {
		return nil, err
	}
	return ctx.Value(bindBodyContextInstance), nil
}

// BindBody 绑定请求数据
func BindBody(ptrStruct interface{}) http.HandlerFunc {
	structType := reflect.TypeOf(ptrStruct)
	if structType.Kind() != reflect.Ptr || structType.Elem().Kind() != reflect.Struct {
		panic("argument ptrStruct must be point of struct")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err error
		)
		ctx := httputil.GetContext(r)
		structBody := reflect.New(structType.Elem())
		contentType := r.Header.Get("Content-Type")
		// 支持json、form、querystring三种方式
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if strings.Contains(contentType, "json") {
				var bodyData []byte
				bodyData, err = ioutil.ReadAll(r.Body)
				if err == nil {
					if err = json.Unmarshal(bodyData, structBody.Interface()); err == nil {
						err = httputil.Validate(structBody.Interface())
					}
				}
			} else if strings.Contains(contentType, "form-urlencoded") {
				err = httputil.UnpackRequest(r, structBody.Interface())
			} else {
				err = errors.New("content type not supported")
			}
		} else {
			err = httputil.UnpackURLValues(r.URL.Query(), structBody.Interface())
		}
		// 标示绑定数据错误
		if err != nil {
			ctx.Set(bindBodyContextError, err)
		} else {
			ctx.Set(bindBodyContextInstance, structBody.Interface())
		}
		ctx.Next()
	}
}
