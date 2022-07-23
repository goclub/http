package xhttp

import (
	"net/url"
	"strconv"
)

type ExampleReplyPost struct {
	Slideshow struct {
		Title  string `json:"title"`
		Author string `json:"author"`
		Date   string `json:"date"`
		Slides []struct {
			Title string   `json:"title"`
			Type  string   `json:"type"`
			Items []string `json:"items,omitempty"`
		} `json:"slides"`
	} `json:"slideshow"`
}

type ExampleSendQuery struct {
	Published bool
	Limit     int
}

// 通过实现结构体  Query() (string, error) 方法后传入 xhttp.SendRequest{}.Query
// 即可设置请求 query 参数
func (r ExampleSendQuery) Query() (string, error) {
	v := url.Values{}
	v.Set("published_eq", strconv.FormatBool(r.Published))
	v.Set("limit", strconv.Itoa(r.Limit))
	return v.Encode(), nil
}
