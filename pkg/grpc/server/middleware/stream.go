package middleware

import (
	"context"

	"google.golang.org/grpc"
)

type ServerStreamWrapper struct {
	grpc.ServerStream

	Ctx context.Context
}

var (
	_ grpc.ServerStream = &ServerStreamWrapper{}
)

func (s *ServerStreamWrapper) Context() context.Context {
	return s.Ctx
}
