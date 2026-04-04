package core_http_server

import (
	"net/http"

	core_http_middleware "github.com/Kosench/golang-todoapp/internal/core/transport/http/middleware"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	MiddleWare []core_http_middleware.Middleware
}

func NewRoute(meth, path string, handler http.HandlerFunc) Route {
	return Route{
		Method:  meth,
		Path:    path,
		Handler: handler,
	}
}

func (r *Route) WithMiddleWare() http.Handler {
	return core_http_middleware.ChainMiddleware(
		r.Handler,
		r.MiddleWare...,
	)
}
