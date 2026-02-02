package collector

import (
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type miningCollector struct {
	client             *rpcclient.Client
	networkHashPS      *prometheus.Desc
	currentBlockSize   *prometheus.Desc
	currentBlockWeight *prometheus.Desc
	currentBlockTx     *prometheus.Desc
	pooledTx           *prometheus.Desc
	logger             log.Logger
	collector          string
}

func NewMiningCollector(rpcClient *rpcclient.Client, logger log.Logger) *miningCollector {
	return &miningCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "mining",
		networkHashPS: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mining", "network_hash_ps"),
			"Estimated network hashes per second",
			nil, nil),
		currentBlockSize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mining", "current_block_size"),
			"Size of the current block",
			nil, nil),
		currentBlockWeight: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mining", "current_block_weight"),
			"Weight of the current block",
			nil, nil),
		currentBlockTx: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mining", "current_block_transactions"),
			"Number of transactions in the current block",
			nil, nil),
		pooledTx: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mining", "pooled_transactions"),
			"Number of transactions in the memory pool",
			nil, nil),
	}
}

func (c *miningCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.networkHashPS
	ch <- c.currentBlockSize
	ch <- c.currentBlockWeight
	ch <- c.currentBlockTx
	ch <- c.pooledTx
}

func (c *miningCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	getMiningInfo, err := c.client.GetMiningInfo()
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get mining info",
			"err", err,
			"method", "GetMiningInfo",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.networkHashPS, prometheus.GaugeValue, getMiningInfo.NetworkHashPS)
	ch <- prometheus.MustNewConstMetric(c.currentBlockSize, prometheus.GaugeValue, float64(getMiningInfo.CurrentBlockSize))
	ch <- prometheus.MustNewConstMetric(c.currentBlockWeight, prometheus.GaugeValue, float64(getMiningInfo.CurrentBlockWeight))
	ch <- prometheus.MustNewConstMetric(c.currentBlockTx, prometheus.GaugeValue, float64(getMiningInfo.CurrentBlockTx))
	ch <- prometheus.MustNewConstMetric(c.pooledTx, prometheus.GaugeValue, float64(getMiningInfo.PooledTx))

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
