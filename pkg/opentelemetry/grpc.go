package otel

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCConn(ctx context.Context, hostAddr string) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(
		ctx,
		hostAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	return conn, err
}
