package detectors

import (
	"context"
	"fmt"
	"testing"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"time"

	"math/rand"

	"github.com/stretchr/testify/assert"
)

func init() {
	registCreator("test", func(ctx context.Context, dm DetectorManager, hp models.Heapster) (detector, error) {
		return &testDetector{}, nil
	})
}

type testDetector struct {
}

func (td *testDetector) plumb(ctx context.Context) (models.Reports, error) {
	select {
	case <-ctx.Done():
		fmt.Println("cancel")
		return nil, fmt.Errorf("plumb timeout")
	case <-time.After(time.Duration((rand.Int() % 4)) * time.Second):
		fmt.Println("ok")
	}
	return nil, nil
}

func TestManager(t *testing.T) {
	var (
		timeout, _  = models.ParseDuration("2s")
		interval, _ = models.ParseDuration("2s")

		hp = models.Heapster{
			ID:        "123456",
			Name:      "test_manager",
			Type:      models.CheckType("test"),
			Port:      5200,
			Timeout:   timeout,
			Interval:  interval,
			Healthy:   3,
			UnHealthy: 3,
		}
	)

	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)

	m := NewManager(ctx)
	looper, err := m.CreateLooper(hp)
	assert.NoError(t, err)

	looper.Run()
	time.Sleep(15 * time.Second)

	m.DropLooper(hp)
	m.Shutdown()
}
