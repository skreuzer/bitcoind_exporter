package collector

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type blockChainCollector struct {
	client               *rpcclient.Client
	blockCount           *prometheus.Desc
	headerCount          *prometheus.Desc
	difficulty           *prometheus.Desc
	sizeOnDisk           *prometheus.Desc
	initialBlockDownload *prometheus.Desc
	logger               log.Logger
}

func NewBlockChainCollector(rpcClient *rpcclient.Client, logger log.Logger) *blockChainCollector {

	return &blockChainCollector{
		client: rpcClient,
		logger: logger,
		blockCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockchain", "blocks_validated_total"),
			"Current number of blocks processed in the server",
			[]string{"chain"}, nil),
		headerCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockchain", "headers_validated_total"),
			"Current number of headers processed in the server",
			[]string{"chain"}, nil),
		difficulty: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockchain", "difficulty"),
			"The proof-of-work difficulty as a multiple of the minimum difficulty.",
			[]string{"chain"}, nil),
		sizeOnDisk: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockchain", "size_bytes"),
			"The estimated size of the block and undo files on disk.",
			[]string{"chain"}, nil),
		initialBlockDownload: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "blockchain", "initial_download"),
			"Estimate of whether this node is in initial block download mode.",
			[]string{"chain"}, nil),
	}
}

func (collector *blockChainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.blockCount
	ch <- collector.headerCount
	ch <- collector.difficulty
}

func (c *blockChainCollector) Collect(ch chan<- prometheus.Metric) {

	getBlockChainInfo, err := c.client.GetBlockChainInfo()
	if err != nil {
		level.Error(c.logger).Log("err", err)
	} else {
		chain := getBlockChainInfo.Chain
		ch <- prometheus.MustNewConstMetric(c.blockCount, prometheus.CounterValue, float64(getBlockChainInfo.Blocks), chain)
		ch <- prometheus.MustNewConstMetric(c.headerCount, prometheus.CounterValue, float64(getBlockChainInfo.Headers), chain)
		ch <- prometheus.MustNewConstMetric(c.difficulty, prometheus.CounterValue, getBlockChainInfo.Difficulty, chain)
		ch <- prometheus.MustNewConstMetric(c.sizeOnDisk, prometheus.CounterValue, float64(getBlockChainInfo.SizeOnDisk), chain)

		var initialDownload float64
		if getBlockChainInfo.InitialBlockDownload {
			initialDownload = 1
		}
		ch <- prometheus.MustNewConstMetric(c.initialBlockDownload, prometheus.GaugeValue, initialDownload, chain)
	}
}
