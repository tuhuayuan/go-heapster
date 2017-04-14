package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"zonst/qipai-golang-libs/httputil"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	ctx := httputil.NewHTTPContext()
	handler := ctx.HandleFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ErrorWrite(w, 401, 100, fmt.Errorf(""))
		})

	req := httptest.NewRequest("GET", "/", bytes.NewReader([]byte{}))
	resp := httptest.NewRecorder()
	handler(resp, req)

	assert.Equal(t, 401, resp.Code)
}
