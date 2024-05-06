package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/felixge/httpsnoop"
	pkgruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	grpcserver "github.com/sdinsure/agent/pkg/grpc/server"
	"github.com/sdinsure/agent/pkg/grpc/server/runtime"
	"github.com/sdinsure/agent/pkg/logger"
)

type HttpMiddlewareHandler func(h http.Handler) http.Handler

type HttpGatewayServerConfigurer interface {
	apply(*HTTPGatewayServerConfig)
}

type HTTPGatewayServerConfig struct {
	middlewares  []HttpMiddlewareHandler
	serveMuxOpts []pkgruntime.ServeMuxOption

	clientTransportCredentials credentials.TransportCredentials
	log                        logger.Logger
}

func newConfig(log logger.Logger, configer ...HttpGatewayServerConfigurer) *HTTPGatewayServerConfig {
	sc := &HTTPGatewayServerConfig{
		log: log,
		middlewares: []HttpMiddlewareHandler{
			withLoggerWrapper(log),
			cors,
		},
		serveMuxOpts: []pkgruntime.ServeMuxOption{
			pkgruntime.WithRoutingErrorHandler(handleRoutingError),
			pkgruntime.WithForwardResponseOption(responseHeaderMatcher),
			pkgruntime.WithIncomingHeaderMatcher(buildinHttpIncomingHeaderMatcher),
			pkgruntime.WithOutgoingHeaderMatcher(buildinHttpOutgoingHeaderMatcher),
			pkgruntime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
				return runtime.ForwardHttpToMetadata(ctx, r)
			}),
		},
		clientTransportCredentials: insecure.NewCredentials(),
	}
	for _, config := range configer {
		config.apply(sc)
	}
	return sc
}

type withLog struct {
	log logger.Logger
}

func (w withLog) apply(c *HTTPGatewayServerConfig) {
	c.log = w.log
}

func WithLogger(log logger.Logger) withLog {
	return withLog{log: log}
}

type httpMiddlewares struct {
	middlewares []HttpMiddlewareHandler
}

func (h httpMiddlewares) apply(c *HTTPGatewayServerConfig) {
	c.middlewares = append(c.middlewares, h.middlewares...)
}

type HTTPGatewayServerConfigOption func(c *HTTPGatewayServerConfig)

func WithHttpMiddlewares(middlewares ...HttpMiddlewareHandler) httpMiddlewares {
	return httpMiddlewares{
		middlewares: middlewares,
	}
}

type serveMuxOption struct {
	serveMuxOpts []pkgruntime.ServeMuxOption
}

func (s serveMuxOption) apply(c *HTTPGatewayServerConfig) {
	c.serveMuxOpts = append(c.serveMuxOpts, s.serveMuxOpts...)
}

func WithServeMuxOption(serveMuxOpts ...pkgruntime.ServeMuxOption) serveMuxOption {
	return serveMuxOption{
		serveMuxOpts: serveMuxOpts,
	}
}

type clientTransportCredentials struct {
	clientTransportCredentials credentials.TransportCredentials
}

func (cc clientTransportCredentials) apply(c *HTTPGatewayServerConfig) {
	c.clientTransportCredentials = cc.clientTransportCredentials
}

func WithTransportCredentials(tc credentials.TransportCredentials) clientTransportCredentials {
	return clientTransportCredentials{
		clientTransportCredentials: tc,
	}
}

func NewHTTPGatewayServer(g *grpcserver.GrpcServer, log logger.Logger, port int, configurers ...HttpGatewayServerConfigurer) (*HTTPGatewayServer, error) {

	sc := newConfig(log, configurers...)

	opts := []grpc.DialOption{
		//grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10 * 1024 * 1024 /*10M for max receive size*/)),
		grpc.WithTransportCredentials(sc.clientTransportCredentials),
	}
	addr, err := g.LocalAddr()
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	serveMux := pkgruntime.NewServeMux(sc.serveMuxOpts...)
	httpMux := http.NewServeMux()
	// register all routes under root
	httpMux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(chainMiddleware(serveMux, sc.middlewares...).ServeHTTP), "otelhandler"))

	return &HTTPGatewayServer{
		log:        log,
		port:       port,
		grpcConn:   conn,
		grpcServer: g,
		ctx:        context.Background(),
		serveMux:   serveMux,
		httpMux:    httpMux,
		httpServer: &http.Server{Handler: httpMux},
	}, nil
}

func chainMiddleware(h http.Handler, m ...HttpMiddlewareHandler) http.Handler {
	if len(m) < 1 {
		return h
	}
	wrapped := h
	// loop in reverse to preserve middleware order
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}
	return wrapped
}

// handleRoutingError handles grpc.status code with http code
func handleRoutingError(ctx context.Context, mux *pkgruntime.ServeMux, marshaler pkgruntime.Marshaler, w http.ResponseWriter, r *http.Request, httpStatus int) {
	pkgruntime.DefaultRoutingErrorHandler(ctx, mux, marshaler, w, r, httpStatus)
}

// responseHeaderMatcher capture Location header in grpc.context and convert it to http 301 Redirection
func responseHeaderMatcher(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	headers := w.Header()
	if location, ok := headers["Grpc-Metadata-Location"]; ok {
		w.Header().Set("Location", location[0])
		w.WriteHeader(http.StatusFound)
	}

	return nil
}

// buildinHttpIncomingHeaderMatcher converts incoming http readers to grpc metadata
func buildinHttpIncomingHeaderMatcher(header string) (string, bool) {
	return pkgruntime.DefaultHeaderMatcher(header)
}

// buildinHttpOutgoingHeaderMatcher allows headers from grpc to http header
func buildinHttpOutgoingHeaderMatcher(header string) (string, bool) {
	var (
		allowedHeaders = map[string]struct{}{
			"x-request-id": {},
		}
	)
	if _, isAllowed := allowedHeaders[header]; isAllowed {
		return strings.ToUpper(header), true
	}
	return header, false
}

// withLogger add logs for each http reqeusts
func withLoggerWrapper(log logger.Logger) HttpMiddlewareHandler {
	return HttpMiddlewareHandler(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			logWriter := &logResponseWriter{actualWriter: writer}
			r1, r2 := cloneRequest(request)
			m := httpsnoop.CaptureMetrics(handler, logWriter, r1)
			// printing exracted data
			log.Info("http[%d]-- %s -- %s\n", m.Code, m.Duration, request.URL.Path)
			if m.Code != 200 {
				log.Info("(conti) body: %s\n", string(logWriter.log.Bytes()))
				fullBody, _ := io.ReadAll(r2.Body)
				log.Info("(conti) request body:%s\n", string(fullBody))
			}
		})
	})
}

func cloneRequest(r *http.Request) (r1 *http.Request, r2 *http.Request) {
	r1 = r

	body, err := io.ReadAll(r1.Body)
	if err != nil {
		// ...
	}
	r2 = r1.Clone(r1.Context())
	// clone body
	r1.Body = io.NopCloser(bytes.NewReader(body))
	r2.Body = io.NopCloser(bytes.NewReader(body))

	return
}

type logResponseWriter struct {
	log          bytes.Buffer
	actualWriter http.ResponseWriter
}

func (l *logResponseWriter) Header() http.Header {
	return l.actualWriter.Header()
}

func (l *logResponseWriter) Write(b []byte) (int, error) {
	mw := io.MultiWriter(l.actualWriter, &l.log)
	return mw.Write(b)
}

func (l *logResponseWriter) WriteHeader(code int) {
	l.actualWriter.WriteHeader(code)
}

func (l *logResponseWriter) Flush() {
	f, ok := l.actualWriter.(http.Flusher)
	if ok {
		f.Flush()
	}
	l.log.Reset()
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigin(r.Header.Get("Origin")) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")
		}
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}

func allowedOrigin(origin string) bool {
	//log.Printf("cors pattern:%s, origin from request:%s\n", viper.GetString("cors"), origin)
	if viper.GetString("cors") == "*" {
		return true
	}
	if matched, _ := regexp.MatchString(viper.GetString("cors"), origin); matched {
		return true
	}
	return false
}

type HTTPGatewayServer struct {
	port       int
	log        logger.Logger
	grpcServer *grpcserver.GrpcServer
	grpcConn   *grpc.ClientConn
	ctx        context.Context
	serveMux   *pkgruntime.ServeMux
	httpMux    *http.ServeMux
	httpServer *http.Server
}

func (h *HTTPGatewayServer) AddRoutes(routes ...*Route) error {
	for _, route := range routes {
		h.log.Info("httpgateway: register path: %s\n", route.Pattern)
		h.httpMux.Handle(route.Pattern, route.Handler)
	}
	return nil
}

type GatewayHandlerFunc func(ctx context.Context, mux *pkgruntime.ServeMux, conn *grpc.ClientConn) error

func (h *HTTPGatewayServer) RegisterHandlers(handlerfuncs ...GatewayHandlerFunc) error {
	for _, handlerfunc := range handlerfuncs {
		if err := handlerfunc(h.ctx, h.serveMux, h.grpcConn); err != nil {
			return err
		}
	}
	return nil
}

func (h *HTTPGatewayServer) ListenAndServe() error {

	go func(svr *grpcserver.GrpcServer) {
		h.log.Info("serving grpc..\n")
		svr.ListenAndServe()
	}(h.grpcServer)

	var lc net.ListenConfig
	addr := fmt.Sprintf(":%d", h.port)
	h.log.Info("httpgateway: listen and serve %s\n", addr)
	li, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return err
	}

	h.httpServer.Serve(li)

	return nil
}

func (h *HTTPGatewayServer) WaitForSIGTERM() error {
	h.log.Info("httpgateway: wait for system interrupt...\n")
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-signalCh

	// handle graceful shutdown for both servers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return h.Shutdown(ctx)
}

func (h *HTTPGatewayServer) Shutdown(ctx context.Context) error {
	if err := h.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	h.grpcServer.GracefulStop()
	return nil
}
