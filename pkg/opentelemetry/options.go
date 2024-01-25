package otel

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
)

type OTELOption struct {
	ctx      context.Context
	conn     *grpc.ClientConn
	resource *resource.Resource

	stdout bool // then stdout=true, we use stdout for tracing/metrics
}

func (o *OTELOption) Validate() error {
	return nil
}

func (o *OTELOption) Conn() *grpc.ClientConn {
	return o.conn
}

func (o *OTELOption) Context() context.Context {
	return o.ctx
}

func (o *OTELOption) Resource() *resource.Resource {
	return o.resource
}

func (o *OTELOption) Stdout() bool {
	return o.stdout
}

func DefaultOption() *OTELOption {
	return &OTELOption{}
}

type OTELOptionFunc func(o *OTELOption)

func WithGRPCConn(conn *grpc.ClientConn) func(o *OTELOption) {
	return func(o *OTELOption) {
		o.conn = conn
	}
}

func WithContext(ctx context.Context) func(o *OTELOption) {
	return func(o *OTELOption) {
		o.ctx = ctx
	}
}

func WithResource(res *resource.Resource) func(o *OTELOption) {
	return func(o *OTELOption) {
		o.resource = res
	}
}

func WithStdout() func(o *OTELOption) {
	return func(o *OTELOption) {
		o.stdout = true
	}
}
