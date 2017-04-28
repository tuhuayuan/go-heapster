package models

import (
	"fmt"
	"testing"
	"zonst/qipai/gamehealthysrv/middlewares"

	"time"

	"context"

	"github.com/stretchr/testify/assert"
)

func TestReportSave(t *testing.T) {
	ctx := middlewares.WithElasticConn(context.Background(), []string{"http://10.0.10.46:9200"}, "", "")

	rps := ProbeLogs{
		ProbeLog{
			Heapster: "test_heapster_model",
			Target:   "10.0.10.46:10000",
			Success:  1,
			Response: "",
			Elapsed:  1 * time.Millisecond,
		},
		ProbeLog{
			Heapster: "test_heapster_model",
			Target:   "10.0.10.47:10000",
			Failed:   1,
			Response: "Connection refused",
			Elapsed:  0,
		},
	}

	assert.NoError(t, rps.Save(ctx))
}

func TestFetchReportsAggs(t *testing.T) {
	ctx := middlewares.WithElasticConn(context.Background(), []string{"http://10.0.10.46:9200"}, "", "")

	fmt.Println(FetchReportsAggs(ctx, "test_heapster_model", time.Now().Add(-100*time.Minute)))
}
