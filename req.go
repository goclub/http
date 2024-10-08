package xhttp

import (
	"bytes"
	"context"
	xerr "github.com/goclub/error"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Core *http.Client
}

func NewClient(core *http.Client) *Client {
	if core == nil {
		core = &http.Client{}
	}
	return &Client{
		Core: core,
	}
}

// Deprecated: Use client.Req
// Do
// 消费 http.Response{}.Body 后必须 Close
// xhttp.Client{}.Do() 的出参提供了一个安全的 bodyClose
// bodyClose 会在 resp 为 nil 时 不调用 resp.Body.Close 以防止 空指针错误
// 请求发起成功时候应该判断 响应的 statusCode，而不是忽视 statusCode
// xhttp.Client{}.Do() 的出参提供了一个等同于 resp.StatusCode 的 statusCode 参数
// 用于提醒开发人员每次调用完 Do 之后需要根据不同的状态码进行相应的处理措施
// xhttp.Client{}.Do() 的实现非常简单，有兴趣的可以直接查看源码帮助理解
func (client *Client) Do(request *http.Request) (resp *http.Response, bodyClose func() error, statusCode int, err error) {
	if client.Core == nil {
		client.Core = http.DefaultClient
	}
	// 防止空指针错误
	bodyClose = func() error { return nil }
	resp, err = client.Core.Do(request)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return
	}
	if resp != nil {
		bodyClose = resp.Body.Close
	}
	statusCode = resp.StatusCode
	return
}

func (client *Client) CloseIdleConnections() {
	client.Core.CloseIdleConnections()
}

func (client *Client) coreSend(ctx context.Context, method Method, url string, sendRequest SendRequest) (httpResult HttpResult, bodyClose func() error, statusCode int, err error) {
	// 防止空指针错误
	bodyClose = func() error { return nil }
	httpResult.Request, err = sendRequest.HttpRequest(ctx, method, url)
	if err != nil {
		return
	}
	if sendRequest.BeforeSend != nil {
		err = sendRequest.BeforeSend(httpResult.Request)
		if err != nil {
			return
		}
	}
	var requestBodyBytes []byte
	if sendRequest.DoNotReturnRequestBody == false && httpResult.Request.Body != nil {
		requestBodyBytes, err = ioutil.ReadAll(httpResult.Request.Body)
		if err != nil {
			return
		}
		httpResult.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBodyBytes))
	}
	httpResult.Response, bodyClose, statusCode, err = client.Do(httpResult.Request)
	if err != nil {
		if bodyClose != nil {
			_ = bodyClose()
		}
	}
	if requestBodyBytes != nil {
		httpResult.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBodyBytes))
	}
	if sendRequest.Debug {
		log.Print(httpResult.Dump())
	}
	return
}

// Deprecated: Use client.Req
// Send
// 发送 query from json 等常见请求
func (client *Client) Send(ctx context.Context, method Method, origin string, path string, request SendRequest) (httpResult HttpResult, bodyClose func() error, statusCode int, err error) {
	path = strings.TrimSpace(path)
	if path == "" {

	} else if strings.HasPrefix(path, "/") == false {
		log.Print("goclub/http: Send(ctx, origin, path) your forget path prefix / path:(" + path + ")")
		path = "/" + path
	}

	if request.Retry.Check == nil {
		request.Retry.Check = DefaultRequestRetryCheck
	}
	// 防止空指针错误
	bodyClose = func() error { return nil }
	if request.Retry.Times == math.MaxUint8 {
		request.Retry.Times = math.MaxUint8 - 1
	}
	requestTimes := request.Retry.Times + 1
	// safe count 用于避免 request.Retry.Check 写错导致的死循环，这种死循环可能在接收请求的服务器出现错误时候才能发现。
	url := origin + path
	for safeCount := 0; safeCount < math.MaxUint8; safeCount++ {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			httpResult, bodyClose, statusCode, err = client.coreSend(ctx, method, url, request)
			requestTimes--
			shouldRetry := request.Retry.Check(httpResult.Response, err)
			// 强制 200 不重试
			if statusCode == 200 {
				shouldRetry = false
			}
			if shouldRetry {
				if requestTimes <= 0 {
					return
				} else {
					if request.Debug {
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						}
						msg := "goclub/http Client{}.Send() " +
							method.String() +
							" " + url +
							"\n\tresponse statusCode(" + strconv.Itoa(statusCode) +
							")\n\terror(" + errMsg +
							")\n\tretry(" + strconv.FormatUint(uint64(requestTimes), 10) +
							")\n\ttry again in " + request.Retry.Interval.String()
						log.Print(msg)
					}
					time.Sleep(request.Retry.Interval)
					if request.Retry.BackupOrigin != "" {
						url = request.Retry.BackupOrigin + path
					}
					continue
				}
			}
			if err != nil {
				return
			}
			return
		}
	}
	return
}

func DumpRequestResponseString(req *http.Request, resp *http.Response, body bool) (data string) {
	return string(DumpRequestResponse(req, resp, body))
}
func DumpRequestResponse(req *http.Request, resp *http.Response, body bool) (data []byte) {
	if req == nil && resp == nil {
		data = []byte(`Request is nil and Response is nil`)
		return
	}
	var reqData []byte
	if req != nil {
		var err error
		reqData, err = httputil.DumpRequest(req, body)
		if err != nil {
			return []byte(err.Error())
		}
	}
	var respData []byte
	if resp != nil {
		var err error
		respData, err = httputil.DumpResponse(resp, body)
		if err != nil {
			return []byte(err.Error())
		}
	}
	if req != nil {
		data = append(data, []byte("Request:\r\n")...)
		data = append(data, reqData...)
	}
	if resp != nil {
		data = append(data, []byte("Response:\r\n")...)
		data = append(data, respData...)
	}
	return
}

type HttpResult struct {
	Request  *http.Request
	Response *http.Response
	elapsed  time.Duration
}

func (v HttpResult) Elapsed() time.Duration {
	return v.elapsed
}
func (v HttpResult) Dump(body ...bool) (dump string) {
	return string(v.DumpBytes(body...))
}
func (v HttpResult) String() (dump string) {
	return v.Dump()
}
func (v HttpResult) DumpBytes(body ...bool) (data []byte) {
	b := false
	if len(body) > 0 {
		b = body[0]
	} else {
		b = true
	}
	return append(DumpRequestResponse(v.Request, v.Response, b), []byte("\r\nElapsed:"+v.elapsed.String())...)
}
func (v HttpResult) GetBodyString() (body string, err error) {
	var b []byte
	if b, err = v.GetBody(); err != nil {
		return
	}
	return string(b), err
}
func (v HttpResult) GetBody() (body []byte, err error) {
	if body, err = ioutil.ReadAll(v.Response.Body); err != nil {
		err = xerr.WrapPrefix("", err)
		return
	}
	v.Response.Body = io.NopCloser(bytes.NewBuffer(body))
	return
}
func (v HttpResult) ReadBody(unmarshal func(data []byte, v interface{}) error, ptr interface{}) (err error) {
	var body []byte
	if body, err = v.GetBody(); err != nil {
		return
	}
	err = unmarshal(body, ptr)
	if err != nil {
		return xerr.WithStack(err)
	}
	return
}
