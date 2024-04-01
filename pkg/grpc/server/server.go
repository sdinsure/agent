package server

import (
	"context"
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

type option struct {
	grpcPort int
}

func newOption(optioners ...Optioner) *option {
	o := &option{}

	for _, optioner := range optioners {
		optioner.apply(o)
	}

	return o
}

type Optioner interface {
	apply(o *option)
}

type withGrpcPort struct {
	grpcPort int
}

func (w withGrpcPort) apply(o *option) {
	o.grpcPort = w.grpcPort
}

func WithGrpcPort(p int) withGrpcPort {
	return withGrpcPort{grpcPort: p}
}

func NewGrpcServer(logger pkglogger.Logger, optioners ...Optioner) *GrpcServer {
	return NewGrpcServerWithInterceptors(logger, nil, nil, nil, optioners...)
}

func NewGrpcServerWithInterceptors(
	logger pkglogger.Logger,
	serverOpts []grpc.ServerOption,
	unaryInterceptors []grpc.UnaryServerInterceptor,
	streamInterceptors []grpc.StreamServerInterceptor,
	optioners ...Optioner,
) *GrpcServer {
	opt := newOption(optioners...)

	beforeMiddlewares := servermiddleware.MultiServerMiddleware(
		[]servermiddleware.ServerMiddleware{
			loggermiddleware.NewTagMiddlware(),
			metricmiddleware.NewMetricMiddleware(),
		})
	afterMiddlewares := servermiddleware.MultiServerMiddleware(
		[]servermiddleware.ServerMiddleware{
			loggermiddleware.NewLoggerMiddleware(logger),
			recoverymiddleware.NewPanicRecoveryMiddleware(),
		})
	unaryInterceptorsSlice := appendMulti(
		beforeMiddlewares.UnaryServerInterceptor(),
		unaryInterceptors,
		afterMiddlewares.UnaryServerInterceptor())
	streamInterceptorsSlice := appendMulti(
		beforeMiddlewares.StreamServerInterceptor(),
		streamInterceptors,
		afterMiddlewares.StreamServerInterceptor())
	var opts []grpc.ServerOption = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptorsSlice...),
		grpc.ChainStreamInterceptor(streamInterceptorsSlice...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}
	opts = append(opts, serverOpts...)
	grpcServer := grpc.NewServer(opts...)

	if *withReflection {
		// Register reflection service on gRPC server.
		reflection.Register(grpcServer)
	}
	return &GrpcServer{
		logger: logger,
		Server: grpcServer,
		opt:    opt,
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
	opt    *option
	logger pkglogger.Logger
	*grpc.Server

	// listenerAddr was setted when the grpc server is up and running
	listenerAddr net.Addr
}

func (g *GrpcServer) ListenAndServe() error {
	var lc net.ListenConfig
	grpcAddr := fmt.Sprintf(":%d", g.opt.grpcPort) // dial any port
	li, err := lc.Listen(context.Background(), "tcp", grpcAddr)
	if err != nil {
		return err
	}
	g.listenerAddr = li.Addr()
	g.logger.Info("grpc: listen and serve %s\n", li.Addr().String())
	return g.Server.Serve(li)
}

func (g *GrpcServer) LocalAddr() string {
	var addr string
	if g.listenerAddr != nil {
		addr = g.listenerAddr.String()
	}
	g.logger.Info("grpc local addr:%s\n", addr)
	return addr
}

func (g *GrpcServer) GracefulStop() {
	g.Server.GracefulStop()
}
