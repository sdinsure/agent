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

func NewGrpcServer(port int, logger pkglogger.Logger) *GrpcServer {
	return NewGrpcServerWithInterceptors(port, logger, nil, nil, nil)
}

func NewGrpcServerWithInterceptors(port int, logger pkglogger.Logger, serverOpts []grpc.ServerOption, unaryInterceptors []grpc.UnaryServerInterceptor, streamInterceptors []grpc.StreamServerInterceptor) *GrpcServer {
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
		port:   port,
		logger: logger,
		Server: grpcServer,
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
	port   int
	logger pkglogger.Logger
	*grpc.Server
}

func (g *GrpcServer) ListenAndServe() error {
	var lc net.ListenConfig
	grpcAddr := fmt.Sprintf(":%d", g.port)
	g.logger.Info("grpc: listen and serve %s\n", grpcAddr)
	li, err := lc.Listen(context.Background(), "tcp", grpcAddr)
	if err != nil {
		return err
	}
	return g.Server.Serve(li)
}

func (g *GrpcServer) LocalAddr() string {
	s := fmt.Sprintf("127.0.0.1:%d", g.port)

	g.logger.Info("grpc local addr:%s\n", s)
	return s
}

func (g *GrpcServer) GracefulStop() {
	g.Server.GracefulStop()
}
