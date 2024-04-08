package identitymiddleware

import (
	"context"

	"github.com/sdinsure/agent/pkg/grpc/server/middleware"
	sdinsureruntime "github.com/sdinsure/agent/pkg/runtime"
	"google.golang.org/grpc"
)

var (
	_ middleware.ServerMiddleware = &UserIdentityMiddleware{}
)

func NewUserIdentityMiddleware(ur sdinsureruntime.UserResolver) *UserIdentityMiddleware {
	return &UserIdentityMiddleware{ur: ur}
}

type UserIdentityMiddleware struct {
	ur sdinsureruntime.UserResolver
}

func (r *UserIdentityMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(r.ur.WithUserInfo(ctx), req)
	}
}

func (r *UserIdentityMiddleware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrappedStream := &middleware.ServerStreamWrapper{
			Ctx:          r.ur.WithUserInfo(stream.Context()),
			ServerStream: stream,
		}

		return handler(srv, wrappedStream)
	}
}
