package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"zonst/qipai-golang-libs/httputil"

	"os"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)

	handler1 := httputil.HandleFunc(ctx, LoggerHandler(5, os.Stdout),
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			httputil.WithValue(ctx, "errno", -1)
			w.WriteHeader(400)
		})
	handler2 := httputil.HandleFunc(ctx, LoggerHandler(5, os.Stdout),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

	req := httptest.NewRequest("GET", "/", bytes.NewReader([]byte{}))
	resp := httptest.NewRecorder()
	handler1(resp, req)
	assert.Equal(t, 400, resp.Code)

	resp = httptest.NewRecorder()
	handler2(resp, req)
	assert.Equal(t, 200, resp.Code)
}
