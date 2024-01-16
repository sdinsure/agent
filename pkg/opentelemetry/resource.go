package otel

import (
	"context"

	"go.opentelemetry.io/contrib/detectors/aws/ec2"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type ServiceName interface {
	Name() string
}

func NewResource(ctx context.Context, serviceName ServiceName, keyValues ...attribute.KeyValue) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			append([]attribute.KeyValue{semconv.ServiceNameKey.String(serviceName.Name())}, keyValues...)...,
		),
		resource.WithDetectors(
			ec2.NewResourceDetector(),
			ecs.NewResourceDetector(),
			//eks.NewResourceDetector(),
			gcp.NewDetector(),
		),
	)
	return res, err
}
