package middleware

import (
	"log"
	"runtime/debug"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func NewPanicRecoveryMiddleware() *PanicRecoveryMiddleware {
	return &PanicRecoveryMiddleware{}
}

type PanicRecoveryMiddleware struct{}

func (p *PanicRecoveryMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_recovery.UnaryServerInterceptor(recoveryOption())
}

func (p *PanicRecoveryMiddleware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_recovery.StreamServerInterceptor(recoveryOption())
}

func recoveryOption() grpc_recovery.Option {
	return grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		log.Printf("panic triggered: %v", p)
		debug.PrintStack()
		return grpc.Errorf(codes.Unknown, "panic triggered: %v", p)
	})
}
