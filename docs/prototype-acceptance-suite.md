> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Prototype Acceptance Suite

This document defines the one-command acceptance check for the Aetra L1 prototype.

The suite proves that a local prototype can be built from the repository, initialized from zero, started with multiple validators, produce blocks, accept signed transactions, update state, expose CLI/gRPC/REST queries, and stop cleanly.

## Command

Default PR-friendly smoke profile:

```powershell
.\tests\e2e\prototype_smoke.ps1
```

Equivalent configurable acceptance command:

```powershell
.\tests\e2e\prototype_acceptance.ps1
```

Full/manual profile:

```powershell
.\tests\e2e\prototype_acceptance.ps1 -Profile Full
.\tests\e2e\prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5
```

Use an already-built binary:

```powershell
.\tests\e2e\prototype_acceptance.ps1 -SkipBuild -Binary build\aetrad.exe
```

Shift ports or choose a specific node endpoint:

```powershell
.\tests\e2e\prototype_acceptance.ps1 -BaseRPCPort 27657 -BaseGRPCPort 9190 -BaseRESTPort 1417 -Node tcp://127.0.0.1:27657
```

## Parameters

- `-Profile Smoke|Full`: `Smoke` is the default PR profile; `Full` adds heavier restart/slippage checks.
- `-ValidatorCount`: defaults to `3`; use `5` for the scale profile.
- `-MinHeight`: first required block height before tx flows start.
- `-TimeoutSeconds`: shared wait timeout for startup, tx inclusion, and health checks.
- `-OutputDir`: ignored localnet directory, default `.localnet`.
- `-Binary`: binary path, default `build\aetrad.exe`.
- `-SkipBuild`: reuse an existing binary instead of running `go build`.
- `-Node`: CLI RPC endpoint; defaults to node0 from the port profile.
- `-KeepLogsOnFailure`: preserve the localnet output directory after a failure. A diagnostic bundle is collected either way.

## Smoke Profile

`Smoke` runs these steps on one fresh localnet:

1. Build `aetrad` unless `-SkipBuild` is set.
2. Stop any previous matching localnet processes.
3. Reset/init localnet through `scripts\localnet\init.ps1`.
4. Validate genesis through `scripts\localnet\validate-genesis.ps1`.
5. Start validators and wait for height, RPC, peers, REST, and gRPC health.
6. Query RPC status, block, native token metadata, fees params, and REST node info.
7. Send `naet` by bank tx and verify balance change.
8. Reject a tx with `testtoken` fee.
9. Create a tokenfactory denom, mint it, and query it through gRPC/REST.
10. Create a DEX pool, swap exact input, and verify output balance plus pool query.
11. Query staking/slashing state, delegate `naet`, and verify delegation and voting power update.
12. Stop localnet.

## Full Profile

`Full` includes all smoke steps and adds:

- DEX slippage failure for excessive `min_amount_out`.
- Stop/start restart check that proves height continues increasing after restart.

The 5-validator full profile is intended for manual or nightly runs, not every PR.

## Failure Behavior

Failures return a non-zero exit code and print the failing step. Before stopping the network, the suite calls:

```powershell
.\scripts\localnet\diagnostics.ps1
```

The bundle is written under ignored `.work\diagnostics\acceptance-*` and includes node logs, safe config files, RPC status, peers, validators, and health output. It excludes keyring data, `priv_validator_key.json`, `priv_validator_state.json`, and `node_key.json`.

If `-KeepLogsOnFailure` is not set, the suite resets the localnet output directory after the diagnostic bundle is collected. The bundle remains in `.work`.

## Security Notes

- The suite uses local `--keyring-backend test` only under ignored localnet homes.
- It does not print mnemonics, private keys, validator keys, node keys, or raw environment secrets.
- Destructive operations go through localnet scripts that validate resolved paths stay inside the repository and are not the repository root.
- The Cosmos consensus-risk checks covered here are integration smoke checks: deterministic startup/export assumptions, valid `naet` fee/staking denom, authorized tokenfactory mint path, DEX reserve movement, and no panic on normal query/tx flows. Deeper ABCI/non-determinism review remains part of the module security baseline.

## CI Contract

Recommended PR check:

```powershell
.\tests\e2e\prototype_smoke.ps1
```

Recommended manual/nightly check:

```powershell
.\tests\e2e\prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5
```
