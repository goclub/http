# xhttp

> 基于 Go 标准库 net/http 扩展出一些安全便捷的方法


```go
package main
import "github.com/goclub/http"
func main () {
    xhttp.
}
```


## http server

**示例**

1. [like gin example](./example/internal/gin/main.go)

**相关的包**

1. [goclub/session](https://github.com/goclub/session)
2. [goclub/validator](https://github.com/goclub/validator)
3. [goclub/error](https://github.com/goclub/error)

## http client

1. [Client.Send](https://pkg.go.dev/github.com/goclub/http#Client.Send)
2. [Client.Do](https://pkg.go.dev/github.com/goclub/http#Client.Do)


## 特性
1. http server 支持 `OnCatchError` `OnCatchPanic` 拦截器，让错误处理更简单，让panic时对客户端更友好。
2. http client  `xhttp.Client{}.Send()` 高易用高性能的发起常见请求（`query` `formurlencoded` `formdata` `json`）

## 特殊说明

### reject error

http server `Router{}.HandleFunc` `Router{}.Use` 函数签名的出参都有 `(reject error)` ，你可以给改成 (err error) ，使用 reject 只是为了与 https://github.com/goclub/error#reject 呼应。

如果没有错误则 `return nil`, 如果则通过 `return reject` 传递 。最终会被 OnCatchError 处理。

### Client{}.Send()

`xhttp.Client{}.Send()` 绑定常见请求 `query` `formUrlencoed` `form-data` `json`
 
是通过实现一个符合 `Query() (url.Values, error)` 接口的结构体完成设置请求。

```go
type ExampleSendQuery struct {
	Published bool
	Limit int
}
// 通过实现结构体  Query() (url.Values, error) 方法后传入 xhttp.SendRequest{}.Query
// 即可设置请求 query 参数
func (r ExampleSendQuery) Query() (url.Values, error) {
	v := url.Values{}
	v.Set("published_eq", strconv.FormatBool(r.Published))
	v.Set("limit", strconv.Itoa(r.Limit))
	return v, nil
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

原因是  `Query() (url.Values, error)` 更加灵活，不使用反射性能更高。

（在一些要求将请求加密后生成 sign 的场景 `Query() (url.Values, error)` 更方便）