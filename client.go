package xhttp

import (
	"context"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)
type Client struct {
	Core *http.Client
}
func NewClient (core *http.Client) *Client {
	if core == nil {
		core = &http.Client{}
	}
	return &Client{
		Core: core,
	}
}
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
	resp, err = client.Core.Do(request) ; if err != nil {
		return
	}
	if resp != nil {
		bodyClose= resp.Body.Close
	}
	statusCode = resp.StatusCode
	return
}

func (client *Client) CloseIdleConnections() {
	client.Core.CloseIdleConnections()
}

func (client *Client) coreSend(ctx context.Context, method Method, url string, request SendRequest) (resp *http.Response, bodyClose func() error, statusCode int, err error)  {
	// 防止空指针错误
	bodyClose = func() error { return nil }
	var httpRequest *http.Request
	httpRequest, err = request.HttpRequest(ctx, method, url) ; if err != nil {return}
	return client.Do(httpRequest)
}


// 发送 query from json 请求等常见下使用 http.Request{} 需要设置 header 等繁琐事项
// 使用 xhttp.Send() 和 xhttp.Request{} 可以高效的创建请求
func (client *Client) Send(ctx context.Context, method Method, url string, request SendRequest) (resp *http.Response, bodyClose func() error, statusCode int, err error)  {
	if request.Retry.Check == nil {
		request.Retry.Check = DefaultRequestRetryCheck
	}
	// 防止空指针错误
	bodyClose = func() error { return nil }
	requestTimes := request.Retry.Times+1
	// safe count 用于避免 request.Retry.Check 写错导致的死循环，这种死循环可能在接收请求的服务器出现错误时候才能发现。
	for safeCount := 0; safeCount<math.MaxUint8;safeCount++ {
		select {
			case <-ctx.Done():
				err = ctx.Err()
				return
		default:
			resp, bodyClose, statusCode, err = client.coreSend(ctx, method, url, request)
			requestTimes--
			shouldRetry := request.Retry.Check(resp, err)
			// 强制 200 不重试
			if statusCode == 200 {
				shouldRetry = false
			}
			if shouldRetry {
				if requestTimes <= 0 {
					return
				} else {
					if request.Debug {
						log.Print("goclub/http Client{}.Send() " + method.String() + " " + url + " response status code("+strconv.Itoa(statusCode)+") retry(" + strconv.FormatUint(uint64(requestTimes), 10) + ")  try again in " + request.Retry.Interval.String())
					}
					time.Sleep(request.Retry.Interval)
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
