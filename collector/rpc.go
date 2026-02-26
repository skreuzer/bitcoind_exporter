package collector

import (
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type rpcCollector struct {
	client          *rpcclient.Client
	activeCalls     *prometheus.Desc
	totalCalls      *prometheus.Desc
	totalCallsByCmd *prometheus.Desc
	logger          log.Logger
	collector       string
}

type rpcMethodInfo struct {
	Name     string `json:"name"`
	Calls    int64  `json:"calls"`
	Duration int64  `json:"duration"`
	LastSeen int64  `json:"lastseen"`
}

type rpcInfo struct {
	ActiveCommands []rpcMethodInfo `json:"active_commands"`
	Commands       []rpcMethodInfo `json:"commands"`
}

func NewRpcCollector(rpcClient *rpcclient.Client, logger log.Logger) *rpcCollector {
	return &rpcCollector{
		client:    rpcClient,
		logger:    logger,
		collector: "rpc",
		activeCalls: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rpc", "active_calls"),
			"Number of active RPC calls currently being processed",
			nil, nil),
		totalCalls: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rpc", "all_calls_total"),
			"Total number of RPC calls made",
			nil, nil),
		totalCallsByCmd: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rpc", "calls_total"),
			"Total number of RPC calls by command",
			[]string{"command"}, nil),
	}
}

func (c *rpcCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.activeCalls
	ch <- c.totalCalls
	ch <- c.totalCallsByCmd
}

func (c *rpcCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()

	rawRpcInfo, err := c.client.RawRequest("getrpcinfo", nil)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to get RPC info",
			"err", err,
			"method", "getrpcinfo",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	var getRpcInfo rpcInfo
	if err := json.Unmarshal(rawRpcInfo, &getRpcInfo); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to unmarshal RPC info",
			"err", err,
			"method", "json.Unmarshal",
		)
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, c.collector)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.activeCalls, prometheus.GaugeValue, float64(len(getRpcInfo.ActiveCommands)))

	var totalCalls int64
	for _, cmd := range getRpcInfo.Commands {
		totalCalls += cmd.Calls
		ch <- prometheus.MustNewConstMetric(c.totalCallsByCmd, prometheus.CounterValue, float64(cmd.Calls), cmd.Name)
	}
	ch <- prometheus.MustNewConstMetric(c.totalCalls, prometheus.CounterValue, float64(totalCalls))

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, c.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), c.collector)
}
