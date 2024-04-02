package client

import (
	"context"
	"sync"

	"github.com/sdinsure/agent/pkg/logger"
	"google.golang.org/grpc"
)

type LocalAddrResolver interface {
	LocalAddr() string
}

func NewGrpcLazyClinet(log logger.Logger, addrResolver LocalAddrResolver, dialOpts ...grpc.DialOption) (*GrpcLazyClient, error) {

	lazyClient := &GrpcLazyClient{
		log:          log,
		addrResolver: addrResolver,
		dialOpts:     dialOpts,
	}
	return lazyClient, nil
}

var (
	_ grpc.ClientConnInterface = &GrpcLazyClient{}
)

type GrpcLazyClient struct {
	log          logger.Logger
	addrResolver LocalAddrResolver
	dialOpts     []grpc.DialOption

	once               sync.Once
	resolvedServerAddr string
	clientConn         *grpc.ClientConn
	dialedErr          error
}

func (l *GrpcLazyClient) dialOnce() error {
	var err error
	l.once.Do(func() {
		l.resolvedServerAddr = l.addrResolver.LocalAddr()
		l.clientConn, l.dialedErr = grpc.Dial(l.resolvedServerAddr, l.dialOpts...)
		l.log.Info("grpc.lazy: resolved addr:%s\n", l.resolvedServerAddr)
		if l.dialedErr != nil {
			l.log.Error("grpc.lazy: dialerr:%+v\n", l.dialedErr)
		}
	})
	return l.dialedErr
}

func (l *GrpcLazyClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	if err := l.dialOnce(); err != nil {
		return err
	}

	return l.clientConn.Invoke(ctx, method, args, reply, opts...)
}

func (l *GrpcLazyClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if err := l.dialOnce(); err != nil {
		return err
	}

	return l.clientConn.NewStream(ctx, desc, method, opts...)
}

func (l *GrpcLazyClient) Close() error {
	if l.clientConn == nil {
		return nil
	}
	return l.clientConn.Close()
}
