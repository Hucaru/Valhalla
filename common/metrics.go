package common

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsPort metrics are served on
var MetricsPort string

var (
	// MetricsGauges available
	MetricsGauges = make(map[string]*prometheus.GaugeVec)

	// MetricsCounters available
	MetricsCounters = make(map[string]*prometheus.CounterVec)
)

// StartMetrics initializes and handles metrics Prometheus endpoint
func StartMetrics() {
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{},
		))
		if err := http.ListenAndServe("0.0.0.0:"+MetricsPort, nil); err != nil {
			log.Fatal(err)
		}
	}()
}
