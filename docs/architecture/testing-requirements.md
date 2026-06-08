# Testing Requirements

Every implementation task must include tests. A feature is not complete without tests.

The expected test layer depends on the feature, but the default stance is strict: if a change touches consensus, staking, economics, governance, slashing, VM execution, module accounts, localnet tooling, or public query surfaces, it must ship with targeted tests in the appropriate layers.

## Required Test Layers

### Unit Tests

Required:

- keeper logic;
- params validation;
- math and accounting;
- cap calculation;
- slashing policy;
- reward split;
- inflation curve;
- score calculation.

Unit tests must be fast, deterministic, and independent of localnet. They should prove pure formulas, state transition helpers, params validation, integer accounting, reward math, validator scoring, caps, and failure behavior.

### Integration Tests

Required:

- staking + custom staking policy;
- slashing + validator score;
- distribution + economics;
- fee collector + burn + treasury;
- nomination pool + delegation + unbonding;
- governance param updates;
- AVM tx flow.

Integration tests must prove cross-module behavior before localnet scripts are treated as acceptance evidence. AVM contract tests are required for genesis; any non-AVM compatibility tests remain post-genesis and explicitly gated.

### E2E/Localnet Tests

Required:

- node startup;
- validator creation;
- delegation;
- redelegation;
- unbonding;
- downtime scenario;
- double-sign evidence scenario where feasible;
- fee burn scenario;
- AVM instantiate/execute/query;
- export/import;
- restart;
- state sync/snapshot where feasible.

E2E tests must start from clean state, write only ignored runtime artifacts, avoid private operator secrets, and produce enough logs or summaries to debug failures. Where a scenario is marked "where feasible", feasibility must be documented; once feasible, the test becomes required.

### Adversarial Tests

Required:

- concentration attack simulation;
- validator overflow stake simulation;
- commission manipulation attempt;
- invalid params proposal;
- malformed evidence;
- jailed validator reward attempt;
- module account abuse attempt;
- contract gas exhaustion;
- contract storage abuse.

Adversarial tests must verify that invalid or hostile inputs fail closed without partial state mutation, supply drift, consensus panic, app hash divergence, or reward leakage.

### Performance Tests

Required:

- 100 validator localnet/profile;
- 150-200 validator simulation/profile if feasible;
- block time under load;
- finality latency measurement;
- mempool pressure;
- AVM execution load;
- state growth profile.

Performance tests may be nightly/manual when expensive, but public testnet readiness must include evidence for the relevant validator-set phase, finality target, load profile, and state growth behavior.

## Completion Rule

A task is not complete when it only compiles. Completion requires:

- implementation code;
- targeted tests for the changed behavior;
- negative tests for invalid inputs or unsafe authority paths;
- docs or test matrix update when the behavior is public or operational;
- explicit deferral only for expensive, infeasible, or gated scenarios.

Explicit deferral must include a reason and target gate. Missing unit tests cannot be deferred for production code.

## Test Acceptance Rule

No module should be considered production-ready unless:

- unit tests pass;
- integration tests pass;
- genesis validation tests pass;
- export/import tests pass;
- deterministic restart tests pass;
- adversarial tests for the relevant module pass;
- CI runs the critical subset automatically.

Production readiness is stricter than implementation readiness. A module can exist behind a feature gate while local/e2e/performance work continues, but it must not be described as production-ready until every acceptance item is green or the module is explicitly scoped out of production.

## Implementation Contract

The implementation gate is `app/params/testing_requirements.go`.

Required catalog properties:

- `DefaultTestingRequirements` must cover unit, integration, e2e/localnet, adversarial, and performance layers;
- every required scenario must have test evidence;
- optional "where feasible" scenarios must become required once feasible;
- `ValidateFeatureTestingEvidence` must reject completed features without tests;
- implementation-ready features without unit tests must fail;
- `ValidateModuleProductionReadiness` must reject production-ready claims unless unit, integration, genesis validation, export/import, deterministic restart, adversarial, and CI critical subset evidence is present.
