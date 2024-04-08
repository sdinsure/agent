package identitymiddleware

import (
	"context"

	"github.com/sdinsure/agent/pkg/grpc/server/middleware"
	sdinsureruntime "github.com/sdinsure/agent/pkg/runtime"
	"google.golang.org/grpc"
)

var (
	_ middleware.ServerMiddleware = &RequestIdentityMiddleware{}
)

func NewRequestIdentityMiddleware() *RequestIdentityMiddleware {
	return &RequestIdentityMiddleware{
		uuidResolver: &sdinsureruntime.UUIDRequestIdentityResolver{},
	}
}

type RequestIdentityMiddleware struct {
	uuidResolver sdinsureruntime.RequestIdentityResolver
}

func (r *RequestIdentityMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(r.uuidResolver.WithRequestID(ctx), req)
	}
}

func (r *RequestIdentityMiddleware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrappedStream := &middleware.ServerStreamWrapper{
			Ctx:          r.uuidResolver.WithRequestID(stream.Context()),
			ServerStream: stream,
		}

		return handler(srv, wrappedStream)
	}
}
