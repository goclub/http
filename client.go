package xhttp

import (
	"context"
	"net/http"
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
	bodyClose = func() error { return nil}
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
// 发送 query from json 请求等常见下使用 http.Request{} 需要设置 header 等繁琐事项
// 使用 xhttp.Send() 和 xhttp.Request{} 可以高效的创建请求
func (client *Client) Send(ctx context.Context, method Method, url string, request SendRequest) (resp *http.Response, bodyClose func() error, statusCode int, err error)  {
	// 防止空指针错误
	bodyClose = func() error { return nil }
	var httpRequest *http.Request
	httpRequest, err = request.HttpRequest(ctx, method, url) ; if err != nil {return}
	return client.Do(httpRequest)
}