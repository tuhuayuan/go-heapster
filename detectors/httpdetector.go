package detectors

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"

	"github.com/Sirupsen/logrus"
)

func init() {
	registCreator(string(models.CheckTypeHTTP), httpDetectorCreator)
}

var httpDetectorCreator detectorCreator = func(ctx context.Context, hp models.Heapster) (detector, error) {
	dtr := &httpDetector{
		model:  hp,
		logger: middlewares.GetLogger(ctx),
	}
	// 获取监控目标组
	groups, err := hp.GetApplyGroups(ctx)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		eps := g.Endpoints.Unfold().Exclude(g.Excluded)
		for _, ep := range eps {
			epURL := fmt.Sprintf("http://%s:%d", string(ep), hp.Port)
			if hp.Location != "" {
				epURL += hp.Location
			}
			req, err := http.NewRequest("GET", epURL, nil)
			if err != nil {
				dtr.logger.Warnf("endpoint %v ignore by error %v", ep, err)
				continue
			}
			if hp.Host != "" {
				req.Host = hp.Host
			}
			dtr.reqs = append(dtr.reqs, req)
		}
	}
	return dtr, nil
}

type httpDetector struct {
	model  models.Heapster
	logger *logrus.Logger
	wg     sync.WaitGroup
	reqs   []*http.Request
}

func (dtr *httpDetector) plumb(ctx context.Context) models.ProbeLogs {
	// 最大并行
	const pageSize = 64
	var (
		total = len(dtr.reqs)
		pages = int(total / pageSize)
	)
	// 按照单次并发计算容量
	probeLogs := make(models.ProbeLogs, 0, pageSize)
	for i := 0; i <= pages; i++ {
		select {
		case <-ctx.Done():
			// 结束
			return probeLogs
		default:
		}
		// 计算分页起始结束偏移
		start := i * pageSize
		end := start + pageSize
		if i == pages {
			end = start + total%pageSize
		}
		for _, req := range dtr.reqs[start:end] {
			// 设置超时上下文
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(dtr.model.Timeout))
			dtr.wg.Add(1)
			// 启动goroutine
			go func(req *http.Request, ctx context.Context, cancel func()) {
				defer dtr.wg.Done()
				defer cancel()
				// 准备报告
				beginAt := time.Now()
				probeLog := models.ProbeLog{
					Heapster:  string(dtr.model.ID),
					Target:    req.URL.String(),
					Timestamp: beginAt,
				}
				// 测试连接
				resp, err := http.DefaultClient.Do(req.WithContext(ctx))
				probeLog.Elapsed = time.Now().Sub(beginAt)
				if err != nil {
					probeLog.Response = err.Error()
					probeLog.Failed = 1
				} else if !dtr.checkResponseCode(resp.StatusCode) {
					probeLog.Response = fmt.Sprintf("http response code %d", resp.StatusCode)
					probeLog.Failed = 1
				} else {
					probeLog.Response = "ok"
					probeLog.Success = 1
				}
				// 添加日志
				probeLogs = append(probeLogs, probeLog)
			}(req, timeoutCtx, cancel)
		}
		dtr.wg.Wait()
	}
	return probeLogs
}

func (dtr *httpDetector) checkResponseCode(c int) bool {
	for _, code := range dtr.model.AcceptCode {
		if code == c {
			return true
		}
	}
	return false
}
