> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Operator Troubleshooting Runbook

This runbook is for local Aetra prototype operation. It uses localnet homes, local endpoints, and `--keyring-backend test` only. Never paste mnemonics, validator keys, node keys, or keyring files into tickets, chats, logs, or diagnostic bundles.

Default variables:

```powershell
$CHAIN_ID = "aetra-local-1"
$OUTPUT = ".localnet"
$VALIDATORS = 3
$NODE = "tcp://127.0.0.1:26657"
$GRPC = "127.0.0.1:9090"
$REST = "http://127.0.0.1:1317"
$HOME = "$OUTPUT\node0\aetrad"
$FROM = "node0"
$KEYRING = "test"
$FEES = "1000000naet"
```

For a 5-validator profile, set `$OUTPUT = ".localnet-5"` and `$VALIDATORS = 5`. If you shift ports, update `$NODE`, `$GRPC`, `$REST`, and the matching `-Base*Port` flags on localnet scripts.

## First Triage

Run these before changing state:

```powershell
git status --short
build\aetrad.exe version --long --output json
.\scripts\localnet\health.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS -Json
.\scripts\localnet\diagnostics.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS
```

Diagnostic bundles are written under ignored `.work\diagnostics`. They redact common secret patterns and exclude keyrings, validator key files, validator state, and node keys. Do not attach bundles from a non-local environment without a separate secret review.

## Common Failures

| Failure | Symptoms | Commands | Fix |
| --- | --- | --- | --- |
| Binary missing | `build\aetrad.exe` is not found; scripts fail before init/start. | `Test-Path build\aetrad.exe`; `.\scripts\build-aetrad.ps1`; `build\aetrad.exe version --long --output json` | Rebuild with `.\scripts\build-aetrad.ps1`. If build fails, run `go test ./...` and check the build output before localnet work. |
| Port in use | `Port ... is already in use`, localnet start cannot bind RPC/REST/gRPC/P2P. | `.\scripts\localnet\health.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS`; `Get-NetTCPConnection -LocalPort 26657,1317,9090 -ErrorAction SilentlyContinue` | Stop the localnet with `.\scripts\localnet\stop.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS`, stop the conflicting local process, or restart with shifted base ports. |
| No blocks | Health reports height stuck or startup timeout. Logs may show consensus, disk, or validator errors. | `.\scripts\localnet\health.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS`; `Get-Content "$OUTPUT\logs\node0.err.log" -Tail 80`; `Invoke-RestMethod "http://127.0.0.1:26657/status"` | Confirm all expected node processes are running. If config is stale, stop and re-run init/start. Destructive reset: `.\scripts\localnet\reset.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS` deletes the generated localnet directory. |
| No peers | Multi-validator health shows zero peers or validator set mismatch. | `Invoke-RestMethod "http://127.0.0.1:26657/net_info"`; `.\scripts\localnet\health.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS -Json` | Use the same `$OUTPUT` and validator count across init/start/health. Stop, validate genesis, then start again with `-NoInit -Wait`. Shift ports if another localnet is already using defaults. |
| Wrong fee denom | Tx is rejected with `fee denom ... not accepted; use naet`. | `build\aetrad.exe query fees params --grpc-addr $GRPC --grpc-insecure --node $NODE --output json`; inspect the failed tx command. | Use `--fees $FEES`, where `$FEES = "1000000naet"`. Do not use `AET`, factory denoms, LP denoms, or multi-denom fees for prototype tx fees. |
| Insufficient funds | Tx fails with `insufficient funds`, balance is too low for amount plus fee. | `$ADDR = build\aetrad.exe keys show $FROM -a --home $HOME --keyring-backend $KEYRING`; `build\aetrad.exe query bank balance $ADDR naet --node $NODE --output json` | Fund local accounts through normal local bank transfer: `.\scripts\localnet\fund.ps1 -OutputDir $OUTPUT -Binary build\aetrad.exe -ChainId $CHAIN_ID -RPCPort 26657 -Recipients @($ADDR) -Amount 1000000naet`. |
| Sequence mismatch | Tx fails with `account sequence mismatch`, `account sequence`, or signature verification error after a retry. | `build\aetrad.exe query auth account $ADDR --node $NODE --output json`; wait for one new block with `.\scripts\localnet\health.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS` | Wait one block and rebuild the tx from the CLI. Do not reuse old signed transaction bytes after a previous tx commits. |
| DEX slippage | `dex add-liquidity` or `dex swap-exact-in` fails with `minted shares below minimum` or `amount out below minimum`. | `build\aetrad.exe query dex pool 1 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json`; query trader balances for both pool denoms. | Lower `min_shares` or `min_amount_out` for the local prototype test, or use current pool reserves to calculate a realistic bound. State should remain unchanged on rejection. |
| Unauthorized tokenfactory | Mint, burn, or change-admin fails with an unauthorized/admin error. | `build\aetrad.exe query tokenfactory denom $DENOM --grpc-addr $GRPC --grpc-insecure --node $NODE --output json`; `build\aetrad.exe keys show $FROM -a --home $HOME --keyring-backend $KEYRING` | Use the current denom admin as `--from`. After `change-admin`, old admin txs should fail and new admin txs should pass. |
| REST down | CLI gRPC queries may pass, but REST returns timeout, 404, or 503. | `Invoke-RestMethod "$REST/cosmos/base/tendermint/v1beta1/blocks/latest"`; `.\scripts\localnet\health.ps1 -OutputDir $OUTPUT -ValidatorCount $VALIDATORS -Json` | Confirm REST and gRPC ports match the generated node config. Restart localnet if the REST gateway started before gRPC readiness. For shifted ports, update `$REST` and `$GRPC`. |

## Port Profiles

Default 3-validator ports:

```text
node0 RPC 26657, REST 1317, gRPC 9090
node1 RPC 26757, REST 1318, gRPC 9091
node2 RPC 26857, REST 1319, gRPC 9092
```

For 5 validators, the scripts continue incrementing from the configured base ports. Use explicit bases when running multiple localnets:

```powershell
.\scripts\localnet\init.ps1 -OutputDir .localnet-5 -ValidatorCount 5 -BaseP2PPort 27656 -BaseRPCPort 27657 -BaseRESTPort 1417 -BaseGRPCPort 9190 -BasePprofPort 6160
.\scripts\localnet\start.ps1 -OutputDir .localnet-5 -ValidatorCount 5 -NoInit -Wait -BaseP2PPort 27656 -BaseRPCPort 27657 -BaseRESTPort 1417 -BaseGRPCPort 9190 -BasePprofPort 6160
$NODE = "tcp://127.0.0.1:27657"
$REST = "http://127.0.0.1:1417"
$GRPC = "127.0.0.1:9190"
```

## Evidence And Links

Runbook commands are covered where practical by:

- `tests\e2e\localnet_smoke.ps1` for start, health, peers, REST, and gRPC.
- `tests\e2e\fees_ante_smoke.ps1` for wrong fee rejection.
- `tests\e2e\mempool_negative_smoke.ps1` for wrong fee, insufficient funds, sequence mismatch, unauthorized mint, and malformed tx failures.
- `tests\e2e\dex_smoke.ps1` for DEX slippage rejection.
- `tests\scripts\observability_scripts_test.ps1` for redacted health and diagnostic bundles.

Deep links:

- [Operator Commands](operator-commands.md)
- [Prototype Observability](observability.md)
- [Mempool And CheckTx Negative Flow](mempool-checktx-negative-flow.md)
- [Fees Ante Policy](fees-ante-policy.md)
- [DEX E2E Flow](dex-e2e-flow.md)
- [Local Funding](local-funding.md)
