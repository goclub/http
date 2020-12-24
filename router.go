package xhttp

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type Router struct {
	router *mux.Router
	OnCatchError func(c *Context, errInterface interface{}) error
	patterns []string
}
type RouterOption struct {
	OnCatchError func(c *Context, errInterface interface{}) error
}
func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}
func NewRouter(opt RouterOption) *Router {
	r := mux.NewRouter()
	return &Router{
		router: r,
		OnCatchError: opt.OnCatchError,
	}
}

func (router Router) LogPatterns() {
	log.Print("\n" + strings.Join(router.patterns, "\n"))
}