package http

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	logger "github.com/sdinsure/agent/pkg/logger"
)

func NewHttpClient(options ...Option) *http.Client {
	return &http.Client{
		Transport: NewHttpTransport(http.DefaultTransport, options...),
	}
}

type config struct {
	Debug bool
	Log   logger.Logger

	HttpTrace bool
}

type Option interface {
	apply(config) config
}

func WithDebug(debug bool, log logger.Logger) Option {
	return withDebug{debug: debug, log: log}
}

type withDebug struct {
	debug bool
	log   logger.Logger
}

func (w withDebug) apply(cfg config) config {
	cfg.Debug = w.debug
	cfg.Log = w.log
	return cfg
}

func WithHttpTrace(httpTrace bool) Option {
	return withHttpTrace{httpTrace}
}

type withHttpTrace struct {
	httpTrace bool
}

func (w withHttpTrace) apply(cfg config) config {
	cfg.HttpTrace = w.httpTrace
	return cfg
}

func newConfig(options ...Option) (config, error) {
	cfg := config{
		HttpTrace: true,
	}
	for _, opt := range options {
		cfg = opt.apply(cfg)
	}
	return cfg, nil
}

func NewHttpTransport(rt http.RoundTripper, options ...Option) http.RoundTripper {
	cfg, _ := newConfig(options...)
	if cfg.HttpTrace {
		rt = otelhttp.NewTransport(
			rt,
			otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
				return otelhttptrace.NewClientTrace(ctx)
			}),
		)
	}
	if !cfg.Debug {
		return rt
	}
	return debugTransport{t: rt, log: cfg.Log}
}

type debugTransport struct {
	t   http.RoundTripper
	log logger.Logger
}

func (d debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return nil, err
	}
	d.log.Infox(req.Context(), "<---req: %s", reqDump)

	resp, err := d.t.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respDump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	d.log.Infox(req.Context(), "--->resp: %s", respDump)
	return resp, nil
}
