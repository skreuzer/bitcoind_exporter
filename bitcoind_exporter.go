// Package main implements a Prometheus exporter for Bitcoin Core (bitcoind).
// It exposes blockchain, network, and mempool metrics via HTTP for Prometheus scraping.
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
	"github.com/skreuzer/bitcoind_exporter/collector"
)

var (
	logger = promlog.New(&promlog.Config{})

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

	disableExporterMetrics = kingpin.Flag(
		"web.disable-exporter-metrics",
		"Exclude metrics about the exporter (promhttp_*, process_*, go_*)",
	).Default("false").Bool()
)

func metricsHandler(client *rpcclient.Client) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()
		registry.MustRegister(collector.NewBlockChainCollector(client, logger))
		registry.MustRegister(collector.NewNetworkCollector(client, logger))
		registry.MustRegister(collector.NewMemPoolCollector(client, logger))

		gatherers := prometheus.Gatherers{registry}
		if !*disableExporterMetrics {
			gatherers = append(gatherers, prometheus.DefaultGatherer)
		}

		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	kingpin.Version(version.Print("bitcoind_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	_ = level.Info(logger).Log("msg", "Starting bitcoind_exporter", "version", version.Info())
	_ = level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	rpcClientConfig := &rpcclient.ConnConfig{
		Host:         *rpcServer,
		User:         *rpcUser,
		Pass:         *rpcPassword,
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	client, err := rpcclient.New(rpcClientConfig, nil)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	defer func() {
		client.Shutdown()
	}()

	http.Handle(*metricsPath, metricsHandler(client))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<head><title>Bitcoin Daemon Exporter</title></head>
	                       <body><h1>Bitcoind Exporter</h1>
	                       <p><a href="` + *metricsPath + `">Metrics</a></p>
	                       </body></html>`))
	})

	_ = level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	server := &http.Server{
		Addr:              *listenAddress,
		ReadHeaderTimeout: 5 * time.Second,
	}
	err = server.ListenAndServe()
	if err != nil {
		_ = level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}
}
