package middlewares

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithInfluxDB(t *testing.T) {
	ctx, err := WithInfluxDB(nil, "http://localhost:8086", "", "")
	assert.NoError(t, err)

	client := GetInfluxDB(ctx)
	_, _, err = client.Ping(1 * time.Second)
	assert.NoError(t, err)
}
