package server

import (
	api "github.com/sdinsure/agent/example/api"
	grpchttpgatewayserver "github.com/sdinsure/agent/pkg/grpc/server/httpgateway"
)

func (s *ServerService) AddGatewayRoutes() error {
	return s.httpGateway.AddRoutes(
		grpchttpgatewayserver.NewSwaggerRoute(),
		grpchttpgatewayserver.NewOpenAPIV2Route("/openapiv2/", api.OpenApiV2HttpHandler),
	)
}
