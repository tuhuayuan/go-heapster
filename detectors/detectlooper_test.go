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
	registCreator("test", func(ctx context.Context, hp models.Heapster) (detector, error) {
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

func TestDetectLooper(t *testing.T) {
	var (
		hp = models.Heapster{
			ID:        "123456",
			Name:      "test_manager",
			Type:      models.CheckType("test"),
			Port:      5200,
			Timeout:   1 * time.Second,
			Interval:  5 * time.Second,
			Threshold: 3,
		}
	)

	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)

	looper, err := NewDetectLooper(ctx, hp)
	assert.NoError(t, err)

	looper.Run()
	time.Sleep(15 * time.Second)
	looper.Stop()
}
