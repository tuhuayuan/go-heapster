package middlewares

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"zonst/qipai-golang-libs/httputil"

	"context"

	"github.com/stretchr/testify/assert"
)

func TestUnicomSend(t *testing.T) {
	config := UnicomConfig{
		SPCode:   "103905",
		Username: "zz_sj",
		Password: "www.zonst.org",
	}
	p, err := CreateSMSProvider("unicom", config)
	assert.NoError(t, err)
	sid := p.SendMessage(context.Background(), "您的验证码为{123456}", []string{"13879156403"})
	assert.Equal(t, 0, sid.Result)
	fmt.Println(sid.SerialNumber)
}

func TestUnicomReceipts(t *testing.T) {
	config := UnicomConfig{
		SPCode:   "",
		Username: "",
		Password: "",
	}
	p, err := CreateSMSProvider("unicom", config)
	assert.NoError(t, err)
	receipts, _ := p.FetchReceipts()
	fmt.Println(receipts)
}

func TestSMSHelper(t *testing.T) {
	config := UnicomConfig{
		SPCode:   "",
		Username: "",
		Password: "",
	}
	ctx := httputil.NewHTTPContext()
	handler := ctx.HandleFunc(SMSHelper("unicom", config),
		func(w http.ResponseWriter, r *http.Request) {
			provider := GetSMSProvider(r, "unicom")
			provider.FetchReceipts()
		})
	req := httptest.NewRequest("GET", "/", bytes.NewReader([]byte{}))
	resp := httptest.NewRecorder()
	handler(resp, req)
	assert.Equal(t, 200, resp.Code)
}
