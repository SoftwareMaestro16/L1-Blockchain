# Implementation Phases

This plan defines the first implementation phases for Aetra Core. Each phase has tasks, deliverables, tests, and acceptance criteria. A phase is not complete because code exists; it is complete only when its tests and evidence are present.

## Phase 0 - Baseline Audit

Tasks:

- inspect current Cosmos SDK and CometBFT versions;
- document current app module graph;
- identify existing modules overlapping with `aetra-staking-policy`, `aetra-validator-score`, and `aetra-economics`;
- decide which modules are renamed, reused, or wrapped;
- verify current `naet` staking denom;
- verify fee collector, burn, treasury, emissions, mint authority wiring;
- verify current localnet scripts and test coverage.

Deliverables:

- module inventory;
- gap analysis;
- risk list;
- updated implementation checklist.

Tests:

- current full unit test run;
- current integration test run;
- current localnet smoke test;
- current export/import test.

Acceptance:

- module inventory exists and maps current app, keeper, store, and module-account ownership;
- overlapping modules have a decision: rename, reuse, wrap, or replace;
- `naet` staking denom and fee denom assumptions are verified;
- fee collector, burn, treasury, emissions, and mint authority wiring have explicit evidence;
- localnet and export/import scripts are known-good or have tracked blockers.

## Phase 1 - Staking Policy and Validator Cap

Tasks:

- implement effective voting power cap;
- implement overflow stake accounting;
- implement commission floor/max/max-change policy;
- add concentration metrics;
- add queries for validator raw/effective/overflow stake;
- add governance params with validation;
- wire module into app lifecycle.

Tests:

- cap math unit tests;
- validator set transition tests;
- concentration query tests;
- commission bounds tests;
- integration tests with staking;
- export/import tests;
- invariant tests.

Acceptance:

- no validator can exceed configured effective power cap;
- excess stake does not increase voting power;
- params cannot be set outside safe bounds;
- state remains deterministic after export/import.

## Phase 2 - Economics and Fee Split

Tasks:

- implement dynamic inflation bounds;
- implement target bonded ratio logic;
- implement fee split to burn/rewards/treasury;
- implement reward smoothing;
- expose APR estimate query;
- expose burned supply and treasury accounting queries;
- add governance param controls.

Tests:

- inflation curve tests;
- bonded ratio tests;
- fee split tests;
- burn accounting tests;
- treasury accounting tests;
- APR query tests;
- supply invariant tests;
- export/import tests.

Acceptance:

- inflation remains within configured bounds;
- fee split sums to 100 percent;
- burned fees reduce spendable/module-held supply according to chain accounting rules;
- treasury receives correct amount;
- rewards are deterministic.

## Phase 3 - Validator Score and Accountability

Tasks:

- implement uptime score;
- implement slash history;
- implement governance participation score;
- implement decentralization score;
- implement public validator metrics queries;
- integrate score with reward modifier only for objective inputs.

Tests:

- uptime window tests;
- missed block tests;
- slash history tests;
- governance participation tests;
- score determinism tests;
- reward modifier tests;
- export/import tests.

Acceptance:

- score is deterministic;
- score cannot be manipulated through subjective inputs;
- score is queryable for explorers and wallets;
- score does not break consensus safety.

## Phase 4 - Slashing Hardening

Tasks:

- configure double-sign slash fraction and tombstone behavior;
- configure downtime windows and jail duration;
- implement progressive downtime if not covered by standard module;
- add timestamp/proposal violation policy where objective;
- document evidence lifecycle and unbonding interaction.

Tests:

- double-sign evidence tests where feasible;
- downtime tests;
- jail/unjail tests;
- progressive downtime tests;
- slashing accounting tests;
- delegator loss tests;
- tombstone tests;
- evidence expiry tests.

Acceptance:

- double-sign leads to severe slash and permanent tombstone;
- downtime penalties are bounded and progressive;
- no subjective slashing path exists;
- slashing cannot underflow stake or corrupt shares.

## Phase 5 - AVM Integration

Tasks:

- finalize AVM module wiring;
- define code upload policy;
- define contract gas limits;
- define contract size limits;
- integrate storage rent or storage pricing;
- expose contract events for indexers;
- document contract developer flow.

Tests:

- instantiate/execute/query tests;
- migration tests;
- gas limit tests;
- storage limit/rent tests;
- malicious contract tests;
- export/import tests with contracts;
- localnet AVM smoke test.

Acceptance:

- contracts are deterministic;
- contract gas is bounded;
- malicious contracts cannot halt chain;
- contract state survives export/import.

## Phase 6 - Finality and Performance Profile

Tasks:

- configure block time targets;
- configure block size/gas limits;
- profile 100 validator localnet;
- profile 150-200 validator scenario if feasible;
- estimate 250-300 validator operational requirements;
- measure finality under load;
- measure finality under partial validator failure.

Tests:

- localnet load profile;
- mempool pressure test;
- block time measurement;
- finality measurement;
- validator failure scenario;
- restart scenario;
- state sync/snapshot scenario.

Acceptance:

- normal finality remains within target;
- stressed finality remains below 120 seconds in healthy majority scenario;
- node requirements remain medium-level;
- no excessive consensus payloads.

## Phase 7 - Public Testnet Readiness

Tasks:

- write validator setup docs;
- write sentry architecture docs;
- write monitoring docs;
- publish genesis parameter explanation;
- publish economic model explanation;
- publish slashing risk explanation;
- publish delegation and nomination pool guide;
- publish AVM developer guide;
- prepare public dashboards;
- prepare incident response process.

Tests:

- clean node bootstrap from docs;
- validator join from docs;
- snapshot restore from docs;
- state sync from docs;
- tx flow smoke tests;
- governance proposal smoke tests;
- public RPC/gRPC/REST smoke tests.

Acceptance:

- new validator can join using docs;
- public endpoints are observable;
- network can recover from node restarts;
- core economic and staking flows work end to end.

## Implementation Contract

The implementation phase gate is `app/params/implementation_phases.go`.

Required catalog properties:

- `DefaultImplementationPhasePlans` must include Phase 0 through Phase 7;
- Phase 0 must include baseline audit tasks, deliverables, and current test evidence;
- Phase 1 must include staking cap tasks, tests, and acceptance checks;
- Phase 2 must include economics, fee split, burn, treasury, APR, reward smoothing, and supply invariant checks;
- Phase 3 must include validator score, accountability metrics, objective reward modifier, and consensus-safety checks;
- Phase 4 must include objective slashing hardening, evidence lifecycle, tombstone, downtime, and stake/share safety checks;
- Phase 5 must include AVM module wiring, upload policy, gas/size limits, storage pricing, indexer events, developer flow, malicious contract safety, and export/import checks;
- Phase 6 must include finality, block timing, load profile, validator-failure profile, restart, and state sync/snapshot checks;
- Phase 7 must include public testnet docs, dashboards, endpoint smoke tests, incident response, and end-to-end economic/staking flow checks;
- phase items require evidence;
- missing tasks, deliverables, tests, or acceptance checks fail validation.
