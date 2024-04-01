package main

import (
	"context"
	"flag"

	"go.opentelemetry.io/otel/attribute"

	appserver "github.com/sdinsure/agent/example/server"
	logger "github.com/sdinsure/agent/pkg/logger"
	otel "github.com/sdinsure/agent/pkg/opentelemetry"
	otelmetrics "github.com/sdinsure/agent/pkg/opentelemetry/metrics"
	oteltraces "github.com/sdinsure/agent/pkg/opentelemetry/traces"
	"github.com/sdinsure/agent/pkg/version"
)

var httpPort = flag.Int("http_port", 50091, "The server http port")
var otelExportAddr = flag.String("otel_export", "localhost:4317", "otel export grpc address")

func main() {
	flag.Parse()

	version.Print()
	log := logger.NewLogger()

	app := &appserver.HelloServiceService{}

	ctx := context.Background()
	conn, err := otel.NewGRPCConn(ctx, *otelExportAddr)
	if err != nil {
		log.Fatal("init grpc conn failed, err:+%v", err)
	}
	res, err := otel.NewResource(ctx, app, attribute.String("version", version.GetVersion()))
	if err != nil {
		log.Fatal("init otel res failed, err:+%v", err)
	}
	otelMetricService, err := otelmetrics.NewOTELMetrics(
		otel.WithContext(ctx),
		otel.WithGRPCConn(conn),
		otel.WithResource(res),
		//otel.WithStdout(),
	)
	if err != nil {
		log.Fatal("init otel metrics failed, err:+%v", err)
	}
	otelTraceService, err := oteltraces.NewOTELTraces(
		otel.WithContext(ctx),
		otel.WithGRPCConn(conn),
		otel.WithResource(res),
		otel.WithStdout(),
	)
	if err != nil {
		log.Fatal("init otel traces failed, err:+%v", err)
	}
	defer func() {
		otel.ShutdownAll(ctx, otelMetricService, otelTraceService)
	}()

	svr, err := appserver.NewServerService(*httpPort, log)
	if err != nil {
		log.Fatal("failed to start server, err:%+v\n", err)
	}
	svr.AddGatewayRoutes()
	appserver.RegisterService(svr, app)
	svr.Start()
}
