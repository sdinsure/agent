package oteltraces

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	otelagent "github.com/sdinsure/agent/pkg/opentelemetry"
)

type OTELTraces struct {
	shutdownFunc func(context.Context) error
}

func NewOTELTraces(optionFuncs ...otelagent.OTELOptionFunc) (*OTELTraces, error) {
	options := otelagent.DefaultOption()

	for _, optFunc := range optionFuncs {
		optFunc(options)
	}

	if err := options.Validate(); err != nil {
		return nil, err
	}
	traceExporter, err := otlptracegrpc.New(
		options.Context(),
		otlptracegrpc.WithGRPCConn(options.Conn()),
	)
	if err != nil {
		return nil, err
	}
	traceProviderOptions := []trace.TracerProviderOption{
		trace.WithResource(options.Resource()),
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second)),
	}

	if options.Stdout() {
		stdoutTraceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		traceProviderOptions = append(traceProviderOptions,
			trace.WithBatcher(stdoutTraceExporter,
				trace.WithBatchTimeout(5*time.Second)))
	}
	tracerProvider := trace.NewTracerProvider(traceProviderOptions...)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return &OTELTraces{shutdownFunc: tracerProvider.Shutdown}, nil
}

func (o *OTELTraces) Shutdown(ctx context.Context) error {
	return o.shutdownFunc(ctx)
}
