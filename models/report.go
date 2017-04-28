package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"zonst/qipai/gamehealthysrv/middlewares"

	elastic "gopkg.in/olivere/elastic.v5"
)

// Report 检测报告,用于API接口和Alert模块
type Report struct {
	Heapster string        `json:"heapster"`
	Target   string        `json:"target"`
	Success  int           `json:"success"`
	Faileds  int           `json:"faileds"`
	MaxDelay time.Duration `json:"max_delay"`
}

func init() {
	elastic.SetTraceLog(log.New(os.Stdout, "****** ", 5))
}

// Reports 报告列表
type Reports []Report

// ProbeLog 持久化文档结构，对应Elastic的gamehealthy-*模版
type ProbeLog struct {
	Timestamp time.Time     `json:"timestamp"`
	Heapster  string        `json:"heapster"`
	Target    string        `json:"target"`
	Response  string        `json:"response"`
	Elapsed   time.Duration `json:"elapsed"`
	Success   int           `json:"success"`
	Failed    int           `json:"failed"`
}

// ProbeLogs ProbeLog列表
type ProbeLogs []ProbeLog

// Validate 验证report必填字段
func (pl ProbeLog) Validate() error {
	if pl.Heapster == "" {
		return fmt.Errorf("Heapster field required")
	}
	if pl.Target == "" {
		return fmt.Errorf("Target field required")
	}
	return nil
}

// Save 保存报告
func (pls ProbeLogs) Save(ctx context.Context) error {
	conn := middlewares.GetElasticConn(ctx)
	for _, doc := range pls {
		if doc.Validate() != nil {
			continue
		}
		doc.Timestamp = time.Now()

		_, err := conn.Index().
			Index("<gamehealthy-{now/d}>").
			Type("probelog").
			BodyJson(doc).Do(ctx)
		if err != nil {
			return fmt.Errorf("save report error %v", err)
		}
	}
	return nil
}

// FetchReportsAggs 获取统计报告
func FetchReportsAggs(ctx context.Context, heapster string, last time.Duration) (Reports, error) {
	conn := middlewares.GetElasticConn(ctx)
	// 查询条件
	queryHeapster := elastic.NewTermQuery("heapster", heapster)
	queryTimestamp := elastic.NewRangeQuery("timestamp").
		Gte(time.Now().Add(-last).Format(time.RFC3339))
	boolQuery := elastic.NewBoolQuery().Filter(queryHeapster, queryTimestamp)
	// 聚集
	aggsSuccess := elastic.NewSumAggregation().Field("success")
	aggsFaileds := elastic.NewSumAggregation().Field("failed")
	aggsElapsed := elastic.NewMaxAggregation().Field("elapsed")
	aggsTarget := elastic.NewTermsAggregation().
		Field("target").OrderByTermAsc().
		SubAggregation("success", aggsSuccess).
		SubAggregation("faileds", aggsFaileds).
		SubAggregation("max_delay", aggsElapsed)

	// 最多检索3天前的数据
	result, err := conn.Search("gamehealthy-*").
		Type("probelog").From(0).Size(1000).
		Query(boolQuery).Aggregation("target", aggsTarget).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	reports := make(Reports, 0, 1024)
	term, ok := result.Aggregations.Terms("target")
	if ok {
		for _, b := range term.Buckets {
			rp := Report{
				Heapster: heapster,
				Target:   b.Key.(string),
			}

			if success, ok := b.Sum("success"); ok {
				rp.Success = int(*success.Value)
			}
			if faileds, ok := b.Sum("faileds"); ok {
				rp.Faileds = int(*faileds.Value)
			}
			if maxDelay, ok := b.Sum("max_delay"); ok {
				rp.MaxDelay = time.Duration(*maxDelay.Value)
			}
			reports = append(reports, rp)
		}
	}
	return reports, nil
}
