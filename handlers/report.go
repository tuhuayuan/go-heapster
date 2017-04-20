package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
)

// FetchReportReq 获取报告请求
type FetchReportReq struct {
	HeaspterID string        `json:"heapster" http:"heapster"`
	LastMinute time.Duration `json:"last" http:"last"`
}

// FetchReportHandler 获取
func FetchReportHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*FetchReportReq)

	rps, err := models.FetchErrorReports(ctx, models.LabelValue(req.HeaspterID), req.LastMinute*time.Minute)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	if len(rps) == 0 {
		middlewares.ErrorWrite(w, 200, 3, fmt.Errorf("not found"))
		return
	}
	data, err := json.Marshal(rps)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 4, err)
	}
	w.WriteHeader(200)
	w.Write(data)
}
