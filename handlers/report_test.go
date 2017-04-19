package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"zonst/qipai-golang-libs/httputil"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/stretchr/testify/assert"
)

func TestFetchReports(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)
	httputil.Use(ctx, middlewares.RedisConnHandler("0.0.0.0:6379", "", 9))
	httputil.Use(ctx, middlewares.InfluxDBHandler("http://localhost:8086", "", ""))

	rp := models.Report{
		Labels: models.LabelSet{
			models.ReportNameFor:    "test_heapster",
			models.ReportNameTarget: "http://localhost",
			models.ReportNameResult: "ok",
		},
	}
	rps := models.Reports{
		rp,
	}
	outCtx, err := middlewares.WithInfluxDB(nil, "http://localhost:8086", "", "")
	assert.NoError(t, err)
	assert.NoError(t, rps.Save(outCtx))

	handler := httputil.HandleFunc(ctx,
		middlewares.BindBody(&FetchReportReq{}),
		FetchReportHandler)

	req := httptest.NewRequest("GET", "/?heapster=test_heapster", nil)
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
