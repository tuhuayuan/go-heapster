package models

import (
	"context"
	"fmt"
	"testing"
	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestReportSave(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)

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

func TestReportFetch(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)

	rps, err := FetchReportsFor(ctx, LabelValue("testreport1"))
	assert.NoError(t, err)
	fmt.Println(rps[0])
}
