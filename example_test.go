package xhttp_test

import (
	"context"
	xhttp "github.com/goclub/http"
	xjson "github.com/goclub/json"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"testing"
)

func TestExample(t *testing.T) {
	ExampleClient_DO()
	ExampleClient_Send()
}
// net/http 的 &http.Client{}.Do() 函数只返回了  resp, err
// 实际上我们一定要记得 resp.Body.Close() ,但是 resp 可能是个 nil ，此时 运行 resp.Body.Close() 会出现空指针错误
// 并且一般情况应该判断 resp.StatusCode != 200 并返回错误
// 所以 &xhttp.Client{}.Do() 的函数签名是 (client *Client) Do(request *http.Request) (resp *http.Response, bodyClose func() error, statusCode int, err error)
// 这样使用者就不容易忘记处理 statusCode 和 bodyClose ，并且 bodyClose 处理了resp 空 nil的情况
func ExampleClient_DO() {
	log.Print("ExampleClient_DO")
	ctx := context.TODO()
	client := xhttp.NewClient(&http.Client{})
	url := "https://mockend.com/goclub/http/posts?views_eq=20"
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil) ; if err != nil {
		panic(err)
	}
	resp, bodyClose, statusCode, err := client.Do(request) ; if err != nil {
		panic(err)
	}
	defer bodyClose()
	if statusCode != 200 {
		panic(errors.New("response " + resp.Status))
	}
	var reply []xhttp.ExampleReplyPost
	err = xjson.NewDecoder(resp.Body).Decode(&reply) ; if err != nil {
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
		resp, bodyClose, statusCode, err := client.Send(ctx, xhttp.GET, "https://mockend.com/goclub/http/posts", xhttp.SendRequest{
			Query:          xhttp.ExampleSendQuery{
				Published: true,
				Limit:     2,
			},
		}) ; if err != nil {
		panic(err)
	}
		defer bodyClose()
		if statusCode != 200 {
			panic(errors.New("status: " + resp.Status))
		}
		var reply []xhttp.ExampleReplyPost
		err = xjson.NewDecoder(resp.Body).Decode(&reply) ; if err != nil {
		panic(err)
	}
		log.Printf("response %+v", reply)
		// [{ID:2 Title:YEBlKOVrgg Views:22 Published:true CreatedAt:1981-05-22T03:42:31Z} {ID:3 Title:sMzheVnQeT Views:43 Published:true CreatedAt:1952-12-30T13:00:17Z}]
	}
}