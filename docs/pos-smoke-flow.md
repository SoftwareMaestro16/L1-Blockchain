# Aetra PoS Smoke Flow

This document defines the operator-visible proof-of-stake smoke scenario for the Aetra prototype localnet.

The flow verifies that a multi-validator localnet exposes staking and slashing state, accepts a `naet` delegation from a funded local account, commits the delegation, and updates validator voting power. It intentionally does not simulate downtime or double-sign slashing on the running localnet; destructive slashing behavior belongs in isolated unit or integration tests.

## Supported Profiles

- Chain ID: `aetra-local-1`
- Bond denom: `naet`
- Default profile: `3` validators under `.localnet`
- Heavy profile: `5` validators under `.localnet-5`
- Delegator key: `node0` from the generated local keyring
- Delegation amount: `5000000naet`
- Fees: `1000000naet`

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
- staking params use `bond_denom = naet`
- slashing params are positive
- slashing signing-infos query returns validator records
- delegation tx from `node0` is accepted
- delegation query shows a `naet` balance
- CometBFT total voting power increases after the delegation
- wrong-denom, insufficient-funds, malformed-validator, missing-delegator, and signed-tx replay delegation attempts fail safely

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
build\aetrad.exe query staking params --node tcp://127.0.0.1:26657 --output json
```

Expected output includes:

```json
{
  "params": {
    "bond_denom": "naet"
  }
}
```

Query validators and select any bonded `operator_address`:

```powershell
build\aetrad.exe query staking validators --node tcp://127.0.0.1:26657 --output json
```

Expected output includes `BOND_STATUS_BONDED` validators with `AE...` operator addresses. Do not assume only one validator or a fixed order.

Show the funded delegator account:

```powershell
build\aetrad.exe keys show node0 -a --home .localnet\node0\aetrad --keyring-backend test
```

Attempt direct delegation from `node0` to a bonded validator. This is a negative
check: normal user staking must use the official liquid staking pool/index, so
the direct validator choice is rejected.

```powershell
build\aetrad.exe tx staking delegate <AE...validator> 5000000naet --from node0 --home .localnet\node0\aetrad --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
```

Expected output includes a `txhash`. Query the transaction until it returns a
non-zero `code` and the pool-only policy error:

```powershell
build\aetrad.exe query tx <txhash> --node tcp://127.0.0.1:26657 --output json
```

Verify no direct delegation was created:

```powershell
build\aetrad.exe query staking delegation <AE...delegator> <AE...validator> --node tcp://127.0.0.1:26657 --output json
```

Expected output is not found or empty; it must not include a delegated `naet`
balance:

```json
{
  "code": 5,
  "message": "not found"
}
```

Query slashing params and signing infos:

```powershell
build\aetrad.exe query slashing params --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query slashing signing-infos --node tcp://127.0.0.1:26657 --output json
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

- Staking, fees, and bank balances in this flow must use `naet`.
- Wrong-denom, insufficient-funds, malformed-validator, missing-delegator, and signed-tx replay delegation paths are covered by app tests and the localnet smoke without changing staking module behavior.
- The e2e flow checks aggregate CometBFT voting power, so it is stable across 3-validator and 5-validator profiles and does not depend on validator ordering.
- Localnet scripts must not print mnemonics or write helper keys outside ignored node homes.
- Slashing smoke is read-only on localnet: query params and signing infos only.

## Required Checks

```powershell
go test ./app
go test ./...
go vet ./...
buf lint
go build -o build/aetrad.exe ./cmd/l1d
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```
