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
	config, ok := model.Config["sms"].(map[string]interface{})
	// 需要解析配置
	if !ok {
		return nil, fmt.Errorf("no config found")
	}
	if val, ok := config["type"].(string); ok {
		spType = val
	}
	// 如果使用联通的短信接口
	if spType == "unicom" {
		if val, ok := config["sp"].(string); ok {
			spID = val
		}
		if val, ok := config["username"].(string); ok {
			spUsername = val
		}
		if val, ok := config["password"].(string); ok {
			spPassword = val
		}
		if val, ok := config["targets"].([]string); ok {
			numbers = val
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
func (sms *smsNotifier) Send(ctx context.Context, reports models.Reports) error {
	var (
		hp      *models.Heapster
		limiter = middlewares.GetRateLimiter(ctx)
		logger  = middlewares.GetLogger(ctx)
	)
	for _, rp := range reports {
		if rp.Validate() != nil {
			continue
		}
		hp = &models.Heapster{
			ID: models.SerialNumber(string(rp.Labels[models.ReportNameFor])),
		}
		if err := hp.Fill(ctx); err != nil {
			continue
		}
		break
	}
	if hp == nil {
		return fmt.Errorf("heapster not found")
	}
	if limiter.TryAccept([]string{string(hp.ID)}, 5*time.Minute, 1) {
		return fmt.Errorf("rate controll by heapster")
	}
	if limiter.TryAccept(sms.numbers, 5*time.Minute, 1) {
		return fmt.Errorf("rate controll by phone")
	}
	tpl := fmt.Sprintf("%s提醒：%s需要%s请查阅%s",
		"监控",
		"（"+hp.Name+"）的实例出现异常",
		"及时处理",
		"监控报告")
	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	go func() {
		result := sms.provider.SendMessage(sendCtx, tpl, sms.numbers)
		cancel()
		if result.Result != 0 {
			logger.Warnf("send sms error %s", result.Error())
		}
	}()

	return nil
}
