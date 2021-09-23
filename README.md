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
| bitcoind_blockchain_blocks_validated_total | Current number of blocks processed in the server. | chain |
| bitcoind_blockchain_headers_validated_total | Current number of headers processed in the server. | chain |
| bitcoind_blockchain_difficulty | The proof-of-work difficulty as a multiple of the minimum difficulty. | chain |
| bitcoind_blockchain_size_bytes | Estimated size of the block and undo files on disk. | chain |
| bitcoind_blockchain_initial_download | Estimate of whether the node is in initial block download mode. | chain |
| bitcoind_network_connections_count | Number of connections to other nodes. | |
| bitcoind_network_receive_bytes_total | Total bytes received over the network. | |
| bitcoind_network_sent_bytes_total | Total bytes sent over the network. | |
| bitcoind_mempool_transactions_count | Number of transcations in the mempool. | |
| bitcoind_exporter_collect_error | Error occured during collection. | collector |
| bitcoind_exporter_collector_duration_seconds | Collector time duration. | collector |

## Labels

| Label | Description |
| ----- | ----------- |
| chain | Current network name as defined in BIP70 (main, test, regtest) |
| collector | Internal name of the collector (mempool, network, blockchain) |

## License
This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.
