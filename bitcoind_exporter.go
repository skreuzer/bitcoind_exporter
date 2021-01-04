package main

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"net/http"
)

type bitcoindCollector struct {
	rpcClientConfig *rpcclient.ConnConfig
	blockCount      *prometheus.Desc
	headerCount     *prometheus.Desc
	difficulty      *prometheus.Desc
	connectionCount *prometheus.Desc
	netSentBytes    *prometheus.Desc
	netRecvBytes    *prometheus.Desc
}

const (
	namespace = "bitcoind"
)

var (
	promlogConfig = &promlog.Config{}
	logger        = promlog.New(promlogConfig)
)

func newBitcoindCollector(rpcUser string, rpcPassword string, rpcServer string) *bitcoindCollector {

	return &bitcoindCollector{
		rpcClientConfig: &rpcclient.ConnConfig{
			Host:         rpcServer,
			User:         rpcUser,
			Pass:         rpcPassword,
			HTTPPostMode: true,
			DisableTLS:   true,
		},
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
		netRecvBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "network", "receive_bytes_total"),
			"Total bytes received.",
			nil, nil),
		netSentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "network", "sent_bytes_total"),
			"Total bytes sent.",
			nil, nil),
		connectionCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "connection", "count"),
			"The number of connections to other nodes.",
			nil, nil),
	}
}

func (collector *bitcoindCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.blockCount
	ch <- collector.headerCount
	ch <- collector.difficulty
	ch <- collector.connectionCount
	ch <- collector.netSentBytes
	ch <- collector.netRecvBytes
}

func (collector *bitcoindCollector) Collect(ch chan<- prometheus.Metric) {

	client, err := rpcclient.New(collector.rpcClientConfig, nil)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	defer client.Shutdown()

	getBlockChainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		level.Error(logger).Log("err", err)
	} else {
		chain := getBlockChainInfo.Chain
		ch <- prometheus.MustNewConstMetric(collector.blockCount, prometheus.CounterValue, float64(getBlockChainInfo.Blocks), chain)
		ch <- prometheus.MustNewConstMetric(collector.headerCount, prometheus.CounterValue, float64(getBlockChainInfo.Headers), chain)
		ch <- prometheus.MustNewConstMetric(collector.difficulty, prometheus.CounterValue, getBlockChainInfo.Difficulty, chain)
	}

	getNetTotals, err := client.GetNetTotals()
	if err != nil {
		level.Error(logger).Log("err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(collector.netRecvBytes, prometheus.CounterValue, float64(getNetTotals.TotalBytesRecv))
		ch <- prometheus.MustNewConstMetric(collector.netSentBytes, prometheus.CounterValue, float64(getNetTotals.TotalBytesSent))
	}

	getConnectionCount, err := client.GetConnectionCount()
	if err != nil {
		level.Error(logger).Log("err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(collector.connectionCount, prometheus.GaugeValue, float64(getConnectionCount))
	}

}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":9960").String()

		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()

		rpcServer = kingpin.Flag(
			"bitcoind.rpc-address",
			"Address of the bitcoind RPC server",
		).OverrideDefaultFromEnvar("BITCOIND_RPC_ADDRESS").Default("localhost:8332").String()

		rpcUser = kingpin.Flag(
			"bitcoind.rpc-user",
			"Username for JSON-RPC connections",
		).OverrideDefaultFromEnvar("BITCOIND_RPC_USER").Required().String()

		rpcPassword = kingpin.Flag(
			"bitcoind.rpc-password",
			"Password for JSON-RPC connections",
		).OverrideDefaultFromEnvar("BITCOIND_RPC_PASSWORD").Required().String()
	)

	kingpin.Version(version.Print("bitcoind_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	exporter := newBitcoindCollector(*rpcUser, *rpcPassword, *rpcServer)
	prometheus.MustRegister(exporter)

	level.Info(logger).Log("msg", "Starting bitcoind_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<head><title>Node Exporter</title></head>
                        <body><h1>Bitcoind Exporter</h1>
                        <p><a href="` + *metricsPath + `">Metrics</a></p>
                        </body></html>`))
	})

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	http.ListenAndServe(*listenAddress, nil)
}
