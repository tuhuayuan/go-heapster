package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"zonst/qipai-golang-libs/httputil"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/stretchr/testify/assert"
)

var notifierTestID string

func TestCreateNotifier(t *testing.T) {
	ctx := httputil.NewHTTPContext()
	ctx.Use(middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := ctx.HandleFunc(
		middlewares.BindBody(&CreateNotifierReq{}),
		CreateNotifierHandler)
	data := []byte(`
    {
            "type": "sms",
            "config": {
                "type": "unicom",
                "sp": "103905",
                "username": "zz_sj",
                "password": "www.zonst.org",
                "targets": [
                    "13879156403"
                ]
            }
        }
    `)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Add("Content-Type", "json")
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	hn := models.HeapsterNotifier{}
	assert.NoError(t, json.Unmarshal(body, &hn))
	notifierTestID = string(hn.ID)
	fmt.Println(notifierTestID)
}

func TestUpdateNotifier(t *testing.T) {
	ctx := httputil.NewHTTPContext()
	ctx.Use(middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := ctx.HandleFunc(
		middlewares.BindBody(&DeleteNotifierReq{}),
		DeleteNotifierHandler)
	data := []byte(`
    {
        "id": "` + notifierTestID + `",
         "type": "sms",
         "config": {
             "type": "unicom",
             "sp": "103905",
             "username": "zz_sj",
             "password": "www.zonst.org",
             "targets": [
                 "13879156403",
                 "13607080910"
             ]
         }
    }
    `)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Add("Content-Type", "json")
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func TestFetchNotifier(t *testing.T) {
	ctx := httputil.NewHTTPContext()
	ctx.Use(middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := ctx.HandleFunc(
		middlewares.BindBody(&FetchNotifierReq{}),
		FetchNotifierHandler)

	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func TestDeleteNotifier(t *testing.T) {
	ctx := httputil.NewHTTPContext()
	ctx.Use(middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := ctx.HandleFunc(
		middlewares.BindBody(&DeleteNotifierReq{}),
		DeleteNotifierHandler)
	data := []byte(`
    {
        "id": "` + notifierTestID + `"
    }
    `)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Add("Content-Type", "json")
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
