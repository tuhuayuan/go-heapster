package middlewares

import (
	"testing"

	"context"

	"github.com/stretchr/testify/assert"
)

func TestRedisClient(t *testing.T) {
	ctx := WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)
	conn := GetRedisConn(ctx)
	defer conn.Close()
	_, err := conn.Do("SET", "testkey", "testvalue")
	assert.NoError(t, err)
}
