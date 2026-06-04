# Operator Commands

This document is the prototype command runbook for `orbitalisd`.

It is scoped to local prototype operation. Commands that use `--keyring-backend test` are for ignored localnet homes only. Do not use the test keyring, generated local validator keys, or localnet mnemonics for public networks.

## Build And Version

Build with the one-command wrapper:

```powershell
.\scripts\build-orbitalisd.ps1
```

The wrapper uses ignored `.work\gocache`, `.work\gotmp`, and `.work\gomodcache` directories. This keeps local builds isolated from a modified global Go module cache.

Check the binary:

```powershell
build\orbitalisd.exe version
build\orbitalisd.exe version --long --output json
build\orbitalisd.exe --help
```

Expected version fields:

- `name = Orbitalis`
- `server_name = orbitalisd`
- `version` is non-empty
- `commit` is non-empty when built from a git checkout
- `cosmos_sdk_version` is non-empty
- `extra_info.cometbft_version` is non-empty
- `extra_info.dirty` is `true`, `false`, or `unknown`

Release-like builds can override the default metadata through the wrapper:

```powershell
$commit = git rev-parse HEAD
.\scripts\build-orbitalisd.ps1 -Version prototype-local -Commit $commit -Force
```

## Localnet

Default localnet profile:

```powershell
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\start.ps1 -NoInit -Wait
```

5-validator profile:

```powershell
.\scripts\localnet\init.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\scripts\localnet\validate-genesis.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\scripts\localnet\start.ps1 -OutputDir .localnet-5 -ValidatorCount 5 -NoInit -Wait
```

Stop or reset:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
```

Use variables for reusable commands:

```powershell
$CHAIN_ID = "orbitalis-local-1"
$NODE = "tcp://127.0.0.1:26657"
$GRPC = "127.0.0.1:9090"
$REST = "http://127.0.0.1:1317"
$HOME = ".localnet\node0\orbitalisd"
$FROM = "node0"
$FEES = "1000000norb"
$KEYRING = "test"
$NODE0 = build\orbitalisd.exe keys show $FROM -a --home $HOME --keyring-backend $KEYRING
```

For `.localnet-5`, set `$HOME = ".localnet-5\node0\orbitalisd"` and keep `$NODE` on node0 unless you intentionally query another node RPC port.

## Common Flags

Use these flags for signed local tx commands:

```powershell
--from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
```

Use JSON output for queries:

```powershell
--node $NODE --output json
```

Prototype examples use `norb` fees. `ORB` is display metadata only and is not a transaction denom.

## Queries

Node and block:

```powershell
build\orbitalisd.exe status --node $NODE
build\orbitalisd.exe query block --node $NODE --output json
```

Native token:

```powershell
build\orbitalisd.exe query bank denom-metadata norb --node $NODE --output json
build\orbitalisd.exe query bank total-supply-of norb --node $NODE --output json
build\orbitalisd.exe query bank balance $NODE0 norb --node $NODE --output json
```

Fees:

```powershell
build\orbitalisd.exe query fees params --grpc-addr $GRPC --grpc-insecure --node $NODE --output json
Invoke-RestMethod "$REST/l1/fees/v1/params"
```

Staking and slashing:

```powershell
build\orbitalisd.exe query staking params --node $NODE --output json
build\orbitalisd.exe query staking validators --node $NODE --output json
build\orbitalisd.exe query slashing params --node $NODE --output json
build\orbitalisd.exe query slashing signing-infos --node $NODE --output json
```

## Staking Tx

Delegate to any bonded validator returned by the validators query:

```powershell
$VALIDATOR = (build\orbitalisd.exe query staking validators --node $NODE --output json | ConvertFrom-Json).validators[0].operator_address
build\orbitalisd.exe tx staking delegate $VALIDATOR 5000000norb --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query staking delegation $NODE0 $VALIDATOR --node $NODE --output json
```

## Bank Tx

Fund local prototype accounts from the genesis-funded `node0` account. This is local-only and uses normal `bank send`, not a faucet mint:

```powershell
$NODE1_HOME = ".localnet\node1\orbitalisd"
$NODE1 = build\orbitalisd.exe keys show node1 -a --home $NODE1_HOME --keyring-backend $KEYRING
.\scripts\localnet\fund.ps1 -OutputDir .localnet -Binary build\orbitalisd.exe -ChainId $CHAIN_ID -RPCPort 26657 -Recipients @($NODE1) -Amount 1000000norb
```

Get a recipient from node1:

```powershell
$NODE1_HOME = ".localnet\node1\orbitalisd"
$NODE1 = build\orbitalisd.exe keys show node1 -a --home $NODE1_HOME --keyring-backend $KEYRING
```

Send `norb`:

```powershell
build\orbitalisd.exe tx bank send $FROM $NODE1 1000norb --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query bank balance $NODE1 norb --node $NODE --output json
```

## Tokenfactory

Create a factory denom:

```powershell
build\orbitalisd.exe tx tokenfactory create-denom gold --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
$GOLD = "factory/$NODE0/gold"
build\orbitalisd.exe query tokenfactory denom $GOLD --grpc-addr $GRPC --grpc-insecure --node $NODE --output json
build\orbitalisd.exe query tokenfactory denoms --limit 50 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json
Invoke-RestMethod "$REST/l1/tokenfactory/v1/denom/$GOLD"
```

Mint and burn:

```powershell
build\orbitalisd.exe tx tokenfactory mint "1000000$GOLD" $NODE0 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query bank balance $NODE0 $GOLD --node $NODE --output json
build\orbitalisd.exe tx tokenfactory burn "1000$GOLD" $NODE0 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
```

Transfer admin:

```powershell
build\orbitalisd.exe tx tokenfactory change-admin $GOLD $NODE1 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query tokenfactory denom $GOLD --node $NODE --output json
```

## DEX

Create and fund a DEX asset:

```powershell
build\orbitalisd.exe tx tokenfactory create-denom dexgold --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
$DEXGOLD = "factory/$NODE0/dexgold"
build\orbitalisd.exe tx tokenfactory mint "100000000$DEXGOLD" $NODE0 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
```

Create pool and query LP balance:

```powershell
build\orbitalisd.exe tx dex create-pool 10000000norb "10000000$DEXGOLD" --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query dex pool 1 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json
Invoke-RestMethod "$REST/l1/dex/v1/pools/1"
build\orbitalisd.exe query bank balance $NODE0 lp/1 --node $NODE --output json
```

Add liquidity, swap, and remove liquidity:

```powershell
build\orbitalisd.exe tx dex add-liquidity 1 1000000norb "1000000$DEXGOLD" 1000000 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe tx dex swap-exact-in 1 100000norb $DEXGOLD 1 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe tx dex remove-liquidity 1 1000000lp/1 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query dex pools --limit 50 --grpc-addr $GRPC --grpc-insecure --node $NODE --output json
```

Slippage examples:

```powershell
build\orbitalisd.exe tx dex add-liquidity 1 1000000norb "1000000$DEXGOLD" 1000001 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe tx dex swap-exact-in 1 100000norb $DEXGOLD 1000000 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
```

Expected rejection logs include `minted shares below minimum` or `amount out below minimum`.

## Diagnose

Run a health check while the localnet is running:

```powershell
.\scripts\localnet\health.ps1 -ValidatorCount 3
.\scripts\localnet\health.ps1 -ValidatorCount 3 -Json
```

Collect a sanitized diagnostic bundle under ignored `.work` paths:

```powershell
.\scripts\localnet\diagnostics.ps1 -ValidatorCount 3
```

Diagnostic bundles include logs, safe config files, RPC status snapshots, and health output. They exclude keyrings, `priv_validator_key.json`, `priv_validator_state.json`, and `node_key.json`.

## Troubleshooting

Use [Operator Troubleshooting Runbook](operator-troubleshooting.md) for symptom-specific commands and fixes. Quick local reminders:

- `account sequence mismatch`: wait one block or re-run the command after the previous tx is committed.
- `fee denom testtoken not accepted; use norb`: use `--fees 1000000norb`.
- `pool already exists`: query `dex pools` and use the existing `pool_id`.
- `REST ... 503`: run `.\scripts\localnet\health.ps1`; node gRPC must be reachable on `127.0.0.1:<grpc-port>`.
- `Port ... is already in use`: run `.\scripts\localnet\stop.ps1`, choose a different base port, or stop the conflicting process.
- `key not found`: check `$HOME`, `$FROM`, and `--keyring-backend test` for localnet commands.

## Required Command Checks

```powershell
build\orbitalisd.exe --help
build\orbitalisd.exe version --long --output json
build\orbitalisd.exe query fees params --grpc-addr 127.0.0.1:9090 --grpc-insecure --node tcp://127.0.0.1:26657 --output json
build\orbitalisd.exe query dex pool 1 --grpc-addr 127.0.0.1:9090 --grpc-insecure --node tcp://127.0.0.1:26657 --output json
.\tests\scripts\operator_commands_doc_test.ps1
.\tests\scripts\prototype_smoke_wrapper_test.ps1
.\tests\e2e\prototype_smoke.ps1
.\tests\e2e\prototype_acceptance.ps1
.\scripts\security\prototype-audit.ps1 -Profile Fast
.\tests\e2e\localnet_smoke.ps1
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\native_token_smoke.ps1
.\tests\e2e\tokenfactory_smoke.ps1
.\tests\e2e\fees_ante_smoke.ps1
.\tests\e2e\dex_smoke.ps1
.\tests\e2e\query_surface_smoke.ps1
```

Build a local prototype package after the checks pass:

```powershell
.\scripts\release\prototype-package.ps1 -Version prototype-local -TargetOS windows -TargetArch amd64
```

See also [Executable Prototype Contract](prototype-contract.md), [Operator Troubleshooting Runbook](operator-troubleshooting.md), [Prototype Acceptance Suite](prototype-acceptance-suite.md), [Prototype Release Package](release/prototype-package.md), [Prototype Query Surface](query-surface.md), and [Prototype Observability](observability.md).
