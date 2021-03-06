package xhttp_test

import (
	"errors"
	xhttp "github.com/goclub/http"
	"log"
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
	router.HandleFunc(xhttp.Pattern{xhttp.POST, "/"}, func(c *xhttp.Context) (reject error) {
		request :=  RequestHome{}
		reject = c.BindRequest(&request) ; if reject != nil {
		    return
		}
		reply := ReplyHome{}
		reply.IDNameAge  = request.ID + ":" + request.Name + ":" + strconv.FormatInt(int64(request.Age), 10)
		return c.WriteJSON(reply)
	})
	router.HandleFunc(xhttp.Pattern{xhttp.GET, "/count"}, func(c *xhttp.Context) (reject error) {
		cookieName := "count"
		var count uint64
		cookie, hasCookie, reject := c.Cookie(cookieName) ; if reject != nil {
		    return
		}
		if hasCookie {
			count, reject = strconv.ParseUint(cookie.Value, 10, 64) ; if reject != nil {
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
	router.HandleFunc(xhttp.Pattern{xhttp.GET, "/error"}, func(c *xhttp.Context) (reject error) {
		return errors.New("abc")
	})
	router.HandleFunc(xhttp.Pattern{xhttp.GET, "/panic"}, func(c *xhttp.Context) (reject error) {
		panic("123")
		return nil
	})
	return router
}

func TestTest(t *testing.T) {
	router := newTestRouter()
	test := xhttp.NewTest(t, router)
	test.RequestJSON(xhttp.Pattern{xhttp.POST, "/"}, RequestHome{
		ID:   "1",
		Name: "nimo",
		Age:  18,
	}).ExpectJSON(200, ReplyHome{IDNameAge:"1:nimo:18"})

	test.RequestJSON(xhttp.Pattern{xhttp.GET, "/count"}, nil).ExpectString(200, "1")

	test.RequestJSON(xhttp.Pattern{xhttp.GET, "/count"}, nil).ExpectString(200, "2")
	test.RequestJSON(xhttp.Pattern{xhttp.POST, "/count"}, nil).ExpectString(405, "")

	test.RequestJSON(xhttp.Pattern{xhttp.GET, "/error"}, nil).ExpectString(500, "error")
	test.RequestJSON(xhttp.Pattern{xhttp.GET, "/panic"}, nil).ExpectString(500, "panic")
}