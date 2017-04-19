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
	var (
		hp = models.Heapster{
			ID:       "testreport1",
			Name:     "test_manager",
			Type:     models.CheckType("test"),
			Port:     5200,
			Timeout:  2 * time.Second,
			Interval: 3 * time.Second,
			Healthy:  3,
		}
	)

	ctx, err := middlewares.WithInfluxDB(context.Background(), "http://localhost:8086", "", "")
	assert.NoError(t, err)
	ctx = middlewares.WithLogger(ctx, 5, os.Stdout)
	ctx = middlewares.WithRedisConn(ctx, "localhost:6379", "", 0)

	al, err := NewAlert(ctx, hp)
	assert.NoError(t, err)

	assert.NoError(t, al.TurnOn())
	time.Sleep(600 * time.Second)
}
