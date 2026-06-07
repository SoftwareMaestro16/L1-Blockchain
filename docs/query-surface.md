> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Prototype Query Surface

Aetra prototype nodes must be observable through CLI, gRPC, REST gateway, and CometBFT RPC. The default localnet exposes node0 at:

```powershell
$NODE = "tcp://127.0.0.1:26657"
$GRPC = "127.0.0.1:9090"
$REST = "http://127.0.0.1:1317"
$HOME = ".localnet\node0\aetrad"
$FROM = "node0"
$KEYRING = "test"
$NODE0 = build\aetrad.exe keys show $FROM -a --home $HOME --keyring-backend $KEYRING
```

## Endpoint Matrix

All examples use JSON output and the default localnet values above. Sample responses are intentionally minimal; real responses include additional Cosmos SDK fields such as pagination and consensus metadata.

| Query | CLI | gRPC method | REST path | Sample request | Sample response |
| --- | --- | --- | --- | --- | --- |
| Latest block | `build\aetrad.exe query block --node $NODE --output json` | `cosmos.base.tendermint.v1beta1.Service/GetLatestBlock` | `GET /cosmos/base/tendermint/v1beta1/blocks/latest` | `{}` | `{"block":{"header":{"height":"12","chain_id":"aetra-local-1"}}}` |
| Node info | `build\aetrad.exe status --node $NODE --output json` | `cosmos.base.tendermint.v1beta1.Service/GetNodeInfo` | `GET /cosmos/base/tendermint/v1beta1/node_info` | `{}` | `{"default_node_info":{"network":"aetra-local-1"}}` |
| Bank balance | `build\aetrad.exe query bank balance $NODE0 naet --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `cosmos.bank.v1beta1.Query/Balance` | `GET /cosmos/bank/v1beta1/balances/{address}/by_denom?denom=naet` | `{"address":"AE...","denom":"naet"}` | `{"balance":{"denom":"naet","amount":"1000000"}}` |
| Bank balances | `build\aetrad.exe query bank balances $NODE0 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `cosmos.bank.v1beta1.Query/AllBalances` | `GET /cosmos/bank/v1beta1/balances/{address}` | `{"address":"AE...","pagination":{"limit":"100"}}` | `{"balances":[{"denom":"naet","amount":"1000000"}]}` |
| Staking validators | `build\aetrad.exe query staking validators --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `cosmos.staking.v1beta1.Query/Validators` | `GET /cosmos/staking/v1beta1/validators?pagination.limit=100` | `{"pagination":{"limit":"100"}}` | `{"validators":[{"operator_address":"AE...","status":"BOND_STATUS_BONDED"}]}` |
| Fees params | `build\aetrad.exe query fees params --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.fees.v1.Query/Params` | `GET /l1/fees/v1/params` | `{}` | `{"params":{"allowed_fee_denoms":["naet"],"validator_rewards_ratio":"0.98","community_pool_ratio":"0.02"}}` |
| Factory denoms | `build\aetrad.exe query tokenfactory denoms --limit 50 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.tokenfactory.v1.Query/Denoms` | `GET /l1/tokenfactory/v1/denoms?pagination.limit=50` | `{"pagination":{"limit":"50"}}` | `{"denoms":[{"denom":"factory/AE.../gold","admin":"AE..."}],"pagination":{"next_key":"..."}}` |
| Factory denom | `build\aetrad.exe query tokenfactory denom $GOLD --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.tokenfactory.v1.Query/Denom` | `GET /l1/tokenfactory/v1/denom/{denom}` | `{"denom":"factory/AE.../gold"}` | `{"metadata":{"denom":"factory/AE.../gold","admin":"AE..."}}` |
| DEX pools | `build\aetrad.exe query dex pools --limit 50 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.dex.v1.Query/Pools` | `GET /l1/dex/v1/pools?pagination.limit=50` | `{"pagination":{"limit":"50"}}` | `{"pools":[{"id":"1","denom0":"factory/AE.../gold","denom1":"naet","lp_denom":"lp/1"}],"pagination":{"next_key":"..."}}` |
| DEX pool | `build\aetrad.exe query dex pool 1 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json` | `l1.dex.v1.Query/Pool` | `GET /l1/dex/v1/pools/{pool_id}` | `{"pool_id":"1"}` | `{"pool":{"id":"1","reserve0":"10000000","reserve1":"10000000","lp_denom":"lp/1"}}` |

CometBFT RPC remains available for node-level checks that are not gRPC services:

| Query | Endpoint | Expected result |
| --- | --- | --- |
| RPC status | `GET http://127.0.0.1:26657/status` | `result.node_info.network = aetra-local-1` |
| RPC peers | `GET http://127.0.0.1:26657/net_info` | peer count for multi-validator localnet |
| RPC validator set | `GET http://127.0.0.1:26657/validators?per_page=100` | expected validator count with positive voting power |

## Error Contract

Custom query servers return gRPC-status compatible errors:

- `InvalidArgument`: nil request, malformed denom, `pool_id = 0`, invalid pagination key, unsupported offset/count/reverse mode, or limit above max
- `NotFound`: valid tokenfactory denom or DEX pool id does not exist

REST gateway maps these to HTTP status codes, for example `400` for invalid pool id and `404` for missing custom module objects.

## Bounded Lists

The custom list endpoints use Cosmos `PageRequest` / `PageResponse` fields:

- `tokenfactory Denoms`
- `dex Pools`

Prototype defaults:

- default limit: `50`
- max limit: `100`
- supported cursor mode: forward `next_key` pagination
- unsupported modes: `offset`, `count_total`, and `reverse`

Clients should request the first page with `--limit N` or `?pagination.limit=N`, then pass the returned base64 `pagination.next_key` back through `--page-key` or `?pagination.key=`. Query handlers read at most `limit + 1` KV entries and reject excessive limits or keys outside the endpoint prefix.

## Compatibility Policy

- Proto changes require the normal proto workflow: edit `.proto`, run generation, run `buf lint`, and review generated Go diffs.
- Generated `.pb.go` and `.pb.gw.go` files must not be edited manually.
- Breaking REST/gRPC path or field changes require versioning or an explicit compatibility exception.
- Query handlers must remain read-only: no state writes, no bank movement, and no business side effects.

## Security Notes

Query endpoints must not expose mnemonics, private keys, local environment secrets, or keyring material. Errors should be short status errors, not dumps of local process state. Iterator code must be deterministic and bounded on public query paths.

## Acceptance Check

```powershell
.\tests\e2e\query_surface_smoke.ps1
.\.work\tools\bin\buf.exe lint
```
