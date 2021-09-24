package collector

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type memPoolCollector struct {
	client    *rpcclient.Client
	txnCount  *prometheus.Desc
	logger    log.Logger
	collector string
}

func NewMemPoolCollector(rpcClient *rpcclient.Client, logger log.Logger) *memPoolCollector {

	return &memPoolCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "mempool",
		txnCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mempool", "transactions_count"),
			"Number of transactions in the mempool",
			nil, nil),
	}
}

func (c *memPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.txnCount
}

func (c *memPoolCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()
	getRawMemPool, err := c.client.GetRawMempool()
	if err != nil {
		level.Error(c.logger).Log("err", err)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.txnCount, prometheus.GaugeValue, float64(len(getRawMemPool)))
	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
