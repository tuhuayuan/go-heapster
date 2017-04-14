package detectors

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"time"

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
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)
	g1 := models.Group{
		ID:   "test_local",
		Name: "test_local",
		Endpoints: models.Endpoints{
			models.Endpoint("127.0.0.1"),
			models.Endpoint("10.0.0.1/28"),
		},
	}
	assert.NoError(t, g1.Save(ctx))
	g2 := models.Group{
		ID:   "test_remote",
		Name: "test_remote",
		Endpoints: models.Endpoints{
			models.Endpoint("118.89.100.129"),
			models.Endpoint("118.89.100.130"),
		},
	}
	assert.NoError(t, g2.Save(ctx))

	hp := models.Heapster{
		ID:         "test_httpdetector_id",
		Name:       "test_httpdetector",
		Type:       models.CheckTypeHTTP,
		Port:       5200,
		Timeout:    2 * time.Second,
		Groups:     []string{string(g1.ID), string(g2.ID)},
		AcceptCode: []int{200, 404},
	}

	ctx, serverCancel := context.WithCancel(ctx)
	serverCtx := WithHTTPTarget(ctx)

	d, err := httpDetectorCreator(ctx, hp)
	assert.NoError(t, err)
	fmt.Println(d.plumb(context.Background()))

	serverCancel()
	<-serverCtx.Done()
}

func TestHttpWithHost(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)
	g1 := models.Group{
		ID:   "test_local1",
		Name: "test_local",
		Endpoints: models.Endpoints{
			models.Endpoint("127.0.0.1"),
		},
	}
	assert.NoError(t, g1.Save(ctx))
	hp := models.Heapster{
		ID:         "test_httpdetector_id",
		Name:       "test_httpdetector",
		Type:       models.CheckTypeHTTP,
		Port:       5050,
		Timeout:    2 * time.Second,
		Groups:     []string{string(g1.ID)},
		AcceptCode: []int{400},
		Location:   "/healthz",
		Host:       "zonst.local2",
	}

	d, err := httpDetectorCreator(ctx, hp)
	assert.NoError(t, err)
	fmt.Println(d.plumb(context.Background()))
}
