package collector

import (
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type memoryCollector struct {
	client        *rpcclient.Client
	totalUsed     *prometheus.Desc
	locked        *prometheus.Desc
	usedTotal     *prometheus.Desc
	usedBlockTree *prometheus.Desc
	usedTxIndex   *prometheus.Desc
	usedSlab      *prometheus.Desc
	logger        log.Logger
	collector     string
}

type memoryInfo struct {
	TotalUsed int64            `json:"total"`
	Locked    interface{}      `json:"locked"`
	Used      map[string]int64 `json:"used"`
}

func NewMemoryCollector(rpcClient *rpcclient.Client, logger log.Logger) *memoryCollector {
	return &memoryCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "memory",
		totalUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "total_used_bytes"),
			"Total memory used in bytes",
			nil, nil),
		locked: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "locked_bytes"),
			"Locked memory in bytes",
			nil, nil),
		usedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "used_bytes"),
			"Memory used in bytes",
			[]string{"category"}, nil),
		usedBlockTree: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "used_block_tree_bytes"),
			"Memory used by block tree in bytes",
			nil, nil),
		usedTxIndex: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "used_tx_index_bytes"),
			"Memory used by transaction index in bytes",
			nil, nil),
		usedSlab: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "used_slab_bytes"),
			"Memory used by slab allocator in bytes",
			nil, nil),
	}
}

func (c *memoryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalUsed
	ch <- c.locked
	ch <- c.usedTotal
	ch <- c.usedBlockTree
	ch <- c.usedTxIndex
	ch <- c.usedSlab
}

func (c *memoryCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	rawMemoryInfo, err := c.client.RawRequest("getmemoryinfo", nil)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get memory info",
			"err", err,
			"method", "getmemoryinfo",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}
	var getMemoryInfo memoryInfo
	if err := json.Unmarshal(rawMemoryInfo, &getMemoryInfo); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to unmarshal memory info",
			"err", err,
			"method", "json.Unmarshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.totalUsed, prometheus.GaugeValue, float64(getMemoryInfo.TotalUsed))

	// Handle locked field - it can be either an int64 (old format) or a map/object (new format)
	var lockedValue float64
	switch v := getMemoryInfo.Locked.(type) {
	case float64:
		lockedValue = v
	case int64:
		lockedValue = float64(v)
	case map[string]interface{}:
		// New format: extract "current" value from the object
		if current, ok := v["current"].(float64); ok {
			lockedValue = current
		}
	}
	ch <- prometheus.MustNewConstMetric(c.locked, prometheus.GaugeValue, lockedValue)

	for category, usage := range getMemoryInfo.Used {
		ch <- prometheus.MustNewConstMetric(c.usedTotal, prometheus.GaugeValue, float64(usage), category)
	}

	if blockTreeUsage, ok := getMemoryInfo.Used["block-tree"]; ok {
		ch <- prometheus.MustNewConstMetric(c.usedBlockTree, prometheus.GaugeValue, float64(blockTreeUsage))
	}
	if txIndexUsage, ok := getMemoryInfo.Used["txindex"]; ok {
		ch <- prometheus.MustNewConstMetric(c.usedTxIndex, prometheus.GaugeValue, float64(txIndexUsage))
	}
	if slabUsage, ok := getMemoryInfo.Used["slab"]; ok {
		ch <- prometheus.MustNewConstMetric(c.usedSlab, prometheus.GaugeValue, float64(slabUsage))
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
