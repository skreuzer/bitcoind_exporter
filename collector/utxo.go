package collector

import (
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type utxoCollector struct {
	client          *rpcclient.Client
	utxoCount       *prometheus.Desc
	utxoTotalAmount *prometheus.Desc
	utxoDiskSize    *prometheus.Desc
	utxoHeight      *prometheus.Desc
	logger          log.Logger
	collector       string
}

func NewUtxoCollector(rpcClient *rpcclient.Client, logger log.Logger) *utxoCollector {
	return &utxoCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "utxo",
		utxoCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "utxo", "count"),
			"Total number of UTXOs in the UTXO set",
			nil, nil),
		utxoTotalAmount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "utxo", "total_amount_satoshis"),
			"Total amount of Bitcoin in all UTXOs (in satoshis)",
			nil, nil),
		utxoDiskSize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "utxo", "disk_size_bytes"),
			"Disk size of the UTXO set in bytes",
			nil, nil),
		utxoHeight: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "utxo", "height"),
			"Blockchain height when the UTXO set was generated",
			nil, nil),
	}
}

func (c *utxoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.utxoCount
	ch <- c.utxoTotalAmount
	ch <- c.utxoDiskSize
	ch <- c.utxoHeight
}

func (c *utxoCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	getUtxoSetInfo, err := c.client.GetTxOutSetInfo()
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get UTXO set info",
			"err", err,
			"method", "GetTxOutSetInfo",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.utxoCount, prometheus.GaugeValue, float64(getUtxoSetInfo.TxOuts))
	ch <- prometheus.MustNewConstMetric(c.utxoTotalAmount, prometheus.GaugeValue, float64(getUtxoSetInfo.TotalAmount))
	ch <- prometheus.MustNewConstMetric(c.utxoDiskSize, prometheus.GaugeValue, float64(getUtxoSetInfo.DiskSize))
	ch <- prometheus.MustNewConstMetric(c.utxoHeight, prometheus.GaugeValue, float64(getUtxoSetInfo.Height))

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
