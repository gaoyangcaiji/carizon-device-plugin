package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"carizon-device-plugin/pkg/logger"

	"github.com/go-resty/resty/v2"
)

const (
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

// Option is the options for client
type Option func(*options)

type options struct {
	retry          int32                             // 重试次数
	retryCheckFunc func(*resty.Response, error) bool // 重试check函数
	timeout        time.Duration                     // 超时时间
	proxy          string                            // 代理
	queryValues    url.Values                        // Get方法参数数组
}

// WithRetry 设置重试次数
func WithRetry(retry int32) Option {
	return func(o *options) {
		o.retry = retry
	}
}

// WithRetryCheckFunc 设置重试check函数
func WithRetryCheckFunc(f func(response *resty.Response, err error) bool) Option {
	return func(o *options) {
		o.retryCheckFunc = f
	}
}

// WithTimeout 设置超时时间(单位:毫秒)
func WithTimeout(t time.Duration) Option {
	return func(o *options) {
		o.timeout = t
	}
}

// WithProxy 设置代理
func WithProxy(p string) Option {
	return func(o *options) {
		o.proxy = p
	}
}

// WithSetQueryValues 设置Get方法数组参数
func WithSetQueryValues(v url.Values) Option {
	return func(o *options) {
		o.queryValues = v
	}
}

type client struct {
	url    string
	header map[string]string
	cli    *resty.Client
	opt    *options
}

// NewIClient 创建HTTP client对象
func NewIClient(opts ...Option) IClient {
	opt := &options{}
	for _, option := range opts {
		option(opt)
	}

	return &client{
		cli: resty.New(),
		opt: opt,
	}
}

type IClient interface {
	DoGet(ctx context.Context, url string, header, query map[string]string, timeout int) (body []byte, err error)
	DoPost(ctx context.Context, url string, header map[string]string, body interface{}, timeout int) (data []byte, err error)
	DoDelete(ctx context.Context, url string, header map[string]string, timeout int) (body []byte, err error)
	GetCli() *resty.Client
	GetOpt() *options
}

func (c *client) GetCli() *resty.Client {
	return c.cli
}

func (c *client) GetOpt() *options {
	return c.opt
}

// DoGet get方法 timeout单位：s
func (c *client) DoGet(ctx context.Context, url string, header, query map[string]string, timeout int) (body []byte, err error) {
	cf := func(response *resty.Response, err error) bool {
		logger.Wrapper.Errorf("[HTTP][Get] retry failed, resp: %+v, err: %+v", response, err)
		return err != nil || response.StatusCode() != http.StatusOK
	}
	if c.opt.retryCheckFunc != nil {
		cf = c.opt.retryCheckFunc
	}

	start := time.Now()
	// https://github.com/go-resty/resty#retries
	req := c.cli.AddRetryCondition(cf).SetRetryCount(int(c.opt.retry)).SetTimeout(time.Second * time.Duration(timeout)).R().SetContext(ctx).EnableTrace().SetHeaders(header).SetQueryParams(query)
	if c.opt.queryValues != nil {
		req.SetQueryParamsFromValues(c.opt.queryValues)
	}
	resp, err := req.Get(url)
	if err != nil {
		return
	}
	duration := time.Since(start)

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
func (c *client) DoPost(ctx context.Context, url string, header map[string]string, body interface{}, timeout int) (data []byte, err error) {

	cf := func(response *resty.Response, err error) bool {
		logger.Wrapper.Errorf("[HTTP][Post] retry failed, resp: %+v, err: %+v", response, err)
		return err != nil || response.StatusCode() != http.StatusOK
	}
	if c.opt.retryCheckFunc != nil {
		cf = c.opt.retryCheckFunc
	}

	// 默认json格式
	if header[HeaderContentType] == "" {
		header[HeaderContentType] = ContentTypeJson
	}

	req := c.cli.AddRetryCondition(cf).
		SetRetryCount(int(c.opt.retry)).
		SetTimeout(time.Second * time.Duration(timeout)).
		R().
		SetContext(ctx).
		EnableTrace().
		SetHeaders(header)

	switch header[HeaderContentType] {
	case ContentTypeFormEncoded:
		q := body.(map[string]string)
		req = req.SetFormData(q)
	case ContentTypeText, ContentTypeJson:
		req = req.SetBody(body)
	default:
		err = errors.New("http/post not support content-type")
		return
	}

	start := time.Now()
	resp, err := req.Post(url)
	if err != nil {
		return
	}
	duration := time.Since(start)

	data = resp.Body()

	fmt.Println(duration.Seconds())
	logger.Wrapper.Debugf("[HTTP][Post] url:%s status:%d body:%+v cost: %fs",
		url, resp.StatusCode(), body, duration.Seconds())

	if resp.StatusCode() != http.StatusOK {
		err = newHttpError(url, int32(resp.StatusCode()))
		return
	}

	return
}

// DoDelete delete方法 timeout单位：s
func (c *client) DoDelete(ctx context.Context, url string, header map[string]string, timeout int) (body []byte, err error) {
	cf := func(response *resty.Response, err error) bool {
		logger.Wrapper.Errorf("[HTTP][Delete] retry failed, resp: %+v, err: %+v", response, err)
		return err != nil || response.StatusCode() != http.StatusOK
	}
	if c.opt.retryCheckFunc != nil {
		cf = c.opt.retryCheckFunc
	}

	start := time.Now()
	// https://github.com/go-resty/resty#retries
	req := c.cli.AddRetryCondition(cf).SetRetryCount(int(c.opt.retry)).SetTimeout(time.Second * time.Duration(timeout)).R().SetContext(ctx).EnableTrace().SetHeaders(header)
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

type APIResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
