package xhttp

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestClient_SendRetry(t *testing.T) {
	http.HandleFunc("/200", func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		writer.WriteHeader(200)
		_, err = writer.Write(body)
		if err != nil {
			panic(err)
		}
	})
	http.HandleFunc("/429", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTooManyRequests)
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
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "https://nosuchhost102923092190311.com", "", SendRequest{
			Retry: RequestRetry{
				Times:    3,
				Interval: time.Millisecond * 100,
			},
		})
		assert.Equal(t, err.Error(), `Get "https://nosuchhost102923092190311.com": dial tcp: lookup nosuchhost102923092190311.com: no such host`)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 0)

	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/429", SendRequest{
			Retry: RequestRetry{
				Times:    3,
				Interval: time.Millisecond * 100,
			},
		})
		assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 429)

	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500", SendRequest{
			Retry: RequestRetry{
				Times:    3,
				Interval: time.Millisecond * 100,
			},
		})
		assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 500)

	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/504", SendRequest{
			Retry: RequestRetry{
				Times:    3,
				Interval: time.Millisecond * 100,
			},
		})
		assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 504)

	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500-504-200", SendRequest{})
		assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 500)

	}
	count1 = 0
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500-504-200", SendRequest{
			Retry: RequestRetry{
				Times:    1,
				Interval: time.Millisecond * 100,
			},
		})
		assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 504)

	}
	count1 = 0
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500-504-200", SendRequest{
			Retry: RequestRetry{
				Times:    2,
				Interval: time.Millisecond * 100,
			},
		})
		assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 200)

	}
	count1 = 0

	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		httpResult, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/200", SendRequest{
			JSON: map[string]interface{}{
				"name": "goclub",
			},
			Retry: RequestRetry{
				Times: 2,
				Check: func(resp *http.Response, requestErr error) (shouldRetry bool) {
					return true
				},
			},
			// Debug: true,
		})
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, bodyClose())
		}()
		assert.Equal(t, statusCode, 200)
		{
			// resp
			body, err := ioutil.ReadAll(httpResult.Response.Body)
			assert.NoError(t, err)
			assert.Equal(t, string(body), `{"name":"goclub"}`)
		}
		{
			// req
			body, err := ioutil.ReadAll(httpResult.Request.Body)
			assert.NoError(t, err)
			assert.Equal(t, string(body), `{"name":"goclub"}`)
		}

	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		_, bodyClose, statusCode, err := client.Send(ctx, GET, "https://apierror.weixin.qq.com", "/cgi-bin/token", SendRequest{
			Retry: RequestRetry{
				Times:        2,
				BackupOrigin: "https://api2.weixin.qq.com",
			},
			// Debug: true,
		})
		assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 200)

	}
	{
		httpResult, bodyClose, _, err := NewClient(nil).Send(context.TODO(), GET, "http://httpbin.org", "/headers", SendRequest{})
		assert.NoError(t, err)
		defer bodyClose()
		d, err := ioutil.ReadAll(httpResult.Response.Body)
		assert.NoError(t, err)
		assert.NotEqual(t, len(d), 0)
	}
}
