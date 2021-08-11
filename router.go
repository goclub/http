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
	patterns []Route
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
			log.Printf("%+v", err)
			debug.PrintStack()
			c.WriteStatusCode(500)
			return nil
		}
	}
	if opt.OnCatchPanic == nil {
		opt.OnCatchPanic = func(c *Context, recoverValue interface{}) error {
			log.Printf("%+v", recoverValue)
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

func (router Router) LogPatterns(server *http.Server) {
	addr := server.Addr
	var messages []string
	messages = append(messages, "Listen http://localhost" + addr)
	for _, route := range router.patterns {
		method := fmt.Sprintf("%-13s", route.Method.String())
		url := "http://localhost" + addr + route.Path
		if route.Method == GET {
			messages = append(messages,  method + " " + url)
		} else {
			messages = append(messages,  `curl -X ` + route.Method.String() + ` '`+ url +`' --header 'Content-Type: application/json' --data-raw '{}'`)
		}
	}
	log.Print("\n" + strings.Join(messages, "\n"))
}

func (router Router) Static(rootPath string, dir string) {
	router.router.PathPrefix(rootPath).Handler(http.StripPrefix("/public", http.FileServer(http.Dir(dir))))
}