# Acceptance Test Matrix

Minimum acceptance matrix before public testnet:

```text
Base node:
  boot single node
  boot multi-validator localnet
  restart
  export/import
  state sync or snapshot restore

Staking:
  create validator
  delegate
  redelegate
  unbond
  withdraw rewards
  validator commission update

Anti-centralization:
  validator reaches cap
  validator exceeds cap
  excess stake reward penalty applied
  top-N concentration query works
  commission floor enforced

Slashing:
  downtime tracked
  downtime jail
  double-sign evidence path where feasible
  tombstone behavior
  delegator slash accounting

Economics:
  inflation update
  fee burn
  treasury allocation
  rewards allocation
  APR query
  supply invariant

AVM:
  upload code
  instantiate
  execute
  query
  migrate if enabled
  gas exhaustion contained

Governance:
  valid param proposal
  invalid param proposal
  treasury proposal
  delayed critical param activation

Observability:
  Prometheus metrics
  CLI queries
  gRPC queries
  events indexable
```

## Implementation Contract

The implementation gate is `DefaultAetraAcceptanceMatrixEvidence` in `app/params/acceptance_test_matrix.go`.

This matrix is the minimum public testnet acceptance scope, not the full production test suite. Each item must map to a documented unit, integration, e2e, localnet, adversarial, load, or operational smoke test. Expensive tests may be nightly/manual before automation is complete, but public testnet readiness must retain explicit evidence for each row.

Base node coverage proves chain boot, validator networking, restart safety, export/import, and state sync or snapshot restore. Staking and anti-centralization coverage proves validator lifecycle, delegation lifecycle, commission control, power cap behavior, overflow penalty, and concentration query behavior. Slashing and economics coverage proves safety-critical accountability and supply/reward accounting.

AVM coverage proves contract upload, instantiate, execute, query, migration if enabled, and malicious gas exhaustion containment. Governance coverage proves valid and invalid param changes, treasury flow, and delayed activation for critical params. Observability coverage proves that Prometheus, CLI, gRPC, and indexable events expose enough public evidence for operators and explorers.

## Acceptance Gate

Required behavior:

- missing base node scenarios fail readiness;
- missing staking scenarios fail readiness;
- missing anti-centralization scenarios fail readiness;
- missing slashing scenarios fail readiness;
- missing economics scenarios fail readiness;
- missing AVM scenarios fail readiness;
- missing governance scenarios fail readiness;
- missing observability scenarios fail readiness;
- duplicate or unexpected scenarios fail readiness.
