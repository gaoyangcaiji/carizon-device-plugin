package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"carizon-device-plugin/pkg/logger"

	"github.com/go-resty/resty/v2"
)

const (
	// TODO: remove this hardcode
	adminUser = ""
	adminUID  = ""

	// HeaderContentType content type
	HeaderContentType = "Content-Type"

	// ContentTypeJson json格式body
	ContentTypeJson = "application/json"
	// ContentTypeFormEncoded form格式body
	ContentTypeFormEncoded = "application/x-www-form-urlencoded"
	// ContentTypeText text格式body
	ContentTypeText = "text/plain; charset=utf-8"
)

var (
	// EmptyHeader 空header头
	EmptyHeader = map[string]string{}
	// EmptyQuery Get请求空的query
	EmptyQuery = map[string]string{}
)

type client struct {
	header map[string]string
	cli    *resty.Client
}

type Result struct {
	Body       []byte
	Err        error
	StatusCode int
	Status     string
	Header     http.Header
}

func NewClient() IClient {
	headers := map[string]string{
		"Content-Type":     "application/json",
		"X-Forwarded-User": adminUser,
		"X-Forwarded-Uid":  adminUID,
	}

	cli := resty.New()
	cli.SetTimeout(10 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(10 * time.Second).
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(3)).
		SetHeaders(headers).
		EnableTrace()

	return &client{
		cli:    cli,
		header: headers,
	}
}

type IClient interface {
	DoGet(ctx context.Context, url string, header, query map[string]string) (data []byte, err error)
	DoPost(ctx context.Context, url string, header map[string]string, body interface{}) (r *Result)
	DoDelete(ctx context.Context, url string, header map[string]string, timeout int) (data []byte, err error)
	GetCli() *resty.Client
}

func (c *client) GetCli() *resty.Client {
	return c.cli
}

// DoGet get方法 timeout单位：s
func (c *client) DoGet(ctx context.Context, url string, header, query map[string]string) (body []byte, err error) {
	req := c.cli.R().SetContext(ctx).SetHeaders(header).SetQueryParams(query)
	resp, err := req.Get(url)
	if err != nil {
		return
	}
	duration := resp.Time()

	logger.Wrapper.Infof("retry times:%d", resp.Request.TraceInfo().RequestAttempt)

	body = resp.Body()

	logger.Wrapper.Debugf("[HTTP][Get] url:%s status:%d query:%s cost: %fs",
		url, resp.StatusCode(), req.RawRequest.URL.RawQuery, duration.Seconds())

	if resp.StatusCode() != http.StatusOK {
		err = newHttpError(url, int32(resp.StatusCode()))
		return
	}

	return
}

// DoPost Post方法，只支持json格式
// 对于body会做校验，使用form格式之前一定要设置header头
//
//	json: []byte
//	form: map[string]string
//
// timeout单位：s
func (c *client) DoPost(ctx context.Context, url string, header map[string]string, body interface{}) (result *Result) {
	// 默认json格式
	if header[HeaderContentType] == "" {
		header[HeaderContentType] = ContentTypeJson
	}

	req := c.cli.R().SetContext(ctx).SetHeaders(header).SetBody(body)

	if len(req.Header) == 0 {
		req.Header = make(http.Header)
	}

	// 删除 Accept-Encoding 避免返回值被压缩
	req.Header.Del("Accept-Encoding")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := req.Post(url)
	if err != nil {
		result.Err = err
		return
	}

	logger.Wrapper.Debugf("[CmdbApiClient] cost: %dms,%s %s with body %s, response status: %s, "+
		"response body: %s", resp.Time().Seconds(),
		resp.Request.Method, url, body, resp.Status, resp.Body())

	result.Body = resp.Body()
	result.StatusCode = resp.StatusCode()
	result.Status = resp.Status()
	result.Header = resp.Header()
	return result
}

// DoDelete delete方法 timeout单位：s
func (c *client) DoDelete(ctx context.Context, url string, header map[string]string, timeout int) (body []byte, err error) {

	start := time.Now()
	// https://github.com/go-resty/resty#retries
	req := c.cli.R().SetContext(ctx).EnableTrace().SetHeaders(header)
	resp, err := req.Delete(url)
	if err != nil {
		return
	}
	duration := time.Since(start)

	body = resp.Body()
	logger.Wrapper.Debugf("[HTTP][Delete] url:%s status:%d query:%s cost: %fs",
		url, resp.StatusCode(), req.RawRequest.URL.RawQuery, duration.Seconds())

	if resp.StatusCode() != http.StatusOK {
		err = newHttpError(url, int32(resp.StatusCode()))
		return
	}

	return
}

type HttpError struct {
	url    string
	status int32
}

// Error 返回error信息
func (he HttpError) Error() string {
	return fmt.Sprintf("[http][get] url: %s status_code: %d", he.url, he.status)
}

// StatusCode 返回 http code
func (he HttpError) StatusCode() int32 {
	return he.status
}

// newHttpError 创建error
func newHttpError(url string, status int32) HttpError {
	return HttpError{url: url, status: status}
}

// Into TODO
func (r *Result) Into(obj interface{}) error {
	if nil != r.Err {
		return r.Err
	}

	if 0 != len(r.Body) {
		err := json.Unmarshal(r.Body, obj)
		if err != nil {
			if r.StatusCode >= 300 {
				return fmt.Errorf("http request err: %s", string(r.Body))
			}
			logger.Wrapper.Errorf("invalid response body, unmarshal json failed, reply:%s, error:%s", r.Body, err.Error())
			return fmt.Errorf("http response err: %v, raw data: %s", err, r.Body)
		}
	} else if r.StatusCode >= 300 {
		return fmt.Errorf("http request failed: %s", r.Status)
	}
	return nil
}
