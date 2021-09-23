package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "bitcoind"
)

var (
	collectDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil)
	collectError = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collect_error"),
		"Error occurred during collection",
		[]string{"collector"}, nil)
)
