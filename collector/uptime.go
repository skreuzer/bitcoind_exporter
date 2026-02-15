package collector

import (
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type uptimeCollector struct {
	client    *rpcclient.Client
	uptime    *prometheus.Desc
	logger    log.Logger
	collector string
}

func NewUptimeCollector(rpcClient *rpcclient.Client, logger log.Logger) *uptimeCollector {
	return &uptimeCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "uptime",
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "uptime", "seconds"),
			"The number of seconds that the Bitcoin server has been running",
			nil, nil),
	}
}

func (c *uptimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.uptime
}

func (c *uptimeCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	rawUptime, err := c.client.RawRequest("uptime", nil)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get uptime",
			"err", err,
			"method", "uptime",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	var uptimeSeconds int64
	if err := json.Unmarshal(rawUptime, &uptimeSeconds); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to unmarshal uptime",
			"err", err,
			"method", "json.Unmarshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.uptime, prometheus.GaugeValue, float64(uptimeSeconds))

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
