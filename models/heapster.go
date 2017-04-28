package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"zonst/qipai/gamehealthysrv/middlewares"

	"github.com/garyburd/redigo/redis"
)

// HealthyStatus 健康状态
type HealthyStatus string

// 三种健康状态常量
const (
	HealthyStatusUnknown HealthyStatus = "unknown"
	HealthyStatusGreen   HealthyStatus = "green"
	HealthyStatusYellow  HealthyStatus = "yellow"
	HealthyStatusRed     HealthyStatus = "red"
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
	ID        SerialNumber  `json:"id"`
	Name      string        `json:"name"`
	Type      CheckType     `json:"type"`
	Port      int           `json:"port"`
	Timeout   time.Duration `json:"timeout"`
	Interval  time.Duration `json:"interval"`
	Threshold int           `json:"threshold"`
	Groups    []string      `json:"groups"`
	Notifiers []string      `json:"notifiers"`
	Mute      bool          `json:"mute"`
	Version   int           `json:"version,omitempty"`
	Status    HealthyStatus `json:"status,omitempty"`

	// for http
	AcceptCode []int  `json:"accept_code,omitempty"`
	Host       string `json:"host,omitempty"`
	Location   string `json:"location,omitempty"`

	// TODO: for https
}

// HeapsterStatusSet 状态集
type HeapsterStatusSet struct {
	ID     SerialNumber  `json:"id"`
	Status HealthyStatus `json:"status"`
}

// Heapsters 列表
type Heapsters []Heapster

// HeapsterSetKey 集合主键
type HeapsterSetKey string

// HeapsterSetKeys 集合主键列表
type HeapsterSetKeys []HeapsterSetKey

// HeapsterSet 集合
type HeapsterSet map[HeapsterSetKey]Heapster

// Diff 计算两个集合的差集
func (newset HeapsterSet) Diff(oldset HeapsterSet) (added HeapsterSetKeys, modified HeapsterSetKeys, deleted HeapsterSetKeys) {
	for key, oldHeapster := range oldset {
		newHeapster, ok := newset[key]
		if !ok {
			// 删除了
			deleted = append(deleted, key)
		} else if newHeapster.Version > oldHeapster.Version {
			// 修改了
			modified = append(modified, key)
		}
	}
	// 找到新添加的
	for key := range newset {
		if _, ok := oldset[key]; !ok {
			added = append(added, key)
		}
	}
	return
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
	case HealthyStatusRed:
		return HealthyStatusRed
	case HealthyStatusYellow:
		return HealthyStatusYellow
	case HealthyStatusGreen:
		return HealthyStatusGreen
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
func (hst *Heapster) GetApplyNotifiers(ctx context.Context) (HeapsterNotifiers, error) {
	ret := make(HeapsterNotifiers, 0, len(hst.Notifiers))
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	for _, id := range hst.Notifiers {
		model := HeapsterNotifier{
			ID: SerialNumber(id),
		}
		if err := model.Fill(ctx); err != nil {
			continue
		}
		ret = append(ret, model)
	}
	return ret, nil
}

// FetchHeapsterStatus 批量获取状态
func FetchHeapsterStatus(ctx context.Context) ([]HeapsterStatusSet, error) {
	statusList := make([]HeapsterStatusSet, 0, 256)

	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	rawKeys, err := redis.Strings(conn.Do("KEYS", "gamehealthy_heapster_*"))
	if err != nil {
		return nil, err
	}
	for _, rawKey := range rawKeys {
		key := strings.TrimPrefix(string(rawKey), "gamehealthy_heapster_")
		heapster := &Heapster{
			ID: SerialNumber(key),
		}
		statusList = append(statusList, HeapsterStatusSet{
			ID:     heapster.ID,
			Status: heapster.GetStatus(ctx),
		})
	}
	return statusList, nil
}

// FetchHeapsters 全部加载
func FetchHeapsters(ctx context.Context) (HeapsterSet, error) {
	var hset = make(HeapsterSet, 256)
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	rawKeys, err := redis.Strings(conn.Do("KEYS", "gamehealthy_heapster_*"))
	if err != nil {
		return nil, err
	}
	for _, rawKey := range rawKeys {
		key := strings.TrimPrefix(string(rawKey), "gamehealthy_heapster_")
		heapster := &Heapster{
			ID: SerialNumber(key),
		}
		if err := heapster.Fill(ctx); err != nil {
			continue
		}
		hset[HeapsterSetKey(heapster.ID)] = *heapster
	}
	return hset, nil
}

// Fill 查询基本信息
func (hst *Heapster) Fill(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	storeKey := fmt.Sprintf("gamehealthy_heapster_%s", hst.ID)
	data, err := redis.Bytes(conn.Do("HGET", storeKey, "meta"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, hst); err != nil {
		return err
	}
	// 获取状态
	status, err := redis.String(conn.Do("HGET", storeKey, "status"))
	if err != nil {
		status = string(HealthyStatusUnknown)
	}
	hst.Status = HealthyStatus(status)
	// 版本计算
	version, err := redis.Int(conn.Do("HGET", storeKey, "version"))
	if err != nil {
		return err
	}
	gs, err := hst.GetApplyGroups(ctx)
	if err == nil {
		for _, g := range gs {
			version += g.Version
		}
	}
	ns, err := hst.GetApplyNotifiers(ctx)
	if err == nil {
		for _, n := range ns {
			version += n.Version
		}
	}
	hst.Version = version

	return nil
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

	storeKey := fmt.Sprintf("gamehealthy_heapster_%s", hst.ID)
	err = conn.Send("MULTI")
	err = conn.Send("HSET", storeKey, "meta", data)
	err = conn.Send("HINCRBY", storeKey, "version", 1)
	err = conn.Send("EXEC")
	err = conn.Flush()
	_, err = conn.Receive()

	return err
}

// Delete 删除
func (hst *Heapster) Delete(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	_, err := conn.Do("DEL", fmt.Sprintf("gamehealthy_heapster_%s", hst.ID))
	return err
}

// HeapsterNotifier 就是自定义的LabelSet
type HeapsterNotifier struct {
	ID      SerialNumber `json:"id"`
	Name    string       `json:"name"`
	Type    string       `json:"type"`
	Version int          `json:"version,omitempty"`
	Config  map[string]interface{}
}

// HeapsterNotifiers 列表
type HeapsterNotifiers []HeapsterNotifier

// Validate 验证
func (hn *HeapsterNotifier) Validate() error {
	if hn.ID == "" {
		return fmt.Errorf("empty id")
	}
	if hn.Type == "" {
		return fmt.Errorf("empty type")
	}
	return nil
}

// Fill 获取notifier模型
func (hn *HeapsterNotifier) Fill(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	storeKey := fmt.Sprintf("gamehealthy_notifier_%s", hn.ID)
	data, err := redis.Bytes(conn.Do("HGET", storeKey, "meta"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, hn); err != nil {
		return err
	}
	hn.Version, err = redis.Int(conn.Do("HGET", storeKey, "version"))
	if err != nil {
		return err
	}
	return nil
}

// Save 保存notifier模型
func (hn *HeapsterNotifier) Save(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()
	if err := hn.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(hn)
	if err != nil {
		return err
	}

	storeKey := fmt.Sprintf("gamehealthy_notifier_%s", hn.ID)
	err = conn.Send("MULTI")
	err = conn.Send("HSET", storeKey, "meta", data)
	err = conn.Send("HINCRBY", storeKey, "version", 1)
	err = conn.Send("EXEC")
	err = conn.Flush()
	_, err = conn.Receive()
	return err
}

// Delete 删除
func (hn *HeapsterNotifier) Delete(ctx context.Context) error {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	_, err := conn.Do("DEL", fmt.Sprintf("gamehealthy_notifier_%s", hn.ID))
	return err
}

// FetchHeapsterNotifiers 获取notifier列表
func FetchHeapsterNotifiers(ctx context.Context) (HeapsterNotifiers, error) {
	conn := middlewares.GetRedisConn(ctx)
	defer conn.Close()

	rawKeys, err := redis.ByteSlices(conn.Do("KEYS", "gamehealthy_notifier_*"))
	if err != nil {
		return nil, err
	}
	var hns = make(HeapsterNotifiers, 0, 256)
	for _, rawKey := range rawKeys {
		key := strings.TrimPrefix(string(rawKey), "gamehealthy_notifier_")
		hn := &HeapsterNotifier{
			ID: SerialNumber(key),
		}
		if hn.Fill(ctx) != nil {
			continue
		}
		hns = append(hns, *hn)
	}
	return hns, nil
}
