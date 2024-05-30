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
	"google.golang.org/protobuf/proto"
)

type ServiceConfigure interface {
	apply(*ServiceConfig)
}

// GrpcMetadataModifier can be used to pass http.Requests into grpc context via metadata.MD
type GrpcMetadataModifier func(context.Context, *http.Request) metadata.MD

// GrpcForwardResponseModifier would be used to modify http.ResponseWriter before the forward process is started
type GrpcForwardResponseModifier func(context.Context, http.ResponseWriter, proto.Message) error

// IncomingHeaderMatcher is used to convert grpc.header to http.header with whitelisted
type IncomingHeaderMatcher runtime.HeaderMatcherFunc

// OutgoingHeaderMatcher is used to convert http.header to grpc.metadata with whitelisted
type OutgoingHeaderMatcher runtime.HeaderMatcherFunc

type CustomizeMarshaler struct {
	mime      string
	marshaler runtime.Marshaler
}

type ServiceConfig struct {
	// paired
	unaryMiddlewares  []grpc.UnaryServerInterceptor
	streamMiddlewares []grpc.StreamServerInterceptor

	// paired
	serverTransportCredentials credentials.TransportCredentials
	clientTransportCredentials credentials.TransportCredentials

	metadataModifiers []GrpcMetadataModifier
	forwardModifiers  []GrpcForwardResponseModifier

	// incoming
	incomingHeaderMatchFunc IncomingHeaderMatcher

	// outgoing
	outgoingHeaderMatchFunc OutgoingHeaderMatcher

	log logger.Logger

	marshalers []CustomizeMarshaler

	maxRecvMsgSize int
}

func newServiceConfig(scs ...ServiceConfigure) *ServiceConfig {
	c := &ServiceConfig{
		serverTransportCredentials: insecure.NewCredentials(), // insecure
		clientTransportCredentials: insecure.NewCredentials(), // insecure

		metadataModifiers: []GrpcMetadataModifier{
			GrpcMetadataModifier(grpcmetadata.HttpCookiesToGrpcMetadata),
		},
		log:                     logger.NewLogger(),
		incomingHeaderMatchFunc: runtime.DefaultHeaderMatcher,
		outgoingHeaderMatchFunc: deniedAll,
		maxRecvMsgSize:          64 * 1024 * 1024, /*64M*/
	}
	for _, sc := range scs {
		sc.apply(c)
	}
	return c
}

func deniedAll(headerString string) (string, bool) {
	return headerString, false
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

type forwardModifierConfigure struct {
	modifier GrpcForwardResponseModifier
}

func (m forwardModifierConfigure) apply(sc *ServiceConfig) {
	sc.forwardModifiers = append(sc.forwardModifiers, m.modifier)
}

func WithForwardModifier(md GrpcForwardResponseModifier) forwardModifierConfigure {
	return forwardModifierConfigure{
		modifier: md,
	}
}

type marshalerOptionConfigure struct {
	mime      string
	marshaler runtime.Marshaler
}

func (m marshalerOptionConfigure) apply(sc *ServiceConfig) {
	sc.marshalers = append(sc.marshalers, CustomizeMarshaler{
		mime:      m.mime,
		marshaler: m.marshaler,
	})
}

func WithMarshaler(mime string, m runtime.Marshaler) marshalerOptionConfigure {
	return marshalerOptionConfigure{
		mime:      mime,
		marshaler: m,
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

type incomingHeaderMatcher struct {
	fn IncomingHeaderMatcher
}

func (g incomingHeaderMatcher) apply(sc *ServiceConfig) {
	sc.incomingHeaderMatchFunc = g.fn
}

func WithIncomingHeaderMatcher(fn IncomingHeaderMatcher) incomingHeaderMatcher {
	return incomingHeaderMatcher{
		fn: fn,
	}
}

type outgoingHeaderMatcher struct {
	fn OutgoingHeaderMatcher
}

func (g outgoingHeaderMatcher) apply(sc *ServiceConfig) {
	sc.outgoingHeaderMatchFunc = g.fn
}

func WithOutgoingHeaderMatcher(fn OutgoingHeaderMatcher) outgoingHeaderMatcher {
	return outgoingHeaderMatcher{
		fn: fn,
	}
}

type maxRecvMsgSize struct {
	maxRecvMsgSize int
}

func (m maxRecvMsgSize) apply(sc *ServiceConfig) {
	sc.maxRecvMsgSize = m.maxRecvMsgSize
}

// WithMaxRecvMsgSize allows server to received this such size
// while allows client to be able to receive this such as well.
// NOTE: there is no limit during send operation in both server/client
func WithMaxRecvMsgSize(size int) maxRecvMsgSize {
	return maxRecvMsgSize{
		maxRecvMsgSize: size,
	}
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
		grpcserver.WithGrpcServerOption(
			grpc.MaxRecvMsgSize(config.maxRecvMsgSize),
			grpc.Creds(config.serverTransportCredentials),
		),
		grpcserver.WithGrpcPort(grpcPort),
	)

	var serveMuxOptions []runtime.ServeMuxOption
	for _, metadataModifier := range config.metadataModifiers {
		serveMuxOptions = append(serveMuxOptions, runtime.WithMetadata(metadataModifier))
	}
	for _, forwardModifier := range config.forwardModifiers {
		serveMuxOptions = append(serveMuxOptions, runtime.WithForwardResponseOption(forwardModifier))
	}
	for _, customizedMarshaler := range config.marshalers {
		serveMuxOptions = append(serveMuxOptions, runtime.WithMarshalerOption(customizedMarshaler.mime, customizedMarshaler.marshaler))
	}
	serveMuxOptions = append(serveMuxOptions,
		runtime.WithIncomingHeaderMatcher(runtime.HeaderMatcherFunc(config.incomingHeaderMatchFunc)),
		runtime.WithOutgoingHeaderMatcher(runtime.HeaderMatcherFunc(config.outgoingHeaderMatchFunc)),
	)

	httpGateway, err := httpgateway.NewHTTPGatewayServer(
		svr,
		config.log,
		httpPort,
		httpgateway.WithMaxCallRecvMsgSize(config.maxRecvMsgSize),
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
