package main

import (
	"bytes"
	"context"
	"fmt"
	xhttp "github.com/goclub/http"
	"github.com/google/uuid"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)
type traceID string
func main () {
	router := NewRouter()
	// request
	RequestBindQuery(router)
	RequestBindFormUrlencoded(router)
	RequestBindFormData(router)
	RequestBindJSON(router)
	RequestBindQueryAndJSON(router)
	RequestBindParam(router)
	RenderFormFile(router)
	RequestFile(router)
	router.Use(func(c *xhttp.Context, next xhttp.Next) (err error) {
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), traceID("traceID"), uuid.New().String()))
		return next()
	})
	RequestTraceID(router)
	// response
	ResponseWriteBytes(router)
	ResponseHTML(router)
	ResponseTemplate(router)
	GetSetCookie(router)
	addr := ":3000"
	serve := &http.Server{
		Handler: router,
		Addr: addr,
	}
	router.LogPatterns(serve)
	go func() {
		listenErr := serve.ListenAndServe() ; if listenErr !=nil {
			if listenErr != http.ErrServerClosed {
				panic(listenErr)
			}
		}
	}()
	xhttp.GracefulClose(func() {
		log.Print("Shuting down server...")
		if err := serve.Shutdown(context.Background()); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}
		log.Println("Server exiting")
	})
}

func NewRouter() *xhttp.Router {
	router := xhttp.NewRouter(xhttp.RouterOption{})
	return router
}



// 打开 http://127.0.0.1:3000/request/query?name=nimoc&age=18
func RequestBindQuery(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/request/query",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		req := struct {
			Name string `query:"name"`
			Age int `query:"age"`
		}{}
		err = c.BindRequest(&req) ; if err != nil {return}
		dump := fmt.Sprintf("%+v", req)
		return c.WriteBytes([]byte(dump))
	})
}

func RequestBindFormUrlencoded(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.POST, "/request/form_urlencoded",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		req := struct {
			Name string `form:"name"`
			Age int `form:"age"`
		}{}
		err = c.BindRequest(&req) ; if err != nil {return}
		dump := fmt.Sprintf("%+v", req)
		return c.WriteBytes([]byte(dump))
	})
}
/*
使用 curl 发起请求
curl --location --request POST 'http://127.0.0.1:3000/request/form_data' \
--form 'name="nimoc"' \
--form 'age="18"'
*/
func RequestBindFormData(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.POST, "/request/form_data",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		req := struct {
			Name string `form:"name"`
			Age int `form:"age"`
		}{}
		err = c.BindRequest(&req) ; if err != nil {return}
		dump := fmt.Sprintf("%+v", req)
		return c.WriteBytes([]byte(dump))
	})
}
/*
使用 curl 发起请求
curl --location --request GET 'http://127.0.0.1:3000/request/json' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":"nimoc",
    "age": 18
}'
*/
func RequestBindJSON(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/request/json",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		req := struct {
			Name string `json:"name"`
			Age int `json:"age"`
		}{}
		err = c.BindRequest(&req) ; if err != nil {return}
		dump := fmt.Sprintf("%+v", req)
		return c.WriteBytes([]byte(dump))
	})
}


/*
curl --location --request GET 'http://127.0.0.1:3000/request/query_and_json?id=11' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":"nimoc",
    "age": 18
}'
*/
func RequestBindQueryAndJSON(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/request/query_and_json",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		req := struct {
			ID string `query:"id"`
			Name string `json:"name"`
			Age int `json:"age"`
		}{}
		err = c.BindRequest(&req) ; if err != nil {return}
		dump := fmt.Sprintf("%+v", req)
		return c.WriteBytes([]byte(dump))
	})
}

// http://127.0.0.1:3000/request/param/11
func RequestBindParam(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/request/param/{userID}",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		req := struct {
			UserID string `param:"userID"`
		}{}
		err = c.BindRequest(&req) ; if err != nil {return}
		dump := fmt.Sprintf("%+v", req)
		return c.WriteBytes([]byte(dump))
	})
}
func RenderFormFile(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/request/file",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		html := []byte(`
		<form action="/request/file" method="post" enctype="multipart/form-data" >
　　　		<input type="file" name="file" /> <br />
　　　		<button type="submit" >上传</button>
		</form>`)
		return c.Render(func(buffer *bytes.Buffer) error {
			buffer.Write(html)
			return nil
		})
	})
}
// 打开 http://127.0.0.1:3000/request/file 上传文件
func RequestFile(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.POST, "/request/file",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		file, fileHeader, err := c.Request.FormFile("file") ; if err != nil {return}
		defer file.Close()
		data, err := ioutil.ReadAll(file) ; if err != nil {return}
		body :=  append([]byte(fileHeader.Filename + ":"), data...)
		return c.WriteBytes(body)
	})
}
func RequestTraceID(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/request/trace_id",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		return c.WriteBytes([]byte(fmt.Sprintf("traceID: %s", c.RequestContext().Value(traceID("traceID")))))
	})
}
func ResponseWriteBytes(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/response/WriteBytes",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		return c.WriteBytes([]byte("goclub"))
	})
}
func ResponseHTML(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/response/html",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		return c.Render(func(buffer *bytes.Buffer) error {
			buffer.WriteString(`<a href="http://github.com/goclub">goclub</a>`)
			return nil
		})
	})
}
var responseTPL =  template.Must(template.New("").Parse("name {{.Name}}"))
func ResponseTemplate(router *xhttp.Router) {
	route := xhttp.Route{
		xhttp.GET, "/response/template",
	}
	router.HandleFunc(route, func(c *xhttp.Context) (err error) {
		return c.Render(func(buffer *bytes.Buffer) error {
			data := struct {
				Name string
			}{Name:"nimoc"}
			return responseTPL.Execute(buffer, data)
		})
	})
}

func GetSetCookie(router *xhttp.Router) {
	router.HandleFunc(xhttp.Route{xhttp.GET, "/cookie"}, func (c *xhttp.Context) (err error) {
		query := c.Request.URL.Query()
		switch query.Get("kind") {
		case "get":
			var nameCookie *http.Cookie
			var hasValue bool
			nameCookie, hasValue, err = c.Cookie("name") ; if err != nil {
				return
			}
			var name string
			if !hasValue {
				name = ""
			} else {
				name = nameCookie.Value
			}
			return c.WriteBytes([]byte("name:" + name))
		case "set":
			name := query.Get("name")
			if name == "" {
				name = time.Now().String()
			}
			c.SetCookie(&http.Cookie{
				Name: "name",
				Value: name,
			})
			return c.WriteBytes([]byte("set cookie done"))
		default:
			return c.WriteBytes([]byte("kind query.must be get or set"))
		}
	})
}