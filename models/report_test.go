package models

import (
	"fmt"
	"testing"
	"zonst/qipai/gamehealthysrv/middlewares"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestReportSave(t *testing.T) {
	ctx, err := middlewares.WithInfluxDB(nil, "http://localhost:8086", "", "")
	assert.NoError(t, err)

	rps := Reports{
		Report{
			Labels: LabelSet{
				ReportNameFor:    LabelValue("testreport1"),
				ReportNameTarget: LabelValue("http://localhost:5000"),
				ReportNameResult: LabelValue(
					fmt.Sprintf("http response code 404"),
				),
			},
		},
		Report{
			Labels: LabelSet{
				ReportNameFor:    LabelValue("testreport1"),
				ReportNameTarget: LabelValue("http://localhost:5001"),
				ReportNameResult: LabelValue(
					"ok",
				),
			},
		},
	}

	assert.NoError(t, rps.Save(ctx))
}

func TestFetchErrorReports(t *testing.T) {
	ctx, err := middlewares.WithInfluxDB(nil, "http://localhost:8086", "", "")
	assert.NoError(t, err)

	rps, err := FetchErrorReports(ctx, LabelValue("testreport1"), 100*time.Minute)
	assert.NoError(t, err)
	fmt.Println(rps)
}

func TestFetchReportsAggregation(t *testing.T) {
	ctx, err := middlewares.WithInfluxDB(nil, "http://localhost:8086", "", "")
	assert.NoError(t, err)

	fmt.Println(FetchReportsAggregation(ctx, "testreport1", 100*time.Minute))
}
