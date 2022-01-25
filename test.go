package xhttp

import (
	"bytes"
	xjson "github.com/goclub/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

type Test struct {
	router *Router
	t *testing.T
	jar *cookiejar.Jar
}
func NewTest(t *testing.T, router *Router) Test {
	jar, err := cookiejar.New(nil) ; testHandleError(err)
	return Test{
		router: router,
		t: t,
		jar: jar,
	}
}
func (test Test) RequestJSON(route Route, jsonValue interface{}) (resp *Response)  {
	request := testNewRequestJSON(route, jsonValue)
	return test.Request(request)
}
func testNewRequestJSON(route Route, jsonValue interface{}) *http.Request {
	data, err := xjson.Marshal(jsonValue) ; testHandleError(err)
	request := httptest.NewRequest(route.Method.String(), route.Path, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	return request
}
func (test *Test) Request(r *http.Request) (resp *Response)  {
	r.URL.Scheme = "http"
	/* request set cookie */{
		cookies := test.jar.Cookies(r.URL)
		for _, cookie := range cookies {
			r.AddCookie(cookie)
		}
	}
	recorder := httptest.NewRecorder()
	test.router.ServeHTTP(recorder, r)
	httpResponse :=  recorder.Result()
	/* response set cookie */ {
		test.jar.SetCookies(r.URL, httpResponse.Cookies())
	}
	return &Response{
		t: test.t,
		recorder: recorder,
		HttpResponse: httpResponse,
	}
}


type Response struct {
	t *testing.T
	recorder *httptest.ResponseRecorder
	HttpResponse *http.Response
}
func testHandleError(err error) {
	// test 场景可以 panic,不用担心子 routine 意外退出
	if err != nil {
		panic(err)
	}
}
func (resp *Response) Bytes(statusCode int) []byte {
	assert.Equal(resp.t, statusCode, resp.HttpResponse.StatusCode)
	b, err := ioutil.ReadAll(resp.recorder.Body) ; testHandleError(err)
	resp.recorder.Body = bytes.NewBuffer(b)
	return b
}
func (resp *Response) String(statusCode int) string {
	return string(resp.Bytes(statusCode))
}
func (resp *Response) ExpectString(statusCode int, s string) {
	assert.Equal(resp.t,s, string(resp.Bytes(statusCode)))
}
func (resp *Response) BindJSON(statusCode int, v interface{})  {
	err := xjson.Unmarshal(resp.Bytes(statusCode), v) ; testHandleError(err)
}
func (resp *Response) ExpectJSON(statusCode int, reply interface{}) {
	data, err := xjson.Marshal(reply) ; testHandleError(err)
	assert.Equal(resp.t, string(data), resp.String(statusCode))
}