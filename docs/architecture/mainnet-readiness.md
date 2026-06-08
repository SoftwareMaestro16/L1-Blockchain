# Mainnet Readiness Criteria

Aetra should not be considered mainnet-ready until all criteria are met. This gate is stricter than public testnet readiness and stricter than module production readiness: it requires implemented behavior, tests, operational evidence, public testnet observation, security audit completion, and complete documentation.

Required criteria:

- validator set policy implemented and tested;
- effective power cap implemented and tested;
- anti-concentration rewards implemented and tested;
- dynamic inflation implemented and tested;
- fee burn/treasury/reward split implemented and tested;
- slashing configured and tested;
- AVM integrated and tested;
- export/import stable;
- state sync/snapshots stable;
- public testnet has run long enough to observe validator behavior;
- load tests demonstrate finality target;
- security audit completed;
- critical findings fixed;
- docs complete for validators, delegators, and contract developers.

## Acceptance Rule

Mainnet readiness is all-or-nothing. Missing evidence for any item means Aetra is not mainnet-ready.

The readiness gate must be backed by:

- passing module tests for staking policy, economics, slashing, AVM, export/import, state sync, and snapshots;
- load reports showing normal and stressed finality targets;
- public testnet reports covering real validator behavior;
- completed security audit report;
- evidence that all critical findings are fixed;
- published docs for validators, delegators, and contract developers.

## Implementation Contract

The implementation gate is `app/params/mainnet_readiness.go`.

Required catalog properties:

- `MainnetReadinessEvidence` must cover every required criterion;
- `BuildMainnetReadinessReport` must report missing criteria;
- `ValidateMainnetReadiness` must reject mainnet-ready claims when any criterion is missing;
- unit tests must cover complete readiness and missing consensus, economics, operational, security, and documentation evidence.
