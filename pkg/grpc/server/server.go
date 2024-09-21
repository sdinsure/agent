package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	servermiddleware "github.com/sdinsure/agent/pkg/grpc/server/middleware"
	loggermiddleware "github.com/sdinsure/agent/pkg/grpc/server/middleware/logger"
	metricmiddleware "github.com/sdinsure/agent/pkg/grpc/server/middleware/metrics"
	recoverymiddleware "github.com/sdinsure/agent/pkg/grpc/server/middleware/recovery"
	pkglogger "github.com/sdinsure/agent/pkg/logger"
)

var withReflection = flag.Bool("with_grpc_reflection", false, "turn on grpc reflection")

type GrpcServerConfigurer interface {
	apply(o *grpcServerConfig)
}

type grpcServerConfig struct {
	grpcPort           int
	serverOpts         []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	logger             pkglogger.Logger
}

type withGrpcPort struct {
	grpcPort int
}

var (
	_ GrpcServerConfigurer = withGrpcPort{}
)

func (w withGrpcPort) apply(o *grpcServerConfig) {
	o.grpcPort = w.grpcPort
}

func WithGrpcPort(p int) withGrpcPort {
	return withGrpcPort{grpcPort: p}
}

type interceptorConfigure struct {
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

var (
	_ GrpcServerConfigurer = interceptorConfigure{}
)

func (i interceptorConfigure) apply(o *grpcServerConfig) {
	o.unaryInterceptors = i.unaryInterceptors
	o.streamInterceptors = i.streamInterceptors
}

func WithInterceptor(unary []grpc.UnaryServerInterceptor, stream []grpc.StreamServerInterceptor) interceptorConfigure {
	return interceptorConfigure{
		unaryInterceptors:  unary,
		streamInterceptors: stream,
	}
}

type grpcServerOption struct {
	serverOpts []grpc.ServerOption
}

var (
	_ GrpcServerConfigurer = grpcServerOption{}
)

func (g grpcServerOption) apply(o *grpcServerConfig) {
	o.serverOpts = g.serverOpts
}

func WithGrpcServerOption(serverOpts ...grpc.ServerOption) grpcServerOption {
	return grpcServerOption{serverOpts: serverOpts}
}

type serverLogger struct {
	logger pkglogger.Logger
}

var (
	_ GrpcServerConfigurer = serverLogger{}
)

func (g serverLogger) apply(o *grpcServerConfig) {
	o.logger = g.logger
}

func WithLogger(logger pkglogger.Logger) serverLogger {
	return serverLogger{logger: logger}
}

func newConfig(cfgs ...GrpcServerConfigurer) *grpcServerConfig {
	o := &grpcServerConfig{}

	for _, cfg := range cfgs {
		cfg.apply(o)
	}
	return o
}

func NewGrpcServer(cfgs ...GrpcServerConfigurer) *GrpcServer {
	config := newConfig(cfgs...)

	beforeMiddlewares := servermiddleware.MultiServerMiddleware(
		[]servermiddleware.ServerMiddleware{
			loggermiddleware.NewTagMiddlware(),
			metricmiddleware.NewMetricMiddleware(),
		})
	afterMiddlewares := servermiddleware.MultiServerMiddleware(
		[]servermiddleware.ServerMiddleware{
			loggermiddleware.NewLoggerMiddleware(config.logger),
			recoverymiddleware.NewPanicRecoveryMiddleware(),
		})
	unaryInterceptorsSlice := appendMulti(
		beforeMiddlewares.UnaryServerInterceptor(),
		config.unaryInterceptors,
		afterMiddlewares.UnaryServerInterceptor())
	streamInterceptorsSlice := appendMulti(
		beforeMiddlewares.StreamServerInterceptor(),
		config.streamInterceptors,
		afterMiddlewares.StreamServerInterceptor())
	var svrOpts []grpc.ServerOption = append(config.serverOpts,
		grpc.ChainUnaryInterceptor(unaryInterceptorsSlice...),
		grpc.ChainStreamInterceptor(streamInterceptorsSlice...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	grpcServer := grpc.NewServer(svrOpts...)

	if *withReflection {
		// Register reflection service on gRPC server.
		reflection.Register(grpcServer)
	}
	return &GrpcServer{
		logger: config.logger,
		Server: grpcServer,
		config: config,
	}
}

func appendMulti[T any](slices ...[]T) []T {
	var result []T
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

type GrpcServer struct {
	config *grpcServerConfig
	logger pkglogger.Logger
	*grpc.Server

	// listenerAddr was setted when the grpc server is up and running
	listenerAddr net.Addr
}

func (g *GrpcServer) ListenAndServe() error {
	var lc net.ListenConfig
	grpcAddr := fmt.Sprintf(":%d", g.config.grpcPort) // dial any port
	li, err := lc.Listen(context.Background(), "tcp", grpcAddr)
	if err != nil {
		return err
	}
	g.listenerAddr = li.Addr()
	g.logger.Info("grpc: listen and serve %s\n", li.Addr().String())
	return g.Server.Serve(li)
}

func (g *GrpcServer) LocalAddr() (string, error) {
	var addr string
	var err error
	if g.listenerAddr != nil {
		addr = g.listenerAddr.String()
	} else if g.config.grpcPort > 0 {
		addr = fmt.Sprintf("127.0.0.1:%d", g.config.grpcPort)
	} else {
		err = errors.New("localaddr is not ready")
	}
	g.logger.Debug("grpc local addr:%s\n", addr)
	return addr, err
}

func (g *GrpcServer) GracefulStop() {
	g.Server.GracefulStop()
}
