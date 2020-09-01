package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Port metrics are served on
var Port string

var (
	// Gauges available
	Gauges = make(map[string]*prometheus.GaugeVec)

	// Counters available
	Counters = make(map[string]*prometheus.CounterVec)
)

// StartMetrics initializes and handles metrics Prometheus endpoint
func StartMetrics() {
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{},
		))
		if err := http.ListenAndServe("0.0.0.0:"+Port, nil); err != nil {
			log.Fatal(err)
		}
	}()
}
