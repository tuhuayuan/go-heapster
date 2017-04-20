package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"

	"zonst/qipai/gamehealthysrv/middlewares"
)

// 报告字段
const (
	ReportNameType      LabelName = "_report_type_"
	ReportNameTarget    LabelName = "_report_target_"
	ReportNameFor       LabelName = "_report_for_"
	ReportNameResult    LabelName = "_report_result_"
	ReportNameTimestamp LabelName = "_report_timestamp"
	ReportNameSuccess   LabelName = "_report_success_"
	ReportNameSubtitle  LabelName = "_report_subtitile_"
)

// Report 检测报告
type Report struct {
	Labels LabelSet `json:"labels"`
}

// Reports 报告列表
type Reports []Report

// Validate 验证report必填字段
func (rp Report) Validate() error {
	if rp.Labels[ReportNameFor] == "" || !rp.Labels[ReportNameFor].IsValid() {
		return fmt.Errorf("ReportNameFor field required")
	}
	if rp.Labels[ReportNameTarget] == "" || !rp.Labels[ReportNameTarget].IsValid() {
		return fmt.Errorf("ReportNameTarget field required")
	}
	return nil
}

// Save 保存报告
func (rps Reports) Save(ctx context.Context) error {
	client := middlewares.GetInfluxDB(ctx)
	defer client.Close()

	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  "gamehealthy",
		Precision: "ns",
	})
	if err != nil {
		return err
	}

	for _, rp := range rps {
		if rp.Validate() != nil {
			continue
		}
		tags := map[string]string{
			"heapster_id":   string(rp.Labels[ReportNameFor]),
			"report_target": string(rp.Labels[ReportNameTarget]),
		}
		fields := make(map[string]interface{}, 2)
		if string(rp.Labels[ReportNameResult]) == "ok" {
			fields["success"] = 1.0
		} else {
			fields["success"] = -1.0
		}

		point, err := influxdb.NewPoint(
			"report",
			tags,
			fields,
			time.Now(),
		)
		if err != nil {
			continue
		}
		bp.AddPoint(point)
	}
	err = client.Write(bp)
	if err != nil {
		return err
	}
	return nil
}

// FetchErrorReports 获取制定目标的报告
func FetchErrorReports(ctx context.Context, reportFor LabelValue, last time.Duration) (Reports, error) {
	var reports Reports

	client := middlewares.GetInfluxDB(ctx)
	defer client.Close()

	req := influxdb.NewQueryWithParameters(
		"SELECT success FROM report WHERE time>=$last and heapster_id=$heapster and success <= 0 group by report_target",
		"gamehealthy",
		"RFC3339",
		map[string]interface{}{
			"last":     time.Now().Add(-last),
			"heapster": string(reportFor),
		})
	resp, err := client.Query(req)
	if err != nil {
		return nil, err
	}

	for _, row := range resp.Results[0].Series {
		for _, val := range row.Values {
			rp := Report{
				Labels: LabelSet{
					ReportNameFor: reportFor,
				},
			}
			rp.Labels[ReportNameTarget] = LabelValue(row.Tags["report_target"])
			rp.Labels[ReportNameTimestamp] = LabelValue(val[0].(json.Number))
			rp.Labels[ReportNameResult] = "error"

			reports = append(reports, rp)
		}
	}
	return reports, nil
}

// FetchReportsAggregation 获取
func FetchReportsAggregation(ctx context.Context, reportFor LabelValue, last time.Duration) (Reports, error) {
	var reports Reports

	client := middlewares.GetInfluxDB(ctx)
	defer client.Close()

	req := influxdb.NewQueryWithParameters(
		"SELECT SUM(success) AS success FROM report WHERE time>=$last and heapster_id=$heapster group by report_target",
		"gamehealthy",
		"RFC3339",
		map[string]interface{}{
			"last":     time.Now().Add(-last),
			"heapster": string(reportFor),
		})
	resp, err := client.Query(req)
	if err != nil {
		return nil, err
	}
	for _, row := range resp.Results[0].Series {
		rp := Report{
			Labels: LabelSet{
				ReportNameFor:    reportFor,
				ReportNameTarget: LabelValue(row.Tags["report_target"]),
			},
		}
		rp.Labels[ReportNameTimestamp] = LabelValue(row.Values[0][0].(json.Number))
		success, ok := row.Values[0][1].(json.Number)
		if !ok {
			success = "0.0"
		}
		rp.Labels[ReportNameSuccess] = LabelValue(success)
		reports = append(reports, rp)
	}
	return reports, nil
}
