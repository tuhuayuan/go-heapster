package detectors

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"zonst/qipai/gamehealthysrv/models"
)

func init() {
	registCreator(string(models.CheckTypeTCP), tcpDetectorCreator)
}

var tcpDetectorCreator detectorCreator = func(ctx context.Context, hp models.Heapster) (detector, error) {
	dtr := &tcpDetector{
		model: hp,
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
				log.Printf("endpoint %v ignore by %v", ep, err)
				continue
			}
			dtr.address = append(dtr.address, addr)
		}
	}
	return dtr, nil
}

type tcpDetector struct {
	model models.Heapster

	wg      sync.WaitGroup
	address []*net.TCPAddr
}

func (dtr *tcpDetector) plumb(ctx context.Context) (models.Reports, error) {
	// 最大并行
	const pageSize = 255
	var (
		reports  models.Reports
		errs     int
		canceled bool
		total    = len(dtr.address)
		pages    = int(total / pageSize)
	)
	for i := 0; i <= pages; i++ {
		// 提前结束
		select {
		case <-ctx.Done():
			canceled = true
			break
		default:
		}
		// 计算分页
		start := i * pageSize
		end := start + pageSize
		if i == pages {
			end = start + total%pageSize
		}
		// 并发访问
		for _, addr := range dtr.address[start:end] {
			dtr.wg.Add(1)
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(dtr.model.Timeout))

			go func(addr *net.TCPAddr, ctx context.Context, cancel func()) {
				defer dtr.wg.Done()
				defer cancel()
				dialer := &net.Dialer{}
				conn, err := dialer.DialContext(ctx, "tcp", addr.String())
				if err != nil {
					reports = append(reports, models.Report{
						Labels: models.LabelSet{
							models.ReportNameFor:    models.LabelValue(dtr.model.ID),
							models.ReportNameTarget: models.LabelValue(addr.String()),
							models.ReportNameResult: models.LabelValue(err.Error()),
						},
					})
					errs++
					return
				}
				conn.Close()
				reports = append(reports, models.Report{
					Labels: models.LabelSet{
						models.ReportNameFor:    models.LabelValue(dtr.model.ID),
						models.ReportNameTarget: models.LabelValue(addr.String()),
						models.ReportNameResult: models.LabelValue("ok"),
					},
				})
			}(addr, timeoutCtx, cancel)
		}
		dtr.wg.Wait()
	}
	if errs > 0 || canceled {
		return reports, fmt.Errorf("found %d errors, canceled %v", errs, canceled)
	}
	return reports, nil
}
