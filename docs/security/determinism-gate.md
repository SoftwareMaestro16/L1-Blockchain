# Deterministic Execution Gate

This gate blocks release when custom consensus code has untriaged nondeterminism.

## Scope

Fast local scope:

- `app`: ABCI wiring, genesis/export, keeper assembly, vote-extension hooks
- `x/*/keeper`: msg servers, ante decorators, keepers, query/list iterators
- `x/*/types`: params, genesis validation, deterministic integer math helpers
- `cmd/l1d/cmd`: CLI/testnet/speedtest findings are scanned and downgraded only when they do not affect AppHash

The gate scans for wall clock use, randomness, floats, unordered map iteration, goroutines, `select`, platform-dependent `int`/`uint`, external API calls, and `panic`.

## Command

```powershell
.\scripts\security\determinism-gate.ps1
```

Outputs are written under `.work\security\determinism-gate-*`:

- `summary.md`: severity/status table with AppHash or consensus-impact triage
- `determinism-findings.json`: machine-readable findings

`-Strict` fails on any untriaged finding. Default mode fails on untriaged `High` or `Critical` findings and permits documented `Low`/`Medium` prototype triage.

## Current Triage

| Finding | Severity | Consensus impact |
| --- | --- | --- |
| `crypto/rand` in `app/abci.go` dummy vote extension | Medium | Does not write app state or AppHash. Replace or disable before public validators. |
| `cmttime.Now()` in `cmd/l1d/cmd/testnet_genesis.go` | Low | Local genesis timestamp. One init run writes identical genesis to all local nodes. |
| `time.Now()` and `math/rand` in `cmd/l1d/cmd/speedtest.go` | Low | CLI benchmark only; not consensus execution. |
| `panic` in app/module wiring and export helpers | Medium | Startup/genesis/export paths only. Tx, ante, and query paths must return errors for malformed input. |
| `app/test_helpers.go` nondeterminism | Low | Test/dev helper surface used by tests; not an operator command or consensus transition. |

## Regression Coverage

- `app/determinism_test.go`: deterministic default genesis JSON, deterministic export for identical state, and identical export after the same empty block sequence.
- `scripts/security/determinism-gate.ps1`: static grep/classification for consensus-risk patterns.
- `tests/scripts/determinism_gate_test.ps1`: verifies the local gate reports no untriaged High/Critical findings.

Cross-architecture replay and heavy state-root comparison remain nightly/manual work until public testnet readiness.
