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
		writer.WriteHeader(200)
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
	count1 :=0
	http.HandleFunc("/500-504-200", func(writer http.ResponseWriter, request *http.Request) {
		count1++
		switch count1 {
		case 1:
			writer.WriteHeader(500)
		case 2:
			writer.WriteHeader(504)
		case 3:
			count1=0
			_,_ = writer.Write([]byte("ok"))// 测试阶段全忽略
			writer.WriteHeader(200)
		}

	})
	go func() {
		log.Print(http.ListenAndServe(":2222", nil))
	}()
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		// dns 出问题的时候会 no such host
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "https://nosuchhost102923092190311.com", "", SendRequest{
			Retry: RequestRetry{
				Times: 3,
				Interval:  time.Millisecond*100,
			},
		})
		assert.Equal(t,err.Error(), `Get "https://nosuchhost102923092190311.com": dial tcp: lookup nosuchhost102923092190311.com: no such host`)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 0)
		_=resp
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/429", SendRequest{
			Retry: RequestRetry{
				Times: 3,
				Interval:  time.Millisecond*100,
			},
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 429)
		_=resp
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500", SendRequest{
			Retry: RequestRetry{
				Times: 3,
				Interval:  time.Millisecond*100,
			},
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 500)
		_=resp
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/504", SendRequest{
			Retry: RequestRetry{
				Times: 3,
				Interval:  time.Millisecond*100,
			},
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 504)
		_=resp
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500-504-200", SendRequest{
		
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 500)
		_=resp
	}
	count1=0
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500-504-200", SendRequest{
			Retry: RequestRetry{
				Times: 1,
				Interval:  time.Millisecond*100,
			},
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 504)
		_=resp
	}
	count1=0
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/500-504-200", SendRequest{
			Retry: RequestRetry{
				Times: 2,
				Interval:  time.Millisecond*100,
			},
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 200)
		_=resp
	}
	count1=0

	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "http://localhost:2222", "/200", SendRequest{
			Retry: RequestRetry{
				Times: 2,
				Check: func(resp *http.Response, requestErr error) (shouldRetry bool) {
					return true
				},
			},
			// Debug: true,
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 200)
		_=resp
	}
	{
		ctx := context.Background()
		client := NewClient(&http.Client{})
		resp, bodyClose, statusCode, err := client.Send(ctx, GET, "https://apierror.weixin.qq.com", "/cgi-bin/token", SendRequest{
			Retry: RequestRetry{
				Times: 2,
				BackupOrigin: "https://api2.weixin.qq.com",
			},
			// Debug: true,
		}) ; assert.NoError(t, err)
		defer assert.NoError(t, bodyClose())
		assert.Equal(t, statusCode, 200)
		_=resp
	}
	{
		resp, bodyClose, _, err := NewClient(nil).Send(context.TODO(), GET, "https://bing.com", "/", SendRequest{
			Debug: true,
		}) ; assert.NoError(t, err)
		defer bodyClose()
		d, err := ioutil.ReadAll(resp.Body) ; assert.NoError(t, err)
		assert.NotEqual(t, len(d), 0)
	}
}
