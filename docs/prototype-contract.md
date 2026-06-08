> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Executable Prototype Contract

This document defines when the Aetra L1 prototype is considered working.

The contract is intentionally narrower than a mainnet readiness checklist. It proves a local or CI operator can build `aetrad`, create a reproducible localnet, produce blocks, run core signed transactions, query state through CLI/gRPC/REST, stop the network cleanly, and collect enough evidence to debug failures.

## Scope

Terms used across prototype docs:

- `prototype`: a prerelease/testnet artifact that proves basic L1 behavior; it is not mainnet-ready.
- `localnet`: generated local node homes under ignored `.localnet*` directories.
- `testnet`: a future public or shared network profile; it must not reuse local test key material.
- `mainnet-ready`: production validator onboarding, economics, upgrade governance, public security posture, and operational SLOs; out of scope here.
- `naet`: base transaction, staking, mint, and fee denom.
- `AET`: display metadata only, with exponent `9`; never use `AET` as a tx fee, bank send, stake, or module accounting denom.

Do not commit or package localnet homes, keyrings, validator keys, mnemonics, diagnostic bundles, `.work`, `.localnet`, or external database credentials. Aetra validator/full nodes do not require Redis, PostgreSQL, or another external database for consensus, mempool, or app state.

## Profiles

| Profile | Command Shape | Contract |
| --- | --- | --- |
| Single-node dev | `-ValidatorCount 1` on localnet scripts | Boot/query/debug only. It does not satisfy the full prototype acceptance suite because bank and multi-validator PoS flows use node1 and peer checks. |
| 3-validator smoke | default localnet and acceptance commands | Required prototype proof for PR/local development. Blocks, peers, validator set, txs, and queries must pass. |
| 5-validator extended | `-OutputDir .localnet-5 -ValidatorCount 5` | Manual/nightly scale profile. Must not require copied scripts or hardcoded node0-only assumptions. |

Default reusable variables:

```powershell
$CHAIN_ID = "aetra-local-1"
$NODE = "tcp://127.0.0.1:26657"
$GRPC = "127.0.0.1:9090"
$REST = "http://127.0.0.1:1317"
$HOME = ".localnet\node0\aetrad"
$FROM = "node0"
$KEYRING = "test"
$FEES = "1000000naet"
$NODE0 = build\aetrad.exe keys show $FROM -a --home $HOME --keyring-backend $KEYRING
$NODE1_HOME = ".localnet\node1\aetrad"
$NODE1 = build\aetrad.exe keys show node1 -a --home $NODE1_HOME --keyring-backend $KEYRING
```

For `.localnet-5`, set `$HOME` and `$NODE1_HOME` under `.localnet-5`. Keep `$NODE` on node0 unless testing another endpoint.

## MUST WORK Flows

| Flow | Commands | Expected Result | Evidence |
| --- | --- | --- | --- |
| Build binary | `go build -o build\aetrad.exe ./cmd/l1d`; `build\aetrad.exe version --long --output json` | Binary exists, starts, and reports app name, commit/version metadata, SDK, CometBFT, build date, and dirty state. | `tests/e2e/prototype_acceptance.ps1`, `tests/scripts/prototype_release_package_test.ps1`, `.github/workflows/prototype-release.yml` |
| Init localnet | `.\scripts\localnet\init.ps1`; `.\scripts\localnet\validate-genesis.ps1` | `.localnet\node*` exists, all nodes share the same genesis hash, chain-id is `aetra-local-1`, staking/mint/fees denom is `naet`, custom modules validate. | `docs/bootstrap-profile.md`, `app/determinism_test.go`, `x/*/types/genesis_test.go`, `tests/e2e/prototype_acceptance.ps1` |
| Start validators | `.\scripts\localnet\start.ps1 -NoInit -Wait` | Validators listen on configured RPC/gRPC/REST ports, peers connect, and blocks are produced. | `scripts/localnet/health.ps1`, `tests/e2e/localnet_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Wait height and health | `.\scripts\localnet\health.ps1 -ValidatorCount 3`; `Invoke-RestMethod http://127.0.0.1:26657/status` | Height is increasing, node status network is `aetra-local-1`, validator set size matches profile, REST and gRPC are reachable. | `docs/observability.md`, `tests/e2e/query_surface_smoke.ps1` |
| Query block | `build\aetrad.exe query block --node $NODE --output json` | Output contains a latest block height/header. | `tests/e2e/prototype_acceptance.ps1`, `tests/e2e/query_surface_smoke.ps1` |
| Bank send | `build\aetrad.exe tx bank send $FROM $NODE1 1000naet --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json`; then `query bank balance $NODE1 naet` | Tx commits with `code = 0`; node1 `naet` balance increases by `1000`. | `tests/e2e/prototype_acceptance.ps1`, `tests/e2e/native_token_smoke.ps1`, `tests/e2e/localnet_smoke.ps1` |
| Tokenfactory create/mint/query | `tx tokenfactory create-denom gold`; `$GOLD = "factory/$NODE0/gold"`; `query tokenfactory denom $GOLD`; `tx tokenfactory mint "1000000$GOLD" $NODE0`; `query bank balance $NODE0 $GOLD` | Factory denom admin is node0, mint succeeds, bank balance and supply reflect minted amount. | `tests/e2e/prototype_acceptance.ps1`, `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/query_surface_smoke.ps1`, `x/tokenfactory/keeper/msg_server_test.go` |
| Tokenfactory burn/change-admin | `tx tokenfactory burn "1000$GOLD" $NODE0`; `tx tokenfactory change-admin $GOLD $NODE1`; `query tokenfactory denom $GOLD` | Burn by current admin from own address succeeds; admin changes to node1; old admin mint/burn is rejected and new admin mint succeeds. | `docs/tokenfactory-lifecycle.md`, `tests/e2e/tokenfactory_smoke.ps1`, `x/tokenfactory/keeper/msg_server_test.go` |
| Fees wrong-denom rejection | `tx bank send $FROM $NODE1 1naet ... --fees 1000testtoken --output json` | Tx is rejected with `fee denom testtoken not accepted; use naet`; no state change. | `docs/fees-ante-policy.md`, `x/fees/keeper/ante_test.go`, `tests/e2e/fees_ante_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Mempool/CheckTx negative flow | wrong fee, insufficient funds, invalid sequence replay, unauthorized mint, invalid DEX pool, malformed denom | Rejection phase and error are documented; target module state and balances remain unchanged, except normal SDK fees for failed DeliverTx. | `docs/mempool-checktx-negative-flow.md`, `tests/e2e/mempool_negative_smoke.ps1` |
| DEX create pool/swap/query | Create/mint a factory denom; `tx dex create-pool 10000000naet "10000000$GOLD"`; `query dex pool 1`; `tx dex swap-exact-in 1 100000naet $GOLD 1`; query balances and pool. | Pool `1` has `lp/1`, reserves are non-zero, swap increases output balance, REST/gRPC pool query returns matching state. | `docs/dex-e2e-flow.md`, `x/dex/keeper/msg_server_test.go`, `tests/e2e/dex_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| DEX slippage failure | `tx dex swap-exact-in 1 100000naet $GOLD 1000000 ...` | Tx is rejected with `amount out below minimum`; balances/reserves are unchanged. | `tests/e2e/dex_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1 -Profile Full` |
| Minimal load profile | `scripts/localnet/load-profile.ps1 -Scenario mixed -Count 12 -RatePerSecond 2` | Local-only summary records block progress, tx latency, successes, failures, failure rate, and per-operation counts for bank/tokenfactory/DEX mixed load. | `docs/minimal-load-profile.md`, `tests/e2e/load_profile_smoke.ps1` |
| Guided demo | `scripts/demo/prototype-demo.ps1` | Human-readable local-only sequence builds/starts localnet, shows height, sends bank tx, creates/mints tokenfactory denom, swaps on DEX, queries REST, prints final balances, and stops. | `docs/prototype-demo.md`, `tests/scripts/prototype_demo_script_test.ps1`, e2e flows above |
| Validator and slashing query with direct delegation rejection | `query staking validators`; select a bonded validator; attempt `tx staking delegate <AE...validator> 5000000naet`; `query staking delegation <AE...delegator> <AE...validator>`; `query slashing params`; `query slashing signing-infos` | Validators are bonded, staking denom is `naet`, direct user delegation is rejected with the pool-only policy error, no delegation is created, validator power stays unchanged, and slashing params are positive. | `docs/pos-smoke-flow.md`, `app/pos_test.go`, `tests/e2e/pos_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |
| Stop/reset | `.\scripts\localnet\stop.ps1`; `.\scripts\localnet\reset.ps1` | Matching node processes stop; reset only deletes a resolved localnet directory inside the repo and never deletes repo root or arbitrary paths. | `scripts/localnet/common.ps1`, `tests/e2e/localnet_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` |

One-command acceptance:

```powershell
.\tests\e2e\prototype_smoke.ps1
.\tests\e2e\prototype_acceptance.ps1
.\tests\e2e\prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5
```

## MUST FAIL SAFELY

| Risk | Required Failure | Evidence |
| --- | --- | --- |
| Wrong fee denom | Any bank/tokenfactory/DEX tx using `testtoken`, factory denom, LP denom, or `AET` as fee is rejected before state mutation. | `x/fees/keeper/ante_test.go`, `tests/e2e/fees_ante_smoke.ps1` |
| Malformed or non-FeeTx fee path | Ante returns an error instead of panic or bypass. | `x/fees/keeper/ante_test.go` |
| Direct user delegation | `MsgDelegate`, `MsgUndelegate`, and `MsgBeginRedelegate` fail with the pool-only staking policy before any validator power or delegation state changes. | `app/pos_test.go`, `tests/integration/pos_lifecycle_test.go` |
| Unauthorized tokenfactory action | Non-admin mint/burn/change-admin and burn from another account are rejected; native `naet`/`AET` spoofing is rejected. | `x/tokenfactory/keeper/msg_server_test.go`, `tests/e2e/tokenfactory_smoke.ps1` |
| DEX accounting attack | Duplicate pair, wrong denom, wrong LP denom, tiny/zero liquidity, excessive slippage, corrupted pool, or reserve/module balance mismatch fails without panic. | `x/dex/keeper/msg_server_test.go`, `x/dex/keeper/math_test.go`, `tests/e2e/dex_smoke.ps1` |
| Query malformed/not found | Custom query servers return gRPC status errors, not panics or local state dumps. | `x/tokenfactory/keeper/query_server_test.go`, `x/dex/keeper/query_server_test.go`, `tests/e2e/query_surface_smoke.ps1` |
| Destructive script target | Reset/init refuses paths outside the workspace or the repository root. | `scripts/localnet/common.ps1`, `tests/e2e/localnet_smoke.ps1` |
| Secret exposure | Staged/history scans and diagnostic bundle checks find no mnemonics, validator keys, keyrings, database URLs, or environment secrets in tracked artifacts. | `docs/security/prototype-audit-gate.md`, `scripts/security/prototype-audit.ps1`, `scripts/localnet/diagnostics.ps1` |

## Audit Gate

Fast PR gate:

```powershell
.\scripts\security\prototype-audit.ps1 -Profile Fast
```

Release gate:

```powershell
.\scripts\security\prototype-audit.ps1 -Profile Full
.\tests\e2e\prototype_acceptance.ps1 -Profile Smoke
```

Manual/nightly scale gate:

```powershell
.\scripts\security\prototype-audit.ps1 -Profile Nightly
.\tests\e2e\prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5
```

The Cosmos-specific manual review must cover:

- no nondeterministic consensus state writes,
- no ABCI/query/genesis panics on malformed input,
- correct `Msg` signer and keeper authorization checks,
- SDK denom validation for user-provided denoms,
- bank supply, DEX reserves, and LP supply consistency,
- integer-only DEX math and safe rounding,
- bounded tx paths and bounded or capped list queries,
- no secrets in tracked files, logs, diagnostics, or release artifacts.

Critical or High audit findings block the prototype unless fixed, regression-tested, or explicitly triaged with owner and mitigation.

## Blocker Classification

MUST FIX before declaring the prototype working:

- `aetrad` does not build or cannot report version metadata.
- Genesis validation fails or nodes in one localnet have different genesis hashes.
- 3-validator localnet cannot produce blocks, form peers, or expose RPC/gRPC/REST health.
- Bank, fees, tokenfactory create/mint/query, DEX create/swap/query, or PoS delegation flow fails in the 3-validator acceptance suite.
- Wrong fee denom is accepted.
- Unauthorized tokenfactory mint/burn/admin transfer succeeds.
- DEX reserve, module balance, or LP supply invariant breaks.
- Any malformed tx/query/genesis path creates an ABCI panic.
- Any untriaged Critical/High security finding remains.
- Secret material or external database credentials enter tracked files or release artifacts.

SHOULD FIX but not a blocker for local prototype declaration:

- 5-validator full profile is manual/nightly rather than mandatory on every PR.
- Minimal load profile is a local prototype baseline, not a production performance claim.
- High-cardinality query load benchmarks should be added before public explorer/API load testing.
- Known dependency advisories require reachability triage or upstream upgrade before release tagging.
- Dummy vote-extension behavior remains unsuitable for public validator networks.

NICE TO HAVE:

- More benchmarks for DEX math, query caps, and localnet startup.
- Automated transcript attachment for release notes.
- Multi-OS e2e localnet runs beyond build/package matrix.

## Evidence Index

Primary docs:

- [Operator Commands](operator-commands.md)
- [Operator Troubleshooting Runbook](operator-troubleshooting.md)
- [Transaction Lifecycle Matrix](transaction-lifecycle-matrix.md)
- [Tx Event Contract](event-contract.md)
- [Bootstrap Profile](bootstrap-profile.md)
- [Prototype Acceptance Suite](prototype-acceptance-suite.md)
- [Query Surface](query-surface.md)
- [Observability](observability.md)
- [Security Audit Gate](security/prototype-audit-gate.md)
- [Prototype Release Package](release/prototype-package.md)
- [Prototype Non-Goals And Limitations](release/prototype-limitations.md)

Required local checks:

```powershell
go test ./...
go vet ./...
buf lint
go build -o build\aetrad.exe ./cmd/l1d
.\tests\e2e\prototype_acceptance.ps1
.\scripts\security\prototype-audit.ps1 -Profile Fast
```

Release-like local package:

```powershell
.\scripts\release\prototype-package.ps1 -Version prototype-local -TargetOS windows -TargetArch amd64
```
