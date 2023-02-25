package xhttp_test

import (
	"context"
	xerr "github.com/goclub/error"
	xhttp "github.com/goclub/http"
	xjson "github.com/goclub/json"
	"log"
	"net/http"
	"testing"
)

func TestExample(t *testing.T) {
	// ExampleClient_Do()
	// ExampleClient_Send()
}

// net/http 的 &http.Client{}.Do() 函数只返回了  resp, err
// 实际上我们一定要记得 resp.Body.Close() ,但是 resp 可能是个 nil ，此时 运行 resp.Body.Close() 会出现空指针错误
// 并且一般情况应该判断 resp.StatusCode != 200 并返回错误
// 所以 &xhttp.Client{}.Do() 的函数签名是 (client *Client) Do(request *http.Request) (resp *http.Response, bodyClose func() error, statusCode int, err error)
// 这样使用者就不容易忘记处理 statusCode 和 bodyClose ，并且 bodyClose 处理了resp 空 nil的情况
func ExampleClient_Do() {
	log.Print("ExampleClient_Do")
	ctx := context.TODO()
	client := xhttp.NewClient(&http.Client{})
	url := "https://httpbin.org/json"
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	resp, bodyClose, statusCode, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer bodyClose()
	if statusCode != 200 {
		panic(xerr.New("response " + resp.Status))
	}
	var reply xhttp.ExampleReplyPost
	err = xjson.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		panic(err)
	}
	log.Printf("response %+v", reply)
	// [{ID:6 Title:OVMVHWVHfA Views:20 Published:false CreatedAt:1936-09-10T20:19:28Z} {ID:57 Title:jOyYTxTfVV Views:20 Published:false CreatedAt:1942-04-18T13:44:54Z} {ID:82 Title:CdnSvYuNzs Views:20 Published:true CreatedAt:1907-05-01T23:53:45Z} {ID:97 Title:CWjFddEmda Views:20 Published:true CreatedAt:1971-01-13T08:22:23Z}]
}

func ExampleClient_Send() {
	{
		log.Print("ExampleClient_Send:query")
		ctx := context.TODO()
		client := xhttp.NewClient(&http.Client{})
		err := func() (err error) {
			httpResult, bodyClose, statusCode, err := client.Send(ctx, xhttp.GET, "https://httpbin.org", "/json", xhttp.SendRequest{
				Query: xhttp.ExampleSendQuery{
					Published: true,
					Limit:     2,
				},
				JSON: map[string]interface{}{"name": "goclub"},
			})
			// 1. 遇到错误向上传递
			if err != nil {
				return
			}
			// 2. bodyClose 防止内存泄露
			defer bodyClose()
			// 3. 检查状态码
			if statusCode != 200 {
				// 状态码错误时候记录日志
				log.Print(httpResult.DumpRequestResponseString(true))
				err = xerr.New("http response statusCode != 200")
				return
			}
			// json解码
			var reply xhttp.ExampleReplyPost
			err = httpResult.ReadResponseBodyAndUnmarshal(xjson.Unmarshal, &reply)
			if err != nil {
				// 解码错误时记录日志
				log.Print(httpResult.DumpRequestResponseString(true))
				return
			}
			// 响应
			log.Print("reply", reply)
			return
		}()
		xerr.PrintStack(err)
	}
}
