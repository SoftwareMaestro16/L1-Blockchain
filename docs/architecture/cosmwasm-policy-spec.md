# CosmWasm Policy Specification

This document defines section 28 of the Aetra architecture backlog.

Aetra's native smart contract runtime is AVM. CosmWasm is an optional gated compatibility layer, not the primary runtime and not a replacement for AVM. The implementation must preserve this boundary in docs, config, tests, and app wiring.

The implementation gate is `app/params/cosmwasm_policy_spec.go`. The runtime policy gate is `app/wasmconfig`.

## 28.1 Contract Permissions

Launch policy:

```text
early testnet:
  permissioned code upload or governance-gated upload

later testnet:
  permissionless upload with strong fees/deposits

mainnet:
  policy decided after security review
```

Requirements:

- CosmWasm remains disabled by default until explicitly enabled by config/governance.
- AVM remains the primary native contract runtime.
- Early testnet upload must be permissioned or governance gated.
- Later testnet may move to permissionless upload only with strong fees/deposits.
- Mainnet upload policy must not be finalized before a completed security review.
- Contract migration must require explicit migration authority rules.
- Pinned code must be disabled by default; if used, pinning must be governance controlled and bounded.
- Permission changes must be governance configurable within safe bounds, emit events, and be covered by tests.

Acceptance gate:

- `BuildAetraCosmWasmLaunchPolicyReport` must pass.
- Missing early testnet, later testnet, or mainnet phase policy must fail validation.
- Missing AVM/CosmWasm boundary must fail validation.
- Mainnet policy without security review gate must fail validation.

## 28.2 Gas And Storage

Required controls:

- max wasm code size;
- max instantiate gas;
- max execute gas per tx;
- max query gas;
- storage rent or storage pricing;
- contract upload fee;
- contract migration authority rules;
- pinned code policy if used.

Runtime requirements:

- `app/wasmconfig.Policy` must expose explicit bounded fields for code size, instantiate gas, execute gas per tx, query gas, upload fee, storage price/rent, migration policy, and pinned code policy.
- Gas and fee math must use deterministic integer accounting.
- Upload, instantiate, execute, query, migrate, pin, and storage accounting must not bypass the native `naet` fee policy.
- Storage pricing must be deterministic, bounded, queryable in future API work, and compatible with export/import.
- Increasing limits requires benchmarks, adversarial tests, and security review updates.
- Contract upload fee must be denominated in `naet` and reject non-native fee assets.
- Query gas, response bytes, and query depth must stay bounded.

Acceptance gate:

- `BuildAetraCosmWasmGasStorageReport` must pass.
- Removing any required gas/storage control must fail validation.
- Removing governance bounds, deterministic accounting, or security/benchmark gates must fail validation.
- `app/wasmconfig` tests must cover instantiate gas, execute gas per tx, upload fee, storage pricing, migration authority, and pinned code policy.

## Required Tests

Every implementation task must include tests. For section 28, required coverage is:

- launch policy tests;
- gas/storage limit tests;
- upload fee tests;
- storage pricing tests;
- migration authority tests;
- pinned code policy tests;
- AVM/CosmWasm boundary tests.

Acceptance gate:

- `BuildAetraCosmWasmTestReport` must pass.
- Missing, duplicate, or unexpected test catalog entries must fail validation.

## Non-Goals

CosmWasm policy work must not:

- enable CosmWasm by default;
- make CosmWasm the primary Aetra runtime;
- bypass AVM architecture requirements;
- allow unbounded contract execution;
- allow non-`naet` base-chain protocol fees;
- finalize mainnet permissionless upload before security review;
- rely on manual smoke tests without unit/integration coverage.
