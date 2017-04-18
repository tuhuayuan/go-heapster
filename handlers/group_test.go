package handlers

import (
	"fmt"
	"testing"

	"bytes"
	"zonst/qipai-golang-libs/httputil"
	"zonst/qipai/gamehealthysrv/middlewares"

	"net/http/httptest"

	"io/ioutil"

	"time"

	"encoding/json"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/stretchr/testify/assert"
)

var testingID string

func TestCreateGroupHandler(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)
	httputil.Use(ctx, middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := httputil.HandleFunc(ctx,
		middlewares.BindBody(&CreateGroupReq{}),
		CreateGroupHandler)

	data := []byte(`
     {
        "name": "监控组例子2",
        "endpoints": [
           "118.89.100.129",	
           "118.89.100.130",	
           "118.89.100.135",	
           "118.89.100.76",
           "118.89.100.77",
           "118.89.100.78",
           "118.89.100.79",
           "118.89.100.122",	
           "118.89.100.86",
           "118.89.100.81",
           "118.89.100.82",
           "118.89.100.85",
           "118.89.99.45",
           "118.89.100.123",	
           "118.89.100.139",	
           "118.89.100.128",	
           "118.89.100.126",	
           "118.89.100.137",	
           "118.89.100.116",	
           "118.89.100.132"
        ],
        "excluded": [
            "118.89.100.132"
        ]
    }
    `)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Add("Content-Type", "json")
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	g := models.Group{}
	assert.NoError(t, json.Unmarshal(body, &g))
	testingID = string(g.ID)
	fmt.Println(testingID)
	fmt.Println(string(body))
}

func TestFetchGroupHandler(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)
	httputil.Use(ctx, middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := httputil.HandleFunc(ctx,
		middlewares.BindBody(&FetchGroupReq{}),
		FetchGroupHandler)

	req := httptest.NewRequest("GET", "/?id="+testingID, nil)
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func TestUpdateGroupHandler(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)
	httputil.Use(ctx, middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := httputil.HandleFunc(ctx,
		middlewares.BindBody(&UpdateGroupReq{}),
		UpdateGroupHandler)
	name := time.Now().String()
	data := []byte(`
     {
        "id": "` + testingID + `",
        "name": "` + name + `",
        "endpoints": [
           "118.89.100.129",	
           "118.89.100.130",	
           "118.89.100.135",	
           "118.89.100.76",
           "118.89.100.77"
        ],
        "excluded": [
            "118.89.100.132"
        ]
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

func TestDeleteGroupHandler(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)
	httputil.Use(ctx, middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	handler := httputil.HandleFunc(ctx,
		middlewares.BindBody(&DeleteGroupReq{}),
		DeleteGroupHandler)
	data := []byte(`
    {
        "id": "` + testingID + `"
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
