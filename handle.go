package xhttp

import (
	"github.com/gorilla/mux"
	"net/http"
)


type HandleFunc func(c *Context) (err error)
func (serve *Router) HandleFunc(pattern Pattern, handler HandleFunc) {
	coreHandleFunc(serve, serve.router, pattern, handler)
}

func coreHandleFunc(serve *Router, router *mux.Router, pattern Pattern,  handler HandleFunc) {
	serve.patterns = append(serve.patterns, pattern)
	router.HandleFunc(pattern.Path, func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods(pattern.Method.String())
}

func (group *Group) HandleFunc(pattern Pattern,action HandleFunc) {
	coreHandleFunc(group.serve, group.router, pattern, action)
}