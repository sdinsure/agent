package server

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sdinsure/agent/pkg/logger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer(t *testing.T) {
	svr := NewGrpcServer(logger.NewLogger())
	defer func() {
		svr.GracefulStop()
	}()

	go func() {
		if err := svr.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)

	fmt.Printf("%s\n", svr.LocalAddr())
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	conn, err := grpc.DialContext(ctx, svr.LocalAddr(), opts...)
	assert.NoError(t, err)
	assert.NoError(t, conn.Close())
}

func TestServer44138(t *testing.T) {
	svr := NewGrpcServer(logger.NewLogger(), WithGrpcPort(44138))
	defer func() {
		svr.GracefulStop()
	}()

	go func() {
		if err := svr.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)

	assert.True(t, strings.HasSuffix(svr.LocalAddr(), ":44138"))
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	conn, err := grpc.DialContext(ctx, svr.LocalAddr(), opts...)
	assert.NoError(t, err)
	assert.NoError(t, conn.Close())
}
