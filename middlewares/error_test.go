package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"zonst/qipai-golang-libs/httputil"

	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	ctx := httputil.NewHTTPContext()
	handler := ctx.HandleFunc(ErrorHandler(),
		func(w http.ResponseWriter, r *http.Request) {
			SetJsonError(r, 401, 100, nil)
		})

	req := httptest.NewRequest("GET", "/", bytes.NewReader([]byte{}))
	resp := httptest.NewRecorder()
	handler(resp, req)

	assert.Equal(t, 401, resp.Code)
}
