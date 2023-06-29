package xhttp

import "github.com/gorilla/mux"

type Group struct {
	serve  *Router
	router *mux.Router
}

func (serve *Router) Group() *Group {
	return &Group{
		serve:  serve,
		router: serve.router.PathPrefix("").Subrouter(),
	}
}
