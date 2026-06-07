> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Aetra Native Token Lifecycle

This document defines the lifecycle contract for the native Aetra token.

## Contract

- Base denom: `naet`
- Display denom: `AET`
- Name: `Aetra`
- Symbol: `AET`
- Display exponent: `9`
- Conversion: `1 AET = 1000000000 naet`
- Staking denom: `naet`
- Fee denom: `naet`
- Mint denom: `naet`
- Supply: uncapped PoS supply through configured inflation and staking rewards

Operators and scripts must use `naet` for balances, fees, staking, and
transaction amounts. `AET` is display metadata only; it is not a transaction
denom.

The local bootstrap profile can give validator accounts `testtoken`. That denom
is a local test asset for module and DEX experiments only. It is not a native
token, staking denom, fee denom, mint denom, or display unit for `AET`.

## One-Command Smoke

Run the default 3-validator native token smoke:

```powershell
.\tests\e2e\native_token_smoke.ps1
```

Run the 5-validator profile:

```powershell
.\tests\e2e\native_token_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Expected result:

- bank metadata exposes base `naet`, display `AET`, symbol `AET`, exponent `9`
- total supply for `naet` is positive
- node0 and node1 have positive `naet` balances
- staking params use `bond_denom = naet`
- fees params allow only `naet`
- mint params use `mint_denom = naet`
- mint params keep `max_supply = 0`, meaning uncapped PoS issuance
- `bank send` from node0 to node1 succeeds with `--fees 1000000naet`
- node1 balance increases by the sent `naet` amount

Recovery:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
.\tests\e2e\native_token_smoke.ps1
```

## Manual CLI Flow

Initialize and start localnet:

```powershell
.\scripts\localnet\init.ps1 -ChainId aetra-local-1
.\scripts\localnet\validate-genesis.ps1 -ChainId aetra-local-1
.\scripts\localnet\start.ps1 -ChainId aetra-local-1 -NoInit -Wait
```

Query metadata:

```powershell
build\aetrad.exe query bank denom-metadata naet --node tcp://127.0.0.1:26657 --output json
```

Expected fields:

```json
{
  "metadata": {
    "base": "naet",
    "display": "AET",
    "name": "Aetra",
    "symbol": "AET",
    "denom_units": [
      { "denom": "naet", "exponent": 0 },
      { "denom": "AET", "exponent": 9 }
    ]
  }
}
```

Query supply and balances:

```powershell
build\aetrad.exe query bank total-supply-of naet --node tcp://127.0.0.1:26657 --output json
$node0 = build\aetrad.exe keys show node0 -a --home .localnet\node0\aetrad --keyring-backend test
$node1 = build\aetrad.exe keys show node1 -a --home .localnet\node1\aetrad --keyring-backend test
build\aetrad.exe query bank balance $node0 naet --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query bank balance $node1 naet --node tcp://127.0.0.1:26657 --output json
```

Send `naet` and pay fees in `naet`:

```powershell
build\aetrad.exe tx bank send node0 $node1 1000naet --home .localnet\node0\aetrad --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
```

Delegate `naet` to a bonded validator:

```powershell
$validator = (build\aetrad.exe query staking validators --node tcp://127.0.0.1:26657 --output json | ConvertFrom-Json).validators[0].operator_address
build\aetrad.exe tx staking delegate $validator 5000000naet --from node0 --home .localnet\node0\aetrad --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query staking delegation $node0 $validator --node tcp://127.0.0.1:26657 --output json
```

Query fee, staking, and mint params:

```powershell
build\aetrad.exe query fees params --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query staking params --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query mint params --node tcp://127.0.0.1:26657 --output json
```

Expected values:

- `fees.params.allowed_fee_denoms = ["naet"]`
- `staking.params.bond_denom = "naet"`
- `mint.params.mint_denom = "naet"`
- `mint.params.max_supply = "0"`

## Audit Notes

- Native metadata is generated from `app/params.NativeTokenMetadata()` and
  injected into default genesis and testnet genesis.
- `x/fees` v1 accepts only `naet`; wrong-denom fees are rejected by the custom
  ante decorator.
- `x/staking` and `x/mint` use `naet` in genesis.
- `x/mint` `max_supply = 0` means AET has no fixed max supply.
- Inflation/reward params must pass SDK mint param validation and remain
  governance-bounded.
- Tokenfactory denoms use `factory/<admin>/<subdenom>` and reject subdenoms that
  directly spoof native names: `naet`, `AET`, or `Aetra`.
- DEX LP tokens use `lp/<pool_id>` and are not display aliases for `AET`.
- `testtoken` is allowed only as a local bootstrap test asset and must not
  appear in fee, staking, mint, or native-token examples.

## Required Checks

```powershell
go test ./app ./app/params ./x/fees/... ./x/tokenfactory/... ./x/dex/...
go test ./...
go vet ./...
buf lint
go build -o build/aetrad.exe ./cmd/l1d
.\tests\e2e\native_token_smoke.ps1
.\tests\e2e\native_token_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```
