package models

import (
	"context"
	"encoding/json"

	"fmt"
	"zonst/qipai/gamehealthysrv/middlewares"

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
	ID                SerialNumber `json:"id"`
	Name              string       `json:"name"`
	FoldedEndpoints   Endpoints    `json:"endpoints,omitempty"`
	UnFoldedEndpoints Endpoints    `json:"-"`
	ExcludedEndpoints Endpoints    `json:"excluded,omitempty"`
	Status            GroupStatus  `json:"status,omitempty"`
}

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

// NewGroup 创建一个Group
func NewGroup(name string, eps Endpoints, excluded Endpoints, st GroupStatus) *Group {
	g := &Group{
		ID:                NewSerialNumber(),
		Name:              name,
		FoldedEndpoints:   eps.Validate(),
		ExcludedEndpoints: excluded.Validate(),
		Status:            st,
	}
	g.UnFoldedEndpoints = g.FoldedEndpoints.Unfold()
	g.UnFoldedEndpoints = g.UnFoldedEndpoints.Exclude(g.ExcludedEndpoints)
	return g
}

// UnmarshalJSON json解码实现
func (g *Group) UnmarshalJSON(b []byte) error {
	type _Group Group
	rawGroup := &_Group{}
	if err := json.Unmarshal(b, rawGroup); err != nil {
		return err
	}
	*g = *NewGroup(rawGroup.Name, rawGroup.FoldedEndpoints, rawGroup.ExcludedEndpoints, rawGroup.Status)
	g.ID = rawGroup.ID
	return nil
}

// Fill 根据ID查询 Group 对象
func (g *Group) Fill(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	data, err := redis.Bytes(conn.Do("GET", fmt.Sprintf("gamehealthy_group_%s", g.ID)))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, g)
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
	_, err = conn.Do("SET", fmt.Sprintf("gamehealthy_group_%s", g.ID), data)
	return err
}

// Groups 组列表
type Groups []Group

// Save 保存所有
func (gs Groups) Save(ctx context.Context) error {
	for _, g := range gs {
		if err := g.Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Fill 查找所有数据
func (gs Groups) Fill(ctx context.Context) error {
	for _, g := range gs {
		if err := g.Fill(ctx); err != nil {
			return err
		}
	}
	return nil
}
