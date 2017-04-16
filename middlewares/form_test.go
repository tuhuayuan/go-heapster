package middlewares

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"zonst/qipai-golang-libs/httputil"

	"bytes"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

type TestFormData struct {
	GameID   int64    `http:"game_id" json:"game_id" validate:"required"`
	GameName string   `http:"game_name" json:"game_name" validate:"required"`
	GameTags []string `http:"game_tags[]" json:"game_tags"`
}

func TestBindBody(t *testing.T) {
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: "/",
		},
		Header: http.Header{
			"Content-Type": []string{"application/form-urlencoded"},
		},
		Form: url.Values{
			"game_id":     []string{"43"},
			"game_name":   []string{"中至南昌麻将"},
			"game_tags[]": []string{"麻将", "休闲"},
		},
	}

	ctx := httputil.WithHTTPContext(nil)
	handler := httputil.HandleFunc(ctx,
		BindBody(&TestFormData{}),
		func(w http.ResponseWriter, r *http.Request) {
			_, err := GetBindBody(r.Context())
			assert.NoError(t, err)
		})

	resp := httptest.NewRecorder()
	handler(resp, req)

	req = &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: "/",
		},
		Header: http.Header{
			"Content-Type": []string{"application/form-urlencoded"},
		},
		Form: url.Values{
			// 缺少game_id
			"game_name":   []string{"中至南昌麻将"},
			"game_tags[]": []string{"麻将", "休闲"},
		},
	}

	handler = httputil.HandleFunc(ctx,
		BindBody(&TestFormData{}),
		func(w http.ResponseWriter, r *http.Request) {
			_, err := GetBindBody(r.Context())
			assert.Error(t, err)
		})

	resp = httptest.NewRecorder()
	handler(resp, req)

	req = &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: "/",
		},
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`
        {
            "game_id": 1,
            "game_name": "中至麻将",
            "game_tags": ["没有ID"]
        }
        `))),
	}
	handler = httputil.HandleFunc(ctx,
		BindBody(&TestFormData{}),
		func(w http.ResponseWriter, r *http.Request) {
			_, err := GetBindBody(r.Context())
			assert.NoError(t, err)
		})

	resp = httptest.NewRecorder()
	handler(resp, req)

	req = &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: "/",
		},
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: ioutil.NopCloser(bytes.NewReader([]byte(`
        {
            "game_name": "中至麻将"
        }
        `))),
	}
	handler = httputil.HandleFunc(ctx,
		BindBody(&TestFormData{}),
		func(w http.ResponseWriter, r *http.Request) {
			_, err := GetBindBody(r.Context())
			assert.Error(t, err)
		})

	resp = httptest.NewRecorder()
	handler(resp, req)
}
