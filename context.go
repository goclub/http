package xhttp

import (
	"net/http"
)

// Context 包含 *http.Request http.ResponseWriter 并封装一些便捷方法
type Context struct {
	Writer        http.ResponseWriter
	Request       *http.Request
	router        *Router
	resolvedParam bool
	param         map[string]string
}

func NewContext(w http.ResponseWriter, r *http.Request, router *Router) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		router:  router,
	}
}

// CheckPanic 让 Router{}.OnCatchError 处理传入的错误
func (c *Context) CheckPanic(r interface{}) {
	err := c.router.OnCatchPanic(c, r)
	if err != nil {
		panic(err)
	}
}

// CheckError 让 Router{}.OnCatchError 处理传入的错误
func (c *Context) CheckError(err error) {
	err = c.router.OnCatchError(c, err)
	if err != nil {
		panic(err)
	}
}
