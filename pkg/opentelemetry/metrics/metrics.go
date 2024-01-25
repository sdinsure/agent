package otelmetrics

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"

	otelagent "github.com/sdinsure/agent/pkg/opentelemetry"
)

type OTELMetrics struct {
	shutdownFunc func(context.Context) error
}

func NewOTELMetrics(optionFuncs ...otelagent.OTELOptionFunc) (*OTELMetrics, error) {
	options := otelagent.DefaultOption()

	for _, optFunc := range optionFuncs {
		optFunc(options)
	}

	if err := options.Validate(); err != nil {
		return nil, err
	}
	metricExporter, err := otlpmetricgrpc.New(
		options.Context(),
		otlpmetricgrpc.WithGRPCConn(options.Conn()),
	)
	if err != nil {
		return nil, err
	}

	meterProviderOptions := []metric.Option{
		metric.WithResource(options.Resource()),
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(15*time.Second))),
	}

	// Wrap the raw grpc connection to OTEL collector with an exporter.
	if options.Stdout() {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		stdoutExporter, err := stdoutmetric.New(
			stdoutmetric.WithEncoder(enc),
			stdoutmetric.WithoutTimestamps(),
		)
		if err != nil {
			return nil, err
		}
		meterProviderOptions = append(meterProviderOptions, metric.WithReader(metric.NewPeriodicReader(stdoutExporter, metric.WithInterval(15*time.Second))))
	}
	meterProvider := metric.NewMeterProvider(meterProviderOptions...)

	// Register controller as global meter provider.
	otel.SetMeterProvider(meterProvider)

	return &OTELMetrics{shutdownFunc: meterProvider.Shutdown}, nil
}

func (o *OTELMetrics) Shutdown(ctx context.Context) error {
	return o.shutdownFunc(ctx)
}
