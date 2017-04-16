package middlewares

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"zonst/qipai-golang-libs/httputil"
)

func TestErrorHandler(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)
	handler := httputil.HandleFunc(ctx,
		func(w http.ResponseWriter, r *http.Request) {
			ErrorWrite(w, 401, 100, fmt.Errorf(""))
		})

	req := httptest.NewRequest("GET", "/", bytes.NewReader([]byte{}))
	resp := httptest.NewRecorder()
	handler(resp, req)

	assert.Equal(t, 401, resp.Code)
}
