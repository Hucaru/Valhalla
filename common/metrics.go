package common

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

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

	// metricsStarted ensures metrics server is only started once
	metricsStarted sync.Once
	
	// metricsServer stores the HTTP server instance for shutdown
	metricsServer *http.Server
)

// StartMetrics initializes and handles metrics Prometheus endpoint
// This function is safe to call multiple times - it will only start the server once
func StartMetrics() {
	metricsStarted.Do(func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{},
		))
		
		metricsServer = &http.Server{
			Addr:    "0.0.0.0:" + MetricsPort,
			Handler: mux,
		}
		
		go func() {
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	})
}

// StopMetrics gracefully shuts down the metrics server
func StopMetrics() {
	if metricsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := metricsServer.Shutdown(ctx); err != nil {
			log.Println("Metrics server shutdown error:", err)
		}
		metricsServer = nil
	}
}
