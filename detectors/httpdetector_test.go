package detectors

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/stretchr/testify/assert"
)

func WithHTTPTarget(ctx context.Context) context.Context {
	endCtx, callDone := context.WithCancel(context.Background())

	go func() {
		l, err := net.Listen("tcp", "0.0.0.0:5200")
		if err != nil {
			panic(err)
		}
		fmt.Println("server start.")
		go func() {
			http.Serve(l, nil)
			fmt.Println("server down.")
			callDone()
		}()

		select {
		case <-ctx.Done():
			l.Close()
			return
		}
	}()
	return endCtx
}

func TestHTTPPlumb(t *testing.T) {
	var (
		timeout, _ = models.ParseDuration("2s")
	)

	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)
	g1 := models.NewGroup(
		"test_local",
		models.Endpoints{
			models.Endpoint("127.0.0.1"),
			models.Endpoint("10.0.0.1/28"),
		},
		nil, models.GroupStatusEnable)
	g1.ID = "test_local_id"
	assert.NoError(t, g1.Save(ctx))
	g2 := models.NewGroup(
		"test_remote",
		models.Endpoints{
			models.Endpoint("118.89.100.129"),
			models.Endpoint("118.89.100.130"),
		},
		nil, models.GroupStatusEnable)
	g2.ID = "test_remote_id"
	assert.NoError(t, g2.Save(ctx))

	hp := models.Heapster{
		ID:         "test_httpdetector_id",
		Name:       "test_httpdetector",
		Type:       models.CheckTypeHTTP,
		Port:       5200,
		Timeout:    timeout,
		Groups:     []string{string(g1.ID), string(g2.ID)},
		AcceptCode: []int{200, 404},
	}

	ctx, serverCancel := context.WithCancel(ctx)
	serverCtx := WithHTTPTarget(ctx)

	d, err := httpDetectorCreator(ctx, nil, hp)
	assert.NoError(t, err)
	fmt.Println(d.plumb(context.Background()))

	serverCancel()
	<-serverCtx.Done()
}
