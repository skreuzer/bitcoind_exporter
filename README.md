# Bitcoin Daemon Exporter

[![Docker Pulls](https://img.shields.io/docker/pulls/skreuzer/bitcoind_exporter.svg?maxAge=604800)](https://hub.docker.com/repository/docker/skreuzer/bitcoind_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/skreuzer/bitcoind_exporter)](https://goreportcard.com/report/github.com/skreuzer/bitcoind_exporter)

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
| bitcoind_blockchain_size_bytes | The estimated size of the block and undo files on disk. | chain |
| bitcoind_blockchain_initial_download | Estimate of whether this node is in initial block download mode. | chain |
| bitcoind_blockchain_verification_progress | Estimate of verification progress [0..1]. | chain |
| bitcoind_blockchain_pruned | Indicates if the node is pruned. | chain |
| bitcoind_network_connections_count | Number of connections to other nodes. | |
| bitcoind_network_receive_bytes_total | Total bytes received. | |
| bitcoind_network_sent_bytes_total | Total bytes sent. | |
| bitcoind_mempool_transactions_count | Number of transactions in the mempool. | |
| bitcoind_mempool_size_bytes | Total size of transactions in the mempool in bytes. | |
| bitcoind_mempool_usage | Memory usage of the mempool in bytes. | |
| bitcoind_mempool_usage_percent | Percentage of mempool usage relative to max mempool size. | |
| bitcoind_mempool_min_fee | Minimum fee rate in the mempool in BTC/KB. | |
| bitcoind_mining_network_hash_ps | Estimated network hashes per second. | |
| bitcoind_mining_current_block_size | Size of the current block. | |
| bitcoind_mining_current_block_weight | Weight of the current block. | |
| bitcoind_mining_current_block_transactions | Number of transactions in the current block. | |
| bitcoind_mining_pooled_transactions | Number of transactions in the memory pool. | |
| bitcoind_utxo_count | Total number of UTXOs in the UTXO set. | |
| bitcoind_utxo_total_amount_satoshis | Total amount of Bitcoin in all UTXOs (in satoshis). | |
| bitcoind_utxo_disk_size_bytes | Disk size of the UTXO set in bytes. | |
| bitcoind_utxo_height | Blockchain height when the UTXO set was generated. | |
| bitcoind_memory_total_used_bytes | Total memory used in bytes. | |
| bitcoind_memory_locked_bytes | Locked memory in bytes. | |
| bitcoind_memory_used_bytes | Memory used in bytes. | category |
| bitcoind_memory_used_block_tree_bytes | Memory used by block tree in bytes. | |
| bitcoind_memory_used_tx_index_bytes | Memory used by transaction index in bytes. | |
| bitcoind_memory_used_slab_bytes | Memory used by slab allocator in bytes. | |
| bitcoind_exporter_collect_error | Error occured during collection. | collector |
| bitcoind_exporter_collector_duration_seconds | Collector time duration. | collector |

## Labels

| Label | Description |
| ----- | ----------- |
| chain | Current network name as defined in BIP70 (main, test, regtest) |
| collector | Internal name of the collector (mempool, network, blockchain, mining, utxo, memory) |

## Using Docker

You can deploy this exporter using the skreuzer/bitcoind_exporter Docker image.

For example:

```bash
docker pull skreuzer/bitcoind_exporter

docker run -d \
  -e BITCOIND_RPC_ADDRESS=172.16.0.1:8332 \
  -e BITCOIND_RPC_PASSWORD=j9JkYnPIxwQgRzG3qAA \
  -e BITCOIND_RPC_USER=bitcoin-rpc-user \
  -p 9960:9960 \
  skreuzer/bitcoind_exporter
```

## License
This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.
