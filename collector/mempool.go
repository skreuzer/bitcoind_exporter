package collector

import (
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type memPoolCollector struct {
	client       *rpcclient.Client
	txnCount     *prometheus.Desc
	logger       log.Logger
	collector    string
	sizeBytes    *prometheus.Desc
	usage        *prometheus.Desc
	usagePercent *prometheus.Desc
	minFee       *prometheus.Desc
}

type mempoolInfo struct {
	Size          int64   `json:"size"`
	Bytes         int64   `json:"bytes"`
	Usage         int64   `json:"usage"`
	MaxMempool    int64   `json:"maxmempool"`
	MempoolMinFee float64 `json:"mempoolminfee"`
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
		sizeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mempool", "size_bytes"),
			"Total size of transactions in the mempool in bytes",
			nil, nil),
		usage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mempool", "usage"),
			"Memory usage of the mempool in bytes",
			nil, nil),
		usagePercent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mempool", "usage_percent"),
			"Percentage of mempool usage relative to max mempool size",
			nil, nil),
		minFee: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mempool", "min_fee"),
			"Minimum fee rate in the mempool in BTC/KB",
			nil, nil),
	}
}

func (c *memPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.txnCount
	ch <- c.sizeBytes
	ch <- c.usage
	ch <- c.usagePercent
	ch <- c.minFee
}

func (c *memPoolCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	rawMempoolInfo, err := c.client.RawRequest("getmempoolinfo", nil)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get mempool info",
			"err", err,
			"method", "getmempoolinfo",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}
	var getMempoolInfo mempoolInfo
	if err := json.Unmarshal(rawMempoolInfo, &getMempoolInfo); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to unmarshal mempool info",
			"err", err,
			"method", "json.Unmarshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.txnCount, prometheus.GaugeValue, float64(getMempoolInfo.Size))
	ch <- prometheus.MustNewConstMetric(c.sizeBytes, prometheus.GaugeValue, float64(getMempoolInfo.Bytes))
	ch <- prometheus.MustNewConstMetric(c.usage, prometheus.GaugeValue, float64(getMempoolInfo.Usage))
	ch <- prometheus.MustNewConstMetric(c.usagePercent, prometheus.GaugeValue, float64(getMempoolInfo.Usage)/float64(getMempoolInfo.MaxMempool)*100)
	ch <- prometheus.MustNewConstMetric(c.minFee, prometheus.GaugeValue, getMempoolInfo.MempoolMinFee)

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
