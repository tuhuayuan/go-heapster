package detectors

import (
	"context"
	"fmt"
	"os"
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

func (td *testDetector) probe(ctx context.Context) models.ProbeLogs {
	select {
	case <-ctx.Done():
		fmt.Println("cancel")
		return nil
	case <-time.After(time.Duration((rand.Int() % 3)) * time.Second):
		fmt.Println("ok")
	}
	return nil
}

func TestDetectLooper(t *testing.T) {
	var (
		hp = models.Heapster{
			ID:        "123456",
			Name:      "test_manager",
			Type:      models.CheckType("test"),
			Port:      5200,
			Timeout:   2 * time.Second,
			Interval:  3 * time.Second,
			Threshold: 3,
		}
	)

	ctx := middlewares.WithRedisConn(context.Background(), "localhost:6379", "", 1)
	ctx = middlewares.WithLogger(ctx, 5, os.Stdout)
	ctx = middlewares.WithElasticConn(ctx, []string{"http://10.0.10.46:9200"}, "", "")

	looper, err := NewDetectLooper(ctx, hp)
	assert.NoError(t, err)

	looper.Run()
	time.Sleep(15 * time.Second)
	looper.Stop()
}
