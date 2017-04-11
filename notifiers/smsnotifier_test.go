package notifiers

import (
	"testing"
	"zonst/qipai/gamehealthysrv/models"

	"context"

	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestSMSNotifier(t *testing.T) {
	sms, err := smsNotifierCreator(models.Notifier{
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
	err = sms.Send(ctx, models.Report{
		Labels: models.LabelSet{
			models.ReportNameFor: "【游戏测试服务器】",
		},
	})
	assert.NoError(t, err)
	err = sms.Send(ctx, models.Report{
		Labels: models.LabelSet{
			models.ReportNameFor: "【游戏测试服务器】",
		},
	})
	assert.Error(t, err)
}
