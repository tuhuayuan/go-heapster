package daemons

import (
	"context"
	"os"

	"sync"
	"time"
	"zonst/qipai/gamehealthysrv/alerts"
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

	// InfluxDB配置
	InfluxURL    string `json:"influx_url"`
	InfluxUser   string `json:"influx_user"`
	InfluxPasswd string `json:"influx_passwd"`
	InfluxDB     string `json:"influx_db"`

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
	alerts  map[models.SerialNumber]alerts.Alert
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
		alerts:  make(map[models.SerialNumber]alerts.Alert),
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
		newSet, err = models.FetchHeapsters(srv.ctx)
		if err != nil {
			logger.Warnf("load heapster error %v", err)
			continue
		}
		added, modified, deleted := newSet.Diff(oldSet)
		total := len(added) + len(modified) + len(deleted)
		if total == 0 {
			continue
		} else {
			logger.Infof("found %d heapster changed", total)
		}
		srv.loadAdded(added)
		srv.loadModified(modified)
		srv.loadDeteted(deleted)

		logger.Infof("heapster reloaded")
		oldSet = newSet
	}
}

func (srv *HealthySrv) installHeapster(looper detectors.DetectLooper, alert alerts.Alert, model models.Heapster) {
	looper.Run()
	alert.TurnOn()
	srv.alerts[model.ID] = alert
	srv.loopers[model.ID] = looper
}

func (srv *HealthySrv) uninstallHeapster(model models.Heapster) {
	alert, ok := srv.alerts[model.ID]
	if ok {
		alert.TurnOn()
		delete(srv.alerts, model.ID)
	}
	looper, ok := srv.loopers[model.ID]
	if ok {
		looper.Stop()
		delete(srv.loopers, model.ID)
	}
}

func (srv *HealthySrv) loadAdded(keys models.HeapsterSetKeys) {
	logger := middlewares.GetLogger(srv.ctx)
	entry := logger.WithField("HealthyServer", "LoadAdded")
	entry.Infof("added %d heapster", len(keys))

	for _, k := range keys {
		model := models.Heapster{
			ID: models.SerialNumber(k),
		}
		// 模型信息
		if err := model.Fill(srv.ctx); err != nil {
			entry.Warnf("load heapster %s error: %v", k, err)
			continue
		}
		// 创建新的
		looper, err := detectors.NewDetectLooper(srv.ctx, model)
		if err != nil {
			entry.Warnf("create looper for heapster %s error: %v", model.ID, err)
			continue
		}
		alert, err := alerts.NewAlert(srv.ctx, model)
		if err != nil {
			entry.Warnf("create alert for heapster %s error: %v", model.ID, err)
			continue
		}
		srv.installHeapster(looper, alert, model)
	}
}

func (srv *HealthySrv) loadModified(keys models.HeapsterSetKeys) {
	logger := middlewares.GetLogger(srv.ctx)
	entry := logger.WithField("HealthyServer", "LoadModified")
	entry.Infof("modified %d heapster", len(keys))

	for _, k := range keys {
		model := models.Heapster{
			ID: models.SerialNumber(k),
		}
		// 模型信息
		if err := model.Fill(srv.ctx); err != nil {
			entry.Warnf("load heapster %s error: %v", k, err)
			continue
		}
		// 创建新的
		looper, err := detectors.NewDetectLooper(srv.ctx, model)
		if err != nil {
			entry.Warnf("create looper for heapster %s error: %v", model.ID, err)
			continue
		}
		alert, err := alerts.NewAlert(srv.ctx, model)
		if err != nil {
			entry.Warnf("create alert for heapster %s error: %v", model.ID, err)
			continue
		}
		srv.uninstallHeapster(model)
		srv.installHeapster(looper, alert, model)
	}
}

func (srv *HealthySrv) loadDeteted(keys models.HeapsterSetKeys) {
	logger := middlewares.GetLogger(srv.ctx)
	entry := logger.WithField("HealthyServer", "LoadDeleted")
	entry.Infof("deleted %d heapster", len(keys))

	for _, k := range keys {
		model := models.Heapster{
			ID: models.SerialNumber(k),
		}
		srv.uninstallHeapster(model)
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
	logger.Infof("stopping all looper and alert")
	var wg = sync.WaitGroup{}

	for _, looper := range srv.loopers {
		wg.Add(1)

		go func(looper detectors.DetectLooper) {
			defer wg.Done()
			looper.Stop()
		}(looper)
	}
	for _, alert := range srv.alerts {
		wg.Add(1)
		go func(alert alerts.Alert) {
			defer wg.Done()
			alert.TurnOff()
		}(alert)
	}

	wg.Wait()
	logger.Infof("healthy server stop")
}

// 服务内部初始化
func (srv *HealthySrv) init() {
	var err error

	srv.ctx = middlewares.WithLogger(context.Background(), srv.LogLevel, os.Stdout)
	srv.ctx = middlewares.WithRedisConn(srv.ctx, srv.RedisHost, srv.RedisPassword, srv.RedisDB)
	srv.ctx = middlewares.WithRateLimiter(srv.ctx, srv.RedisHost, srv.RedisPassword, srv.RedisDB)
	srv.ctx, err = middlewares.WithInfluxDB(srv.ctx, srv.InfluxURL, srv.InfluxUser, srv.InfluxPasswd, srv.InfluxDB)
	if err != nil {
		panic(err)
	}
	srv.ctx, srv.cancel = context.WithCancel(srv.ctx)
}
