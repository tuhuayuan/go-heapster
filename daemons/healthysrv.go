package daemons

import (
	"zonst/qipai/logagent/utils"

	"github.com/Sirupsen/logrus"
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

	LogLevel   uint8    `json:"log_level"`
	AccessKeys []string `json:"accesskeys"`

	// 私有配置
	logger *logrus.Logger
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
	return nil, nil
}

// Start 组件启动
func (srv *HealthySrv) Start() {
	srv.logger.Infof("healthy server start.")
}

// Stop 组件停止
func (srv *HealthySrv) Stop() {
	srv.logger.Infof("healthy server stop.")
}

// 服务内部初始化
func (srv *HealthySrv) init() {

}
