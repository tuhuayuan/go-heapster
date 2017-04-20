package detectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
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

	worker detector
}

// Run 启动循环
func (dl *defaultLooper) Run() error {
	logger := middlewares.GetLogger(dl.ctx)

	// 确保线程安全，只能启动一次
	dl.mtx.Lock()
	if dl.running {
		dl.mtx.Unlock()
		return fmt.Errorf("already running")
	}
	dl.running = true
	dl.mtx.Unlock()

	// 启动主循环
	go func() {
		defer close(dl.done)
		defer func() {
			// 设置looper状态
			dl.mtx.Lock()
			defer dl.mtx.Unlock()
			dl.running = false
		}()

		ticker := time.NewTicker(dl.model.Interval)
		defer ticker.Stop()

		for {
			var (
				timeoutCtx, cancel = context.WithTimeout(dl.ctx, time.Duration(dl.model.Timeout))
			)
			reports, _ := dl.worker.plumb(timeoutCtx)
			cancel()
			// 写入报告
			if err := reports.Save(dl.ctx); err != nil {
				logger.Warnf("pass report save %v", err)
			}
			// 结束或者下一次循环开始时间到了
			select {
			case <-dl.ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

// Stop 停止
func (dl *defaultLooper) Stop() {
	dl.cancel()
	<-dl.done
}
