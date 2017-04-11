package middlewares

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"context"
	"zonst/qipai-golang-libs/httputil"
)

func init() {
	RegistSMSProvider("unicom", unicomCreator)
}

// GetSMSProvider 获取短信发送器
func GetSMSProvider(r *http.Request, name string) SMSProvider {
	ctx := httputil.GetContext(r)
	config := ctx.Value(name)
	provider, err := CreateSMSProvider(name, config)
	if err != nil {
		// 设计错误经快抛出
		panic(err)
	}
	return provider
}

// SMSHelper 短信发送助手
func SMSHelper(name string, config interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := httputil.GetContext(r)
		ctx.Set(name, config)
		ctx.Next()
	}
}

// SMSReceipt 短信回执
type SMSReceipt struct {
	SerialNumber int64
	Number       string
	Status       int
}

// SMSResult 短信发送结果
type SMSResult struct {
	Result       int
	Desc         string
	SerialNumber int64
}

// Error 实现标准错误
func (r SMSResult) Error() string {
	return r.Desc
}

// SMSProvider 短信服务提供者
type SMSProvider interface {
	// 按照指定模版发送一条消息到号码列表
	SendMessage(ctx context.Context, tpl string, numbers []string) SMSResult
	// 批量获取短信回执
	FetchReceipts() ([]SMSReceipt, error)
}

var (
	providerFactoryMap = make(map[string]interface{})
)

// CreateSMSProvider 创建短信提供者
func CreateSMSProvider(name string, config interface{}) (SMSProvider, error) {
	creater, ok := providerFactoryMap[name]
	if !ok {
		return nil, fmt.Errorf("sms provider name %s not found", name)
	}
	inj := httputil.NewInjector()
	inj.Map(config)
	inj.Map(name)
	v, err := inj.Invoke(creater)
	if err != nil {
		return nil, err
	}
	if err = httputil.CheckError(v); err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return nil, fmt.Errorf("no provider created by name %s", name)
	}
	return v[0].Interface().(SMSProvider), nil
}

// RegistSMSProvider 注册
func RegistSMSProvider(name string, creator interface{}) {
	if !httputil.IsFunction(creator) {
		panic("creator must be a function")
	}
	providerFactoryMap[name] = creator
}

// UnicomConfig 联通短信接口配置
type UnicomConfig struct {
	SPCode   string
	Username string
	Password string
}

// UnicomProvider 联通短信接口服务者
type UnicomProvider struct {
	config UnicomConfig
}

// SendMessage 实现接口
func (unicom *UnicomProvider) SendMessage(ctx context.Context, tpl string, numbers []string) (result SMSResult) {
	var (
		gbkMessage []byte
		phones     string
		err        error
	)
	defer func() {
		if err != nil {
			result.Result = -1
			result.Desc = err.Error()
			return
		}
	}()
	// 组建号码列表
	for _, n := range numbers {
		phones += n + ","
	}
	// 解码模版内容
	reader := transform.NewReader(bytes.NewReader([]byte(tpl)), simplifiedchinese.GBK.NewEncoder())
	gbkMessage, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	// 设置序列号
	result.SerialNumber = time.Now().UnixNano()
	postData := make(url.Values)
	postData.Add("SpCode", unicom.config.SPCode)
	postData.Add("LoginName", unicom.config.Username)
	postData.Add("Password", unicom.config.Password)
	postData.Add("MessageContent", string(gbkMessage))
	postData.Add("SerialNumber", strconv.FormatInt(result.SerialNumber, 10))
	postData.Add("UserNumber", phones)

	req, err := http.NewRequest(
		"POST",
		"http://gd.ums86.com:8899/sms/Api/Send.do",
		bytes.NewReader([]byte(postData.Encode())))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	respValues, err := url.ParseQuery(string(respData))
	if err != nil {
		return
	}
	code, err := strconv.Atoi(respValues.Get("result"))
	if code != 0 {
		result.Result = code
		reader := transform.NewReader(
			bytes.NewReader([]byte(respValues.Get("description"))), simplifiedchinese.GBK.NewDecoder())
		decoded, err := ioutil.ReadAll(reader)
		if err == nil {
			result.Desc = string(decoded)
		}
	}
	return
}

// FetchReceipts 实现接口
func (unicom *UnicomProvider) FetchReceipts() ([]SMSReceipt, error) {
	var (
		receipts []SMSReceipt
	)
	postData := make(url.Values)
	postData.Add("SpCode", unicom.config.SPCode)
	postData.Add("LoginName", unicom.config.Username)
	postData.Add("Password", unicom.config.Password)
	resp, err := http.PostForm("http://smsapi.ums86.com:8888/sms/Api/report.do", postData)
	if err != nil {
		return nil, err
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respValues, err := url.ParseQuery(string(respData))
	if err != nil {
		return nil, err
	}
	code, err := strconv.Atoi(respValues.Get("result"))
	if err != nil {
		return nil, err
	}
	out := respValues.Get("out")
	if code == 0 && out != "" {
		receiptsData := strings.Split(out, ";")
		for _, receiptData := range receiptsData {
			receiptFields := strings.Split(receiptData, ",")
			if len(receiptFields) != 3 {
				continue
			}
			receipt := SMSReceipt{
				Number: receiptFields[1],
			}
			if sn, err := strconv.ParseInt(receiptFields[0], 10, 64); err == nil {
				receipt.SerialNumber = sn
			}
			if status, err := strconv.ParseInt(receiptFields[2], 10, 64); err == nil {
				receipt.Status = int(status)
			}
			receipts = append(receipts, receipt)
		}
	}
	return receipts, nil
}

// 注册到工厂的方法
func unicomCreator(config UnicomConfig) SMSProvider {
	return &UnicomProvider{
		config: config,
	}
}
