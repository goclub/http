package xhttp

import (
	"fmt"
	xerr "github.com/goclub/error"
	"github.com/gorilla/mux"
	"log"
	"mime"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type Router struct {
	router       *mux.Router
	OnCatchError func(c *Context, err error) error
	OnCatchPanic func(c *Context, recoverValue interface{}) error
	patterns     []Route
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
		router:       r,
		OnCatchError: opt.OnCatchError,
		OnCatchPanic: opt.OnCatchPanic,
	}
}

func (router Router) LogPatterns(server *http.Server) {
	addr := server.Addr
	var messages []string
	messages = append(messages, "Listen http://localhost"+addr)
	for _, route := range router.patterns {
		method := fmt.Sprintf("%-13s", route.Method.String())
		url := "http://localhost" + addr + route.Path
		if route.Method == GET {
			messages = append(messages, method+" "+url)
		} else {
			messages = append(messages, `curl -X `+route.Method.String()+` '`+url+`' --header 'Content-Type: application/json' --data-raw '{}'`)
		}
	}
	log.Print("\n" + strings.Join(messages, "\n"))
}
func init() {
	var err error
	err = mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	if err != nil {
		xerr.PrintStack(err)
	}
	err = mime.AddExtensionType(".css", "text/css; charset=utf-8")
	if err != nil {
		xerr.PrintStack(err)
	}
}

// example:
// dir := path.Join(os.Getenv("GOPATH"), "src/github.com/goclub/http/example/internal/gin/public")
// defer r.FileServer("/public", dir, true)
// noCache 不使用缓存
func (router *Router) FileServer(prefix string, dir string, noCache bool, middleware func(c *Context, next Next) (err error)) {
	if middleware == nil {
		middleware = func(c *Context, next Next) (err error) {
			return next()
		}
	}
	if strings.HasPrefix("internal", dir) {
		panic(xerr.New("xhttp.Router{}.Static(prefix, dir) prefix maybe contains golang code"))
	}
	if prefix == "/" {
		panic(xerr.New("xhttp.Router{}.Static(prefix, dir) prefix can not be /, is unsafe"))
	}
	router.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, fileServer{
		NoCache:    noCache,
		handler:    http.FileServer(http.Dir(dir)),
		router:     router,
		middleware: middleware,
	}))

}

type fileServer struct {
	NoCache    bool
	handler    http.Handler
	middleware func(c *Context, next Next) (err error)
	router     *Router
}

func (f fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.NoCache {
		w.Header().Add("Cache-control", "no-cache")
		w.Header().Add("Cache-control", "no-store")
		w.Header().Add("Expires", "0")
		w.Header().Add("Last-Modified", time.Now().String())
		w.Header().Add("Pragma", "no-cache")
	}
	c := NewContext(w, r, f.router)
	err := f.middleware(c, func() error {
		f.handler.ServeHTTP(w, r)
		return nil
	})
	if err != nil {
		c.WriteStatusCode(500)
		cacheErr := c.router.OnCatchError(c, err)
		if err != nil {
			panic(cacheErr)
		}
	}
}
func (router Router) PrefixHandler(prefix string, handler http.Handler) {
	router.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, handler))
}
