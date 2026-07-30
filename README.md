# Bitcoin Daemon Exporter

[![Docker Pulls](https://img.shields.io/docker/pulls/skreuzer/bitcoind_exporter.svg?maxAge=604800)](https://hub.docker.com/repository/docker/skreuzer/bitcoind_exporter)

Export statistics from bitcoind to [Prometheus](https://prometheus.io).

Metrics are retrieved using calls to the JSON-RPC interface of the bitcoin
daemon.

To run it:

    go build
    ./bitcoind_exporter [flags]

## Collectors

Most collectors are enabled by default. The UTXO collector is disabled by default because it performs an expensive operation.

```bash
# Enable the UTXO collector (disabled by default)
./bitcoind_exporter --collector.utxo

# Disable a collector that's enabled by default
./bitcoind_exporter --no-collector.blockstats

# Disable multiple collectors
./bitcoind_exporter --no-collector.utxo --no-collector.rpc --no-collector.mining
```

### Available Collectors

| Collector | Flag | Default | Description |
| --------- | ---- | ------- | ----------- |
| blockchain | `--collector.blockchain` | ✓ | Blockchain info (blocks, headers, difficulty) |
| blockstats | `--collector.blockstats` | ✓ | Statistics for the latest block (fees, transactions, UTXO changes) |
| network | `--collector.network` | ✓ | Network connections and traffic |
| mempool | `--collector.mempool` | ✓ | Mempool size and transactions |
| mining | `--collector.mining` | ✓ | Mining info and hash rate |
| utxo | `--collector.utxo` | ✗ | UTXO set info (warning: expensive operation) |
| memory | `--collector.memory` | ✓ | Memory usage |
| peer | `--collector.peer` | ✓ | Connected peer statistics |
| rpc | `--collector.rpc` | ✓ | RPC call statistics |
| version | `--collector.version` | ✓ | Bitcoin Core version info |
| uptime | `--collector.uptime` | ✓ | Server uptime |

**Note:** The UTXO collector calls `gettxoutsetinfo` which can take several minutes to complete on a full node. It is disabled by default for this reason.

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
| bitcoind_blockstats_utxo_increase | Net increase in UTXOs in the latest block. | |
| bitcoind_blockstats_utxo_size_increase_bytes | Net increase in UTXO size in the latest block. | |
| bitcoind_blockstats_total_fee_satoshis | Total transaction fees in the latest block. | |
| bitcoind_blockstats_total_out_satoshis | Total output value in the latest block. | |
| bitcoind_blockstats_total_in_satoshis | Total input value in the latest block. | |
| bitcoind_blockstats_total_subsidy_satoshis | Block subsidy in the latest block. | |
| bitcoind_blockstats_transactions | Number of transactions in the latest block. | |
| bitcoind_blockstats_height | Height of the latest block. | |
| bitcoind_network_connections_count | Number of connections to other nodes. | |
| bitcoind_network_receive_bytes_total | Total bytes received. | |
| bitcoind_network_sent_bytes_total | Total bytes sent. | |
| bitcoind_mempool_transactions_count | Number of transactions in the mempool. | |
| bitcoind_mempool_size_bytes | Total size of transactions in the mempool in bytes. | |
| bitcoind_mempool_usage | Memory usage of the mempool in bytes. | |
| bitcoind_mempool_usage_percent | Percentage of mempool usage relative to max mempool size. Returns 0 if no max mempool limit is configured. | |
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
| bitcoind_peer_count | Total number of connected peers. | |
| bitcoind_peer_count_by_connection | Number of peers by connection type. | connection_type |
| bitcoind_peer_count_by_network | Number of peers by network type. | network |
| bitcoind_peer_count_by_sync_status | Number of peers by sync status. | sync_status |
| bitcoind_peer_at_latest_block | Number of peers at the latest block height. | |
| bitcoind_peer_average_ping_seconds | Average ping time across all peers in seconds. | |
| bitcoind_peer_average_ping_by_connection_seconds | Average ping time by connection type in seconds. | connection_type |
| bitcoind_peer_min_ping_seconds | Minimum ping time across all peers in seconds. | |
| bitcoind_peer_max_ping_seconds | Maximum ping time across all peers in seconds. | |
| bitcoind_rpc_active_calls | Number of active RPC calls currently being processed. | |
| bitcoind_rpc_all_calls_total | Total number of RPC calls made. | |
| bitcoind_rpc_calls_total | Total number of RPC calls by command. | command |
| bitcoind_version_info | Bitcoin Core version information (always 1). | version, subversion, protocol_version |
| bitcoind_uptime_seconds | The number of seconds that the Bitcoin server has been running. | |
| bitcoind_exporter_collect_error | Error occured during collection. | collector |
| bitcoind_exporter_collector_duration_seconds | Collector time duration. | collector |

## Labels

| Label | Description |
| ----- | ----------- |
| chain | Current network name as defined in BIP70 (main, test, regtest) |
| collector | Internal name of the collector (blockstats, mempool, network, blockchain, mining, utxo, memory, peer, rpc, version, uptime) |
| connection_type | Type of peer connection (inbound, outbound, unknown) |
| network | Network protocol of peer (ipv4, ipv6, onion, unknown) |
| sync_status | Sync status of peer (synced, not_synced) |
| command | RPC command name |
| version | Bitcoin Core version number (e.g., 260000) |
| subversion | Bitcoin Core user agent string (e.g., /Satoshi:26.0/) |
| protocol_version | Bitcoin protocol version (e.g., 70016) |

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

## TLS Configuration

By default, the exporter connects to bitcoind without TLS. To enable TLS:

```bash
./bitcoind_exporter --bitcoind.rpc-tls --bitcoind.rpc-user=user --bitcoind.rpc-password=pass
```

For self-signed certificates, provide the CA certificate:

```bash
./bitcoind_exporter --bitcoind.rpc-tls --bitcoind.rpc-ca-cert=/path/to/ca.crt
```

Or via environment variables:

```bash
export BITCOIND_RPC_TLS=true
export BITCOIND_RPC_CA_CERT=/path/to/ca.crt
./bitcoind_exporter --bitcoind.rpc-user=user --bitcoind.rpc-password=pass
```

### Extracting Certificate from Server

To extract the certificate from a server with a self-signed cert:

```bash
openssl s_client -connect localhost:8332 -showcerts </dev/null 2>/dev/null | openssl x509 -outform PEM > server.crt
```

Then use it with:

```bash
./bitcoind_exporter --bitcoind.rpc-tls --bitcoind.rpc-ca-cert=server.crt
```

### TLS Flags

| Flag | Environment Variable | Description |
| ---- | -------------------- | ----------- |
| `--bitcoind.rpc-tls` | `BITCOIND_RPC_TLS` | Enable TLS for RPC connections |
| `--bitcoind.rpc-ca-cert` | `BITCOIND_RPC_CA_CERT` | Path to CA certificate file for self-signed certs |

## License
This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.
