package xhttp

import (
	"bytes"
	"context"
	xerr "github.com/goclub/error"
	xjson "github.com/goclub/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Req struct {
	QueryEncode        func(q url.Values) (encode string)
	FormUrlencoded     func(f url.Values) (encode string)
	FormData           func(w *multipart.Writer) (err error)
	Header             func(h http.Header) http.Header
	JSON               interface{}
	Body               io.Reader
	Debug              bool
	Before             func(r *http.Request) (err error)
	NotCheckStatusCode bool
}

func (c *Client) Req(ctx context.Context, method Method, u string, req Req) (result HttpResult, err error) {
	// make request
	var rBody io.Reader
	var rUrl *url.URL
	if rUrl, err = url.Parse(u); err != nil {
		return
	}
	rHeader := http.Header{}
	if req.Header != nil {
		rHeader = req.Header(http.Header{})
	}
	if req.QueryEncode != nil {
		resultQuery := req.QueryEncode(rUrl.Query())
		rUrl.RawQuery = resultQuery
	}
	if req.FormUrlencoded != nil {
		f := url.Values{}
		resultForm := req.FormUrlencoded(f)
		rHeader.Set("Content-Type", "application/x-www-form-urlencoded")
		rBody = strings.NewReader(resultForm)
	}
	if req.FormData != nil {
		bufferData := bytes.NewBuffer(nil)
		formWriter := multipart.NewWriter(bufferData)
		if err = req.FormData(formWriter); err != nil {
			return
		}
		if err = formWriter.Close(); err != nil {
			return
		}
		rBody = bufferData
		rHeader.Set("Content-Type", formWriter.FormDataContentType())
	}
	if req.JSON != nil {
		var jsonb []byte
		if jsonb, err = xjson.Marshal(req.JSON); err != nil {
			return
		}
		rBody = bytes.NewBuffer(jsonb)
		rHeader.Set("Accept", "application/json")
		rHeader.Set("Content-Type", "application/json")
	}
	if req.Body != nil {
		rBody = req.Body
	}
	var httpRequest *http.Request
	if httpRequest, err = http.NewRequestWithContext(ctx, method.String(), rUrl.String(), rBody); err != nil {
		return
	}
	httpRequest.Header = rHeader
	if req.Before != nil {
		if err = req.Before(httpRequest); err != nil {
			return
		}
	}
	result.Request = httpRequest
	// debug
	defer func() {
		if req.Debug {
			log.Print(result.DumpRequestResponseString(true))
		}
	}()
	// send request
	var resp *http.Response
	if resp, err = c.Core.Do(httpRequest); err != nil {
		return
	}
	result.Response = resp
	// check status
	if req.NotCheckStatusCode == false {
		if resp.StatusCode != 200 {
			err = xerr.New("goclub/http: statusCode != 200, is " + strconv.Itoa(resp.StatusCode))
			return
		}
	}
	// set nop close body
	var body []byte
	if body, err = ioutil.ReadAll(result.Response.Body); err != nil {
		return
	}
	result.SetNopCloserBody(body)
	return
}
