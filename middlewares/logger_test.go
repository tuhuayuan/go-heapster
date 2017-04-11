package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"zonst/qipai-golang-libs/httputil"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	ctx := httputil.NewHTTPContext()

	handler1 := ctx.HandleFunc(Logger(logrus.New()),
		func(w http.ResponseWriter, r *http.Request) {
			ctx := httputil.GetContext(r)
			ctx.Set("errno", -1)
			w.WriteHeader(400)
		})
	handler2 := ctx.HandleFunc(Logger(logrus.New()),
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
