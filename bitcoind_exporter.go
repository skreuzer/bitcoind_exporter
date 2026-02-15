package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/skreuzer/bitcoind_exporter/collector"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	logger = level.NewFilter(log.NewLogfmtLogger(os.Stderr), level.AllowInfo())

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
		registry.MustRegister(collector.NewMiningCollector(client, logger))
		registry.MustRegister(collector.NewUtxoCollector(client, logger))
		registry.MustRegister(collector.NewMemoryCollector(client, logger))
		registry.MustRegister(collector.NewPeerCollector(client, logger))
		registry.MustRegister(collector.NewRpcCollector(client, logger))

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

	level.Info(logger).Log("msg", "Starting bitcoind_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	rpcClientConfig := &rpcclient.ConnConfig{
		Host:         *rpcServer,
		User:         *rpcUser,
		Pass:         *rpcPassword,
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	client, err := rpcclient.New(rpcClientConfig, nil)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Failed to create RPC client")
		os.Exit(1)
	}
	defer client.Shutdown()

	// Setup HTTP server
	server := &http.Server{
		Addr:    *listenAddress,
		Handler: nil,
	}

	// Setup routes
	http.Handle(*metricsPath, metricsHandler(client))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<head><title>Bitcoin Daemon Exporter</title></head>
                        <body><h1>Bitcoind Exporter</h1>
                        <p><a href="` + *metricsPath + `">Metrics</a></p>
                        </body></html>`))
	})

	// Handle graceful shutdown
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		level.Info(logger).Log("msg", "Shutting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			level.Error(logger).Log("msg", "Server shutdown error", "err", err)
		}
		client.Shutdown()
		os.Exit(0)
	}()

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		level.Error(logger).Log("msg", "Server error", "err", err)
		os.Exit(1)
	}
}
