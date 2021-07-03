package xhttp

import (
	"bytes"
	"context"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type RequestQuery interface {
	Query() (url.Values, error)
}
type RequestFormUrlencoded interface {
	FormUrlencoded() (url.Values, error)
}
type RequestFormData interface {
	FormData(bufferData *bytes.Buffer) (*multipart.Writer, error)
}
type RequestHeader interface {
	Header() (http.Header, error)
}
type SendRequest struct {
	Query RequestQuery
	FormUrlencoded RequestFormUrlencoded
	FormData RequestFormData
	Header RequestHeader
	JSON io.Reader
	Body io.Reader
	Debug bool
	Retry RequestRetry
}
type RequestRetry struct {
	Times uint8
	Interval time.Duration `eg:"time.Millisecond*100"`
	BackupOrigin string `note:"灾备接口域名必须以 http:// 或 https:// 开头"`
	Check func(resp *http.Response, requestErr error) (shouldRetry bool) `note:"if Check == nil { Check = xhttp.DefaultRequestRetryCheck }"`
}
func DefaultRequestRetryCheck (resp *http.Response, err error) (shouldRetry bool) {
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
func (request SendRequest) HttpRequest(ctx context.Context, method Method, url string) (*http.Request, error) {
	var bodyReader io.Reader
	if request.Body != nil {
		bodyReader = request.Body
	}
	// json
	if request.JSON != nil {
		bodyReader = request.JSON
	}
	// x-www-form-urlencoded
	if request.FormUrlencoded != nil {
		values, err := request.FormUrlencoded.FormUrlencoded() ; if err != nil {return nil, err}
		bodyReader = strings.NewReader(values.Encode())
	}
	// form data
	var formWriter *multipart.Writer
	if formData := request.FormData; formData != nil {
		bufferData := bytes.NewBuffer(nil)
		formWriter, err := request.FormData.FormData(bufferData) ; if err != nil {return nil, err}
		err = formWriter.Close() ; if err != nil {return nil, err}
		bodyReader = bufferData
	}
	httpRequest, err := http.NewRequestWithContext(ctx, method.String(), url, bodyReader) ; if err != nil {return nil, err}
	// header
	{
		if request.Header != nil {
			header, err := request.Header.Header() ; if err != nil {return nil, err}
			httpRequest.Header = header
		}
		if request.FormUrlencoded != nil {
			httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
		if request.FormData != nil {
			httpRequest.Header.Add("Content-Type", formWriter.FormDataContentType())
		}
		if request.JSON != nil {
			httpRequest.Header.Set("Content-Type", "application/json")
		}
	}
	// query
	if request.Query != nil {
		values, err := request.Query.Query() ; if err != nil {return nil, err}
		httpRequest.URL.RawQuery = values.Encode()
	}
	if request.Debug {
		data, dumpErr := httputil.DumpRequest(httpRequest, true) ; if dumpErr != nil {
			log.Print(dumpErr)
		}
		log.Print(string(data))
	}
	return httpRequest, nil
}