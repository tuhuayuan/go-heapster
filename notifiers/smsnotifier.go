package notifiers

import (
	"context"
	"fmt"
	"time"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
)

func init() {
	registCreator("sms", smsNotifierCreator)
}

var smsNotifierCreator = func(model models.HeapsterNotifier) (Notifier, error) {
	var (
		spType     string
		spID       string
		spUsername string
		spPassword string
		numbers    []string
	)

	if val, ok := model.Config["type"].(string); ok {
		spType = val
	}
	// 如果使用联通的短信接口
	if spType == "unicom" {
		if val, ok := model.Config["sp"].(string); ok {
			spID = val
		}
		if val, ok := model.Config["username"].(string); ok {
			spUsername = val
		}
		if val, ok := model.Config["password"].(string); ok {
			spPassword = val
		}
		if vals, ok := model.Config["targets"].([]interface{}); ok {
			for _, v := range vals {
				numbers = append(numbers, v.(string))
			}
		}
		smsConfig := middlewares.UnicomConfig{
			SPCode:   spID,
			Username: spUsername,
			Password: spPassword,
		}
		p, err := middlewares.CreateSMSProvider(string(spType), smsConfig)
		if err != nil {
			return nil, err
		}
		return &smsNotifier{
			provider: p,
			numbers:  numbers,
		}, nil
	}
	// 暂时不支持其它的
	return nil, fmt.Errorf("sms provider type %s not support", spType)
}

type smsNotifier struct {
	provider middlewares.SMSProvider
	numbers  []string
}

// Send 短信不能发那么多字, 只能发一个大概的描述
func (sms *smsNotifier) Send(ctx context.Context, report models.Report) error {
	var (
		limiter = middlewares.GetRateLimiter(ctx)
		logger  = middlewares.GetLogger(ctx)
	)
	// 获取heapster信息
	hp := &models.Heapster{
		ID: models.SerialNumber(report.Heapster),
	}
	if err := hp.Fill(ctx); err != nil {
		return fmt.Errorf("report missing heapster")
	}
	// 流量控制
	if limiter.TryAccept([]string{string(hp.ID)}, 5*time.Minute, 1) {
		return fmt.Errorf("rate controll by heapster")
	}
	if limiter.TryAccept(sms.numbers, 5*time.Minute, 1) {
		return fmt.Errorf("rate controll by phone")
	}
	// 构建消息
	tpl := fmt.Sprintf("%s提醒：%s需要%s请查阅%s",
		"监控",
		"("+hp.Name+")中的("+report.Target+")出现异常",
		"及时处理",
		"监控报告")
	// 发送超时默认5秒
	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	result := sms.provider.SendMessage(sendCtx, tpl, sms.numbers)
	cancel()
	if result.Result != 0 {
		logger.Warnf("send sms error %s", result.Error())
	}
	return nil
}
