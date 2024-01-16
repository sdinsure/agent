package metricmiddleware

import (
	"context"
	"strconv"

	"github.com/sdinsure/agent/pkg/metrics"
	"github.com/zeromicro/go-zero/core/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var serverNamespace = metrics.NewTypeNamespace("rpc_server")
var subsystem = metrics.NewTypeSubsystem("requests")

var (
	metricServerReqDur = metrics.NewTimeDurationHistogramVec(
		serverNamespace,
		subsystem,
		metrics.NewTypeMetricName("duration_ms"),
		"method",
	)
	metricServerReqCodeTotal = metrics.NewCounterVec(
		serverNamespace,
		subsystem,
		metrics.NewTypeMetricName("code_total"),
		"method", "rpccode",
	)
)

func NewMetricMiddleware() *MetricMiddleware {
	return &MetricMiddleware{}
}

type MetricMiddleware struct {
}

func (m *MetricMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := timex.Now()
		resp, err := handler(ctx, req)
		metricServerReqDur.Observe(ctx, timex.Since(startTime), info.FullMethod)
		metricServerReqCodeTotal.Inc(ctx, info.FullMethod, strconv.Itoa(int(status.Code(err))))
		return resp, err
	}
}

func (m *MetricMiddleware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// no request duration for streaming
		err := handler(srv, stream)
		metricServerReqCodeTotal.Inc(context.Background(), info.FullMethod, strconv.Itoa(int(status.Code(err))))
		return err
	}
}
