# Definition of Done

No task is complete until:

- code is implemented;
- params are validated;
- genesis import/export works;
- query surface exists;
- events exist where operationally relevant;
- unit tests pass;
- integration tests pass;
- e2e/localnet test exists for user-facing flow;
- docs describe operator/user impact;
- failure modes are documented;
- security implications are reviewed.

For consensus/economics/staking changes, also required:

- adversarial tests;
- invariant tests;
- export/import test;
- deterministic restart test;
- migration test if state changed.

## Implementation Contract

The implementation gate is `DefaultAetraDefinitionOfDoneEvidence` in `app/params/definition_of_done.go`.

Every feature must be reviewed as a full delivery unit, not as isolated code. Keeper logic without params validation, genesis handling, query surface, events, tests, docs, and failure-mode documentation is incomplete. User-facing flow means any flow exposed to validators, delegators, wallets, explorers, governance, operators, or contract developers.

Security review must be proportional to risk. Consensus, economics, staking, slashing, migration, app wiring, fee handling, and contract execution changes require the stricter critical-change gate because a small logic error can affect liveness, safety, supply, voting power, rewards, or app hash determinism.

## Acceptance Gate

Required behavior:

- missing code implementation fails readiness;
- missing params validation fails readiness;
- missing genesis import/export fails readiness;
- missing query surface fails readiness;
- missing operational events fail readiness where relevant;
- missing unit, integration, or e2e/localnet flow tests fail readiness;
- missing operator/user docs fail readiness;
- missing failure-mode documentation fails readiness;
- missing security review fails readiness;
- consensus/economics/staking changes must also include adversarial, invariant, export/import, deterministic restart, and migration tests where state changed.
