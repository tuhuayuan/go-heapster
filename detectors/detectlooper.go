package detectors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"zonst/qipai/gamehealthysrv/models"
	"zonst/qipai/gamehealthysrv/notifiers"
)

var (
	namedDetectors = make(map[string]detectorCreator)
)

// 注册detector类型
func registCreator(name string, creator detectorCreator) {
	namedDetectors[name] = creator
}

// detector 探测器接口
type detector interface {
	plumb(ctx context.Context) (models.Reports, error)
}

// detector工厂方法
type detectorCreator func(ctx context.Context, hp models.Heapster) (detector, error)

// DetectLooper 循环接口
type DetectLooper interface {
	Run() error
	Stop()
}

// NewDetectLooper 创建一个looper
func NewDetectLooper(ctx context.Context, model models.Heapster) (DetectLooper, error) {
	creator, ok := namedDetectors[string(model.Type)]
	if !ok {
		return nil, fmt.Errorf("detector type not found")
	}
	// looper对象
	lp := &defaultLooper{
		model:  model,
		status: models.HealthyStatusUnknown,
		done:   make(chan struct{}),
	}
	// 设置上下文
	lp.ctx, lp.cancel = context.WithCancel(ctx)
	// 创建detector
	detector, err := creator(ctx, model)
	if err != nil {
		return nil, err
	}
	lp.worker = detector
	// 创建notifier
	notifierModels, err := lp.model.GetApplyNotifiers(ctx)
	if err == nil {
		for _, model := range notifierModels {
			notifier, err := notifiers.NewNotifier(model)
			if err != nil {
				continue
			}
			lp.notifiers = append(lp.notifiers, notifier)
		}
	}
	return lp, nil
}

// 默认looper实现
type defaultLooper struct {
	ctx    context.Context
	model  models.Heapster
	status models.HealthyStatus

	mtx     sync.RWMutex
	cancel  func()
	done    chan struct{}
	running bool

	worker    detector
	notifiers []notifiers.Notifier
}

// Run 启动循环
func (dl *defaultLooper) Run() error {
	// 确保线程安全，只能启动一次
	dl.mtx.Lock()
	if dl.running {
		dl.mtx.Unlock()
		return fmt.Errorf("already running")
	}
	dl.running = true
	dl.mtx.Unlock()

	// 更新当前状态
	dl.status = dl.model.GetStatus(dl.ctx)
	// 启动主循环
	go func() {
		defer close(dl.done)

		ticker := time.NewTicker(time.Duration(dl.model.Interval))
		defer ticker.Stop()
		var (
			// 健康和不健康指数
			healthy   = 0
			unhealthy = 0
		)
		for {
			var (
				timeoutCtx, cancel = context.WithTimeout(dl.ctx, time.Duration(dl.model.Timeout))
			)
			reports, err := dl.worker.plumb(timeoutCtx)
			cancel()
			if err != nil {
				unhealthy++
				healthy = 0
				if unhealthy >= dl.model.UnHealthy {
					// 不健康了
					if dl.status != models.HealthyStatusBad {
						// 改变状态
						dl.status = models.HealthyStatusBad
						err := dl.model.SetStatus(dl.ctx, dl.status)
						if err != nil {
							log.Printf("pass status change HealthyStatusBad %v", err)
						}
						// 发送状态变化通知
						for _, nt := range dl.notifiers {
							nt.Send(dl.ctx, reports)
						}
					}
					// 处理报告
					if err := reports.Save(dl.ctx); err != nil {
						log.Printf("pass report save %v", err)
					}
					unhealthy = 0
				}
			} else {
				healthy++
				unhealthy = 0
				// 达到健康标准
				if healthy >= dl.model.Healthy {
					if dl.status != models.HealthyStatusGood {
						// 改变状态
						dl.status = models.HealthyStatusGood
						err := dl.model.SetStatus(dl.ctx, dl.status)
						if err != nil {
							log.Printf("pass status change HealthyStatusGood %v", err)
						}
					}
					healthy = 0
				}
			}
			// 结束或者下一次循环开始时间到了
			select {
			case <-dl.ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	// 设置looper状态
	dl.mtx.Lock()
	defer dl.mtx.Unlock()
	dl.running = false

	return nil
}

// Stop 停止
func (dl *defaultLooper) Stop() {
	dl.cancel()
	<-dl.done
}
