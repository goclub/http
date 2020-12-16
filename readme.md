# xhttp

> 基于 Go 标准库 net/http 扩展出一些安全便捷的方法

```go
package main
import "github.com/goclub/http"
func main () {
    xhttp.
}
```

## 特性

1. 封装 `xhttp.Client{}` 使用 `client.Do()` 时 不会忘记处理 `resp.Body.Close` 和 `resp.StatusCode` 
2. 新增 `xhttp.Client{}.Send()` 方法高性能更便捷的发起常见请求（`query` `formurlencoded` `formdata` `json`）

## 特殊说明

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