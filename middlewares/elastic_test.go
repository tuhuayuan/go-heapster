package middlewares

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"zonst/qipai-golang-libs/httputil"

	"github.com/stretchr/testify/assert"
)

type elasticTestData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestElasticConn(t *testing.T) {
	ctx := httputil.WithHTTPContext(nil)
	handler := httputil.HandleFunc(ctx, ElasticConnHandler([]string{"http://localhost:9200"}, "", ""),
		func(w http.ResponseWriter, r *http.Request) {
			conn := GetElasticConn(r.Context())
			// 健康测试
			resp, err := conn.ClusterHealth().Do(context.Background())
			assert.NoError(t, err)
			fmt.Println(resp)
			var wg sync.WaitGroup
			// 并发测试
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					data := &elasticTestData{
						ID:   id,
						Name: "tuhuayuan",
					}
					resp, err := conn.Index().
						Index("test_middleware").
						Type("test_doc").
						BodyJson(data).Do(context.Background())
					if err != nil {
						fmt.Println("id:", id, err)
					} else {
						fmt.Println(resp)
					}
				}(i)
			}
			wg.Wait()
		})
	req := httptest.NewRequest("GET", "/", bytes.NewReader([]byte{}))
	handler(httptest.NewRecorder(), req)
}
