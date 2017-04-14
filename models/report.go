package models

import (
	"context"
	"encoding/json"
	"fmt"

	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/garyburd/redigo/redis"
)

// 报告字段
const (
	ReportNameType     LabelName = "_report_type_"
	ReportNameTarget   LabelName = "_report_target_"
	ReportNameFor      LabelName = "_report_for_"
	ReportNameResult   LabelName = "_report_result_"
	ReportNameErrs     LabelName = "_report_errs_"
	ReportNameSubtitle LabelName = "_report_subtitile_"
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
	if rp.Labels[ReportNameResult] == "" || !rp.Labels[ReportNameResult].IsValid() {
		return fmt.Errorf("ReportNameResult field required")
	}
	return nil
}

// Save 保存报告
func (rps Reports) Save(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	for _, rp := range rps {
		if err := rp.Validate(); err != nil {
			return err
		}
		data, err := json.Marshal(rp)
		if err != nil {
			return err
		}
		reportKey := fmt.Sprintf("gamehealthy_report_%s", rp.Labels[ReportNameFor])

		if _, err := conn.Do("HSET", reportKey, rp.Labels[ReportNameTarget], data); err != nil {
			return err
		}
	}

	return nil
}

// FetchReportsFor 获取制定目标的报告
func FetchReportsFor(ctx context.Context, reportFor LabelValue) (Reports, error) {
	var reports Reports

	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	all, err := redis.ByteSlices(conn.Do("HGETALL", fmt.Sprintf("gamehealthy_report_%s", reportFor)))
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(all); i += 2 {
		data := all[i+1]
		rp := &Report{}
		if err := json.Unmarshal(data, rp); err != nil {
			continue
		}
		reports = append(reports, *rp)
	}
	return reports, nil
}
