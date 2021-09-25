package main

import (
	"bytes"
	"context"
	"fmt"
	xhttp "github.com/goclub/http"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"time"
)

func main() {
	r := xhttp.NewRouter(xhttp.RouterOption{
		// error 拦截器 会在 r.Use() r.HandleFunc() 返回 err 不为 nil 时执行
		OnCatchError: func(c *xhttp.Context, err error) error {
			log.Print(err)
			debug.PrintStack()
			return nil
		},
		// panic 拦截器 会在 r.Use() r.HandleFunc() 中出现 panic 时通过 recover 捕获并触发
		OnCatchPanic: func(c *xhttp.Context, recoverValue interface{}) error {
			log.Print(recoverValue)
			debug.PrintStack()
			return nil
		},
	})
	// 静态资源

	r.FileServer("/public", path.Join(os.Getenv("GOPATH"), "src/github.com/goclub/http/example/internal/gin/public"), true)
	r.HandleFunc(xhttp.Route{xhttp.GET, "/user/{name}"}, func(c *xhttp.Context) (reject error) {
		name, reject := c.Param("name") ; if reject != nil {
			return
		}
		return c.WriteBytes([]byte(name))
	})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/welcome"}, func(c *xhttp.Context) (reject error) {
		query := c.Request.URL.Query()
		firstName := query.Get("firstname")
		if firstName == "" {firstName = "Guest"}
		lastName := query.Get("lastname")
		return c.WriteBytes([]byte("hello " + firstName + " " + lastName))
	})
	/*
	curl --location --request POST 'http://127.0.0.1:1111/from_post' \
		--form 'message="abc"' \
		--form 'nike="123"'
	*/
	r.HandleFunc(xhttp.Route{xhttp.POST, "/from_post"}, func(c *xhttp.Context) (reject error) {
		message := c.Request.FormValue("message")
		nick := c.Request.FormValue("nick")
		if nick == "" {nick = "anonymous"}
		return c.WriteJSON(map[string]interface{}{
			"status":  "posted",
			"message": message,
			"nick":    nick,
		})
	})
	/*
	curl --location --request POST 'http://127.0.0.1:1111/post?id=1&page=2' \
		--form 'name="goclub"' \
		--form 'message="abc"'
	*/
	r.HandleFunc(xhttp.Route{xhttp.POST, "/post"}, func(c *xhttp.Context) (reject error) {
		query := c.Request.URL.Query()
		id := query.Get("id")
		pageStr := query.Get("page")
		if pageStr == "" { pageStr = "0" }
		page, reject :=  strconv.ParseUint(pageStr, 10, 64) ; if reject != nil {
			return
		}
		name := c.Request.FormValue("name")
		message := c.Request.FormValue("message")
		return c.WriteJSON(map[string]interface{}{
			"id": id,
			"page": page,
			"name": name,
			"message": message,
		})
	})
	/*
	curl -X POST http://localhost:1111/upload \
	-F "file=@/Users/nimo/Desktop/1.txt" \
	-H "Content-Type: multipart/form-data"
	*/
	r.HandleFunc(xhttp.Route{xhttp.POST, "/upload"}, func(c *xhttp.Context) (reject error) {
		file, fileHeader, reject := c.Request.FormFile("file") ; if reject != nil {
			return
		}
		log.Print(fileHeader.Filename)
		data, reject := ioutil.ReadAll(file) ; if reject != nil {
			return
		}
		log.Print(string(data))
		return c.WriteBytes([]byte(fileHeader.Filename))
	})
	/*
	curl -X POST http://localhost:1111/multi_file_upload \
	-F "upload[]=@/Users/nimo/Desktop/1.txt" \
	-F "upload[]=@/Users/nimo/Desktop/2.txt" \
	-H "Content-Type: multipart/form-data"
	*/
	r.HandleFunc(xhttp.Route{xhttp.POST, "/multi_file_upload"}, func(c *xhttp.Context) (reject error) {
		reject = c.Request.ParseMultipartForm(8 << 20) ; if reject != nil {
			return
		}
		form := c.Request.MultipartForm

		files := form.File["upload[]"]
		for _, fileHeader := range files {
			log.Println(fileHeader.Filename)
			var data []byte
			file, err := fileHeader.Open() ; if err != nil {
				return
			}
			data, reject = ioutil.ReadAll(file) ; if reject != nil {
				return
			}
			log.Print(string(data))
		}
		return c.WriteBytes([]byte(fmt.Sprintf("%d files uploaded!", len(files))))
	})
	// goclub/http 不希望在 Group() 中传递前缀，这样会导致url分散
	v1 := r.Group()
	v1.HandleFunc(xhttp.Route{xhttp.GET, "/v1/login"}, func(c *xhttp.Context) (reject error) {
		return c.WriteBytes([]byte("v1 login"))
	})
	v2 := r.Group()
	v2.HandleFunc(xhttp.Route{xhttp.GET, "/v2/login"}, func(c *xhttp.Context) (reject error) {
		return c.WriteBytes([]byte("v2 login"))
	})
	// 中间件
	r.Use(func(c *xhttp.Context, next xhttp.Next) (reject error) {
		requestTime := time.Now()
		log.Print("Request: ", c.Request.Method, c.Request.URL.String())
		reject = next() ; if reject != nil {
			return
		}
		responseTime := time.Now().Sub(requestTime)
		log.Print("Response: (" , responseTime.String(), ") ", c.Request.Method, c.Request.URL.String())
		return nil
	})
	// goclub/http 自己实现了绑定器，用于绑定各种 http 请求
	// 使用自定义结构绑定表单数据
	/*
	curl --location --request POST 'http://127.0.0.1:1111/bind_query_form/1?name=goclub' \
		--form 'age=18'
	*/
	r.HandleFunc(xhttp.Route{xhttp.POST, "/bind_query_form/{id}"}, func(c *xhttp.Context) (reject error) {
		request := struct {
			ID string `param:"id"`
			Name string `query:"name"`
			// 会将 string int 自动转换
			Age int `form:"age"`
		}{}
		reject = c.BindRequest(&request) ; if reject != nil {
			return
		}
		return c.WriteJSON(request)
	})
	/*
	curl --location --request POST 'http://127.0.0.1:1111/bind_query_json?name=goclub' \
	--header 'Content-Type: application/json' \
	--data-raw '{"id": "1", "age": 18}'
	*/
	r.HandleFunc(xhttp.Route{xhttp.POST, "/bind_query_json"}, func(c *xhttp.Context) (reject error) {
		request := struct {
			Name string `query:"name"`
			// 会将 string int 自动转换
			Age int `json:"age"`
			// 会将 string int 自动转换
			ID int `json:"id"`
		}{}
		reject = c.BindRequest(&request) ; if reject != nil {
			return
		}
		return c.WriteJSON(request)
	})
	// 绑定uri， 绑定Get参数或者Post参数 均可以使用 c.BindRequest(ptr) 实现
	// 绑定 xml 使用 xjson.NewDecoder(c.Request.Body).Decode(&data) 科技

	// 请求验证由 goclub/validator 实现， 注意数据验证应当在业务逻辑层(biz/service)验证，而不是协议层(http/rpc)验证
	// 返回第三方获取的数据 goclub/http 单独提供了一些函数来支持
	// https://pkg.go.dev/github.com/goclub/http#example-Client.Send
	// https://pkg.go.dev/github.com/goclub/http#example-Client.Do
	// https://cn.bing.com/search?q=go+html+%E6%A8%A1%E6%9D%BF%E6%B8%B2%E6%9F%93%E5%BA%93%E9%80%89%E5%9E%8B
	r.HandleFunc(xhttp.Route{xhttp.GET, "/index"}, func(c *xhttp.Context) (reject error) {
		// 为了更方面的支持各种模板引擎， goclub/http 提供 render 接口让使用者自己组合
		return c.Render(func(buffer *bytes.Buffer) error {
			buffer.WriteString(`<a href="https://github.com/goclub">goclub</a>`)
			return nil
		})
	})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/redirect"}, func(c *xhttp.Context) (reject error) {
		http.Redirect(c.Writer, c.Request, "https://goclub.vip", 302)
		return nil
	})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/cookie/set"}, func(c *xhttp.Context) (reject error) {
		c.SetCookie(&http.Cookie{
			Name: "time",
			Value:  time.Now().Format("15:04:05"),
		})
		return c.WriteBytes([]byte("set"))
	})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/cookie/get"}, func(c *xhttp.Context) (reject error) {
		cookie, hasCookie, reject := c.Cookie("time") ; if reject != nil {
			return
		}
		message := ""
		if hasCookie == false {
			message = "no cookie"
		} else {
			message = cookie.Value
		}
		return c.WriteBytes([]byte(message))
	})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/cookie/clear"}, func(c *xhttp.Context) (reject error) {
		c.ClearCookie(&http.Cookie{Name: "time"})
		return c.WriteBytes([]byte("clear"))
	})

	// 为了演示，将测试放在注释中，实际上应该在 xxx_test.go 文件中写
	// goclub/http 的测试是不需要 ListenAndServe 就能测试的，因为测试阶段监听端口会有安全隐患
	// NewTest 只需要传入 *testing.T 和 *xhttp.Router
	// goclub/http test 自带 CookieJar
	// https://pkg.go.dev/github.com/goclub/http#Test
	/*
		test := xhttp.NewTest(t, r)
		request, err := http.NewRequest("GET", "/user/nimo", nil) ; if err != nil {
			panic(err)
		}
		resp := test.Request(request)
		resp.ExpectString(200, "nimo")

	*/
	addr := ":1111"
	server := &http.Server{
		Addr:  addr,
		Handler:  r,
		// 自定义HTTP配置
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	r.LogPatterns(server)
	log.Print("http://localhost" + addr)
	// 如果需要使用 https 证书 则通过 server.ListenAndServeTLS() 启动服务
	// 或者在 cdn 环节部署 https
	go func() {
		listenErr := server.ListenAndServe() ; if listenErr !=nil {
			if listenErr != http.ErrServerClosed {
				panic(listenErr)
			}
		}
	}()
	// 优雅重启或停止
	xhttp.GracefulClose(func() {
		log.Print("Shuting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}
		log.Println("Server exiting")
	})
}
