package xhttp_test

import (
	xhttp "github.com/goclub/http"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func init () {
	go func() {
		http.HandleFunc("/query", func(writer http.ResponseWriter, request *http.Request) {
			query := request.URL.Query()
			_, err := writer.Write([]byte("hello " + query.Get("name"))) ; if err != nil {panic(err)}
		})
		addr :=  ":1212"
		log.Print(http.ListenAndServe(addr, nil))
	}()
}
func TestExample(t *testing.T) {
	ExampleClient_DO()
}
// net/http 的 &http.Client{}.Do() 函数只返回了  resp, err
// 实际上我们一定要记得 resp.Body.Close() ,但是 resp 可能是个 nil ，此时 运行 resp.Body.Close() 会出现空指针错误
// 并且一般情况应该判断 resp.StatusCode != 200 并返回错误
// 所以 &xhttp.Client{}.Do() 的函数签名是 (client *Client) Do(request *http.Request) (resp *http.Response, bodyClose func() error, statusCode int, err error)
// 这样使用者就不容易忘记处理 statusCode 和 bodyClose ，并且 bodyClose 处理了resp 空 nil的情况
func ExampleClient_DO() {
	client := xhttp.NewClient(&http.Client{})
	request, err := http.NewRequest("GET", "https://unpkg.com/goclub@0.0.1/package.json", nil) ; if err != nil {
		panic(err)
	}
	resp, bodyClose, statusCode, err := client.Do(request) ; if err != nil {
		panic(err)
	}
	defer bodyClose()
	if statusCode != 200 {
		panic(errors.New("response " + resp.Status))
	}
	data, err := ioutil.ReadAll(resp.Body) ; if err != nil {
		panic(err)
	}
	log.Print("response ", string(data))
}