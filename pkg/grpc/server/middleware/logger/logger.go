package logger

import (
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"

	"github.com/sdinsure/agent/pkg/logger"
)

func NewLoggerMiddleware(l logger.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		l: l,
	}
}

type LoggerMiddleware struct {
	l logger.Logger
}

func (l *LoggerMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_zap.UnaryServerInterceptor(logger.GetUnderlyingZapLoggerOrDie(l.l))
}

func (l *LoggerMiddleware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_zap.StreamServerInterceptor(logger.GetUnderlyingZapLoggerOrDie(l.l))
}

func NewTagMiddlware() *TagMiddlware {
	return &TagMiddlware{}
}

type TagMiddlware struct{}

func (t *TagMiddlware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_ctxtags.UnaryServerInterceptor(tagMiddlewareOption())
}

func (t *TagMiddlware) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_ctxtags.StreamServerInterceptor(tagMiddlewareOption())
}

func tagMiddlewareOption() grpc_ctxtags.Option {
	return grpc_ctxtags.WithFieldExtractor(
		grpc_ctxtags.CodeGenRequestFieldExtractor,
	)

}
