package xhttp_test

import (
	"context"
	"errors"
	xhttp "github.com/goclub/http"
	"github.com/stretchr/testify/assert"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"testing"
)

type RequestHome struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
}
type ReplyHome struct {
	IDNameAge string `json:"idNameAge"`
}
func newTestRouter() *xhttp.Router {
	router := xhttp.NewRouter(xhttp.RouterOption{
		OnCatchError: func(c *xhttp.Context, err error) error {
			c.WriteStatusCode(500)
			log.Print(err)
			return c.WriteBytes([]byte("error"))
		},
		OnCatchPanic: func(c *xhttp.Context, recoverValue interface{}) error {
			c.WriteStatusCode(500)
			log.Print(recoverValue)
			return c.WriteBytes([]byte("panic"))
		},
	})
	router.HandleFunc(xhttp.Route{xhttp.POST, "/"}, func(c *xhttp.Context) (err error) {
		request :=  RequestHome{}
		err = c.BindRequest(&request) ; if err != nil {
		    return
		}
		reply := ReplyHome{}
		reply.IDNameAge  = request.ID + ":" + request.Name + ":" + strconv.FormatInt(int64(request.Age), 10)
		return c.WriteJSON(reply)
	})
	router.HandleFunc(xhttp.Route{xhttp.GET, "/count"}, func(c *xhttp.Context) (err error) {
		cookieName := "count"
		var count uint64
		cookie, hasCookie, err := c.Cookie(cookieName) ; if err != nil {
		    return
		}
		if hasCookie {
			count, err = strconv.ParseUint(cookie.Value, 10, 64) ; if err != nil {
			    return
			}
		}
		count++
		c.SetCookie(&http.Cookie{
			Name: cookieName,
			Value: strconv.FormatUint(count, 10),
		})
		return c.WriteBytes([]byte(strconv.FormatUint(count, 10)))
	})
	router.HandleFunc(xhttp.Route{xhttp.GET, "/error"}, func(c *xhttp.Context) (err error) {
		return errors.New("abc")
	})
	router.HandleFunc(xhttp.Route{xhttp.GET, "/panic"}, func(c *xhttp.Context) (err error) {
		panic("123")
		return nil
	})
	router.HandleFunc(xhttp.Route{xhttp.POST, "/form"}, func(c *xhttp.Context) (err error) {
		return c.WriteBytes([]byte(c.Request.FormValue("name")))
	})
	return router
}

func TestTest(t *testing.T) {
	router := newTestRouter()
	test := xhttp.NewTest(t, router)
	test.RequestJSON(xhttp.Route{xhttp.POST, "/"}, RequestHome{
		ID:   "1",
		Name: "nimo",
		Age:  18,
	}).ExpectJSON(200, ReplyHome{IDNameAge:"1:nimo:18"})

	test.RequestJSON(xhttp.Route{xhttp.GET, "/count"}, nil).ExpectString(200, "1")

	test.RequestJSON(xhttp.Route{xhttp.GET, "/count"}, nil).ExpectString(200, "2")
	test.RequestJSON(xhttp.Route{xhttp.POST, "/count"}, nil).ExpectString(405, "")

	test.RequestJSON(xhttp.Route{xhttp.GET, "/error"}, nil).ExpectString(500, "error")
	test.RequestJSON(xhttp.Route{xhttp.GET, "/panic"}, nil).ExpectString(500, "panic")
	{
		r, err := xhttp.SendRequest{
			FormData: TestFormDataReq{
				Name: "nimo",
			},
		}.HttpRequest(context.TODO(), xhttp.POST, "/form") ; assert.NoError(t, err)
		test.Request(r).ExpectString(200, "nimo")
	}

}
type TestFormDataReq struct {
	Name string
}
func (v TestFormDataReq) FormData(formWriter *multipart.Writer) (err error) {
	err = formWriter.WriteField("name", v.Name) ; if err != nil {
	    return
	}
	return
}