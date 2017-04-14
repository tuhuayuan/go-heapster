package models

import (
	"context"
	"encoding/json"
	"testing"

	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	g1 := &Group{
		ID:   "testgroup1",
		Name: "测试游戏服务器",
		Endpoints: Endpoints{
			Endpoint("10.0.0.0/24"),
		},
		Excluded: Endpoints{
			Endpoint("10.0.0.1"),
			Endpoint("10.0.0.254"),
		},
	}
	data, err := json.Marshal(g1)
	assert.NoError(t, err)

	g2 := Group{}
	assert.NoError(t, json.Unmarshal(data, &g2))
	assert.Len(t, g1.Endpoints.Unfold().Exclude(g2.Excluded), 252)
}

func TestGroupPersistent(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 9)

	g1 := &Group{
		ID:   "testgroup1",
		Name: "测试游戏服务器",
		Endpoints: Endpoints{
			Endpoint("10.0.0.0/28"),
		},
		Excluded: Endpoints{
			Endpoint("10.0.0.1"),
			Endpoint("10.0.0.254"),
		},
	}
	assert.NoError(t, g1.Save(ctx))
	g2 := Group{ID: g1.ID}
	assert.NoError(t, g2.Fill(ctx))
	assert.Equal(t, g1.Name, g2.Name)
}

func TestFetchGroups(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 9)

	gs, err := FetchGroups(ctx)
	assert.NoError(t, err)
	assert.True(t, len(gs) > 0)
}
