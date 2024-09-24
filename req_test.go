package xhttp

import (
	"context"
	xjson "github.com/goclub/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestClient_Req(t *testing.T) {
	http.HandleFunc("/200", func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		writer.WriteHeader(200)
		_, err = writer.Write([]byte("resp:"))
		_, err = writer.Write(body)
		if err != nil {
			panic(err)
		}
	})
	http.HandleFunc("/429", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTooManyRequests)
		writer.Write([]byte(`s429`))
	})
	http.HandleFunc("/500", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(500)
	})
	http.HandleFunc("/504", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(504)
	})
	count1 := 0
	http.HandleFunc("/500-504-200", func(writer http.ResponseWriter, request *http.Request) {
		count1++
		switch count1 {
		case 1:
			writer.WriteHeader(500)
		case 2:
			writer.WriteHeader(504)
		case 3:
			count1 = 0
			_, _ = writer.Write([]byte("ok")) // 测试阶段全忽略
		}

	})
	go func() {
		log.Print(http.ListenAndServe(":2222", nil))
	}()
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		// dns 出问题的时候会 no such host
		result, err := client.Req(ctx, GET, "https://nosuchhost102923092190311.com", Req{})
		assert.Equal(t, err.Error(), `Get "https://nosuchhost102923092190311.com": dial tcp: lookup nosuchhost102923092190311.com: no such host`)
		assert.True(t, strings.HasPrefix(result.Dump(), "Request:\r\nGET / HTTP/1.1\r\nHost: nosuchhost102923092190311.com\r\n\r\n"))
		assert.True(t, strings.HasSuffix(result.Dump(), "ms"))
		assert.True(t, strings.Contains(result.Dump(), "Elapsed:"))

	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		result, err := client.Req(ctx, GET, "http://localhost:2222/429", Req{})
		assert.Equal(t, err.Error(), `goclub/http: statusCode != 200, is 429`)
		var b []byte
		if b, err = result.GetBody(); err != nil {
			assert.NoError(t, err)
		}
		// 即使 err 了result.body 还是要有内容
		assert.Equal(t, string(b), "s429")
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		result, err := client.Req(ctx, GET, "http://localhost:2222/429", Req{
			NotCheckStatusCode: true,
		})
		assert.NoError(t, err)
		var b []byte
		if b, err = result.GetBody(); err != nil {
			assert.NoError(t, err)
		}
		// 即使 err 了result.body 还是要有内容
		assert.Equal(t, string(b), "s429")
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, err := client.Req(ctx, GET, "http://localhost:2222/500", Req{})
		assert.Equal(t, err.Error(), `goclub/http: statusCode != 200, is 500`)
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, err := client.Req(ctx, GET, "http://localhost:2222/504", Req{})
		assert.Equal(t, err.Error(), `goclub/http: statusCode != 200, is 504`)
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		httpResult, err := client.Req(ctx, GET, "http://localhost:2222/200", Req{
			JSON: map[string]interface{}{
				"name": "goclub",
			},
		})
		assert.NoError(t, err)
		{
			// resp
			body, err := ioutil.ReadAll(httpResult.Response.Body)
			assert.NoError(t, err)
			assert.Equal(t, string(body), `resp:{"name":"goclub"}`)
		}
		{
			// req
			body, err := ioutil.ReadAll(httpResult.Request.Body)
			assert.NoError(t, err)
			assert.Equal(t, string(body), `{"name":"goclub"}`)
		}

	}
	{
		err := func() (err error) {
			client := NewClient(nil)
			url := "http://httpbin.org/headers"
			var result HttpResult
			if result, err = client.Req(context.TODO(), GET, url, Req{
				Defer: func(r HttpResult, err error) {
					log.Print(r, err)
				},
				// ...
			}); err != nil {
				return
			}
			type Reply struct {
				Headers struct {
					AcceptEncoding string `json:"Accept-Encoding"`
					Host           string `json:"Host"`
					UserAgent      string `json:"User-Agent"`
					XAmznTraceID   string `json:"X-Amzn-Trace-Id"`
				} `json:"headers"`
			}
			reply := Reply{}
			if err = result.ReadBody(xjson.Unmarshal, &reply); err != nil {
				return
			}
			assert.Equal(t, "gzip", reply.Headers.AcceptEncoding)
			assert.Equal(t, "httpbin.org", reply.Headers.Host)
			assert.Equal(t, "Go-http-client/1.1", reply.Headers.UserAgent)
			return
		}()
		assert.NoError(t, err)
	}
}

func TestReqNumGoroutine(t *testing.T) {
	ctx := context.TODO()
	client := NewClient(nil)
	log.Print("t1 ", runtime.NumGoroutine())
	var err error
	for i := 0; i < 10; i++ {
		var r HttpResult
		if r, err = client.Req(ctx, GET, "https://baidu.com/"+strconv.Itoa(i), Req{
			NotCheckStatusCode: true,
			JSON:               map[string]interface{}{"mockbody": 1},
		}); err != nil {
			panic(err)
		}
		_ = r
		// log.Print(r.GetBodyString())
		// r.Response.Body.Close()
		log.Print("t2 ", i, runtime.NumGoroutine())
	}
	time.Sleep(time.Second)
	log.Print("t3 ", runtime.NumGoroutine())
	time.Sleep(time.Second * 3)
	log.Print("t4 ", runtime.NumGoroutine())
	assert.Equal(t, runtime.NumGoroutine(), 6)
}
