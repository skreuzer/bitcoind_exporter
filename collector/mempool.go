package collector

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type memPoolCollector struct {
	client   *rpcclient.Client
	txnCount *prometheus.Desc
	logger   log.Logger
}

func NewMemPoolCollector(rpcClient *rpcclient.Client, logger log.Logger) *memPoolCollector {

	return &memPoolCollector{
		client: rpcClient,
		logger: logger,
		txnCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mempool", "transactions_count"),
			"Number of transcations in the mempool",
			nil, nil),
	}
}

func (c *memPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.txnCount
}

func (c *memPoolCollector) Collect(ch chan<- prometheus.Metric) {

	getRawMemPool, err := c.client.GetRawMempool()
	if err != nil {
		level.Error(c.logger).Log("err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.txnCount, prometheus.GaugeValue, float64(len(getRawMemPool)))
	}

}
