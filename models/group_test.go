package models

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	g1 := NewGroup(
		"测试游戏服务器",
		Endpoints{
			Endpoint("10.0.0.0/28"),
		},
		Endpoints{
			Endpoint("10.0.0.1"),
			Endpoint("10.0.0.254"),
		}, GroupStatusEnable)
	data, err := json.Marshal(g1)
	assert.NoError(t, err)
	g2 := Group{}
	assert.NoError(t, json.Unmarshal(data, &g2))
	assert.Len(t, g1.UnFoldedEndpoints.Exclude(g2.UnFoldedEndpoints), 0)
}

func TestGroupPersistent(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)

	g1 := NewGroup(
		"测试游戏服务器",
		Endpoints{
			Endpoint("10.0.0.0/28"),
		},
		Endpoints{
			Endpoint("10.0.0.1"),
			Endpoint("10.0.0.254"),
		}, GroupStatusEnable)
	assert.NoError(t, g1.Save(ctx))
	g2 := Group{ID: g1.ID}
	assert.NoError(t, g2.Fill(ctx))
	assert.Equal(t, g1, &g2)
}

func TestGroups(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 1)

	gs1 := Groups{
		*NewGroup(
			"测试游戏服务器",
			Endpoints{
				Endpoint("10.0.0.0/28"),
			},
			Endpoints{
				Endpoint("10.0.0.1"),
				Endpoint("10.0.0.254"),
			}, GroupStatusEnable),
		*NewGroup(
			"测试Web服务器",
			Endpoints{
				Endpoint("10.0.0.0/28"),
			},
			Endpoints{
				Endpoint("10.0.0.1"),
				Endpoint("10.0.0.254"),
			}, GroupStatusEnable),
	}
	data, err := gs1.MarshalJSON()
	assert.NoError(t, err)
	assert.NoError(t, gs1.Save(ctx))
	gs2 := &Groups{}
	assert.NoError(t, json.Unmarshal(data, gs2))
	assert.NoError(t, gs2.Fill(ctx))
	fmt.Println((*gs2)[0])
}
