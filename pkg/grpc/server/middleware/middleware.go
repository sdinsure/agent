package middleware

import "google.golang.org/grpc"

type ServerMiddleware interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
	StreamServerInterceptor() grpc.StreamServerInterceptor
}

type MultiServerMiddleware []ServerMiddleware

func (m MultiServerMiddleware) UnaryServerInterceptor() []grpc.UnaryServerInterceptor {
	var interceptors []grpc.UnaryServerInterceptor
	for _, middleware := range m {
		interceptors = append(interceptors, middleware.UnaryServerInterceptor())
	}
	return interceptors
}

func (m MultiServerMiddleware) StreamServerInterceptor() []grpc.StreamServerInterceptor {
	var interceptors []grpc.StreamServerInterceptor
	for _, middleware := range m {
		interceptors = append(interceptors, middleware.StreamServerInterceptor())
	}
	return interceptors
}
