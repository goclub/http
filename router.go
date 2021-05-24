package xhttp

import (
	"fmt"
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
	patterns []Pattern
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

func (router Router) LogPatterns(addr string) {
	var messages []string
	messages = append(messages, "Listen http://localhost" + addr)
	for _, pattern := range router.patterns {
		method := fmt.Sprintf("%-7s", pattern.Method.String())
		url := "http://localhost" + addr + pattern.Path
		if pattern.Method == GET {
			messages = append(messages,  method + " " + url)
		} else {
			messages = append(messages,  `curl  --request \` + "\n" +fmt.Sprintf("%-6s", pattern.Method.String())+` '`+ url +`' --header 'Content-Type: application/json' --data-raw '{}'`)
		}
	}
	log.Print("\n" + strings.Join(messages, "\n"))
}

func (router Router) Static(rootPath string, handler http.Handler) {
	router.router.PathPrefix(rootPath).Handler(http.StripPrefix(rootPath, handler))
}