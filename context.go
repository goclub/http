package xhttp

import (
	"bytes"
	"context"
	xjson "github.com/goclub/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
)
// 包含 *http.Request http.ResponseWriter 并封装一些便捷方法
type Context struct {
	Writer http.ResponseWriter
	Request *http.Request
	router *Router
	resolvedParam bool
	param map[string]string
}
func NewContext (w http.ResponseWriter, r *http.Request, router *Router) *Context {
	return &Context{
		Writer: w,
		Request: r,
		router: router,
	}
}
// 等同于 c.Request.Context()
func (c *Context) RequestContext() context.Context {
	return c.Request.Context()
}
// 获取格式为: /user/{userID} URL的参数
// 列如 /user/11 通过 c.Param("userID") 获取
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
		return "", errors.New(`not found param (` + name + `)`)
	}
	return param, nil
}
func (c *Context) WriteStatusCode(statusCode int) {
	c.Writer.WriteHeader(statusCode)
}
// 等同于 writer.Write(data) ,但函数签名返回 error 不返回 int
func (c *Context) WriteBytes(b []byte) error {
	_, err := c.Writer.Write(b)
	if err != nil {
		return err
	}
	return nil
}
// 设置 header Content-Type text/html; charset=UTF-8 并输出 buffer
func (c *Context) Render(render func(buffer *bytes.Buffer) error) error {
	buffer := bytes.NewBuffer(nil)
	err := render(buffer) ; if err != nil {return err}
	c.Writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
	return c.WriteBytes(buffer.Bytes())
}
// 响应 json
func (c *Context) WriteJSON(v interface{}) error {
	data, err := xjson.Marshal(v)
	if err != nil {
		return err
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	return c.WriteBytes(data)
}
// 绑定请求，支持自定义结构体表 `query` `form` `param`
func (c *Context) BindRequest(ptr interface{}) error {
	return BindRequest(ptr, c.Request)
}
// 让 Router{}.OnCatchError 处理传入的错误
func (c *Context) CheckPanic(r interface{}) {
	err := c.router.OnCatchPanic(c, r)
	if err != nil {
		panic(err)
	}
}
// 让 Router{}.OnCatchError 处理传入的错误
func (c *Context) CheckError(err error) {
	err = c.router.OnCatchError(c, err)
	if err != nil {
		panic(err)
	}
}