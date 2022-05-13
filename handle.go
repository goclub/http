package xhttp

import (
	xerr "github.com/goclub/error"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type HandleFunc func(c *Context) (err error)

func (serve *Router) HandleFunc(route Route, handler HandleFunc) {
	coreHandleFunc(serve, serve.router, route, handler)
}
func (group *Group) HandleFunc(route Route, handler HandleFunc) {
	coreHandleFunc(group.serve, group.router, route, handler)
}

func coreHandleFunc(serve *Router, router *mux.Router, route Route, handler HandleFunc) {
	if route.Path == "" {
		panic(xerr.New("goclub/http: HandleFunc(route) route.Path can not be empty string"))
	}
	if strings.HasPrefix(route.Path, "/") == false {
		route.Path = "/" + route.Path
		xerr.PrintStack(xerr.New("goclub/http: HandleFunc(route) route.Path must has prefix /"))
	}
	serve.patterns = append(serve.patterns, route)
	router.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(w, r, serve)
		defer func() {
			r := recover()
			if r != nil {
				c.CheckPanic(r)
				return
			}
		}()
		err := handler(c)
		if err != nil {
			c.CheckError(err)
			return
		}
	}).Methods(route.Method.String())
}

func (serve *Router) Handle(path string, handler http.Handler) {
	coreHandle(serve, serve.router, path, handler)
}
func (group *Group) Handle(path string, handler http.Handler) {
	coreHandle(group.serve, group.router, path, handler)
}

func coreHandle(serve *Router, router *mux.Router, path string, handler http.Handler) {
	serve.patterns = append(serve.patterns, Route{"*", path})
	router.Handle(path, handler)
}
