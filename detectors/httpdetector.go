package detectors

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"zonst/qipai/gamehealthysrv/models"
)

func init() {
	registCreator(string(models.CheckTypeHTTP), httpDetectorCreator)
}

var httpDetectorCreator detectorCreator = func(ctx context.Context, hp models.Heapster) (detector, error) {
	dtr := &httpDetector{
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
			dtr.target = fmt.Sprintf("http://%s:%d", string(ep), hp.Port)
			if hp.Location != "" {
				dtr.target += hp.Location
			}
			req, err := http.NewRequest("GET", dtr.target, nil)
			if err != nil {
				log.Printf("endpoint %v ignore by %v", ep, err)
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
	model models.Heapster

	wg     sync.WaitGroup
	reqs   []*http.Request
	target string
}

func (dtr *httpDetector) plumb(ctx context.Context) (models.Reports, error) {
	// 最大并行
	const pageSize = 255
	var (
		reports  models.Reports
		errs     int
		canceled bool
		total    = len(dtr.reqs)
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
		for _, req := range dtr.reqs[start:end] {
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(dtr.model.Timeout))
			dtr.wg.Add(1)
			go func(req *http.Request, ctx context.Context, cancel func()) {
				defer dtr.wg.Done()
				defer cancel()

				resp, err := http.DefaultClient.Do(req.WithContext(ctx))
				if err != nil {
					reports = append(reports, models.Report{
						Labels: models.LabelSet{
							models.ReportNameFor:    models.LabelValue(dtr.model.ID),
							models.ReportNameTarget: models.LabelValue(dtr.target),
							models.ReportNameResult: models.LabelValue(err.Error()),
						},
					})
					errs++
				} else if !dtr.checkResponseCode(resp.StatusCode) {
					reports = append(reports, models.Report{
						Labels: models.LabelSet{
							models.ReportNameFor:    models.LabelValue(dtr.model.ID),
							models.ReportNameTarget: models.LabelValue(dtr.target),
							models.ReportNameResult: models.LabelValue(
								fmt.Sprintf("http response code %d", resp.StatusCode),
							),
						},
					})
					errs++
				} else {
					reports = append(reports, models.Report{
						Labels: models.LabelSet{
							models.ReportNameFor:    models.LabelValue(dtr.model.ID),
							models.ReportNameTarget: models.LabelValue(dtr.target),
							models.ReportNameResult: "ok",
						},
					})
				}
			}(req, timeoutCtx, cancel)
		}
		dtr.wg.Wait()
	}
	if errs > 0 || canceled {
		return reports, fmt.Errorf("found %d errors, canceled %v", errs, canceled)
	}
	return reports, nil
}

func (dtr *httpDetector) checkResponseCode(c int) bool {
	for _, code := range dtr.model.AcceptCode {
		if code == c {
			return true
		}
	}
	return false
}
