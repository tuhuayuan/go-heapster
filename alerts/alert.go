package alerts

import (
	"context"
	"fmt"
	"sync"
	"time"

	"strconv"
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

	go func() {
		defer close(al.done)
		defer func() {
			// 设置looper状态
			al.mtx.Lock()
			defer al.mtx.Unlock()
			al.running = false
		}()
		// 计算采样间隔，放大一倍
		threshold := float64(al.model.Threshold) * 2
		sampleInterval := time.Duration(threshold*2) * al.model.Interval
		// 警报器运行间隔固定
		ticker := time.NewTicker(sampleInterval)
		defer ticker.Stop()

		for {
			// 循环控制
			select {
			case <-al.ctx.Done():
				return
			case <-ticker.C:
			}

			// 查询统计报告
			rps, err := models.FetchReportsAggregation(al.ctx, models.LabelValue(al.model.ID), sampleInterval)
			if err != nil || len(rps) == 0 {
				if err := al.model.SetStatus(al.ctx, models.HealthyStatusUnknown); err != nil {
					logger.Warnf("healthy status change pass. %v", err)
				}
				continue
			}

			for _, rp := range rps {
				success, err := strconv.ParseFloat(string(rp.Labels[models.ReportNameSuccess]), 64)
				if err != nil {
					continue
				}
				if success < 0 && success <= threshold {
					// RED
					if err := al.model.SetStatus(al.ctx, models.HealthyStatusRed); err != nil {
						logger.Warnf("healthy status change pass. %v", err)
					}
					// 发通知
					if !al.mute {
						for _, nt := range al.notifiers {
							if err := nt.Send(al.ctx, rps); err != nil {
								logger.Warnf("send report error %v", err)
							}
						}
					}
					break
				} else if success >= 0 && success >= threshold {
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
	// TODO 应该把静音配置写到heapter模型里面
	al.mute = isMute
}
