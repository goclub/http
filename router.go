package xhttp

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

type Router struct {
	router *mux.Router
	OnCatchError func(c *Context, err error) error
	OnCatchPanic func(c *Context, recoverValue interface{}) error
	patterns []string
}
type RouterOption struct {
	OnCatchError func(c *Context, err error) error
	OnCatchPanic func(c *Context, recoverValue interface{}) error
}
func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}
func NewRouter(opt RouterOption) *Router {
	r := mux.NewRouter()
	if opt.OnCatchError == nil {
		opt.OnCatchError = func(c *Context, err error) error {
			log.Print(err)
			debug.PrintStack()
			c.WriteStatusCode(500)
			return nil
		}
	}
	if opt.OnCatchPanic == nil {
		opt.OnCatchPanic = func(c *Context, recoverValue interface{}) error {
			log.Print(recoverValue)
			debug.PrintStack()
			c.WriteStatusCode(500)
			return nil
		}
	}
	return &Router{
		router: r,
		OnCatchError: opt.OnCatchError,
		OnCatchPanic: opt.OnCatchPanic,
	}
}

func (router Router) LogPatterns() {
	log.Print("\n" + strings.Join(router.patterns, "\n"))
}