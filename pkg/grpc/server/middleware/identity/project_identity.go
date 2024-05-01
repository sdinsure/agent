package identitymiddleware

import (
	"context"

	"github.com/sdinsure/agent/pkg/grpc/server/middleware"
	sdinsuregrpcserverruntime "github.com/sdinsure/agent/pkg/grpc/server/runtime"
	sdinsureruntime "github.com/sdinsure/agent/pkg/runtime"
	"google.golang.org/grpc"
)

var (
	_ middleware.ServerMiddleware = &ProjectIdentityMiddleware{}
)

func NewProjectIdentityMiddleware(p sdinsureruntime.ProjectResolver) *ProjectIdentityMiddleware {
	return &ProjectIdentityMiddleware{pr: p}
}

type ProjectIdentityMiddleware struct {
	pr sdinsureruntime.ProjectResolver
}

func (r *ProjectIdentityMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		httpPath, _ := sdinsuregrpcserverruntime.HttpPath(ctx)
		return handler(r.pr.WithProjectInfo(ctx, httpPath), req)
	}
}

func (r *ProjectIdentityMiddleware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		httpPath, _ := sdinsuregrpcserverruntime.HttpPath(stream.Context())
		wrappedStream := &middleware.ServerStreamWrapper{
			Ctx:          r.pr.WithProjectInfo(stream.Context(), httpPath),
			ServerStream: stream,
		}

		return handler(srv, wrappedStream)
	}
}
