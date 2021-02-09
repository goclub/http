package xhttp

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Next func() error
type Middleware func(c *Context, next Next) (reject error)
func (serve *Router) Use(middleware Middleware) {
	middlewareUse(serve, serve.router, middleware)
}
func middlewareUse(serve *Router, router *mux.Router, middleware Middleware) {
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := NewContext(w,r, serve)
			defer func() {
				r := recover()
				if r  != nil {
					c.CheckPanic(r) ; return
				}
			}()
			mwErr := middleware(c, func() error {
				handler.ServeHTTP(w, r)
				return nil
			})
			if mwErr != nil {
				c.CheckError(mwErr) ; return
			}
		})
	})
}
func (group *Group) Use(middleware Middleware) {
	middlewareUse(group.serve, group.router, middleware)
}
