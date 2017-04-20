package models

import (
	"context"
	"fmt"
	"testing"

	"zonst/qipai/gamehealthysrv/middlewares"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestHeapterNotifier(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 9)

	n1 := HeapsterNotifier{
		ID:   "testnotifier1",
		Type: "sms",
		Config: map[string]interface{}{
			"type":     "unicom",
			"sp":       "103905",
			"username": "zz_sj",
			"password": "www.zonst.org",
			"targets":  []string{"13879156403", "15507911970"},
		},
	}

	n2 := HeapsterNotifier{
		ID:   "testnotifier2",
		Type: "sms",
		Config: map[string]interface{}{
			"type":     "unicom",
			"sp":       "103905",
			"username": "zz_sj",
			"password": "www.zonst.org",
			"targets":  []string{"13879156403", "15507911970"},
		},
	}

	assert.NoError(t, n1.Save(ctx))
	assert.NoError(t, n2.Save(ctx))

	n11 := HeapsterNotifier{
		ID: "testnotifier1",
	}

	n22 := HeapsterNotifier{
		ID: "testnotifier2",
	}

	assert.NoError(t, n11.Fill(ctx))
	assert.NoError(t, n22.Fill(ctx))

	assert.Equal(t, n1.Type, n11.Type)
}

func TestHeapsterPersistent(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 9)

	h1 := Heapster{
		ID:        "test_heapster_1",
		Type:      CheckTypeTCP,
		Port:      10000,
		Timeout:   3 * time.Second,
		Interval:  5 * time.Second,
		Notifiers: []string{"testnotifier1", "testnotifier2"},
	}

	assert.NoError(t, h1.Save(ctx))
	h2 := Heapster{
		ID: "test_heapster_1",
	}

	assert.NoError(t, h2.Fill(ctx))
	fmt.Println(h2)
}

func TestApplyNotifiers(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 9)

	hp1 := Heapster{ID: "test_heapster_1"}
	assert.NoError(t, hp1.Fill(ctx))
	fmt.Println(hp1)

	notifiers, err := hp1.GetApplyNotifiers(ctx)
	assert.NoError(t, err)
	fmt.Println(notifiers)
}

func TestHeapsterLoad(t *testing.T) {
	ctx := middlewares.WithRedisConn(context.Background(), "0.0.0.0:6379", "", 9)

	hs, err := FetchHeapsters(ctx)
	assert.NoError(t, err)
	fmt.Println(hs)
}

func TestHeapsterDiff(t *testing.T) {
	hp1 := Heapster{
		ID: "hp1",
	}
	hp2 := Heapster{
		ID: "hp2",
	}
	hp3 := Heapster{
		ID: "hp3",
	}
	hp4 := Heapster{
		ID: "hp4",
	}

	hset1 := HeapsterSet{
		HeapsterSetKey(hp1.ID): hp1,
		HeapsterSetKey(hp2.ID): hp2,
		HeapsterSetKey(hp3.ID): hp3,
	}
	hp3.Name = "newname"
	hset2 := HeapsterSet{
		HeapsterSetKey(hp3.ID): hp3,
		HeapsterSetKey(hp2.ID): hp2,
		HeapsterSetKey(hp4.ID): hp4,
	}

	fmt.Println(hset1.Diff(hset2))
}
