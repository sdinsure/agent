package server

import (
	"net/http"

	"github.com/sdinsure/agent/pkg/swagger"
)

type Route struct {
	Pattern string
	Handler http.Handler
}

func NewSwaggerRoute() *Route {
	return &Route{
		Pattern: "/swaggerui/",
		Handler: swagger.SwaggerUIAssetsHttpHandler,
	}
}

func NewOpenAPIV2Route(prefix string, handler http.Handler) *Route {
	return &Route{
		Pattern: prefix,
		Handler: handler,
	}
}
