package notifiers

import (
	"testing"
	"zonst/qipai/gamehealthysrv/models"

	"context"

	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestSMSNotifier(t *testing.T) {
	sms, err := smsNotifierCreator(models.HeapsterNotifier{
		Type: "sms",
		Config: map[string]interface{}{
			"sms": map[string]interface{}{
				"type":     "unicom",
				"sp":       "103905",
				"username": "zz_sj",
				"password": "www.zonst.org",
				"targets":  []string{"13879156403", "15507911970"},
			},
		},
	})
	assert.NoError(t, err)
	ctx := middlewares.WithRateLimiter(context.Background(), "0.0.0.0:6379", "", 8)
	err = sms.Send(ctx, models.Reports{
		models.Report{
			Labels: models.LabelSet{
				models.ReportNameFor:    "游戏测试服务器",
				models.ReportNameTarget: "10.0.10.46:10000",
				models.ReportNameResult: "timeout",
			},
		},
		models.Report{

			Labels: models.LabelSet{
				models.ReportNameFor:    "游戏测试服务器",
				models.ReportNameTarget: "10.0.10.47:10000",
				models.ReportNameResult: "ok",
			},
		},
	})
	assert.NoError(t, err)
}
