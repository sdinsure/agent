package server

import (
	"context"

	apppb "github.com/sdinsure/agent/example/api/pb"
	otelagent "github.com/sdinsure/agent/pkg/opentelemetry"
)

type HelloServiceService struct {
	apppb.UnimplementedHelloServiceServer
}

func (h *HelloServiceService) SayHello(_ context.Context, req *apppb.HelloRequest) (*apppb.HelloResponse, error) {
	return &apppb.HelloResponse{
		Reply: "reply from " + req.Greeting,
	}, nil
}

var (
	_ otelagent.ServiceName = &HelloServiceService{}
)

// Name implements otelagent.ServiceName interface
func (*HelloServiceService) Name() string {
	return "hello-service-service"
}
