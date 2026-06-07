> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# DEX Prototype End-To-End Flow

This document defines the prototype user flow for the Aetra DEX module.

## Contract

- AMM type: constant-product pool
- Pool fee: `30` bps
- Native side: `naet`
- Prototype external asset: a factory denom created by `x/tokenfactory`
- LP denom format: `lp/<pool_id>`
- Default first pool ID: `1`
- Prototype fee coin for all tx examples: `1000000naet`

`lp/<pool_id>` is a bank denom minted by the `dex` module. Users see LP balances with the normal bank balance query:

```powershell
build\aetrad.exe query bank balance $node0 lp/1 --node tcp://127.0.0.1:26657 --output json
```

## One-Command Smoke

Run the default 3-validator flow:

```powershell
.\tests\e2e\dex_smoke.ps1
```

Run the 5-validator profile:

```powershell
.\tests\e2e\dex_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Expected result:

- localnet reaches the requested height with the expected validator set
- node0 creates `factory/<node0>/dexgold`
- node0 mints the factory asset to itself
- `tx dex create-pool` creates pool `1` with `naet` and the factory denom
- `query dex pool 1` returns reserves, total shares, and `lp/1`
- duplicate pair pool creation is rejected
- `add-liquidity` succeeds with a realistic `min_shares`
- `add-liquidity` fails when `min_shares` is too high
- `swap-exact-in` succeeds and the output balance increases
- `swap-exact-in` fails when `min_amount_out` is too high
- `remove-liquidity` burns LP shares and decreases reserves
- wrong liquidity denoms and wrong LP denoms are rejected

Recovery:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
.\tests\e2e\dex_smoke.ps1
```

## Manual CLI Flow

Initialize and start localnet:

```powershell
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\start.ps1 -NoInit -Wait
```

Load node0:

```powershell
$node = "tcp://127.0.0.1:26657"
$home = ".localnet\node0\aetrad"
$node0 = build\aetrad.exe keys show node0 -a --home $home --keyring-backend test
$denom = "factory/$node0/dexgold"
```

Create and fund the factory denom:

```powershell
build\aetrad.exe tx tokenfactory create-denom dexgold --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
build\aetrad.exe tx tokenfactory mint "100000000$denom" $node0 --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
build\aetrad.exe query bank balance $node0 $denom --node $node --output json
```

Create a pool:

```powershell
build\aetrad.exe tx dex create-pool 10000000naet "10000000$denom" --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
build\aetrad.exe query dex pool 1 --node $node --output json
build\aetrad.exe query bank balance $node0 lp/1 --node $node --output json
```

Expected pool fields:

```json
{
  "pool": {
    "id": "1",
    "denom0": "factory/<node0>/dexgold",
    "denom1": "naet",
    "reserve0": "10000000",
    "reserve1": "10000000",
    "total_shares": "10000000",
    "lp_denom": "lp/1"
  }
}
```

Add liquidity with slippage protection:

```powershell
build\aetrad.exe tx dex add-liquidity 1 1000000naet "1000000$denom" 1000000 --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
```

Expected failure when `min_shares` is too high:

```powershell
build\aetrad.exe tx dex add-liquidity 1 1000000naet "1000000$denom" 1000001 --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
```

Expected rejection log includes:

```text
minted shares below minimum
```

Swap exact amount in:

```powershell
build\aetrad.exe tx dex swap-exact-in 1 100000naet $denom 1 --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
build\aetrad.exe query bank balance $node0 $denom --node $node --output json
```

Expected failure when `min_amount_out` is too high:

```powershell
build\aetrad.exe tx dex swap-exact-in 1 100000naet $denom 1000000 --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
```

Expected rejection log includes:

```text
amount out below minimum
```

Remove liquidity:

```powershell
build\aetrad.exe tx dex remove-liquidity 1 1000000lp/1 --from node0 --home $home --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node $node --output json
build\aetrad.exe query dex pool 1 --node $node --output json
build\aetrad.exe query bank balance $node0 lp/1 --node $node --output json
```

## Audit Notes

- Duplicate pair pools are rejected with a bounded pair index. `create-pool` checks `denom0/denom1` directly instead of scanning all pools.
- Swap, add liquidity, and remove liquidity operate by `pool_id`; they do not scan all pools on the transaction path.
- Genesis validation rejects duplicate pool IDs, duplicate pairs, non-canonical denoms, invalid LP denoms, zero reserves, zero shares, and malformed integer state.
- Keeper tests cover corrupted pool state without panic, wrong liquidity denoms, wrong LP denom, reserve/module-account balance consistency, LP supply consistency, tiny swap rounding, and slippage guards.
- DEX reserves are checked against the `dex` module account balance in keeper tests.
- LP denoms use `lp/<pool_id>` and are not aliases for native `naet` or display `AET`.
- Query `dex pools` uses bounded `next_key` pagination with default limit `50` and max limit `100`.
- There is no external price oracle, sandwich protection, or production routing. Multi-hop swaps and exchange-grade routing are out of scope for this prototype.

## Required Checks

```powershell
go test ./x/dex/...
go test ./...
go vet ./...
buf lint
go build -o build/aetrad.exe ./cmd/l1d
.\tests\e2e\dex_smoke.ps1
.\tests\e2e\dex_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```
