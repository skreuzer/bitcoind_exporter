# Bitcoin Daemon Exporter

[![CI](https://github.com/skreuzer/bitcoind_exporter/workflows/CI/badge.svg)](https://github.com/skreuzer/bitcoind_exporter/actions?query=workflow%3ACI)
[![Docker Pulls](https://img.shields.io/docker/pulls/skreuzer/bitcoind_exporter.svg?maxAge=604800)](https://hub.docker.com/repository/docker/skreuzer/bitcoind_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/skreuzer/bitcoind_exporter)](https://goreportcard.com/report/github.com/skreuzer/bitcoind_exporter)
[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](LICENSE)

Export statistics from bitcoind to [Prometheus](https://prometheus.io).

Metrics are retrieved using calls to the JSON-RPC interface of the bitcoin
daemon.

To run it:

    go build
    ./bitcoind_exporter [flags]

## Exported Metrics

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| bitcoind_blockchain_blocks_validated_total | Current number of blocks processed in the server. | chain |
| bitcoind_blockchain_headers_validated_total | Current number of headers processed in the server. | chain |
| bitcoind_blockchain_difficulty | The proof-of-work difficulty as a multiple of the minimum difficulty. | chain |
| bitcoind_blockchain_size_bytes | Estimated size of the block and undo files on disk. | chain |
| bitcoind_blockchain_initial_download | Estimate of whether the node is in initial block download mode. | chain |
| bitcoind_network_connections_count | Number of connections to other nodes. | |
| bitcoind_network_receive_bytes_total | Total bytes received over the network. | |
| bitcoind_network_sent_bytes_total | Total bytes sent over the network. | |
| bitcoind_mempool_transactions_count | Number of transactions in the mempool. | |
| bitcoind_exporter_collect_error | Error occured during collection. | collector |
| bitcoind_exporter_collector_duration_seconds | Collector time duration. | collector |

## Labels

| Label | Description |
| ----- | ----------- |
| chain | Current network name as defined in BIP70 (main, test, regtest) |
| collector | Internal name of the collector (mempool, network, blockchain) |

## Using Container Image

You can deploy this exporter using a container runtime like Podman or Docker.

Example with Podman:

```bash
podman pull docker.io/skreuzer/bitcoind_exporter

podman run -d \
  -e BITCOIND_RPC_ADDRESS=172.16.0.1:8332 \
  -e BITCOIND_RPC_PASSWORD=ilzbylavkypurfvbyvchsapul \
  -e BITCOIND_RPC_USER=bitcoin-rpc-user \
  -p 9960:9960 \
  skreuzer/bitcoind_exporter
```

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git

### Build

```bash
git clone https://github.com/skreuzer/bitcoind_exporter.git
cd bitcoind_exporter
go build
```

### Run Tests

```bash
go test -v ./...
```

### Run Linter

```bash
golangci-lint run
```

## Container Build

Build the container image locally:

```bash
podman build -t bitcoind_exporter -f Containerfile .
```

## Configuration

The exporter can be configured using command-line flags or environment variables:

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--web.listen-address` | - | `:9960` | Address to listen on for web interface and telemetry |
| `--web.telemetry-path` | - | `/metrics` | Path under which to expose metrics |
| `--bitcoind.rpc-address` | `BITCOIND_RPC_ADDRESS` | `localhost:8332` | Address of the bitcoind RPC server |
| `--bitcoind.rpc-user` | `BITCOIND_RPC_USER` | - | Username for JSON-RPC connections (required) |
| `--bitcoind.rpc-password` | `BITCOIND_RPC_PASSWORD` | - | Password for JSON-RPC connections (required) |
| `--web.disable-exporter-metrics` | - | `false` | Exclude metrics about the exporter itself |

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'bitcoind'
    static_configs:
      - targets: ['localhost:9960']
```

## License
This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.