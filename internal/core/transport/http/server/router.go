package core_http_server

import (
	"fmt"
	"net/http"
)

type APIVersion string

var (
	APIVersion1 = APIVersion("v1")
	APIVersion2 = APIVersion("v2")
	APIVersion3 = APIVersion("v3")
)

type APIVersionRouter struct {
	*http.ServeMux
	apiVersion APIVersion
}

func NewVersionAPI(apiVersion APIVersion) *APIVersionRouter {
	return &APIVersionRouter{
		ServeMux:   http.NewServeMux(),
		apiVersion: apiVersion,
	}
}
func (r *APIVersionRouter) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)
		r.Handle(pattern, route.Handler)
	}
}
