package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Limiter interface {
	// Limit return the expected duration that client can make another call
	// 0 stands the request can be proceed
	Limit(rpcFullMethod string, req interface{}) error
}

// UnaryServerInterceptor returns a new unary server interceptors that performs request rate limiting.
func UnaryServerInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := limiter.Limit(info.FullMethod, req); err != nil {
			return nil, status.Errorf(codes.ResourceExhausted, "%s too many requests, please retry later. details: %s", info.FullMethod, err.Error())
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that performs rate limiting on the request.
func StreamServerInterceptor(limiter Limiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := limiter.Limit(info.FullMethod, nil); err != nil {
			return status.Errorf(codes.ResourceExhausted, "%s too many requests, please retry later. details: %s", info.FullMethod, err.Error())
		}
		return handler(srv, stream)
	}
}

type NoLimiter struct{}

func (n NoLimiter) Limit(rpcFullMethod string) error {
	return nil
}
