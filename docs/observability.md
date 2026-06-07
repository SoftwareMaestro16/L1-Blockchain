# Prototype Observability

This runbook defines the minimum diagnostic surface for local Aetra prototype nodes.

## Health Checks

Default 3-validator localnet:

```powershell
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\start.ps1 -NoInit
.\scripts\localnet\health.ps1 -ValidatorCount 3
```

The health command checks:

- CometBFT RPC `/status`
- block height increasing
- CometBFT validator set count and voting power
- peer availability for multi-validator localnet
- REST gateway `/cosmos/base/tendermint/v1beta1/blocks/latest`
- gRPC TCP readiness
- tracked `aetrad` process ids
- generated telemetry mode from node `app.toml`
- recent redacted node log tails

Use `-Json` when a script or CI job needs machine-readable output:

```powershell
.\scripts\localnet\health.ps1 -ValidatorCount 3 -Json
```

## Logs

Localnet writes one stdout and stderr file per node:

```text
.localnet\logs\node0.out.log
.localnet\logs\node0.err.log
.localnet\logs\node1.out.log
.localnet\logs\node1.err.log
.localnet\logs\node2.out.log
.localnet\logs\node2.err.log
```

Normal state:

- blocks are finalized and committed regularly
- validator set count matches the bootstrap profile
- multi-validator peers are connected
- REST and gRPC endpoints answer after startup

Unhealthy state:

- RPC port is not listening
- height does not increase
- validator set count is wrong or voting power is zero
- peer count is zero in a multi-validator profile
- REST returns 503 after gRPC is ready
- tx logs show fee policy rejection, sequence mismatch, insufficient funds, or malformed input
- consensus logs contain `CONSENSUS FAILURE` or disk write errors

## Diagnostic Bundle

Collect a local diagnostic bundle in an ignored `.work` path:

```powershell
.\scripts\localnet\diagnostics.ps1 -ValidatorCount 3
```

The bundle includes:

- localnet manifest
- recent redacted node logs
- safe redacted config files: `app.toml`, `config.toml`, `genesis.json`
- RPC snapshots: `status`, `net_info`, `validators`
- health output
- process list and recent-log JSON

The bundle excludes keyring directories, `priv_validator_key.json`, `priv_validator_state.json`, and `node_key.json`. Log and config snapshots redact common mnemonic, private key, password, token, seed, and secret patterns. Do not attach bundles from non-local environments without a separate secret review.

## Minimal Metrics Policy

SDK telemetry mode is reported by `health.ps1` from generated app config. CometBFT Prometheus export is not a release gate for this prototype profile. If Prometheus is enabled in a later profile, labels must stay bounded: do not add user addresses, tx hashes, denoms, pool ids, or other unbounded values as labels without an explicit metrics cardinality review.

## Quick Triage

For symptom-specific operator commands and fixes, see [Operator Troubleshooting Runbook](operator-troubleshooting.md).

- `Port ... is already in use`: run `.\scripts\localnet\stop.ps1` or shift base ports.
- `REST ... 503`: check `.\scripts\localnet\health.ps1`; localnet gRPC address must be `127.0.0.1:<grpc-port>`.
- `height does not increase`: inspect `node*.out.log` for consensus timeout, disk, or validator errors.
- `fee denom ... not accepted`: prototype tx examples must use `--fees 1000000naet`.
- `account sequence mismatch`: wait one block and retry after the previous tx commits.
- `There is not enough space on the disk`: stop localnet and clean generated `.localnet*` and `.work\gocache` if needed.
