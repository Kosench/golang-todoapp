package core_http_server

import (
	"net/http"

	core_http_middleware "github.com/Kosench/golang-todoapp/internal/core/transport/http/middleware"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []core_http_middleware.Middleware
}

// WithMiddleware применяет middleware маршрута к обработчику и возвращает готовый http.Handler.
func (r *Route) WithMiddleware() http.Handler {
	return core_http_middleware.ChainMiddleware(
		r.Handler,
		r.Middleware...,
	)
}
