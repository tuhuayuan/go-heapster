package detectors

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
)

func init() {
	registCreator(string(models.CheckTypeTCP), tcpDetectorCreator)
}

var tcpDetectorCreator detectorCreator = func(ctx context.Context, hp models.Heapster) (detector, error) {
	dtr := &tcpDetector{
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
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", string(ep), hp.Port))
			if err != nil {
				dtr.logger.Warnf("tpc endpoint %v ignore by error %v", ep, err)
				continue
			}
			dtr.address = append(dtr.address, addr)
		}
	}
	return dtr, nil
}

type tcpDetector struct {
	model   models.Heapster
	logger  *logrus.Logger
	wg      sync.WaitGroup
	address []*net.TCPAddr
}

func (dtr *tcpDetector) plumb(ctx context.Context) models.ProbeLogs {
	// 最大并行
	const pageSize = 64
	var (
		total = len(dtr.address)
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
		for _, addr := range dtr.address[start:end] {
			// 设置超时上下文
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(dtr.model.Timeout))
			dtr.wg.Add(1)
			// 启动goroutine
			go func(addr *net.TCPAddr, ctx context.Context, cancel func()) {
				defer dtr.wg.Done()
				defer cancel()
				// 准备报告
				beginAt := time.Now()
				probeLog := models.ProbeLog{
					Heapster:  string(dtr.model.ID),
					Target:    addr.String(),
					Timestamp: beginAt,
				}
				// 测试连接
				dialer := &net.Dialer{}
				conn, err := dialer.DialContext(ctx, "tcp", addr.String())
				probeLog.Elapsed = time.Now().Sub(beginAt)
				if err != nil {
					probeLog.Response = err.Error()
					probeLog.Failed = 1
				} else {
					probeLog.Response = "ok"
					probeLog.Success = 1
					conn.Close()
				}
				// 添加日志
				probeLogs = append(probeLogs, probeLog)
			}(addr, timeoutCtx, cancel)
		}
		dtr.wg.Wait()
	}
	return probeLogs
}
