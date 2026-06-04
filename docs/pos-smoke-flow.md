# Orbitalis PoS Smoke Flow

This document defines the operator-visible proof-of-stake smoke scenario for the Orbitalis prototype localnet.

The flow verifies that a multi-validator localnet exposes staking and slashing state, accepts a `norb` delegation from a funded local account, commits the delegation, and updates validator voting power. It intentionally does not simulate downtime or double-sign slashing on the running localnet; destructive slashing behavior belongs in isolated unit or integration tests.

## Supported Profiles

- Chain ID: `orbitalis-local-1`
- Bond denom: `norb`
- Default profile: `3` validators under `.localnet`
- Heavy profile: `5` validators under `.localnet-5`
- Delegator key: `node0` from the generated local keyring
- Delegation amount: `5000000norb`
- Fees: `1000000norb`

Generated keys and node homes live under ignored localnet directories. They are prototype helper accounts only and have no production privileges.

## One-Command Smoke

Run the default 3-validator PoS smoke:

```powershell
.\tests\e2e\pos_smoke.ps1
```

Run the 5-validator profile:

```powershell
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Expected result:

- localnet reaches the requested height
- CometBFT and staking validator queries return the requested validator count
- every staking validator is bonded
- staking params use `bond_denom = norb`
- slashing params are positive
- slashing signing-infos query returns validator records
- delegation tx from `node0` is accepted
- delegation query shows a `norb` balance
- CometBFT total voting power increases after the delegation
- wrong-denom, insufficient-funds, malformed-validator, and missing-delegator delegation attempts fail safely

Recovery:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
.\tests\e2e\pos_smoke.ps1
```

Use `-OutputDir .localnet-5 -ValidatorCount 5` for the 5-validator cleanup and rerun.

## Manual CLI Flow

Initialize and start the default localnet:

```powershell
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\start.ps1 -NoInit -Wait
```

Query staking params:

```powershell
build\orbitalisd.exe query staking params --node tcp://127.0.0.1:26657 --output json
```

Expected output includes:

```json
{
  "params": {
    "bond_denom": "norb"
  }
}
```

Query validators and select any bonded `operator_address`:

```powershell
build\orbitalisd.exe query staking validators --node tcp://127.0.0.1:26657 --output json
```

Expected output includes `BOND_STATUS_BONDED` validators with `orbvaloper...` operator addresses. Do not assume only one validator or a fixed order.

Show the funded delegator account:

```powershell
build\orbitalisd.exe keys show node0 -a --home .localnet\node0\orbitalisd --keyring-backend test
```

Delegate from `node0` to a bonded validator:

```powershell
build\orbitalisd.exe tx staking delegate <orbvaloper...> 5000000norb --from node0 --home .localnet\node0\orbitalisd --chain-id orbitalis-local-1 --keyring-backend test --fees 1000000norb --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
```

Expected output includes a `txhash`. Query the transaction until it returns `code = 0`:

```powershell
build\orbitalisd.exe query tx <txhash> --node tcp://127.0.0.1:26657 --output json
```

Verify the delegation:

```powershell
build\orbitalisd.exe query staking delegation <orb1...delegator> <orbvaloper...> --node tcp://127.0.0.1:26657 --output json
```

Expected output includes:

```json
{
  "delegation_response": {
    "balance": {
      "denom": "norb",
      "amount": "5000000"
    }
  }
}
```

Query slashing params and signing infos:

```powershell
build\orbitalisd.exe query slashing params --node tcp://127.0.0.1:26657 --output json
build\orbitalisd.exe query slashing signing-infos --node tcp://127.0.0.1:26657 --output json
```

Expected output:

- `signed_blocks_window` is positive
- `min_signed_per_window` is positive
- `slash_fraction_double_sign` is positive
- `slash_fraction_downtime` is positive
- signing infos include validator consensus records

Stop the network:

```powershell
.\scripts\localnet\stop.ps1
```

## Audit Notes

- Staking, fees, and bank balances in this flow must use `norb`.
- Wrong-denom, insufficient-funds, malformed-validator, and missing-delegator delegation paths are covered by app tests and the localnet smoke without changing staking module behavior. Sequence/replay protection stays in the shared `x/auth` boundary and is not modeled by this PoS-specific smoke.
- The e2e flow checks aggregate CometBFT voting power, so it is stable across 3-validator and 5-validator profiles and does not depend on validator ordering.
- Localnet scripts must not print mnemonics or write helper keys outside ignored node homes.
- Slashing smoke is read-only on localnet: query params and signing infos only.

## Required Checks

```powershell
go test ./app
go test ./...
go vet ./...
buf lint
go build -o build/orbitalisd.exe ./cmd/l1d
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```
