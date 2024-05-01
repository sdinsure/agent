package server

import (
	"context"

	apppb "github.com/sdinsure/agent/example/api/pb"
	grpcserver "github.com/sdinsure/agent/pkg/grpc/server"
	grpchttpgatewayserver "github.com/sdinsure/agent/pkg/grpc/server/httpgateway"
	"github.com/sdinsure/agent/pkg/logger"
)

func NewServerService(
	httpPort int,
	log logger.Logger,
) (*ServerService, error) {

	svr := grpcserver.NewGrpcServer(grpcserver.WithLogger(log))
	httpGateway, err := grpchttpgatewayserver.NewHTTPGatewayServer(svr, log, httpPort)
	if err != nil {
		return nil, err
	}
	return &ServerService{
		svr:         svr,
		httpGateway: httpGateway,
	}, nil
}

type ServerService struct {
	svr         *grpcserver.GrpcServer
	httpGateway *grpchttpgatewayserver.HTTPGatewayServer
}

func (s *ServerService) Start() error {
	go func() {
		s.httpGateway.ListenAndServe()
	}()

	return s.httpGateway.WaitForSIGTERM()
}

func (s *ServerService) Stop() error {
	return s.httpGateway.Shutdown(context.Background())
}

func RegisterService(svr *ServerService, a *HelloServiceService) {
	apppb.RegisterHelloServiceServer(svr.svr, a)
	svr.httpGateway.RegisterHandlers(apppb.RegisterHelloServiceHandler)
}
