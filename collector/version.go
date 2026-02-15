package collector

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type versionCollector struct {
	client      *rpcclient.Client
	versionInfo *prometheus.Desc
	logger      log.Logger
	collector   string
}

type getNetworkInfoResult struct {
	Version         int    `json:"version"`
	SubVersion      string `json:"subversion"`
	ProtocolVersion int    `json:"protocolversion"`
}

func NewVersionCollector(rpcClient *rpcclient.Client, logger log.Logger) *versionCollector {
	return &versionCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "version",
		versionInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "version", "info"),
			"Bitcoin Core version information (always 1)",
			[]string{"version", "subversion", "protocol_version"}, nil),
	}
}

func (c *versionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.versionInfo
}

func (c *versionCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	rawNetworkInfo, err := c.client.RawRequest("getnetworkinfo", nil)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get network info",
			"err", err,
			"method", "getnetworkinfo",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	var networkInfo getNetworkInfoResult
	if err := json.Unmarshal(rawNetworkInfo, &networkInfo); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to unmarshal network info",
			"err", err,
			"method", "json.Unmarshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.versionInfo,
		prometheus.GaugeValue,
		1,
		strconv.Itoa(networkInfo.Version),
		networkInfo.SubVersion,
		strconv.Itoa(networkInfo.ProtocolVersion),
	)

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
