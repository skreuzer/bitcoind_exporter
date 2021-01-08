package collector

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type networkCollector struct {
	client          *rpcclient.Client
	connectionCount *prometheus.Desc
	netSentBytes    *prometheus.Desc
	netRecvBytes    *prometheus.Desc
	logger          log.Logger
}

func NewNetworkCollector(rpcClient *rpcclient.Client, logger log.Logger) *networkCollector {

	return &networkCollector{
		client: rpcClient,
		logger: logger,
		netRecvBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "network", "receive_bytes_total"),
			"Total bytes received.",
			nil, nil),
		netSentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "network", "sent_bytes_total"),
			"Total bytes sent.",
			nil, nil),
		connectionCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "network", "connections_count"),
			"The number of connections to other nodes.",
			nil, nil),
	}
}

func (c *networkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.connectionCount
	ch <- c.netSentBytes
	ch <- c.netRecvBytes
}

func (c *networkCollector) Collect(ch chan<- prometheus.Metric) {

	getNetTotals, err := c.client.GetNetTotals()
	if err != nil {
		level.Error(c.logger).Log("err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.netRecvBytes, prometheus.CounterValue, float64(getNetTotals.TotalBytesRecv))
		ch <- prometheus.MustNewConstMetric(c.netSentBytes, prometheus.CounterValue, float64(getNetTotals.TotalBytesSent))
	}

	getConnectionCount, err := c.client.GetConnectionCount()
	if err != nil {
		level.Error(c.logger).Log("err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.connectionCount, prometheus.GaugeValue, float64(getConnectionCount))
	}

}
