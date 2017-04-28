package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"

	"zonst/qipai-golang-libs/httputil"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/stretchr/testify/assert"
)

func TestFetchReports(t *testing.T) {
	ctx := middlewares.WithLogger(context.Background(), 5, os.Stdout)
	ctx = middlewares.WithElasticConn(ctx, []string{"http://10.0.10.46:9200"}, "", "")
	ctx = httputil.WithHTTPContext(ctx)
	httputil.Use(ctx, middlewares.RedisConnHandler("localhost:6379", "", 1))
	httputil.Use(ctx, middlewares.ElasticConnHandler([]string{"http://10.0.10.46:9200"}, "", ""))
	httputil.Use(ctx, middlewares.LoggerHandler(5, os.Stdout))

	pls := models.ProbeLogs{
		models.ProbeLog{
			Heapster: "testheapster",
			Target:   "10.0.10.46:10000",
			Success:  1,
		},
		models.ProbeLog{
			Heapster: "testheapster",
			Target:   "10.0.10.46:10000",
			Success:  1,
		},
		models.ProbeLog{
			Heapster: "testheapster",
			Target:   "10.0.10.46:10000",
			Success:  1,
		},
	}

	assert.NoError(t, pls.Save(ctx))

	handler := httputil.HandleFunc(ctx,
		middlewares.BindBody(&FetchReportReq{}),
		FetchReportHandler)

	req := httptest.NewRequest("GET", "/?heapster=testheapster&last=120", nil)
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
