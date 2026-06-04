# Tokenfactory Lifecycle

This document defines the prototype tokenfactory user flow. It is localnet/testnet functionality for factory assets and must not be confused with the native `ORB/norb` token.

## Contract

- Factory denom format: `factory/<admin-address>/<subdenom>`
- Default local example subdenom: `gold`
- Admin controls create-time metadata, mint, burn from own account, and admin transfer.
- Factory token bank metadata uses the full factory denom as base/display/symbol.
- Factory denoms must not spoof native `norb`, `ORB`, `Orbitalis`, or LP denoms.
- Query-all is capped by the prototype `MaxQueryDenoms` limit; production pagination is future hardening work.

## One-Command Smoke

Run the default 3-validator tokenfactory smoke:

```powershell
.\tests\e2e\tokenfactory_smoke.ps1
```

Run the 5-validator profile:

```powershell
.\tests\e2e\tokenfactory_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Expected result:

- fresh localnet starts with no factory denoms
- node0 creates `factory/<node0>/gold`
- CLI and REST denom queries return node0 as admin
- factory bank metadata does not use native `norb` or `ORB`
- duplicate `gold` and native-spoof `norb` subdenoms are rejected
- node0 mints `gold`, then bank balance and supply increase
- burn from another account is rejected
- node0 burns own `gold`, then bank balance and supply decrease
- invalid admin address is rejected
- admin transfers to node1
- old admin mint is rejected
- new admin mints successfully

## Manual CLI Flow

Initialize and start localnet:

```powershell
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\start.ps1 -NoInit -Wait
```

Set local variables:

```powershell
$NODE = "tcp://127.0.0.1:26657"
$REST = "http://127.0.0.1:1317"
$CHAIN_ID = "orbitalis-local-1"
$KEYRING = "test"
$HOME0 = ".localnet\node0\orbitalisd"
$HOME1 = ".localnet\node1\orbitalisd"
$NODE0 = build\orbitalisd.exe keys show node0 -a --home $HOME0 --keyring-backend $KEYRING
$NODE1 = build\orbitalisd.exe keys show node1 -a --home $HOME1 --keyring-backend $KEYRING
$GOLD = "factory/$NODE0/gold"
$FEES = "1000000norb"
```

Create and query:

```powershell
build\orbitalisd.exe tx tokenfactory create-denom gold --from node0 --home $HOME0 --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query tokenfactory denom $GOLD --node $NODE --output json
Invoke-RestMethod "$REST/l1/tokenfactory/v1/denom/$GOLD"
build\orbitalisd.exe query bank denom-metadata $GOLD --node $NODE --output json
```

Mint, query balance and supply, then burn:

```powershell
build\orbitalisd.exe tx tokenfactory mint "1000000$GOLD" $NODE0 --from node0 --home $HOME0 --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query bank balance $NODE0 $GOLD --node $NODE --output json
build\orbitalisd.exe query bank total-supply-of $GOLD --node $NODE --output json
build\orbitalisd.exe tx tokenfactory burn "250000$GOLD" $NODE0 --from node0 --home $HOME0 --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
```

Transfer admin and prove authorization:

```powershell
build\orbitalisd.exe tx tokenfactory change-admin $GOLD $NODE1 --from node0 --home $HOME0 --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe query tokenfactory denom $GOLD --node $NODE --output json
build\orbitalisd.exe tx tokenfactory mint "1$GOLD" $NODE0 --from node0 --home $HOME0 --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
build\orbitalisd.exe tx tokenfactory mint "100$GOLD" $NODE1 --from node1 --home $HOME1 --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees $FEES --yes --broadcast-mode sync --node $NODE --output json
```

The old-admin mint command must fail with an authorization error; the new-admin mint command must commit.

## Audit Notes

- `MsgCreateDenom` validates creator address, bounded subdenom syntax, duplicate denom state, and native-name spoofing.
- `MsgMint`, `MsgBurn`, and `MsgChangeAdmin` require the current admin signer.
- `MsgBurn` only burns from the signer account, so an admin cannot burn another account's factory balance.
- Bank keeper mint/burn/send errors are returned directly; local supply bookkeeping is not duplicated.
- Factory bank metadata uses the full factory denom and must not replace native `norb` metadata.
- The e2e smoke checks bank balance and `total-supply-of` after mint and burn.

## Required Checks

```powershell
go test ./x/tokenfactory/...
.\tests\e2e\tokenfactory_smoke.ps1
.\tests\e2e\tokenfactory_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\scripts\security\prototype-audit.ps1 -Profile Fast
```
