package metrics

import (
	"context"
	"log"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	ubergoatomic "go.uber.org/atomic"
)

type Counter interface {
	Inc(ctx context.Context, labels ...string)
}

type Gauge interface {
	Set(ctx context.Context, v float64, labels ...string)
}

type Histogram interface {
	Observe(ctx context.Context, v time.Duration, labels ...string)
}

type ValueHistogram interface {
	Observe(ctx context.Context, v int64, labels ...string)
}

// Option applies a configuration option value to a MeterProvider.
type Option interface {
	apply(config) config
}

// optionFunc applies a set of options to a config.
type optionFunc func(config) config

// apply returns a config with option(s) applied.
func (o optionFunc) apply(conf config) config {
	return o(conf)
}

type config struct {
	meterName string
}

func WithMeterName(meterName string) Option {
	return withMeterName{meterName}
}

type withMeterName struct {
	meterName string
}

func (w withMeterName) apply(c config) config {
	c.meterName = w.meterName
	return c
}

type Client struct {
	defaultExporter      *prometheus.Exporter
	defaultMeterProvider *metric.MeterProvider
	pkgMeter             otelmetric.Meter
	config               config
}

func NewClient(options ...Option) (*Client, error) {
	config := newConfig(options)
	defaultExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	defaultMeterProvider := metric.NewMeterProvider(metric.WithReader(defaultExporter))
	pkgMeter := defaultMeterProvider.Meter(config.meterName)

	return &Client{
		defaultExporter:      defaultExporter,
		defaultMeterProvider: defaultMeterProvider,
		pkgMeter:             pkgMeter,
		config:               config,
	}, nil
}

func (c *Client) Shutdown(ctx context.Context) error {
	return c.defaultMeterProvider.Shutdown(ctx)
}

func newConfig(options []Option) config {
	c := config{
		meterName: "github.com/sdinsure/agent/pkg/metrics",
	}
	for _, option := range options {
		c = option.apply(c)
	}
	return c

}

var (
	_ Counter = &CounterVec{}
)

type CounterVec struct {
	v          otelmetric.Float64Counter
	labelNames []string
}

func (c *Client) NewCounterVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *CounterVec {
	v, err := c.pkgMeter.Float64Counter(normalizedNames(ns, ss, name))
	fatalIfNotNil(err)

	return &CounterVec{
		v:          v,
		labelNames: labelNames,
	}
}

func fatalIfNotNil(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Inc increments labels.
func (c *CounterVec) Inc(ctx context.Context, labelValues ...string) {
	c.v.Add(ctx, 1, otelmetric.WithAttributes((makeAttributes(c.labelNames, labelValues))...))
}

func makeAttributes(names []string, values []string) []attribute.KeyValue {
	labels := make([]attribute.KeyValue, len(names), len(values))
	for idx, name := range names {
		labels[idx] = attribute.String(name, values[idx])
	}
	return labels
}

var (
	_ Gauge = &GaugeVec{}
)

type GaugeVec struct {
	v            otelmetric.Float64ObservableGauge
	labelNames   []string
	gaugeValue   *ubergoatomic.Float64
	registerOnce *sync.Once
	pkgMeter     otelmetric.Meter
}

func (c *Client) NewGaugeVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *GaugeVec {
	v, err := c.pkgMeter.Float64ObservableGauge(normalizedNames(ns, ss, name))
	fatalIfNotNil(err)

	return &GaugeVec{
		v:            v,
		registerOnce: &sync.Once{},
		labelNames:   labelNames,
		gaugeValue:   ubergoatomic.NewFloat64(0),
		pkgMeter:     c.pkgMeter,
	}
}

// Set sets v to labels.
func (g *GaugeVec) Set(ctx context.Context, v float64, labelValues ...string) {
	g.gaugeValue.Store(v)

	g.registerOnce.Do(func() {
		cb := func(_ context.Context, o otelmetric.Observer) error {
			o.ObserveFloat64(g.v, g.gaugeValue.Load(), otelmetric.WithAttributes((makeAttributes(g.labelNames, labelValues))...))
			return nil
		}
		if _, err := g.pkgMeter.RegisterCallback(cb, g.v); err != nil {
			log.Printf("register callback failed, err:%+v\n", err)
		}
	})
}

var (
	_ Histogram = &TimeDurationHistogramVec{}
)

type TimeDurationHistogramVec struct {
	v          otelmetric.Float64Histogram
	labelNames []string
}

func (c *Client) NewTimeDurationHistogramVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *TimeDurationHistogramVec {

	v, err := c.pkgMeter.Float64Histogram(
		normalizedNames(ns, ss, name),
		otelmetric.WithExplicitBucketBoundaries(
			prom.ExponentialBuckets(32, 2, 10)...,
		),
		otelmetric.WithUnit("millisecond"),
	)
	fatalIfNotNil(err)

	return &TimeDurationHistogramVec{
		v:          v,
		labelNames: labelNames,
	}
}

func (t *TimeDurationHistogramVec) Observe(ctx context.Context, v time.Duration, labelValues ...string) {
	t.v.Record(ctx, float64(v.Microseconds()), otelmetric.WithAttributes((makeAttributes(t.labelNames, labelValues))...))
}

var (
	_ ValueHistogram = &ValueHistogramVec{}
)

type ValueHistogramVec struct {
	v          otelmetric.Float64Histogram
	labelNames []string
}

func (c *Client) NewValueHistogramVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *ValueHistogramVec {
	v, err := c.pkgMeter.Float64Histogram(
		normalizedNames(ns, ss, name),
		otelmetric.WithExplicitBucketBoundaries(
			prom.ExponentialBuckets(2, 2, 10)...,
		),
		otelmetric.WithUnit("millisecond"),
	)
	fatalIfNotNil(err)

	return &ValueHistogramVec{
		v:          v,
		labelNames: labelNames,
	}
}

func (t *ValueHistogramVec) Observe(ctx context.Context, v int64, labelValues ...string) {
	t.v.Record(ctx, float64(v), otelmetric.WithAttributes((makeAttributes(t.labelNames, labelValues))...))
}

var defaultClient *Client
var initOnce sync.Once

func GetDefaultClient() *Client {
	initOnce.Do(func() {
		var err error
		defaultClient, err = NewClient()
		if err != nil {
			panic(err)
		}
	})
	return defaultClient
}

func NewCounterVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *CounterVec {
	return GetDefaultClient().NewCounterVec(ns, ss, name, labelNames...)
}

func NewGaugeVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *GaugeVec {
	return GetDefaultClient().NewGaugeVec(ns, ss, name, labelNames...)
}

func NewTimeDurationHistogramVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *TimeDurationHistogramVec {
	return GetDefaultClient().NewTimeDurationHistogramVec(ns, ss, name, labelNames...)
}

func NewValueHistogramVec(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName, labelNames ...string) *ValueHistogramVec {
	return GetDefaultClient().NewValueHistogramVec(ns, ss, name, labelNames...)
}
