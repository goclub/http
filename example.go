package xhttp

import (
	"net/url"
	"strconv"
)
type ExampleReplyPost struct{
	ID int `json:"id"`
	Title string `json:"title"`
	Views int `json:"views"`
	Published bool `json:"published"`
	CreatedAt string `json:"createdAt"`
}
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
