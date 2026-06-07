# Repository-Level Work Breakdown

This section maps work to likely repository areas. Exact paths may change, but responsibilities should remain.

The purpose is to prevent architecture tasks from becoming vague cross-repository work. Each area must have concrete ownership, deliverables, tests, and acceptance gates.

## 32.1 `proto/`

Tasks:

- define protobuf messages for new modules;
- define query services;
- define tx services;
- define genesis messages;
- define params messages;
- run code generation;
- add proto breaking-change checks if available.

Tests:

- generated code compiles;
- proto lint passes if configured;
- query/tx service registration tested.

The `proto/` tree owns public wire contracts. Query, tx, genesis, and params definitions are not internal implementation details; they are API commitments for CLI, gRPC, REST gateway, wallets, explorers, dashboards, indexers, validators, and automation.

Code generation must be reproducible. Generated Go must compile, must match the selected code generation workflow, and must not be manually edited. If proto breaking-change checks are available, they must run before public testnet and on every protocol-facing proto change.

Service registration tests must prove that query and tx services are reachable through the app module wiring, not only that generated files exist. If a module has proto definitions but no registered query/tx service where one is required, the module is not production-ready.

## 32.2 `x/`

Tasks:

- implement keepers;
- implement message servers;
- implement query servers;
- implement genesis;
- implement params validation;
- implement invariants;
- implement hooks where needed;
- implement events;
- implement module interfaces.

Tests:

- keeper unit tests;
- msg server tests;
- query server tests;
- genesis tests;
- invariant tests;
- fuzz/property tests for math.

The `x/` tree owns module behavior. Keepers must be deterministic, avoid floating point, validate authority boundaries, and use bounded iteration where state size can grow. Message servers must validate signers, authority, denoms, addresses, and params before mutating state. Query servers must return stable response shapes, handle not-found cases clearly, and support pagination where lists are exposed.

Genesis code must support default genesis, validation, import, export, and migration compatibility. Params validation must reject unsafe values before they reach consensus state. Invariants must cover accounting, share math, supply movement, module account permissions, and any module-specific safety rule. Hooks and events must be deterministic and tested where they affect staking, slashing, rewards, governance, or contract execution.

Fuzz/property tests are required for math-heavy logic such as shares, rewards, slashing fractions, fee split, inflation, APR estimates, voting power caps, overflow stake, storage rent, and any rounding-sensitive calculation.

## 32.3 `app/`

Tasks:

- wire keepers;
- wire modules;
- wire module account permissions;
- wire begin/end/preblock order;
- wire simulation manager if used;
- wire API routes;
- wire AutoCLI if used;
- validate startup.

Tests:

- app startup;
- module account permissions;
- begin/end order;
- export/import;
- deterministic restart;
- API service registration.

The `app/` tree owns whole-chain assembly. Keeper wiring must pass the exact keeper dependencies used by modules and must avoid nil keepers, wrong authorities, wrong codecs, or mismatched store keys. Module wiring must register services, invariants, genesis, migrations, hooks, and begin/end/preblock order consistently with the module manager.

Module account permissions are consensus-sensitive and must be tested. Begin/end/preblock order must be explicit because staking, slashing, distribution, fee burn, treasury, validator score, and contract execution can depend on hook order. API routes, gRPC services, REST gateway routes, and AutoCLI wiring must be tested through app-level registration, not only package-level constructors.

Startup validation must reject unsafe module account permissions, missing stores, duplicate blocked addresses, invalid params, missing authority wiring, and invalid genesis. Export/import and deterministic restart tests are mandatory before public testnet because they catch app hash drift and migration mistakes.

## 32.4 `tests/`

Tasks:

- integration test suites;
- e2e localnet smoke tests;
- adversarial tests;
- load profile tests;
- documentation path tests;
- CI scripts.

Required:

- tests must be runnable from documented commands;
- Windows PowerShell local scripts should remain usable if current project supports them;
- Linux CI path should remain primary for production confidence.

The `tests/` tree owns cross-module confidence. Integration suites must prove keeper, app, module, governance, staking, economics, slashing, contract, and migration behavior across boundaries. E2E localnet smoke tests must prove that a node can start, produce blocks, submit transactions, query state, restart, and preserve state across export/import where feasible.

Adversarial tests must cover malformed txs, invalid params, module account abuse attempts, concentration and overflow edge cases, slashing/evidence edge cases, contract gas/storage abuse, and migration failure paths. Load profile tests must measure block time, finality, mempool pressure, contract execution load, and state growth where feasible.

Documentation path tests must keep architecture docs, runbooks, command examples, and production gates synchronized with code. CI scripts must keep Linux as the production-confidence path while preserving Windows PowerShell local scripts when the project supports them.

## Acceptance Gate

The implementation gates are `DefaultAetraRepoProtoWorkEvidence`, `DefaultAetraRepoXWorkEvidence`, `DefaultAetraRepoAppWorkEvidence`, and `DefaultAetraRepoTestsWorkEvidence` in `app/params/repository_work_breakdown.go`.

Required behavior:

- missing protobuf message definitions fail readiness;
- missing query service definitions fail readiness;
- missing tx service definitions fail readiness;
- missing genesis or params messages fail readiness;
- missing code generation step fails readiness;
- missing breaking-change checks fail readiness where available;
- missing generated-code compile, proto lint, or service-registration tests fail readiness.
- missing keeper, msg server, query server, genesis, params validation, invariant, hook, event, or module interface work fails readiness;
- missing keeper, msg server, query server, genesis, invariant, or fuzz/property math tests fails readiness.
- missing app keeper wiring, module wiring, module account permissions, begin/end/preblock order, simulation manager, API routes, AutoCLI, or startup validation work fails readiness;
- missing app startup, module account permissions, begin/end order, export/import, deterministic restart, or API service-registration tests fails readiness.
- missing integration, e2e localnet, adversarial, load profile, documentation path, or CI script work fails readiness;
- missing documented commands, Windows PowerShell local-script usability where supported, or Linux CI production-confidence path fails readiness.
