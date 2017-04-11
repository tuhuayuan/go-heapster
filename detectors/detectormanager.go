package detectors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"zonst/qipai/gamehealthysrv/models"
	alert "zonst/qipai/gamehealthysrv/notifiers"
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
type detectorCreator func(ctx context.Context, manager DetectorManager, hp models.Heapster) (detector, error)

// DetectorLooper 循环接口
type DetectorLooper interface {
	Run() error
	Stop()
}

// DetectorManager 管理接口
type DetectorManager interface {
	// 创建或则返回已有looper
	CreateLooper(hp models.Heapster) (DetectorLooper, error)
	// 停止并且销毁已有looper
	DropLooper(hp models.Heapster)
	// 相当于Drop所有
	Shutdown()
}

type defaultManager struct {
	mtx     sync.RWMutex
	dead    bool
	ctx     context.Context
	loopers map[models.SerialNumber]DetectorLooper
}

// NewManager 新建管理器
func NewManager(ctx context.Context) DetectorManager {
	dm := &defaultManager{
		ctx:     ctx,
		loopers: make(map[models.SerialNumber]DetectorLooper, 256),
		dead:    false,
	}
	return dm
}

// CreateLooper 创建
func (dm *defaultManager) CreateLooper(hp models.Heapster) (DetectorLooper, error) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()

	// 错误状态
	if dm.dead {
		return nil, fmt.Errorf("manager is shutdown")
	}
	// 返回已有
	if lp, ok := dm.loopers[hp.ID]; ok {
		return lp, nil
	}
	// 创建一个
	lp := &defaultLooper{
		model:  hp,
		status: models.HealthyStatusUnknown,
		done:   make(chan struct{}),
	}
	lp.ctx, lp.cancel = context.WithCancel(dm.ctx)
	creator, ok := namedDetectors[string(lp.model.Type)]
	if !ok {
		return nil, fmt.Errorf("detector type not found")
	}
	// 获取配置的通知器
	notifiers, err := lp.model.GetApplyNotifiers(dm.ctx)
	if err == nil {
		for _, ntr := range notifiers {
			al, err := alert.DefaultManager.CreateNotifier(ntr)
			if err != nil {
				continue
			}
			lp.alerts = append(lp.alerts, al)
		}
	}
	// 创建detector
	if lp.worker, err = creator(dm.ctx, dm, lp.model); err != nil {
		return nil, err
	}
	// 保留looper指针
	dm.loopers[lp.model.ID] = lp
	return lp, nil
}

// DropLooper 停止并且销毁
func (dm *defaultManager) DropLooper(hp models.Heapster) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()
	lp, ok := dm.loopers[hp.ID]
	if !ok {
		return
	}
	lp.Stop()
	delete(dm.loopers, hp.ID)
}

// Stop 停止所有looper
func (dm *defaultManager) Shutdown() {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()

	dm.dead = true
	wg := sync.WaitGroup{}
	for _, lp := range dm.loopers {
		wg.Add(1)
		go func(lp DetectorLooper) {
			defer wg.Done()
			lp.Stop()
		}(lp)
	}
	wg.Wait()
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
	alerts []alert.Notifier
}

// Run 启动循环
func (dl *defaultLooper) Run() error {
	// 确保线程安全，只能启动一次
	var keeper = func() error {
		dl.mtx.Lock()
		defer dl.mtx.Unlock()
		if dl.running {
			return fmt.Errorf("already running")
		}
		dl.running = true
		return nil
	}
	if err := keeper(); err != nil {
		return err
	}

	// 更新当前状态
	dl.status = dl.model.GetStatus(dl.ctx)
	// 启动主循环
	go func() {
		defer close(dl.done)

		ticker := time.NewTicker(time.Duration(dl.model.Interval))
		defer ticker.Stop()
		var (
			healthy   = 0
			unhealthy = 0
		)
		for {
			var (
				ctx, cancel = context.WithTimeout(dl.ctx, time.Duration(dl.model.Timeout))
			)
			// TODO 不要忽略plumb报告
			reports, err := dl.worker.plumb(ctx)
			cancel()
			if err != nil {
				unhealthy++
				healthy = 0
				// 不健康了
				if unhealthy >= dl.model.UnHealthy {
					if dl.status != models.HealthyStatusBad {
						// 改变状态
						dl.status = models.HealthyStatusBad
						err := dl.model.SetStatus(dl.ctx, dl.status)
						if err != nil {
							log.Printf("pass status change HealthyStatusBad %v", err)
						}
						// 处理Reports
						for _, al := range dl.alerts {
							al.Send(dl.ctx, reports[0])
						}
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

	// 设置looper运行结束
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
