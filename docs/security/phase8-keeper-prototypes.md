# Phase 8 Keeper Prototype Security Notes

## Scope

Phase 8 converts approved executable specs into keeper-shaped prototypes for:

- `x/load`
- `x/routing`
- `x/zones`
- `x/mesh`

These keepers are intentionally not wired into `app.go`, do not register public
Msg services, and do not add module accounts or KVStore keys. They are audit and
test scaffolding for future SDK module integration.

## Security Model

- Feature gates are disabled by default.
- Testnet enablement uses explicit testnet params.
- Production enablement requires an explicit software version gate.
- Mutating keeper methods call `RequireEnabled`.
- Governance-style params updates call `Authorize`.
- Query helpers use bounded pagination with a shared max limit.
- Export paths clone and canonically sort state through the underlying specs.
- Migration skeletons validate current state and perform no mutation.

## Consensus Safety

The prototype keepers cannot alter production consensus behavior until app
wiring is added deliberately. Current implementation avoids:

- wall-clock time;
- randomness;
- goroutines;
- external APIs;
- node-local latency;
- validator-local preferences;
- unbounded query response allocation.

State iteration is bounded by request limits for query helpers. Spec-level
export/import functions canonicalize ordering before state is returned.

## Audit Findings

### Load

- Mutating metrics updates are disabled by default.
- EMA and history validation reject out-of-order or out-of-range scores.
- History query pagination is bounded.

### Routing

- Routing table mutation requires both feature enablement and authority.
- Shard config import rejects duplicate, zero, unknown, and non-canonical state.
- Runtime routing remains pure and deterministic.

### Zones

- Zone registration and activation are disabled by default.
- Duplicate zones and non-canonical imported state are rejected.
- Commitment chains are validated by hash and previous commitment linkage.

### Mesh

- Destination, source commitment, and message application are disabled by default.
- Replay markers and receipts round-trip through export/import.
- Duplicate destinations, duplicate receipts, and replay attempts are rejected.

## Remaining Risks Before Production Wiring

- Replace in-memory prototype state with SDK KVStores only after prefix-bounded
  iterators and pagination are reviewed.
- Add module accounts only if a state transition requires custody.
- Add Msg and Query services only after protobuf contracts are finalized.
- Run benchmarks before enabling high-volume testnet profiles.
- Perform another consensus-critical audit when keepers are connected to app
  lifecycle hooks.

## Verification Commands

```powershell
go test ./...
go vet ./...
buf lint
go test ./x/routing/types ./x/zones/types ./x/mesh/types -bench . -run '^$'
```
