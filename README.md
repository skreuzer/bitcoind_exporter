# Bitcoin Daemon Exporter

Export statistics from bitcoind to [Prometheus](https://prometheus.io).

Metrics are retrieved using calls to the JSON-RPC interface of the bitcoin
daemon.

To run it:

    go build
    ./bitcoind_exporter [flags]

## Exported Metrics

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| bitcoind_block_count | Number of blocks in the longest blockchain. | |
| bitcoind_difficulty  _| The proof-of-work difficulty as a multiple of the minimum difficulty. | |
| bitcoind_connection_count | Number of connections to other nodes. | |
