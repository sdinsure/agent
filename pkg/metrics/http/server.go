package httpsvr

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricPort = flag.Int("metric_port", 50999, "port for metrics")
)

func init() {
	m := &metricSvr{port: *metricPort}
	m.ListenAndServe()
}

type metricSvr struct {
	port int
}

func (m *metricSvr) ListenAndServe() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d".m.port), nil)
	if err != nil {
		log.Fatalf(err)
	}
}
