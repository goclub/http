package xhttp

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Next func() error

func (serve *Router) Use(middleware func(c *Context, next Next) (err error)) {
	middlewareUse(serve, serve.router, middleware)
}
func middlewareUse(serve *Router, router *mux.Router, middleware func(c *Context, next Next) (err error)) {
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := NewContext(w, r, serve)
			defer func() {
				recoverValue := recover()
				if recoverValue != nil {
					c.CheckPanic(recoverValue)
					return
				}
			}()
			next := func() error {
				handler.ServeHTTP(c.Writer, c.Request)
				return nil
			}
			mwErr := middleware(c, next)
			if mwErr != nil {
				c.CheckError(mwErr)
				return
			}
		})
	})
}
func (group *Group) Use(middleware func(c *Context, next Next) (err error)) {
	middlewareUse(group.serve, group.router, middleware)
}
