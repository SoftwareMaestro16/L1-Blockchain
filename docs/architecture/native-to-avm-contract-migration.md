> This document describes a future possibility; no token/NFT/DEX standards or native modules exist in the current project.
>
> Note: historical native asset-factory and native exchange modules have been removed from the active app graph.
# Native to AVM Contract Migration Plan

Status: Phase 1 audit and safe marking.

This document defines the migration boundary for L1 native `x/` packages. It does
not authorize deleting any existing native application module. Deletion is only
allowed after an AVM contract replacement exists, is wired through runtime
messages, and passes compatibility tests against the native behavior it replaces.

## Target Architecture

Protocol safety remains native. Application and product logic moves to AVM
contracts. AVM runtime infrastructure remains native, but should be consolidated
behind a future `x/avm` boundary instead of spreading as independent application
modules.

Native staking is protocol core. Native staking, validator registry/election,
slashing, evidence, unbonding, redelegation, commission, validator insurance,
delegator protection, rewards, minting, and emissions must stay native because
they directly affect validator power, CometBFT validator updates, supply
accounting, and chain safety. Contract staking may exist only as an
application-layer delegation product, such as liquid staking or delegation
vaults, with true validator power still calculated by native protocol modules.

## Classification

- `KEEP_NATIVE`: protocol core, consensus safety, native chain accounting, or
  root-level system control.
- `CONSOLIDATE_AVM_RUNTIME`: native runtime package that should become part of
  the `x/avm` actor/message/storage/runtime stack.
- `MOVE_TO_CONTRACT`: application or economic logic whose target implementation
  is an AVM contract.
- `MOVE_OFFCHAIN`: indexing, memo enrichment, or tooling that should not be an
  independent native protocol module.
- `DEPRECATED_LATER`: keep temporarily, mark as deprecated native app logic, and
  remove only after replacement contracts and compatibility tests exist.

## Static Audit

The audit below was generated from the current `x/` tree. `App wiring` means
the module is referenced from app code or app tests. `Proto` checks
`proto/l1/<module>` and `proto/l1/<module-without-dashes>`.

| Module | Current role | App wiring | Keeper | Proto | CLI | Tests | Target role | Action | Blockers before deletion | Tests required |
|---|---|---:|---:|---:|---:|---:|---|---|---|---|
| x/actor-registry | Actor/system registry | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native and integrate with `x/avm` deploy/execute routing | none | genesis, actor authority, AVM deploy integration |
| x/actors | Actor types/runtime helpers | yes | no | no | no | 1 | CONSOLIDATE_AVM_RUNTIME | Move under `x/avm` runtime stack or internal runtime package later | stable `x/avm` API | actor lifecycle and compatibility tests |
| x/aetracore | Core execution/system package | yes | yes | no | no | 32 | KEEP_NATIVE | Keep native protocol core | none | existing app and keeper tests |
| x/aetravm | AVM implementation and standards | no | no | no | no | 15 | CONSOLIDATE_AVM_RUNTIME | Keep native, consolidate with `x/avm` | `x/avm` boundary finalized | AVM runtime, async, AFT/ANFT/AW tests |
| x/avm-scheduler | AVM scheduler | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native runtime infrastructure | none | scheduler genesis/export/import, gas limits |
| x/bridge-hub | Cross-chain bridge coordination | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native if it owns trust roots and bridge safety | none if only trust registry | bridge trust and replay tests |
| x/burn | Native burn/accounting | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native if tied to supply accounting | none | supply accounting and authority tests |
| x/compute | Compute type helpers | yes | no | no | no | 1 | CONSOLIDATE_AVM_RUNTIME | Fold into runtime capability model | `x/avm` resource model | compute limit tests |
| x/config | Protocol config | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native | none | config bounds, authority, migration tests |
| x/config-voting | Config voting flow | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native if it controls protocol params | none | quorum/delay/authority tests |
| x/constitution | Constitutional bounds | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native | none | protected limits and emergency expiry tests |
| x/cross-chain-registry | Cross-chain trust registry | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native trust root registry | none | trust registry and deterministic iteration tests |
| x/dex | Native AMM/DEX app logic | yes | yes | yes | yes | 7 | MOVE_TO_CONTRACT | Mark deprecated-after-contract-replacement, keep until AVM AMM replacement passes | AMM factory/pool contracts, LP AFT-44, compatibility tests | native DEX suite plus contract swap/liquidity/e2e |
| x/dynamic-commission | Validator commission policy | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native validator economics | none | commission bounds and validator integration |
| x/emissions | Native emissions accounting | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native protocol supply accounting if/when landed | none | emission bounds and supply tests |
| x/epoch | Epoch helper module | yes | yes | no | no | 1 | KEEP_NATIVE | Keep native if used by protocol scheduling | none | epoch transition tests |
| x/events | Event types | yes | no | no | no | 1 | CONSOLIDATE_AVM_RUNTIME | Consolidate events under receipts/runtime where appropriate | receipt schema finalized | event determinism tests |
| x/evidence | Validator evidence | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native | none | slash/evidence/adversarial tests |
| x/execution | Execution type helpers | yes | no | no | no | 1 | CONSOLIDATE_AVM_RUNTIME | Fold into `x/avm` execute API | `x/avm` API | external/internal/bounced execution tests |
| x/fee-collector | Fee collection/distribution | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native if it owns protocol fee settlement | none | fee accounting and authority tests |
| x/fees | Fee policy and ante logic | yes | yes | yes | yes | 8 | KEEP_NATIVE | Keep native protocol fee policy | none | ante, denom, bounds, fee market tests |
| x/identity | DNS/app identity logic | yes | no | no | no | 39 | MOVE_TO_CONTRACT | Split: root protocol remains native in `x/identity-root`; resolver/NFT/subdomain/auction move to contracts | ANFT-66 domain collection, resolver, subdomain manager, auction contracts | identity root invariants plus contract resolver/NFT tests |
| x/identity-root | Root identity registry | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native root-level protocol | none | root uniqueness, reserved names, expiry bounds |
| x/indexer | Indexer types | yes | no | no | no | 1 | MOVE_OFFCHAIN | Move to off-chain indexer or internal non-protocol package | external indexer replacement | indexer compatibility tests |
| x/internal | Internal helpers | yes | no | no | no | 0 | KEEP_NATIVE | Keep internal only | none | covered by dependent package tests |
| x/load | Protocol load model | yes | yes | no | no | 3 | KEEP_NATIVE | Keep native | none | load bounds and routing tests |
| x/market | Market app logic | yes | no | no | no | 1 | MOVE_TO_CONTRACT | Mark deprecated native app logic; replace by contracts or scheduler/fee policy | contract or protocol fee-policy split | market contract/unit tests |
| x/memo | Memo/indexing helpers | yes | no | no | no | 2 | MOVE_OFFCHAIN | Move to ante/fee policy/app params or off-chain indexing | ante/app-param replacement | memo policy and app tests |
| x/mesh | Mesh system coordination | yes | yes | no | no | 3 | KEEP_NATIVE | Keep native system infrastructure | none | mesh routing and persistence tests |
| x/messages | Message model | yes | no | no | no | 10 | CONSOLIDATE_AVM_RUNTIME | Fold into `x/avm` message stack | `x/avm` message envelope | message routing, retry, receipt tests |
| x/messaging | Cross-zone messaging types | no | no | no | no | 3 | CONSOLIDATE_AVM_RUNTIME | Fold into `x/avm` async messaging stack | `x/avm` queue model | cross-zone queue tests |
| x/networking | Native networking system | yes | yes | no | no | 6 | KEEP_NATIVE | Keep native protocol/system networking | none | DNL, routing, deterministic tests |
| x/nominator-pool | Shared staking pool product | yes | yes | yes | no | 1 | MOVE_TO_CONTRACT | Prefer contract if it is product-level and does not directly own validator set; keep native until replacement | contract pool, native staking interface, slash/unbond compatibility | pool accounting and contract integration tests |
| x/payments | Payments/settlement app logic | yes | yes | no | no | 10 | MOVE_TO_CONTRACT | Move app-specific payment channels/settlement to contracts unless protocol fee settlement | contract settlement replacement | payment contract/e2e tests |
| x/performance | Performance evidence/types | yes | no | no | no | 11 | KEEP_NATIVE | Keep native when feeding validator scoring/evidence | none | performance oracle/evidence tests |
| x/permissions | App permission model | yes | no | no | no | 1 | MOVE_TO_CONTRACT | Mark deprecated native app logic; protocol authorities stay in config/system registry | contract permission standard | permission contract tests |
| x/pos | Native PoS model | yes | no | no | no | 3 | KEEP_NATIVE | Keep native staking/validator power domain | none | staking, power, slashing integration tests |
| x/proofregistry | Proof registry types | no | no | no | no | 1 | KEEP_NATIVE | Keep native only if it is on-chain proof/system registry | decide on-chain vs off-chain scope | proof verification and registry tests |
| x/queue | Queue types | no | no | no | no | 1 | CONSOLIDATE_AVM_RUNTIME | Fold into `x/avm` async queue | queue boundary finalized | queue determinism and bounded loop tests |
| x/reporter | Reporter/fisherman registry | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native evidence support | none | reporter bond/reward/slash tests |
| x/reputation | Validator reputation types | yes | no | no | no | 2 | KEEP_NATIVE | Keep native when used for validator safety | none | reputation bounds and evidence integration |
| x/routing | System routing | yes | yes | no | no | 3 | KEEP_NATIVE | Keep native | none | route admission and determinism tests |
| x/scheduler | Protocol scheduler | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native | none | bounded task and genesis tests |
| x/schedulerv2 | Scheduler specs/types | no | no | no | no | 1 | CONSOLIDATE_AVM_RUNTIME | Merged into x/scheduler (DAG + MailboxPlan + ReplayHash) | done | merged |
| x/services | Service registry/app payments | yes | yes | yes | no | 20 | MOVE_TO_CONTRACT | Mark deprecated native app logic; service registry products move to contracts | contract service registry/retry/payment replacement | service contract and compatibility tests |
| x/sharding | Sharding simulation/types | yes | no | no | no | 2 | KEEP_NATIVE | Keep native sharding coordination support | none | sharding sim/coordinator tests |
| x/sharding-coordinator | Native sharding coordinator | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native | none | coordinator genesis and invariant tests |
| x/single-nominator-pool | Single nominator pool | yes | yes | yes | no | 1 | MOVE_TO_CONTRACT | Product-level staking facade may become contract if validator power remains native | contract facade and native staking interface | owner, slash, unbond compatibility tests |
| x/storage | Storage types | yes | no | no | no | 7 | CONSOLIDATE_AVM_RUNTIME | Fold into `x/avm` storage stack | code/storage state API | storage invariants and rent integration |
| x/storage-rent | Contract storage rent | yes | yes | yes | no | 2 | KEEP_NATIVE | Keep native AVM safety infrastructure | none | rent bounds and state growth tests |
| x/system-registry | System entity registry | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native | none | dependency graph and authority tests |
| x/taskgroups | Task groups | no | yes | no | no | 4 | KEEP_NATIVE | Keep native if scheduler/protocol task coordination | none | task bounded loop and state tests |
| x/tokenfactory | Native token factory app logic | yes | yes | yes | yes | 5 | MOVE_TO_CONTRACT | Mark deprecated-after-contract-replacement, keep until AFT-44 replacement passes | AFT-44 TokenFactory, TokenMaster, TokenWallet contracts and compatibility tests | native tokenfactory suite plus contract mint/burn/admin/metadata |
| x/treasury | Protocol treasury | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native protocol treasury | none | spend limits, distribution, authority tests |
| x/validator-economy | Validator economy types | no | no | no | no | 3 | KEEP_NATIVE | Keep native validator economics | none | market/commission/safety tests |
| x/validator-election | Validator election | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native | none | validator set transition and power cap tests |
| x/validator-insurance | Validator insurance | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native validator safety module | none | insurance, slash-first, claims tests |
| x/validator-registry | Validator registry | yes | yes | yes | no | 1 | KEEP_NATIVE | Keep native | none | key separation, status, export/import tests |
| x/vm | VM/runtime type stack | yes | no | no | no | 29 | CONSOLIDATE_AVM_RUNTIME | Fold under `x/avm` runtime API | `x/avm` boundary finalized | VM gas, storage, actor, receipt tests |
| x/workflow | Workflow app logic | no | no | no | no | 1 | MOVE_TO_CONTRACT | Mark deprecated native app logic; move to contracts or SDK tooling | contract workflow/tooling replacement | workflow contract/tool tests |
| x/zones | Zone system infrastructure plus app-specific financial zone | yes | yes | no | no | 9 | KEEP_NATIVE with extraction | Keep zone infrastructure; extract financial app logic to contracts | contract replacements for tokenfactory/dex/payment channel state | zone infra tests plus financial extraction compatibility |

## Safe Migration Phases

### Phase 1: Audit and Marking

Completed by this document:

- No native application module is deleted.
- `x/tokenfactory`, `x/dex`, `x/services`, `x/market`, `x/workflow`, and
  `x/permissions` are marked as deprecated native app logic.
- `x/zones/types/financial_zone.go` is marked for extraction to AVM contracts.
- The staking boundary is fixed: native staking is protocol core; contract
  staking is only an application-layer facade over native staking.

### Phase 2: `x/avm` Runtime Boundary

Create a native `x/avm` boundary that owns the runtime surface while reusing the
existing `x/aetravm`, `x/vm`, `x/actors`, `x/messages`, `x/messaging`,
`x/execution`, `x/queue`, and `x/storage` packages as implementation layers.

Required `x/avm` protocol surface:

- deploy contract;
- execute external message;
- execute internal message;
- bounced message handling;
- contract storage and code storage;
- actor registry integration;
- gas accounting;
- storage rent integration;
- receipts and deterministic events;
- genesis/export/import;
- versioned migration path.

The boundary must expose a narrow keeper/API so application contracts do not
depend on scattered runtime packages directly.

### Phase 3: Contract Replacement Specs

Before any native app module can be removed, write and test replacement specs.

Token replacement:

- AFT-44 TokenFactory contract;
- TokenMaster contract;
- TokenWallet contract;
- mint, burn, admin, metadata through AVM external/internal messages;
- compatibility vectors for existing native `x/tokenfactory` behavior.

DEX replacement:

- AMM factory contract;
- pool contract;
- LP token through AFT-44;
- swap, add liquidity, remove liquidity through AVM messages;
- compatibility vectors for existing native `x/dex` behavior.

Identity/DNS replacement:

- native root-level protocol remains `x/identity-root`;
- native keeps `.aet` root, name uniqueness, reserved names, normalization,
  expiry bounds, and canonical root registry;
- contracts implement domain NFT, resolver, subdomain manager, domain auction,
  DAO-controlled domain policy, and NFT binding customization.

Zones financial extraction:

- remove tokenfactory roots from native zone state only after contract registry
  replacement exists;
- remove DEX pool/swap state only after AMM contracts exist;
- remove payment channel app-state only after payment contracts exist;
- preserve `x/zones` as native system zone infrastructure.

### Phase 4: Removal Gates

Native app modules can be removed only after all gates pass:

- replacement contract specs are implemented;
- contract deployment/execution paths use `x/avm` and pass gas/rent checks;
- genesis/export/import migration exists for live native state;
- compatibility tests prove old native state can be migrated or bridged;
- old native tests are either ported to contract tests or retained as migration
  compatibility tests;
- app wiring removal is tested with `go test ./...`;
- chain upgrade migration documents exact state transitions.

## Deprecated Native App Logic Markers

The following native packages are kept for compatibility but marked
`deprecated-after-contract-replacement`:

- `x/tokenfactory`;
- `x/dex`;
- `x/services`;
- `x/market`;
- `x/workflow`;
- `x/permissions`;
- app-specific financial logic inside `x/zones/types/financial_zone.go`.

These markers are not removal instructions. They are guardrails preventing more
application logic from being added to native modules while the AVM contract
replacement is being built.

## Test Plan

Phase 1 validation:

- `go test -count=1 ./app`
- `go test -count=1 ./x/aetravm/...`
- `go test -count=1 ./x/vm/...`
- `go test -count=1 ./x/tokenfactory/... ./x/dex/...`
- `go test -count=1 ./x/config/... ./x/constitution/... ./x/actor-registry/... ./x/avm-scheduler/...`

Future phase validation must add:

- `x/avm` genesis/export/import and migration tests;
- contract deployment and message execution tests;
- AFT-44 token replacement compatibility tests;
- AMM factory/pool compatibility tests;
- identity/DNS root plus contract resolver/NFT tests;
- financial zone extraction compatibility tests.
