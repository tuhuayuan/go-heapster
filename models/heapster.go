package models

import (
	"context"
	"encoding/json"
	"fmt"
	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/garyburd/redigo/redis"
)

// HealthyStatus 健康状态
type HealthyStatus string

// 三种健康状态常量
const (
	HealthyStatusUnknown HealthyStatus = "unknown"
	HealthyStatusGood    HealthyStatus = "up"
	HealthyStatusBad     HealthyStatus = "down"
)

// MarshalJSON json编码实现
func (hs HealthyStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(hs))
}

// UnmarshalJSON json解码实现
func (hs *HealthyStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*hs = HealthyStatus(str)
	return nil
}

// CheckType 健康检查的类型
type CheckType string

// 支持的检查类型
const (
	CheckTypeHTTP CheckType = "http"
	CheckTypeTCP  CheckType = "tcp"
)

// MarshalJSON json编码实现
func (ct CheckType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ct))
}

// UnmarshalJSON json解码实现
func (ct *CheckType) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*ct = CheckType(str)
	return nil
}

// Heapster 健康检查结构体
type Heapster struct {
	ID         SerialNumber `json:"id"`
	Name       string       `json:"name"`
	Type       CheckType    `json:"type"`
	Port       int          `json:"port"`
	AcceptCode []int        `json:"accept_code,omitempty"`
	Timeout    Duration     `json:"timeout"`
	Interval   Duration     `json:"interval"`
	Healthy    int          `json:"healthy_threshold"`
	UnHealthy  int          `json:"unhealthy_threshold"`
	Groups     []string     `json:"groups"`
	// 太长了
	Notifiers map[string]interface{} `json:"notifiers"`
}

// Validate 验证
func (hst *Heapster) Validate() error {
	if hst.ID == "" {
		return fmt.Errorf("empty id")
	}
	if hst.Port <= 0 || hst.Port >= 65536 {
		return fmt.Errorf("port must > 0  and < 65536")
	}
	return nil
}

// GetStatus 获取状态
func (hst *Heapster) GetStatus(ctx context.Context) HealthyStatus {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	status, err := redis.String(conn.Do("HGET", fmt.Sprintf("gamehealthy_heapster_%s", hst.ID), "status"))
	if err != nil {
		return HealthyStatusUnknown
	}
	switch HealthyStatus(status) {
	case HealthyStatusBad:
		return HealthyStatusBad
	case HealthyStatusGood:
		return HealthyStatusGood
	default:
		return HealthyStatusUnknown
	}
}

// SetStatus 设置状态
func (hst *Heapster) SetStatus(ctx context.Context, status HealthyStatus) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	_, err := conn.Do("HSET", fmt.Sprintf("gamehealthy_heapster_%s", hst.ID), "status", status)
	return err
}

// GetApplyGroups 从基本信息里面获取Group列表
func (hst *Heapster) GetApplyGroups(ctx context.Context) (Groups, error) {
	gs := make(Groups, 0, len(hst.Groups))
	for _, gid := range hst.Groups {
		g := Group{
			ID: SerialNumber(gid),
		}
		if err := g.Fill(ctx); err != nil {
			return nil, err
		}
		gs = append(gs, g)
	}
	return gs, nil
}

// GetApplyNotifiers 从配置的notifier字段提取出通知器
func (hst *Heapster) GetApplyNotifiers(ctx context.Context) (Notifiers, error) {
	ret := make(Notifiers, 0, len(hst.Notifiers))

	for t, v := range hst.Notifiers {
		config, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		model := Notifier{
			Type:   t,
			Config: config,
		}
		ret = append(ret, model)
	}
	return ret, nil
}

// Fill 查询基本信息
func (hst *Heapster) Fill(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	data, err := redis.Bytes(conn.Do("HGET", fmt.Sprintf("gamehealthy_heapster_%s", hst.ID), "meta"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, hst)
}

// Save 保存基本信息
func (hst *Heapster) Save(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	if err := hst.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(hst)
	if err != nil {
		return err
	}
	_, err = conn.Do("HSET", fmt.Sprintf("gamehealthy_heapster_%s", hst.ID), "meta", data)
	return err
}
