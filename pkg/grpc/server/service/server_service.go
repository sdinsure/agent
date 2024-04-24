package service

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcserver "github.com/sdinsure/agent/pkg/grpc/server"
	httpgateway "github.com/sdinsure/agent/pkg/grpc/server/httpgateway"
	grpcmetadata "github.com/sdinsure/agent/pkg/grpc/server/metadata"
	"github.com/sdinsure/agent/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ServiceConfigure interface {
	apply(*ServiceConfig)
}

type GrpcMetadataModifier func(context.Context, *http.Request) metadata.MD

type ServiceConfig struct {
	// paired
	unaryMiddlewares  []grpc.UnaryServerInterceptor
	streamMiddlewares []grpc.StreamServerInterceptor

	// paired
	serverTransportCredentials credentials.TransportCredentials
	clientTransportCredentials credentials.TransportCredentials

	metadataModifiers []GrpcMetadataModifier
	log               logger.Logger
}

func newServiceConfig(scs ...ServiceConfigure) *ServiceConfig {
	c := &ServiceConfig{
		serverTransportCredentials: insecure.NewCredentials(), // insecure
		clientTransportCredentials: insecure.NewCredentials(), // insecure

		metadataModifiers: []GrpcMetadataModifier{
			GrpcMetadataModifier(grpcmetadata.HttpCookiesToGrpcMetadata),
		},
		log: logger.NewLogger(),
	}
	for _, sc := range scs {
		sc.apply(c)
	}
	return c
}

type middlewareConfigure struct {
	unaryMiddlewares  []grpc.UnaryServerInterceptor
	streamMiddlewares []grpc.StreamServerInterceptor
}

func (m middlewareConfigure) apply(sc *ServiceConfig) {
	sc.unaryMiddlewares = append(sc.unaryMiddlewares, m.unaryMiddlewares...)
	sc.streamMiddlewares = append(sc.streamMiddlewares, m.streamMiddlewares...)
}

func WithMiddlewareConfigure(unaryMiddlewares []grpc.UnaryServerInterceptor, streamMiddlewares []grpc.StreamServerInterceptor) middlewareConfigure {
	return middlewareConfigure{
		unaryMiddlewares:  unaryMiddlewares,
		streamMiddlewares: streamMiddlewares,
	}
}

type transportCredentialConfigure struct {
	server credentials.TransportCredentials
	client credentials.TransportCredentials
}

func (t transportCredentialConfigure) apply(sc *ServiceConfig) {
	sc.serverTransportCredentials = t.server
	sc.clientTransportCredentials = t.client
}

func WithTransportCredential(server, client credentials.TransportCredentials) transportCredentialConfigure {
	return transportCredentialConfigure{
		server: server,
		client: client,
	}
}

type metadataModifierConfigure struct {
	modifier GrpcMetadataModifier
}

func (m metadataModifierConfigure) apply(sc *ServiceConfig) {
	sc.metadataModifiers = append(sc.metadataModifiers, m.modifier)
}

func WithMetadataModifier(md GrpcMetadataModifier) metadataModifierConfigure {
	return metadataModifierConfigure{
		modifier: md,
	}
}

type loggerConfigure struct {
	log logger.Logger
}

func (l loggerConfigure) apply(sc *ServiceConfig) {
	sc.log = l.log
}

func WithLogger(log logger.Logger) loggerConfigure {
	return loggerConfigure{log: log}
}

func NewServerService(
	grpcPort int,
	httpPort int,
	configures ...ServiceConfigure,
) (*ServerService, error) {

	config := newServiceConfig(configures...)

	svr := grpcserver.NewGrpcServer(
		grpcserver.WithLogger(config.log),
		grpcserver.WithInterceptor(config.unaryMiddlewares, config.streamMiddlewares),
		grpcserver.WithGrpcServerOption(grpc.Creds(config.serverTransportCredentials)),
		grpcserver.WithGrpcPort(grpcPort),
	)

	var serveMuxOptions []runtime.ServeMuxOption
	for _, metadataModifier := range config.metadataModifiers {
		serveMuxOptions = append(serveMuxOptions, runtime.WithMetadata(metadataModifier))
	}

	httpGateway, err := httpgateway.NewHTTPGatewayServer(
		svr,
		config.log,
		httpPort,
		httpgateway.WithTransportCredentials(config.clientTransportCredentials),
		httpgateway.WithServeMuxOption(serveMuxOptions...),
	)
	if err != nil {
		return nil, err
	}
	return &ServerService{
		log:         config.log,
		svr:         svr,
		httpGateway: httpGateway,
	}, nil
}

type ServerService struct {
	log         logger.Logger
	svr         *grpcserver.GrpcServer
	httpGateway *httpgateway.HTTPGatewayServer
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

func (s *ServerService) RegisterService(sd *grpc.ServiceDesc, serviceImpl any, handlers ...httpgateway.GatewayHandlerFunc) error {
	s.svr.RegisterService(sd, serviceImpl)
	return s.httpGateway.RegisterHandlers(handlers...)
}

func (s *ServerService) AddGatewayRoutes(routes ...*httpgateway.Route) error {
	return s.httpGateway.AddRoutes(routes...)
}
