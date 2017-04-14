package models

import (
	"context"
	"encoding/json"

	"fmt"
	"zonst/qipai/gamehealthysrv/middlewares"

	"strings"

	"github.com/garyburd/redigo/redis"
)

// GroupStatus 监控组状态
type GroupStatus string

// 组状态
const (
	GroupStatusEnable  GroupStatus = "enable"
	GroupStatusDisable GroupStatus = "disable"
)

// MarshalJSON json编码实现
func (gs GroupStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(gs))
}

// UnmarshalJSON json解码实现
func (gs *GroupStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*gs = GroupStatus(str)
	return nil
}

// Group 监控组数据结构
type Group struct {
	ID        SerialNumber `json:"id"`
	Name      string       `json:"name"`
	Endpoints Endpoints    `json:"endpoints,omitempty"`
	Excluded  Endpoints    `json:"excluded,omitempty"`
	Status    GroupStatus  `json:"status,omitempty"`
	Version   int          `json:"version,omitempty"`
}

// Groups 组列表
type Groups []Group

// Validate 验证合法性
func (g *Group) Validate() error {
	if g.ID == "" {
		return fmt.Errorf("id can't be empty")
	}
	if g.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	return nil
}

// Fill 根据ID查询 Group 对象
func (g *Group) Fill(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	storeKey := fmt.Sprintf("gamehealthy_group_%s", g.ID)
	data, err := redis.Bytes(conn.Do("HGET", storeKey, "meta"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, g); err != nil {
		return err
	}
	g.Version, err = redis.Int(conn.Do("HGET", storeKey, "version"))
	return err
}

// Save 持久化 Group 对象
func (g *Group) Save(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	if err := g.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(g)
	if err != nil {
		return err
	}

	storeKey := fmt.Sprintf("gamehealthy_group_%s", g.ID)
	err = conn.Send("MULTI")
	err = conn.Send("HSET", storeKey, "meta", data)
	err = conn.Send("HINCRBY", storeKey, "version", 1)
	err = conn.Send("EXEC")
	err = conn.Flush()
	_, err = conn.Receive()
	return err
}

// Delete 删除
func (g *Group) Delete(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	storeKey := fmt.Sprintf("gamehealthy_group_%s", g.ID)
	_, err := conn.Do("DEL", storeKey)

	return err
}

// FetchGroups 获取group列表
func FetchGroups(ctx context.Context) (Groups, error) {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	rawKeys, err := redis.ByteSlices(conn.Do("KEYS", "gamehealthy_group_*"))
	if err != nil {
		return nil, err
	}
	gs := make(Groups, 0, 256)
	for _, rawKey := range rawKeys {
		key := strings.TrimPrefix(string(rawKey), "gamehealthy_group_")
		g := &Group{
			ID: SerialNumber(key),
		}
		if g.Fill(ctx) != nil {
			continue
		}
		gs = append(gs, *g)
	}
	return gs, nil
}
