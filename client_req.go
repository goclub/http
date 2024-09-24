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
	"time"
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
	Defer              func(result HttpResult, err error)
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
	var requestBody []byte
	if rBody != nil {
		if requestBody, err = ioutil.ReadAll(rBody); err != nil {
			return
		}
	}
	var httpRequest *http.Request
	if httpRequest, err = http.NewRequestWithContext(ctx, method.String(), rUrl.String(), bytes.NewBuffer(requestBody)); err != nil {
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
			log.Print(result.Dump())
		}
	}()
	// send request
	var resp *http.Response
	startTime := time.Now()
	if req.Defer != nil {
		defer func() {
			req.Defer(result, err)
		}()
	}
	resp, err = c.Core.Do(httpRequest)
	result.Response = resp
	result.elapsed = time.Now().Sub(startTime)
	if err == nil {
		var b []byte
		if b, err = result.GetBody(); err != nil {
			return
		}
		resp.Body.Close()
		result.Response.Body = io.NopCloser(bytes.NewBuffer(b))
	}
	if err != nil {
		return
	}
	// 让 Request.Body 可读
	if requestBody != nil {
		result.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	// check status
	if req.NotCheckStatusCode == false {
		if resp.StatusCode != 200 {
			err = xerr.New("goclub/http: statusCode != 200, is " + strconv.Itoa(resp.StatusCode))
			return
		}
	}
	return
}
