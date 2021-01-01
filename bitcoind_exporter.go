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
	difficulty      *prometheus.Desc
}

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
		blockCount: prometheus.NewDesc("bitcoind_block_count",
			"Number of blocks in the longest blockchain.",
			nil, nil),
		difficulty: prometheus.NewDesc("bitcoind_difficulty",
			"The proof-of-work difficulty as a multiple of the minimum difficulty.",
			nil, nil),
	}
}

func (collector *bitcoindCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.blockCount
	ch <- collector.difficulty
}

func (collector *bitcoindCollector) Collect(ch chan<- prometheus.Metric) {

	client, err := rpcclient.New(collector.rpcClientConfig, nil)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	defer client.Shutdown()

	getBlockCount, err := client.GetBlockCount()
	if err != nil {
		level.Error(logger).Log("err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(collector.blockCount, prometheus.CounterValue, float64(getBlockCount))
	}

	getDifficulty, err := client.GetDifficulty()
	if err != nil {
		level.Error(logger).Log("err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(collector.difficulty, prometheus.CounterValue, getDifficulty)
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

	kingpin.Version(version.Print(exporter))
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
