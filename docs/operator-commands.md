# Operator Commands

This document is the prototype command runbook for `orbitalisd`.

It is scoped to local prototype operation. Commands that use `--keyring-backend test` are for ignored localnet homes only. Do not use the test keyring, generated local validator keys, or localnet mnemonics for public networks.

## Build And Version

Use the repo-local Go toolchain when it exists:

```powershell
$env:PATH = "$PWD\.work\tools\go1.25.11\go\bin;$env:PATH"
go build -o build\orbitalisd.exe ./cmd/l1d
```

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

Release-like builds can override the default metadata with ldflags:

```powershell
$commit = git rev-parse HEAD
$date = Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ"
$dirty = if ((git status --porcelain) -eq $null) { "false" } else { "true" }
go build -o build\orbitalisd.exe -ldflags "-X github.com/sovereign-l1/l1/cmd/l1d/cmd.appVersion=prototype -X github.com/sovereign-l1/l1/cmd/l1d/cmd.gitCommit=$commit -X github.com/sovereign-l1/l1/cmd/l1d/cmd.buildDate=$date -X github.com/sovereign-l1/l1/cmd/l1d/cmd.dirty=$dirty" ./cmd/l1d
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
build\orbitalisd.exe query fees params --node $NODE --output json
Invoke-RestMethod "$REST/l1/fees/v1/params"
```

Staking and slashing:

```powershell
build\orbitalisd.exe query staking params --node $NODE --output json
build\orbitalisd.exe query staking validators --node $NODE --output json
build\orbitalisd.exe query slashing params --node $NODE --output json
build\orbitalisd.exe query slashing signing-infos --node $NODE --output json
```

## Bank Tx

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
build\orbitalisd.exe query tokenfactory denom $GOLD --node $NODE --output json
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
build\orbitalisd.exe query dex pool 1 --node $NODE --output json
build\orbitalisd.exe query bank balance $NODE0 lp/1 --node $NODE --output json
```

Add liquidity, swap, and remove liquidity:

```powershell
build\orbitalisd.exe tx dex add-liquidity 1 1000000norb "1000000$DEXGOLD" 1000000 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe tx dex swap-exact-in 1 100000norb $DEXGOLD 1 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe tx dex remove-liquidity 1 1000000lp/1 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query dex pools --node $NODE --output json
```

Slippage examples:

```powershell
build\orbitalisd.exe tx dex add-liquidity 1 1000000norb "1000000$DEXGOLD" 1000001 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe tx dex swap-exact-in 1 100000norb $DEXGOLD 1000000 --from $FROM --home $HOME --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
```

Expected rejection logs include `minted shares below minimum` or `amount out below minimum`.

## Troubleshooting

- `account sequence mismatch`: wait one block or re-run the command after the previous tx is committed.
- `fee denom testtoken not accepted; use norb`: use `--fees 1000000norb`.
- `pool already exists`: query `dex pools` and use the existing `pool_id`.
- `REST ... 503`: use CLI/gRPC/RPC first; the local REST gateway may lag during startup.
- `Port ... is already in use`: run `.\scripts\localnet\stop.ps1`, choose a different base port, or stop the conflicting process.
- `key not found`: check `$HOME`, `$FROM`, and `--keyring-backend test` for localnet commands.

## Required Command Checks

```powershell
build\orbitalisd.exe --help
build\orbitalisd.exe version --long --output json
build\orbitalisd.exe query fees params --node tcp://127.0.0.1:26657 --output json
build\orbitalisd.exe query dex pool 1 --node tcp://127.0.0.1:26657 --output json
.\tests\e2e\localnet_smoke.ps1
.\tests\e2e\native_token_smoke.ps1
.\tests\e2e\fees_ante_smoke.ps1
.\tests\e2e\dex_smoke.ps1
```
