package metrics

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Port metrics are served on
const Port = 9000

var (
	// Gauges available
	Gauges = make(map[string]*prometheus.GaugeVec)

	// Counters available
	Counters = make(map[string]*prometheus.CounterVec)
)

func init() {
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{},
		))
		log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(Port), nil))
	}()
}
