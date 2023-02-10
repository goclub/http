package xhttp

import (
	xerr "github.com/goclub/error"
	xjson "github.com/goclub/json"
	"github.com/gorilla/mux"
	"github.com/tomasen/realip"
	"reflect"
	"strings"
)

// BindRequest - 绑定请求，支持自定义结构体标签 `query` `form` `param`
func (c *Context) BindRequest(ptr interface{}) error {
	return BindRequest(ptr, c.Request)
}

// UnmarshalJSONFromQuery 从 query 中读取json并解析
func (c *Context) UnmarshalJSONFromQuery(queryKey string, ptr interface{}) (err error) {
	// 判断ptr 必须是指针
	if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
		return xerr.New("goclub/http: Context.UnmarshalJSONFromQuery(queryKey, ptr) ptr not be pointer")
	}
	value := c.Request.URL.Query().Get(queryKey)
	if len(value) != 0 {
		if err = xjson.Unmarshal([]byte(value), ptr); err != nil {
			return err
		}
	}
	return nil
}

// Param - 获取路由参数
// 比如路由是 xhttp.Route{xhttp.GET, "/user/{userID}"}
// 当请求 "/user/11" 时候通过 c.Param("userID") 可以获取到 "11"
func (c *Context) Param(name string) (param string, err error) {
	data := map[string]string{}
	if c.resolvedParam {
		data = c.param
	} else {
		data = mux.Vars(c.Request)
	}
	var has bool
	param, has = data[name]
	if !has {
		return "", xerr.New(`not found param (` + name + `)`)
	}
	return param, nil
}

// RealIP return "20.205.243.166" - 获取当前请求的客户端ip
func (c *Context) RealIP() (ip string) {
	return realip.FromRequest(c.Request)
}

// AbsURL return "http://domain.com/path?q=1"
func (c *Context) AbsURL() (url string) {
	r := c.Request
	url = r.URL.String()
	if strings.HasPrefix(url, "/") {
		url = c.Scheme() + "://" + c.Host() + url
	}
	return
}

// Scheme return "http" or "https", support CDN Forwarded https
func (c *Context) Scheme() (scheme string) {
	r := c.Request
	scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}
	forwardedProto := r.Header.Get("X-Forwarded-Proto")
	if forwardedProto == "https" {
		scheme = forwardedProto
	}
	return
}

// Host return "goclub.run" or "127.0.0.1:1111"
func (c *Context) Host() (host string) {
	r := c.Request
	host = r.Host
	if host == "" {
		host = r.URL.Host
	}
	return
}
