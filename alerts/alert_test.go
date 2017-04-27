package alerts

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
)

func TestDefaultAlert(t *testing.T) {
	hp := models.Heapster{
		ID:        "testreport1",
		Name:      "test_manager",
		Type:      models.CheckType("test"),
		Port:      5200,
		Timeout:   2 * time.Second,
		Interval:  3 * time.Second,
		Threshold: 3,
	}
	ctx := middlewares.WithRedisConn(context.Background(), "localhost:6379", "", 1)
	ctx = middlewares.WithLogger(ctx, 5, os.Stdout)
	ctx = middlewares.WithElasticConn(ctx, []string{"http://10.0.10.46:9200"}, "", "")

	assert.NoError(t, hp.Save(ctx))

	al, err := NewAlert(ctx, hp)
	assert.NoError(t, err)

	assert.NoError(t, al.TurnOn())
	time.Sleep(600 * time.Second)
}
