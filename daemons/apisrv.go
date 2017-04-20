package daemons

import (
	"context"
	"net/http"
	"os"
	"time"

	"zonst/qipai-golang-libs/httputil"
	"zonst/qipai/gamehealthysrv/handlers"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/logagent/utils"

	"github.com/gorilla/mux"
)

// HealthyAPISrv 配置结构体
type HealthyAPISrv struct {
	utils.InputPluginConfig

	// http配置
	Host string `json:"host"`

	// Redis配置
	RedisHost     string `json:"redis_host"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// InfluxDB配置
	InfluxURL    string `json:"influx_url"`
	InfluxUser   string `json:"influx_user"`
	InfluxPasswd string `json:"influx_passwd"`

	// 通用配置
	LogLevel   int      `json:"log_level"`
	AccessKeys []string `json:"accesskeys"`

	// 内部变量
	server *http.Server
	ctx    context.Context
}

// 服务名称
const (
	HealthyAPISrvName = "gamehealthyapisrv"
)

func init() {
	utils.RegistInputHandler(HealthyAPISrvName, HealthyAPISrvInit)
}

// HealthyAPISrvInit Healthy API 配置服务初始化
func HealthyAPISrvInit(part *utils.ConfigPart) (srv *HealthyAPISrv, err error) {
	config := HealthyAPISrv{
		InputPluginConfig: utils.InputPluginConfig{
			TypePluginConfig: utils.TypePluginConfig{
				Type: HealthyAPISrvName,
			},
		},
	}
	if err = utils.ReflectConfigPart(part, &config); err != nil {
		return
	}
	config.init()
	srv = &config
	return
}

// Start 组件启动
func (srv *HealthyAPISrv) Start() {
	logger := middlewares.GetLogger(srv.ctx)
	logger.Infof("gamesms server listen at %s", srv.Host)
	if err := srv.server.ListenAndServe(); err != nil {
		logger.Fatalf("gamesms server error %s", err)
	}
}

// Stop 组件停止
func (srv *HealthyAPISrv) Stop() {
	logger := middlewares.GetLogger(srv.ctx)
	logger.Infof("gamesms server stop.")
}

// 服务内部初始化
func (srv *HealthyAPISrv) init() {
	r := mux.NewRouter()
	srv.server = &http.Server{
		Handler:      r,
		Addr:         srv.Host,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ctx = httputil.WithHTTPContext(nil)
	srv.ctx = middlewares.WithLogger(srv.ctx, srv.LogLevel, os.Stdout)
	httputil.Use(srv.ctx, middlewares.LoggerHandler(srv.LogLevel, os.Stdout))
	httputil.Use(srv.ctx, middlewares.RedisConnHandler(srv.RedisHost, srv.RedisPassword, srv.RedisDB))
	httputil.Use(srv.ctx, middlewares.InfluxDBHandler(srv.InfluxURL, srv.InfluxUser, srv.InfluxPasswd))

	// api 版本
	v1 := r.PathPrefix("/v1").Subrouter()

	r.NotFoundHandler = middlewares.ErrorNotFoundHandler()
	// group
	v1.HandleFunc("/gamehealthy/group",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.CreateGroupReq{}),
			handlers.CreateGroupHandler)).Methods("POST")
	v1.HandleFunc("/gamehealthy/group",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.UpdateGroupReq{}),
			handlers.UpdateGroupHandler)).Methods("PATCH", "PUT")
	v1.HandleFunc("/gamehealthy/group",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.DeleteGroupReq{}),
			handlers.DeleteGroupHandler)).Methods("DELETE")
	v1.HandleFunc("/gamehealthy/group",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.FetchGroupReq{}),
			handlers.FetchGroupHandler)).Methods("GET")

	// notifier
	v1.HandleFunc("/gamehealthy/notifier",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.CreateNotifierReq{}),
			handlers.CreateNotifierHandler)).Methods("POST")
	v1.HandleFunc("/gamehealthy/notifier",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.UpdateNotifierReq{}),
			handlers.UpdateNotifierHandler)).Methods("PATCH", "PUT")
	v1.HandleFunc("/gamehealthy/notifier",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.DeleteNotifierReq{}),
			handlers.DeleteNotifierHandler)).Methods("DELETE")
	v1.HandleFunc("/gamehealthy/notifier",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.FetchNotifierReq{}),
			handlers.FetchNotifierHandler)).Methods("GET")

	// heapster
	v1.HandleFunc("/gamehealthy/heapster",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.CreateHeapsterReq{}),
			handlers.CreateHeapsterHandler)).Methods("POST")
	v1.HandleFunc("/gamehealthy/heapster",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.UpdateHeapsterReq{}),
			handlers.UpdateHeapsterHandler)).Methods("PATCH", "PUT")
	v1.HandleFunc("/gamehealthy/heapster",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.DeleteHeapsterReq{}),
			handlers.DeleteHeapsterHandler)).Methods("DELETE")
	v1.HandleFunc("/gamehealthy/heapster",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.FetchHeapsterReq{}),
			handlers.FetchHeapsterHandler)).Methods("GET")
	v1.HandleFunc("/gamehealthy/heapster/status",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.FetchHeapsterStatusReq{}),
			handlers.FetchHeapsterStatusHandler)).Methods("GET")

	// report
	v1.HandleFunc("/gamehealthy/report",
		httputil.HandleFunc(srv.ctx,
			middlewares.BindBody(&handlers.FetchReportReq{}),
			handlers.FetchReportHandler)).Methods("GET")
}
