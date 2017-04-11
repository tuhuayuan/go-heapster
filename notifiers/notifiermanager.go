package notifiers

import (
	"context"
	"fmt"
	"zonst/qipai/gamehealthysrv/models"
)

var (
	namedNotifiers = make(map[string]notifierCreator)
)

// notifier工厂方法
type notifierCreator func(model models.Notifier) (Notifier, error)

// 注册工厂
func registCreator(name string, creator notifierCreator) {
	namedNotifiers[name] = creator
}

// Notifier 健康状态通知者接口
type Notifier interface {
	Send(ctx context.Context, report models.Report) error
}

// NotifierManager 管理接口
type NotifierManager interface {
	CreateNotifier(model models.Notifier) (Notifier, error)
}

// DefaultManager 默认管理器
var DefaultManager = defaultManager{}

type defaultManager struct {
}

// CreateNotifier 实现管理接口
func (dm *defaultManager) CreateNotifier(model models.Notifier) (Notifier, error) {
	name := model.Type
	creator, ok := namedNotifiers[string(name)]
	if !ok {
		return nil, fmt.Errorf("nofitier %v not suppore ", name)
	}
	notifier, err := creator(model)
	if err != nil {
		return nil, err
	}
	return notifier, nil
}
