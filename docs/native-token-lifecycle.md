# Orbitalis Native Token Lifecycle

This document defines the prototype lifecycle contract for the native Orbitalis token.

## Contract

- Base denom: `norb`
- Display denom: `ORB`
- Name: `Orbitalis`
- Symbol: `ORB`
- Display exponent: `9`
- Conversion: `1 ORB = 1000000000 norb`
- Staking denom: `norb`
- Fee denom: `norb`
- Mint denom: `norb`

Operators and scripts must use `norb` for balances, fees, staking, and transaction amounts. `ORB` is display metadata only; it is not used as a transaction denom.

The local bootstrap profile also gives validator accounts `testtoken`. That denom is a local test asset for module and DEX experiments only. It is not a native token, staking denom, fee denom, or display unit for `ORB`.

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

- bank metadata exposes base `norb`, display `ORB`, symbol `ORB`, exponent `9`
- total supply for `norb` is positive
- node0 and node1 have positive `norb` balances
- staking params use `bond_denom = norb`
- fees params allow only `norb`
- mint params use `mint_denom = norb`
- `bank send` from node0 to node1 succeeds with `--fees 1000000norb`
- node1 balance increases by the sent `norb` amount

Recovery:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
.\tests\e2e\native_token_smoke.ps1
```

## Manual CLI Flow

Initialize and start localnet:

```powershell
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\start.ps1 -NoInit -Wait
```

Query metadata:

```powershell
build\orbitalisd.exe query bank denom-metadata norb --node tcp://127.0.0.1:26657 --output json
```

Expected fields:

```json
{
  "metadata": {
    "base": "norb",
    "display": "ORB",
    "name": "Orbitalis",
    "symbol": "ORB",
    "denom_units": [
      { "denom": "norb", "exponent": 0 },
      { "denom": "ORB", "exponent": 9 }
    ]
  }
}
```

Query supply:

```powershell
build\orbitalisd.exe query bank total-supply-of norb --node tcp://127.0.0.1:26657 --output json
```

Query balances:

```powershell
$node0 = build\orbitalisd.exe keys show node0 -a --home .localnet\node0\orbitalisd --keyring-backend test
$node1 = build\orbitalisd.exe keys show node1 -a --home .localnet\node1\orbitalisd --keyring-backend test
build\orbitalisd.exe query bank balance $node0 norb --node tcp://127.0.0.1:26657 --output json
build\orbitalisd.exe query bank balance $node1 norb --node tcp://127.0.0.1:26657 --output json
```

Send `norb` and pay fees in `norb`:

```powershell
build\orbitalisd.exe tx bank send node0 $node1 1000norb --home .localnet\node0\orbitalisd --chain-id orbitalis-local-1 --keyring-backend test --fees 1000000norb --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
```

Delegate `norb` to a bonded validator:

```powershell
$validator = (build\orbitalisd.exe query staking validators --node tcp://127.0.0.1:26657 --output json | ConvertFrom-Json).validators[0].operator_address
build\orbitalisd.exe tx staking delegate $validator 5000000norb --from node0 --home .localnet\node0\orbitalisd --chain-id orbitalis-local-1 --keyring-backend test --fees 1000000norb --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
build\orbitalisd.exe query staking delegation $node0 $validator --node tcp://127.0.0.1:26657 --output json
```

Query fee, staking, and mint params:

```powershell
build\orbitalisd.exe query fees params --node tcp://127.0.0.1:26657 --output json
build\orbitalisd.exe query staking params --node tcp://127.0.0.1:26657 --output json
build\orbitalisd.exe query mint params --node tcp://127.0.0.1:26657 --output json
```

Expected values:

- `fees.params.allowed_fee_denoms = ["norb"]`
- `staking.params.bond_denom = "norb"`
- `mint.params.mint_denom = "norb"`

## Audit Notes

- Native metadata is generated from `app/params.NativeTokenMetadata()` and injected into default genesis and testnet genesis.
- `x/fees` v1 accepts only `norb`; wrong-denom fees are rejected by the custom ante decorator.
- `x/staking` and `x/mint` use `norb` in genesis.
- Tokenfactory denoms use `factory/<admin>/<subdenom>` and reject subdenoms that directly spoof native names: `norb`, `ORB`, or `Orbitalis`.
- DEX LP tokens use `lp/<pool_id>` and are not display aliases for `ORB`.
- `testtoken` is allowed only as a local bootstrap test asset and must not appear in fee, staking, mint, or native-token examples.

## Required Checks

```powershell
go test ./app ./app/params ./x/fees/... ./x/tokenfactory/... ./x/dex/...
go test ./...
go vet ./...
buf lint
go build -o build/orbitalisd.exe ./cmd/l1d
.\tests\e2e\native_token_smoke.ps1
.\tests\e2e\native_token_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\pos_smoke.ps1
```
