package daemons

import (
	"context"
	"os"

	"sync"
	"time"
	"zonst/qipai/gamehealthysrv/detectors"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
	"zonst/qipai/logagent/utils"
)

// HealthySrv 配置结构体
type HealthySrv struct {
	utils.InputPluginConfig

	// Redis配置
	RedisHost     string `json:"redis_host"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// 联通短信配置
	UnicomSP       string `json:"unicom_sp"`
	UnicomUsername string `json:"unicom_username"`
	UnicomPassword string `json:"unicom_password"`

	LogLevel   int      `json:"log_level"`
	AccessKeys []string `json:"accesskeys"`

	// 私有配置
	ctx     context.Context
	cancel  func()
	done    chan struct{}
	loopers map[models.SerialNumber]detectors.DetectLooper
}

// 服务名称
const (
	HealthySrvName = "gamehealthysrv"
)

func init() {
	utils.RegistInputHandler(HealthySrvName, HealthySrvInit)
}

// HealthySrvInit Healthy配置服务初始化
func HealthySrvInit(part *utils.ConfigPart) (srv *HealthySrv, err error) {
	config := HealthySrv{
		InputPluginConfig: utils.InputPluginConfig{
			TypePluginConfig: utils.TypePluginConfig{
				Type: HealthySrvName,
			},
		},

		loopers: make(map[models.SerialNumber]detectors.DetectLooper),
		done:    make(chan struct{}),
	}
	if err = utils.ReflectConfigPart(part, &config); err != nil {
		return
	}
	config.init()
	srv = &config
	return
}

// Start 组件启动
func (srv *HealthySrv) Start() {
	defer close(srv.done)

	var (
		logger      = middlewares.GetLogger(srv.ctx)
		ratelimiter = middlewares.GetRateLimiter(srv.ctx)
		limiterKey  = string(models.NewSerialNumber())

		oldSet, newSet models.HeapsterSet
		err            error
	)

	logger.Info("start main looper")
	for {
		select {
		case <-srv.ctx.Done():
			return
		default:
		}
		ratelimiter.Accept([]string{limiterKey}, 5*time.Second, 1)
		// 加载所有模型
		logger.Infof("reload heapters")
		newSet, err = models.FetchHeapsters(srv.ctx)
		if err != nil {
			logger.Warnf("load heapster error %v", err)
			continue
		}
		diffkeys := newSet.Diff(oldSet)

		if len(diffkeys) == 0 {
			logger.Infof("no changed")
			continue
		} else {
			logger.Infof("found %d heapster changed", len(diffkeys))
		}
		// 重新加载差异数据
		for _, k := range diffkeys {
			model := models.Heapster{
				ID: models.SerialNumber(k),
			}
			// 重新读取模型信息
			if err := model.Fill(srv.ctx); err != nil {
				logger.Warnf("load heapster %s error: %v", k, err)
				continue
			}
			// 创建新的looper丢弃老looper
			looper, err := detectors.NewDetectLooper(srv.ctx, model)
			if err != nil {
				logger.Warnf("create detect looper for heapster %s error: %v", model.ID, err)
				continue
			}
			oldModel, ok := srv.loopers[model.ID]
			if ok {
				oldModel.Stop()
			}
			looper.Run()
			srv.loopers[model.ID] = looper
		}
		logger.Infof("heapster reloaded")
		oldSet = newSet
	}
}

// Stop 组件停止
func (srv *HealthySrv) Stop() {
	logger := middlewares.GetLogger(srv.ctx)
	// 停止主循环
	logger.Infof("stopping main looper")
	srv.cancel()
	<-srv.done

	// 停止looper
	logger.Infof("stopping all detect looper")
	var wg = sync.WaitGroup{}

	for _, looper := range srv.loopers {
		wg.Add(1)
		go func(looper detectors.DetectLooper) {
			defer wg.Done()
		}(looper)
	}
	wg.Wait()

	logger.Infof("healthy server stop")
}

// 服务内部初始化
func (srv *HealthySrv) init() {
	srv.ctx = middlewares.WithLogger(context.Background(), srv.LogLevel, os.Stdout)
	srv.ctx = middlewares.WithRedisConn(srv.ctx, srv.RedisHost, srv.RedisPassword, srv.RedisDB)
	srv.ctx = middlewares.WithRateLimiter(srv.ctx, srv.RedisHost, srv.RedisPassword, srv.RedisDB)

	srv.ctx, srv.cancel = context.WithCancel(srv.ctx)
}
