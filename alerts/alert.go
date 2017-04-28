package alerts

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
	"zonst/qipai/gamehealthysrv/notifiers"
)

// Alert 警报器接口
type Alert interface {
	TurnOn() error
	TurnOff()
	Mute(isMute bool)
}

// NewAlert 新建警报器
func NewAlert(ctx context.Context, model models.Heapster) (Alert, error) {
	logger := middlewares.GetLogger(ctx)
	al := &defaultAlert{
		model: model,
		done:  make(chan struct{}),
		mute:  model.Mute,
	}
	al.ctx, al.cancel = context.WithCancel(ctx)
	// 创建notifier
	notifierModels, err := al.model.GetApplyNotifiers(ctx)
	if err == nil {
		for _, model := range notifierModels {
			notifier, err := notifiers.NewNotifier(model)
			if err != nil {
				logger.Warnf("load notifier error %v, heapster %s", err, model.ID)
			}
			al.notifiers = append(al.notifiers, notifier)
		}
	}
	return al, nil
}

// 默认警报器，不支持查询语句
type defaultAlert struct {
	model     models.Heapster
	ctx       context.Context
	mute      bool
	notifiers []notifiers.Notifier
	mtx       sync.RWMutex
	cancel    func()
	done      chan struct{}
	running   bool
}

// TurnOn 启动警报器
func (al *defaultAlert) TurnOn() error {
	logger := middlewares.GetLogger(al.ctx)

	// 确保线程安全，只能启动一次
	al.mtx.Lock()
	if al.running {
		al.mtx.Unlock()
		return fmt.Errorf("already turnon")
	}
	al.running = true
	al.mtx.Unlock()

	// 启动监控循环
	go func() {
		defer close(al.done)
		// 循环结束调整状态
		defer func() {
			al.mtx.Lock()
			defer al.mtx.Unlock()
			al.running = false
		}()
		// 查询报告间隔
		sampleInterval := time.Duration(al.model.Threshold)*al.model.Interval + al.model.Interval
		ticker := time.NewTicker(sampleInterval)
		defer ticker.Stop()
		// 循环查询报告
		for {
			// 循环控制
			select {
			case <-al.ctx.Done():
				return
			case <-ticker.C:
			}
			// 查询统计报告
			rps, err := models.FetchReportsAggs(al.ctx, string(al.model.ID), sampleInterval)
			if err != nil || len(rps) == 0 {
				al.model.SetStatus(al.ctx, models.HealthyStatusUnknown)
				logger.Warnf("no report data for %s", string(al.model.ID))
				continue
			}
			// 处理报告
			for _, rp := range rps {
				if rp.Faileds >= al.model.Threshold {
					// RED
					if err := al.model.SetStatus(al.ctx, models.HealthyStatusRed); err != nil {
						logger.Warnf("healthy status change pass. %v", err)
					}
					// 发通知
					if !al.mute {
						for _, nt := range al.notifiers {
							if err := nt.Send(al.ctx, rp); err != nil {
								logger.Warnf("send report error %v", err)
							}
						}
					}
					break
				} else if rp.Success >= al.model.Threshold {
					// GREEN
					if err := al.model.SetStatus(al.ctx, models.HealthyStatusGreen); err != nil {
						logger.Warnf("healthy status change pass. %v", err)
					}
					break
				} else {
					// YELLOW
					if err := al.model.SetStatus(al.ctx, models.HealthyStatusYelow); err != nil {
						logger.Warnf("healthy status change pass. %v", err)
					}
					break
				}
			}
		}
	}()
	return nil
}

// TurnOff 关闭警报器
func (al *defaultAlert) TurnOff() {
	al.cancel()
	<-al.done
}

// Mute 设置是否静音
func (al *defaultAlert) Mute(isMute bool) {
	al.mute = isMute
}
