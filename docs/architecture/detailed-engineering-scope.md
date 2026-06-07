# Detailed Engineering Scope

This section translates the architecture into a concrete engineering backlog. Every item below must be implemented as production-grade functionality, not as a placeholder.

General rule for every task:

```text
feature = code + params + genesis validation + queries + events + tests + docs
```

If a feature has code but does not have tests, query surface, genesis validation, or acceptance criteria, the feature is not complete.

## Core Chain Configuration

Tasks:

- define chain id naming policy for devnet, testnet, and mainnet;
- define staking denom `naet`;
- define display denom `AET`;
- verify coin metadata for `naet/AET`;
- verify address prefix and reserved system address policy;
- verify module account permissions;
- verify blocked address policy;
- verify mint authority;
- verify burn authority;
- verify fee collector authority;
- verify treasury authority;
- verify genesis validation for all Aetra modules;
- verify app export/import with all modules enabled.

Expected deliverables:

- `app` wiring review;
- genesis params table;
- module accounts table;
- authority matrix;
- CLI command matrix;
- query matrix;
- event matrix;
- tests for startup validation.

Required tests:

- app boots with default genesis;
- app rejects invalid denom metadata;
- app rejects missing module accounts;
- app rejects duplicate reserved addresses;
- app rejects unsafe module account permissions;
- export/import preserves app hash where expected;
- simulation or integration test covers module initialization order.

## Consensus Parameter Policy

Tasks:

- define target block time range;
- define max block bytes;
- define max block gas;
- define evidence max age by blocks;
- define evidence max age by duration;
- define validator public key types;
- define CometBFT timeout profile for 100, 200, and 300 validators;
- define snapshot interval;
- define state sync parameters;
- define pruning profiles.

Recommended initial values must be conservative. Do not maximize block size early.

Example target policy:

```text
block_time_target:
  100 validators: 5-6s
  200 validators: 6s
  300 validators: 7-8s

max_block_gas:
  start conservative
  increase only after load tests

max_block_bytes:
  keep below values that increase propagation delay
  change only through governance after testnet evidence
```

Expected deliverables:

- conservative initial values;
- block time target table for 100, 200, and 300 validators;
- max block gas bounds;
- max block bytes bounds;
- evidence window table;
- timeout profile table;
- state sync, snapshot, and pruning table;
- governance safety bounds.

Required tests:

- localnet remains stable under configured timeout profile;
- oversized blocks are rejected;
- invalid consensus params are rejected;
- governance cannot set unsafe block gas/bytes outside bounds;
- evidence remains valid through configured evidence period.

## Implementation Contract

The implementation gate is `app/params/engineering_scope.go`.

Required catalog properties:

- `FeatureCompletionEvidence` must require code, params, genesis validation, queries, events, tests, and docs;
- `ValidateFeatureCompletion` must reject incomplete features;
- `DefaultCoreChainConfigurationScopePlan` must include all core chain configuration tasks, expected deliverables, and required tests;
- `DefaultConsensusParameterPolicyScopePlan` must include all consensus parameter tasks, conservative deliverables, and required tests;
- `ValidateEngineeringScopePlan` must reject missing evidence, missing required items, unknown scopes, and unexpected items.
