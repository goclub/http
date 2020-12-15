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
