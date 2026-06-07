# Refactor Audit Report

This note records the scoped audit performed during the async execution
refactor. It is not a replacement for the full determinism gate or independent
security review.

## Scope

- `x/aetravm/async`
- async contract standard callers under `x/aetravm/standards`
- repository-wide scan for obvious Cosmos consensus-risk patterns

## Refactor Result

The former monolithic `x/aetravm/async/async.go` was split into focused
files:

- `types.go`
- `params.go`
- `address.go`
- `validation.go`
- `executor.go`
- `queue.go`
- `process.go`
- `export.go`
- `clone.go`
- `test_helpers_test.go`

Tests remain in `_test.go` files. Shared async test setup moved into
`test_helpers_test.go`.

## Determinism Audit

Checked patterns:

- `range` over maps / range over maps in consensus-adjacent code;
- `time.Now`;
- `rand`;
- goroutines;
- `select`;
- floating point.

Changes made:

- `Params.Validate` now checks fee-like integer params through an ordered slice
  instead of ranging over a map.
- `ValidateExportedState` now validates inbox/outbox maps by sorted owner keys
  instead of ranging over maps directly.

Remaining scanner hits are not consensus-critical in the reviewed context:

- deterministic test random source in `tests/testutil`;
- `ed25519.GenerateKey(rand.Reader)` in tests;
- `time.Now` in tests and local service startup timing;
- generated gRPC gateway goroutines in `*.pb.gw.go`;
- fixed slice of memo index maps, with deterministic sorting inside each
  bucket.

## Verification

Required checks after this refactor:

```powershell
go test ./...
go vet ./...
buf lint
scripts\proto\verify-generated.ps1
tests\scripts\*.ps1
git diff --check
```

No PR is required for this work. Direct push is an operator decision and should
only happen from the target branch after local checks pass.
