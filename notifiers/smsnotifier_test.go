package notifiers

import (
	"testing"
	"time"
	"zonst/qipai/gamehealthysrv/models"

	"context"

	"zonst/qipai/gamehealthysrv/middlewares"

	"os"

	"github.com/stretchr/testify/assert"
)

func TestSMSNotifier(t *testing.T) {
	sms, err := smsNotifierCreator(models.HeapsterNotifier{
		Type: "sms",
		Config: map[string]interface{}{
			"type":     "unicom",
			"sp":       "103905",
			"username": "zz_sj",
			"password": "www.zonst.org",
			"targets":  []interface{}{"13879156403"},
		},
	})
	assert.NoError(t, err)

	ctx := middlewares.WithRateLimiter(context.Background(), "localhost:6379", "", 0)
	ctx = middlewares.WithRedisConn(ctx, "localhost:6379", "", 0)
	ctx = middlewares.WithLogger(ctx, 5, os.Stdout)

	hp := &models.Heapster{
		ID:        models.NewSerialNumber(),
		Name:      "smsnotifier_test",
		Type:      models.CheckTypeTCP,
		Port:      10000,
		Timeout:   1 * time.Second,
		Interval:  5 * time.Second,
		Threshold: 3,
	}
	assert.NoError(t, hp.Save(ctx))

	err = sms.Send(ctx, models.Report{
		Heapster: string(hp.ID),
		Target:   "10.0.10.46:10000",
		Success:  1,
	})
	assert.NoError(t, err)
}
