package xhttp

import (
	"bytes"
	"context"
	xjson "github.com/goclub/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type SendRequest struct {
	Query interface {
		Query() (string, error)
	}
	FormUrlencoded interface {
		FormUrlencoded() (string, error)
	}
	FormData interface {
		FormData(w *multipart.Writer) (err error)
	}
	Header interface {
		Header() (http.Header, error)
	}
	JSON       interface{}
	Body       io.Reader
	Debug      bool
	Retry      RequestRetry
	BeforeSend func(r *http.Request) (err error)
	// DoNotReturnRequestBody 控制返回的 httpResult.Request{}.Body 是否为空
	// 在请求 Body 大时可以设置为 true 以提高性能
	DoNotReturnRequestBody bool
}
type RequestRetry struct {
	Times        uint8
	Interval     time.Duration                                                  `eg:"time.Millisecond*100"`
	BackupOrigin string                                                         `note:"灾备接口域名必须以 http:// 或 https:// 开头"`
	Check        func(resp *http.Response, requestErr error) (shouldRetry bool) `note:"if Check == nil { Check = xhttp.DefaultRequestRetryCheck }"`
}

func DefaultRequestRetryCheck(resp *http.Response, err error) (shouldRetry bool) {
	if err != nil {
		return true
	}
	// 列出需要重试的条件
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return true
	case 0:
		return true
	default:
		if resp.StatusCode >= 500 {
			return true
		} else {
			return false
		}
	}
}
func (request SendRequest) HttpRequest(ctx context.Context, method Method, requestURL string) (httpRequest *http.Request, err error) {
	var bodyReader io.Reader
	if request.Body != nil {
		bodyReader = request.Body
	}
	// json
	if request.JSON != nil {
		var jsonb []byte
		jsonb, err = xjson.Marshal(request.JSON)
		if err != nil {
			return
		}
		bodyReader = bytes.NewBuffer(jsonb)
	}
	// x-www-form-urlencoded
	if request.FormUrlencoded != nil {
		var formUrlencoded string
		formUrlencoded, err = request.FormUrlencoded.FormUrlencoded()
		if err != nil {
			return
		}
		bodyReader = strings.NewReader(formUrlencoded)
	}
	// form data
	var formWriter *multipart.Writer
	if formData := request.FormData; formData != nil {
		bufferData := bytes.NewBuffer(nil)
		formWriter = multipart.NewWriter(bufferData)
		err = request.FormData.FormData(formWriter)
		if err != nil {
			return
		}
		err = formWriter.Close()
		if err != nil {
			return nil, err
		}
		bodyReader = bufferData
	}
	httpRequest, err = http.NewRequestWithContext(ctx, method.String(), requestURL, bodyReader)
	if err != nil {
		return
	}
	// header
	{
		if request.Header != nil {
			var header http.Header
			header, err = request.Header.Header()
			if err != nil {
				return
			}
			httpRequest.Header = header
		}
		if request.FormUrlencoded != nil {
			httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if request.FormData != nil {
			httpRequest.Header.Set("Content-Type", formWriter.FormDataContentType())
		}
		if request.JSON != nil {
			httpRequest.Header.Set("Accept", "application/json")
			httpRequest.Header.Set("Content-Type", "application/json")
		}
	}
	// query
	if request.Query != nil {
		var queryValue string
		queryValue, err = request.Query.Query()
		if err != nil {
			return
		}
		httpRequest.URL.RawQuery = queryValue
	}
	if request.Debug {
		data, dumpErr := httputil.DumpRequest(httpRequest, true)
		if dumpErr != nil {
			log.Print(dumpErr)
		}
		log.Print("Request:", string(data))
	}
	return httpRequest, nil
}
