package collector

import (
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type blockStatsCollector struct {
	rpcClient           *rpcclient.Client
	utxoIncrease        *prometheus.Desc
	utxoSizeIncrease    *prometheus.Desc
	totalFee            *prometheus.Desc
	totalOut            *prometheus.Desc
	totalIn             *prometheus.Desc
	totalSubsidy        *prometheus.Desc
	transactions        *prometheus.Desc
	height              *prometheus.Desc
	logger              log.Logger
	collector           string
}

type blockStatsInfo struct {
	Height        int64   `json:"height"`
	UtxoIncrease  int64   `json:"utxo_increase"`
	UtxoSizeInc   int64   `json:"utxo_size_inc"`
	TotalFee      float64 `json:"total_fee"`
	TotalOut      float64 `json:"total_out"`
	TotalIn       float64 `json:"total_in"`
	TotalSubsidy  float64 `json:"total_subsidy"`
	Transactions  int64   `json:"txs"`
}

func NewBlockStatsCollector(rpcClient *rpcclient.Client, logger log.Logger) *blockStatsCollector {
	return &blockStatsCollector{
		rpcClient: rpcClient,
		logger:    logger,
		collector: "blockstats",
		utxoIncrease: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "utxo_increase"),
			"Net increase in UTXOs in the latest block",
			nil, nil),
		utxoSizeIncrease: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "utxo_size_increase_bytes"),
			"Net increase in UTXO size in the latest block",
			nil, nil),
		totalFee: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "total_fee_satoshis"),
			"Total transaction fees in the latest block",
			nil, nil),
		totalOut: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "total_out_satoshis"),
			"Total output value in the latest block",
			nil, nil),
		totalIn: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "total_in_satoshis"),
			"Total input value in the latest block",
			nil, nil),
		totalSubsidy: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "total_subsidy_satoshis"),
			"Block subsidy in the latest block",
			nil, nil),
		transactions: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "transactions"),
			"Number of transactions in the latest block",
			nil, nil),
		height: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockstats", "height"),
			"Height of the latest block",
			nil, nil),
	}
}

func (c *blockStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.utxoIncrease
	ch <- c.utxoSizeIncrease
	ch <- c.totalFee
	ch <- c.totalOut
	ch <- c.totalIn
	ch <- c.totalSubsidy
	ch <- c.transactions
	ch <- c.height
}

func (c *blockStatsCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	// Get the best block hash to get stats for the latest block
	bestBlockHash, err := c.rpcClient.GetBestBlockHash()
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get best block hash",
			"err", err,
			"method", "GetBestBlockHash",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	// Get block stats for the latest block
	hashBytes, err := json.Marshal(bestBlockHash.String())
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to marshal block hash",
			"err", err,
			"method", "json.Marshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}
	rawBlockStats, err := c.rpcClient.RawRequest("getblockstats", []json.RawMessage{hashBytes})
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get block stats",
			"err", err,
			"method", "getblockstats",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	var getBlockStats blockStatsInfo
	if err := json.Unmarshal(rawBlockStats, &getBlockStats); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to unmarshal block stats",
			"err", err,
			"method", "json.Unmarshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.utxoIncrease, prometheus.GaugeValue, float64(getBlockStats.UtxoIncrease))
	ch <- prometheus.MustNewConstMetric(c.utxoSizeIncrease, prometheus.GaugeValue, float64(getBlockStats.UtxoSizeInc))
	ch <- prometheus.MustNewConstMetric(c.totalFee, prometheus.GaugeValue, getBlockStats.TotalFee)
	ch <- prometheus.MustNewConstMetric(c.totalOut, prometheus.GaugeValue, getBlockStats.TotalOut)
	ch <- prometheus.MustNewConstMetric(c.totalIn, prometheus.GaugeValue, getBlockStats.TotalIn)
	ch <- prometheus.MustNewConstMetric(c.totalSubsidy, prometheus.GaugeValue, getBlockStats.TotalSubsidy)
	ch <- prometheus.MustNewConstMetric(c.transactions, prometheus.GaugeValue, float64(getBlockStats.Transactions))
	ch <- prometheus.MustNewConstMetric(c.height, prometheus.GaugeValue, float64(getBlockStats.Height))

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
