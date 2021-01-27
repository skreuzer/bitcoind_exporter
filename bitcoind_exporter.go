package main

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
	"github.com/skreuzer/bitcoind_exporter/collector"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
)

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

	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)

	rpcClientConfig := &rpcclient.ConnConfig{
		Host:         *rpcServer,
		User:         *rpcUser,
		Pass:         *rpcPassword,
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	client, err := rpcclient.New(rpcClientConfig, nil)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	defer client.Shutdown()

	prometheus.MustRegister(collector.NewBlockChainCollector(client, logger))
	prometheus.MustRegister(collector.NewNetworkCollector(client, logger))
	prometheus.MustRegister(collector.NewMemPoolCollector(client, logger))

	level.Info(logger).Log("msg", "Starting bitcoind_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<head><title>Bitcoin Daemon Exporter</title></head>
                        <body><h1>Bitcoind Exporter</h1>
                        <p><a href="` + *metricsPath + `">Metrics</a></p>
                        </body></html>`))
	})

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}

}
