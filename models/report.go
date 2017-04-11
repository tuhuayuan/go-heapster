package models

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
