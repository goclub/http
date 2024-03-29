---
permalink: /
sidebarBasedOnContent: true
---

# goclub/http

## 介绍

goclub/http 基于 [gorilla/mux](https://github.com/gorilla/mux) 实现了接口友好并且安全便捷的`http router`和`http client`

## 特性

### 错误处理

对错误处理支持良好,在 http 请求中发生 error/panic 时可以捕获错误并进行自定义的处理.

### 路由分组和中间件

支持路由分组和中间件,可满足常见的的请求日志,登录鉴权.

### 请求校验

配套类型安全的验证器: [goclub/validator](https://github.com/goclub/validator)

## 运行 

```go
package main

import (
	xhttp "github.com/goclub/http"
	"net/http"
)

func main() {
	r := xhttp.NewRouter(xhttp.RouterOption{})
	server := &http.Server{
		Addr:    ":2222",
		Handler: r,
	}
	r.HandleFunc(xhttp.Route{xhttp.GET, "/"}, func(c *xhttp.Context) (err error) {
		return c.WriteJSON(map[string]interface{}{"name": "goclub/http"})
	})
	r.LogPatterns(server)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
```


## server

**示例**

1. [基础示例](./example/internal/basic/main.go)
1. [请求响应示例](./example/internal/request_response/main.go)
1. [给前端用的模拟服务器](./example/internal/mock/main.go)

**相关的包**

1. [goclub/session](https://github.com/goclub/session)
2. [goclub/validator](https://github.com/goclub/validator)
3. [goclub/error](https://github.com/goclub/error)

## client

1. [Client.Send](https://pkg.go.dev/github.com/goclub/http#Client.Send)
2. [Client.Do](https://pkg.go.dev/github.com/goclub/http#Client.Do)


## 特性
1. http server 支持 `OnCatchError` `OnCatchPanic` 拦截器，让错误处理更简单，让panic时对客户端更友好。
2. http client  `xhttp.Client{}.Send()` 高易用高性能的发起常见请求（`query` `formurlencoded` `formdata` `json`）


### Client{}.Send()

`xhttp.Client{}.Send()` 绑定常见请求 `query` `formUrlencoed` `form-data` `json`
 
是通过实现一个符合 `Query() (url.Values, error)` 接口的结构体完成设置请求。

```go
type ExampleSendQuery struct {
	Published bool
	Limit int
}
// 通过实现结构体  Query() (string, error) 方法后传入 xhttp.SendRequest{}.Query
// 即可设置请求 query 参数
func (r ExampleSendQuery) Query() (string, error) {
	v := url.Values{}
	v.Set("published_eq", strconv.FormatBool(r.Published))
	v.Set("limit", strconv.Itoa(r.Limit))
	return v.Encode(), nil
}
client.Send(ctx, xhttp.GET, url, xhttp.SendRequest{
    Query:          xhttp.ExampleSendQuery{
        Published: true,
        Limit:     2,
    },
})
```

而没有使用结构体标签的设计
 
```go
type ExampleSendQuery struct {
    Published bool `query:"published_eq"`
    Limit int `query:"Limit"`
}
```

原因是  `Query() (string, error)` 更加灵活，不使用反射性能更高。

（在一些要求将请求加密后生成 sign 的场景 `Query() (string, error)` 更方便）

## test

你可以使用 test xhttp.NewTest 去创建测试代码.

```go
func TestTest(t *testing.T) {
	router := newTestRouter()
	test := xhttp.NewTest(t, router)
	test.RequestJSON(xhttp.Route{xhttp.POST, "/"}, RequestHome{
		ID:   "1",
		Name: "nimo",
		Age:  18,
	}).ExpectJSON(200, ReplyHome{IDNameAge:"1:nimo:18"})

	test.RequestJSON(xhttp.Route{xhttp.GET, "/count"}, nil).ExpectString(200, "1")

	test.RequestJSON(xhttp.Route{xhttp.GET, "/count"}, nil).ExpectString(200, "2")
	test.RequestJSON(xhttp.Route{xhttp.POST, "/count"}, nil).ExpectString(405, "")

	test.RequestJSON(xhttp.Route{xhttp.GET, "/error"}, nil).ExpectString(500, "error")
	test.RequestJSON(xhttp.Route{xhttp.GET, "/panic"}, nil).ExpectString(500, "panic")
	{
		r, err := xhttp.SendRequest{
			FormData: TestFormDataReq{
				Name: "nimo",
			},
		}.HttpRequest(context.TODO(), xhttp.POST, "/form") ; assert.NoError(t, err)
		test.Request(r).ExpectString(200, "nimo")
	}

}
type TestFormDataReq struct {
	Name string
}
func (v TestFormDataReq) FormData(formWriter *multipart.Writer) (err error) {
	err = formWriter.WriteField("name", v.Name) ; if err != nil {
	    return
	}
	return
}
```
