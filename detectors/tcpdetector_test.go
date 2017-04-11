package detectors

import (
	"context"
	"fmt"
	"testing"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/stretchr/testify/assert"

	"net"
	"time"
)

func WithTCPTarget(ctx context.Context) context.Context {
	endCtx, callDone := context.WithCancel(context.Background())
	go func() {
		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:10000")
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Println("server start.")
			for {
				conn, err := l.Accept()
				if err != nil {
					break
				}
				conn.Close()
			}
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

func TestTCPPlumb(t *testing.T) {
	var (
		timeout, _ = models.ParseDuration("1s")
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
			models.Endpoint("118.89.100.76"),
			models.Endpoint("118.89.100.77"),
			models.Endpoint("118.89.100.78"),
			models.Endpoint("118.89.100.79"),
		},
		nil, models.GroupStatusEnable)
	g2.ID = "test_remote_id"
	assert.NoError(t, g2.Save(ctx))

	hp := models.Heapster{
		ID:      "test_tcpdetector_id",
		Name:    "test_tcpdetector",
		Type:    models.CheckTypeTCP,
		Port:    10000,
		Timeout: timeout,
		Groups:  []string{string(g1.ID), string(g2.ID)},
	}

	ctx, serverCancel := context.WithCancel(ctx)
	serverCtx := WithTCPTarget(ctx)

	d, err := tcpDetectorCreator(ctx, nil, hp)
	assert.NoError(t, err)

	plumbCtx, plumbCancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(2 * time.Second)
		plumbCancel()
	}()
	fmt.Println(d.plumb(plumbCtx))
	fmt.Println(d.plumb(plumbCtx))
	serverCancel()
	<-serverCtx.Done()
}
