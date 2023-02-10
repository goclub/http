package main

import (
	xhttp "github.com/goclub/http"
	"net/http"
)

func main() {
	r := xhttp.NewRouter(xhttp.RouterOption{})
	server := &http.Server{
		Addr:    ":2222",
		Handler: r,
	}
	r.HandleFunc(xhttp.Route{xhttp.GET, "/"}, func(c *xhttp.Context) (err error) {
		return c.WriteJSON(map[string]interface{}{"name": "goclub/http"})
	})
	r.LogPatterns(server)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
