package collector

import (
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type peerCollector struct {
	client                  *rpcclient.Client
	peerCount               *prometheus.Desc
	peerCountByConnection   *prometheus.Desc
	peerCountByNetwork      *prometheus.Desc
	peerCountBySyncStatus   *prometheus.Desc
	peersAtLatestBlock      *prometheus.Desc
	averagePingSeconds      *prometheus.Desc
	averagePingByConnection *prometheus.Desc
	minPingSeconds          *prometheus.Desc
	maxPingSeconds          *prometheus.Desc
	logger                  log.Logger
	collector               string
}

type peerInfo struct {
	ID             int64   `json:"id"`
	Addr           string  `json:"addr"`
	ConnectionType string  `json:"connection_type"`
	Network        string  `json:"network"`
	Version        int     `json:"version"`
	SubVer         string  `json:"subver"`
	StartingHeight int64   `json:"starting_height"`
	SyncedBlocks   int64   `json:"synced_height"`
	PingTime       float64 `json:"pingtime"`
	MinPingTime    float64 `json:"minping"`
}

func NewPeerCollector(rpcClient *rpcclient.Client, logger log.Logger) *peerCollector {
	return &peerCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "peer",
		peerCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "count"),
			"Total number of connected peers",
			nil, nil),
		peerCountByConnection: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "count_by_connection"),
			"Number of peers by connection type",
			[]string{"connection_type"}, nil),
		peerCountByNetwork: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "count_by_network"),
			"Number of peers by network type",
			[]string{"network"}, nil),
		peerCountBySyncStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "count_by_sync_status"),
			"Number of peers by sync status",
			[]string{"sync_status"}, nil),
		peersAtLatestBlock: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "at_latest_block"),
			"Number of peers at the latest block height",
			nil, nil),
		averagePingSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "average_ping_seconds"),
			"Average ping time across all peers in seconds",
			nil, nil),
		averagePingByConnection: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "average_ping_by_connection_seconds"),
			"Average ping time by connection type in seconds",
			[]string{"connection_type"}, nil),
		minPingSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "min_ping_seconds"),
			"Minimum ping time across all peers in seconds",
			nil, nil),
		maxPingSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "peer", "max_ping_seconds"),
			"Maximum ping time across all peers in seconds",
			nil, nil),
	}
}

func (c *peerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.peerCount
	ch <- c.peerCountByConnection
	ch <- c.peerCountByNetwork
	ch <- c.peerCountBySyncStatus
	ch <- c.peersAtLatestBlock
	ch <- c.averagePingSeconds
	ch <- c.averagePingByConnection
	ch <- c.minPingSeconds
	ch <- c.maxPingSeconds
}

func (c *peerCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	rawPeerInfo, err := c.client.RawRequest("getpeerinfo", nil)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get peer info",
			"err", err,
			"method", "getpeerinfo",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	var peers []peerInfo
	if err := json.Unmarshal(rawPeerInfo, &peers); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to unmarshal peer info",
			"err", err,
			"method", "json.Unmarshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	peerCount := len(peers)
	ch <- prometheus.MustNewConstMetric(c.peerCount, prometheus.GaugeValue, float64(peerCount))

	if peerCount == 0 {
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
		ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
		return
	}

	var totalPing float64
	var minPing float64 = -1
	var maxPing float64
	peersAtLatest := 0

	connectionTypeCount := make(map[string]int)
	networkCount := make(map[string]int)
	syncStatusCount := map[string]int{"synced": 0, "not_synced": 0}
	connectionPingSum := make(map[string]float64)
	connectionPingCount := make(map[string]int)

	for _, peer := range peers {
		totalPing += peer.PingTime

		if minPing == -1 || peer.PingTime < minPing {
			minPing = peer.PingTime
		}
		if peer.PingTime > maxPing {
			maxPing = peer.PingTime
		}

		if peer.SyncedBlocks >= peer.StartingHeight {
			peersAtLatest++
		}

		connectionType := peer.ConnectionType
		if connectionType == "" {
			connectionType = "unknown"
		}
		connectionTypeCount[connectionType]++
		connectionPingSum[connectionType] += peer.PingTime
		connectionPingCount[connectionType]++

		network := peer.Network
		if network == "" {
			network = "unknown"
		}
		networkCount[network]++

		if peer.SyncedBlocks >= peer.StartingHeight {
			syncStatusCount["synced"]++
		} else {
			syncStatusCount["not_synced"]++
		}
	}

	ch <- prometheus.MustNewConstMetric(c.averagePingSeconds, prometheus.GaugeValue, totalPing/float64(peerCount))
	ch <- prometheus.MustNewConstMetric(c.minPingSeconds, prometheus.GaugeValue, minPing)
	ch <- prometheus.MustNewConstMetric(c.maxPingSeconds, prometheus.GaugeValue, maxPing)
	ch <- prometheus.MustNewConstMetric(c.peersAtLatestBlock, prometheus.GaugeValue, float64(peersAtLatest))

	for connType, count := range connectionTypeCount {
		ch <- prometheus.MustNewConstMetric(c.peerCountByConnection, prometheus.GaugeValue, float64(count), connType)
		avgPing := connectionPingSum[connType] / float64(connectionPingCount[connType])
		ch <- prometheus.MustNewConstMetric(c.averagePingByConnection, prometheus.GaugeValue, avgPing, connType)
	}

	for network, count := range networkCount {
		ch <- prometheus.MustNewConstMetric(c.peerCountByNetwork, prometheus.GaugeValue, float64(count), network)
	}

	for syncStatus, count := range syncStatusCount {
		ch <- prometheus.MustNewConstMetric(c.peerCountBySyncStatus, prometheus.GaugeValue, float64(count), syncStatus)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
