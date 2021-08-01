package xhttp

import (
	"github.com/gorilla/mux"
	"net/http"
)


type HandleFunc func(c *Context) (err error)
func (serve *Router) HandleFunc(route Route, handler HandleFunc) {
	coreHandleFunc(serve, serve.router, route, handler)
}

func coreHandleFunc(serve *Router, router *mux.Router, route Route,  handler HandleFunc) {
	serve.patterns = append(serve.patterns, route)
	router.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(w, r, serve)
		defer func() {
			r := recover()
			if r  != nil {
				c.CheckPanic(r) ; return
			}
		}()
		err := handler(c)
		if err != nil {
			c.CheckError(err) ; return
		}
	}).Methods(route.Method.String())
}

func (group *Group) HandleFunc(route Route,action HandleFunc) {
	coreHandleFunc(group.serve, group.router, route, action)
}