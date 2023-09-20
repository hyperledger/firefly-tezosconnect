[![codecov](https://codecov.io/gh/hyperledger/firefly-tezosconnect/branch/main/graph/badge.svg)](https://codecov.io/gh/hyperledger/firefly-tezosconnect)
[![Go Reference](https://pkg.go.dev/badge/github.com/hyperledger/firefly-tezosconnect.svg)](https://pkg.go.dev/github.com/hyperledger/firefly-tezosconnect)

# Hyperledger FireFly Tezos Connector

This repo provides a reference implementation of the FireFly Connector API (FFCAPI)
for Tezos blockchain.

See the [Hyperledger Firefly Documentation](https://hyperledger.github.io/firefly/overview/public_vs_permissioned.html#firefly-architecture-for-public-chains)
and the [FireFly Transaction Manager](https://github.com/hyperledger/firefly-transaction-manager) repository for
more information.

# License

Apache 2.0

## Transaction signing

Tezosconnect leverages remote transaction signing through a powerful 'signatory' service, offering compatibility with multiple key management solutions. With the flexibility to use AWS KMS, Azure KMS, GCP KMS, Yubi HSM, etc. for transaction signing, you can secure your blockchain transactions efficiently and conveniently.

More info at: https://signatory.io/

## Configuration

For a full list of configuration options see [config.md](./config.md)

## Example configuration

```yaml
connector:
  blockchain:
    rpc: https://rpc.ghost.tzstats.com
    network: ghostnet
    signatory: http://localhost:6732
```

## Blockchain node compatibility

For Tezos connector to function properly, you should check the blockchain node supports the following RPC Methods:

### Chains
- `GET /chains/<chain_id>/blocks/<block_id>/hash`
- `GET /chains/<chain_id>/blocks/<block_id>/operations/<list_offset>`
- `GET /chains/<chain_id>/blocks/<block_id>/operations/<list_offset>/<operation_offset>`
- `POST /chains/<chain_id>/blocks/<block_id>/helpers/forge/operations`
- `POST /chains/<chain_id>/blocks/<block_id>/helpers/scripts/simulate_operation`
- `POST /chains/<chain_id>/blocks/<block_id>/helpers/scripts/run_operation`

### Block monitoring
- `GET /monitor/heads/<chain_id>`

### Injection
- `POST /injection/operation`
