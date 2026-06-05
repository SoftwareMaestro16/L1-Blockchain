# Aetheris Roadmap V4: Modular L1 Execution OS Implementation

This roadmap turns the target design in
`docs/architecture/aetheris-modular-execution-os.md` and the safety gates in
`docs/architecture/sharding-rd.md` into implementation tasks.

V4 is an implementation backlog, not a production claim. Aetheris must not
claim production sharding, production Execution Zones, production Aether Mesh,
or production AVM/CosmWasm state mutation until the simulator, prototype,
adversarial tests, long-run testnet, independent audit, and consensus-safety
proof are complete.

## Source Documents

- `docs/architecture/aetheris-modular-execution-os.md`
- `docs/architecture/sharding-rd.md`
- `docs/architecture/execution-os.md`
- `docs/architecture/async-smart-contract-execution.md`
- `docs/security/cosmos-security-checklist.md`
- `docs/security/determinism-gate.md`
- `docs/security/cosmwasm-readiness.md`

## Non-Negotiable Rules

- Aether Core remains the control plane only.
- Aether Core must not execute smart contracts.
- Aether Core must not process application-specific business logic.
- All protocol fees remain native-only in `naet`.
- Zero address remains forbidden by default.
- All consensus-critical math uses deterministic integer or fixed-point decimal
  math; no floats in state transitions.
- No consensus path may use local wall-clock time, local latency, goroutine
  races, random values, map iteration order, external APIs, or machine learning.
- Every new persistent state type must support deterministic validate,
  init-genesis, export-genesis, import, migration, and corruption tests.
- All prototype zone/shard/Mesh functionality remains feature-gated until its
  production gate is accepted.
- DEX is out of scope for V4 except as an Application Zone example.

## Phase 0: Baseline Audit And Protection

Goal: prevent the roadmap implementation from destabilizing the already
working base chain.

Implementation tasks:

- Inventory all currently wired modules in `app/modules.go`, `app/keepers.go`,
  and `app/app.go`.
- Label each existing package as one of:
  - production-wired consensus module;
  - pure executable spec;
  - simulator;
  - docs/test/tooling only.
- Add or update docs explaining that `x/sharding/sim`, `x/execution/types`,
  `x/vm/types`, `x/aetherisvm/*`, `x/identity/types`, `x/messaging/types`,
  `x/queue/types`, and related packages are not wired consensus modules unless
  explicitly mounted later.
- Confirm current dirty files before each implementation branch.
- Keep untracked localnet state, keys, `.work`, and generated node homes out of
  commits.

Tests and checks:

- `go test ./...`
- `go vet ./...`
- `buf lint`
- `powershell -NoProfile -ExecutionPolicy Bypass -File tests\scripts\determinism_gate_test.ps1`

Acceptance:

- Auditors can tell which code affects AppHash and which code is only an
  executable specification.
- No roadmap implementation begins from an unknown dirty state.

## Phase 1: Deterministic LOAD_SCORE Executable Spec

Goal: implement the load model as a pure deterministic package before it is
used for routing, fees, or shard activation.

Recommended package:

- `x/load/types`

Core types:

- `Params`
  - `WindowBlocks uint64`
  - `AlphaNumerator uint64`
  - `AlphaDenominator uint64`
  - `MaxDeltaBps uint32`
  - `TargetMempoolSize uint64`
  - `TargetBlockGas uint64`
  - `TargetLatencyBlocks uint64`
  - `TargetExecutionSteps uint64`
  - metric weights in basis points
- `Metrics`
  - `CanonicalMempoolSize uint64`
  - `UsedBlockGas uint64`
  - `AverageInclusionDelayBlocks uint64`
  - `FailedTxCount uint64`
  - `TotalTxCount uint64`
  - `ExecutionStepCount uint64`
- `EMAState`
  - one fixed-point EMA value per metric
  - previous `LoadScoreBps`
  - window height
- `LoadBand`
  - `LOW`
  - `MEDIUM`
  - `HIGH`

Implementation tasks:

- Represent all scores as basis points from `0` to `10000`.
- Encode constants from the architecture doc:
  - `N = 60`
  - `alpha = 2 / (N + 1)`
  - `MAX_DELTA = 0.05`, represented as `500` bps.
  - weights: `2000`, `3000`, `2000`, `1000`, `2000`.
- Implement metric normalization:
  - `mempool_size_score = min(10000, mempool_size * 10000 / target_mempool_size)`
  - `block_utilization_score = min(10000, used_block_gas * 10000 / target_block_gas)`
  - `tx_latency_score = min(10000, avg_inclusion_delay_blocks * 10000 / target_latency_blocks)`
  - `failure_rate_score = failed_tx_count * 10000 / max(1, total_tx_count)`
  - `execution_time_score = min(10000, execution_step_count * 10000 / target_execution_steps)`
- Implement EMA with rational integer math:
  - `ema = (alpha_num * value + (alpha_den - alpha_num) * previous_ema) / alpha_den`
  - default `alpha_num = 2`
  - default `alpha_den = 61`
- Implement `ComputeLoadScore(params, previous, metrics)`.
- Apply per-block final score cap:
  - `abs(load_score_t - load_score_t_minus_1) <= MaxDeltaBps`.
- Implement load band mapping:
  - low: `0 <= score < 3000`
  - medium: `3000 <= score < 7000`
  - high: `7000 <= score <= 10000`
- Reject invalid params:
  - zero targets;
  - zero window;
  - weights not summing to `10000`;
  - `MaxDeltaBps > 10000`;
  - invalid alpha numerator/denominator.

Tests:

- Default params validate.
- Every metric normalizes to `[0,10000]`.
- Weight sum must equal `10000`.
- EMA is deterministic across repeated runs.
- One-block spike is capped by `MAX_DELTA`.
- Load score moves gradually across low, medium, high thresholds.
- Failure-rate division handles zero total tx count.
- Extreme uint64 inputs do not overflow.
- Export/import of `EMAState` preserves exact score.

Benchmarks:

- `BenchmarkLoadScoreUpdate1k`
- `BenchmarkLoadScoreUpdate10k`

Acceptance:

- Load score is fully deterministic and independent of local node behavior.
- Package has no dependency on Cosmos SDK context, CometBFT, wall-clock time,
  networking, goroutines, or external APIs.

## Phase 2: Deterministic Routing Engine Executable Spec

Goal: implement routing as a pure deterministic classifier before any app
module consumes it.

Recommended package:

- `x/routing/types`

Core types:

- `TxClass`
  - `CRITICAL_SYSTEM`
  - `STAKING_GOV_SECURITY`
  - `FINANCIAL`
  - `IDENTITY`
  - `CONTRACT`
  - `APPLICATION`
  - `ASYNC_MESSAGE`
- `ZoneID`
- `ShardID`
- `ReputationClass`
- `FeeClass`
- `RouteInput`
- `RouteDecision`
- `PriorityKey`

Implementation tasks:

- Implement `ClassifyTx` over stable message type strings, not Go reflection
  order.
- Implement locality extraction:
  - account key;
  - contract address;
  - `.aet` domain key;
  - asset denom key;
  - async message destination key.
- Implement zone selection:
  - financial tx -> Financial Zone;
  - identity tx -> Identity Zone;
  - contract tx -> Contract Zone;
  - app tx -> Application Zone;
  - staking/gov/security tx -> Aether Core handling path.
- Implement shard assignment:
  - `shard_id = hash(zone_id || primary_actor || routing_epoch) % active_shards(zone_id)`.
- Use SHA-256 or another already accepted deterministic hash helper.
- Implement priority key:
  - priority class;
  - effective fee class;
  - bounded reputation class;
  - admission height;
  - tx hash.
- Implement deterministic compare/sort for priority keys.
- Reject:
  - missing zone;
  - zero active shards;
  - empty primary actor for sharded classes;
  - unknown tx class;
  - invalid zero address actor;
  - non-`naet` protocol fee class.

Tests:

- Same input always produces same route.
- Different map insertion order produces same route and priority ordering.
- High-priority system tx sorts before normal user tx.
- Fee overpayment cannot exceed bounded fee class.
- Reputation class is bounded.
- Tx hash tie-breaker is deterministic.
- Zero active shards fails safely.
- Unknown tx class fails safely.
- Zero address primary actor is rejected.

Acceptance:

- Routing can be audited without running a node.
- Routing has no dependency on local mempool order, local latency, wall clock,
  randomness, or validator preference.

## Phase 3: Zone Registry Simulator

Goal: model Execution Zones before keepers or stores are added.

Recommended package:

- `x/zones/types`

Core types:

- `Zone`
  - `ID`
  - `Kind`
  - `VMPolicy`
  - `FeePolicy`
  - `GenesisStateHash`
  - `StateTransitionID`
  - `UpgradePolicy`
  - `DataAvailabilityPolicy`
  - `AuditStatus`
  - `ActivationHeight`
- `ZoneCommitment`
  - `ZoneID`
  - `ZoneHeight`
  - `StateRoot`
  - `ReceiptRoot`
  - `MessageRoot`
  - `ExecutionResultRoot`
  - `PreviousCommitment`
  - `CommitmentHash`
- `ZoneRegistryState`

Implementation tasks:

- Implement deterministic zone id validation.
- Implement governance-proposal-like zone registration as a pure state
  transition.
- Implement activation height checks.
- Implement zone commitment hashing:
  - `hash(zone_id, zone_height, state_root, receipt_root, message_root, execution_result_root, previous_zone_commitment)`.
- Implement commitment chain validation.
- Implement export/import validation:
  - duplicate zone id rejected;
  - missing previous commitment rejected;
  - invalid root formats rejected;
  - non-canonical ordering rejected.
- Require all zones to use `naet` protocol fee policy.
- Require VM policy to be one of:
  - `AVM`
  - `COSMWASM_GATED`
  - `NATIVE_MODULE`

Tests:

- Register and activate Financial, Identity, Application, Contract zones.
- Duplicate zone id rejected.
- Zone cannot activate before activation height.
- Commitment hash is stable.
- Commitment chain detects tampering.
- Export/import round trip is deterministic.
- Non-`naet` fee policy rejected.
- Unknown VM rejected.

Acceptance:

- Zone registry semantics are clear before SDK keeper storage is introduced.

## Phase 4: Compute Shards And Load-Driven Shard Activation

Goal: extend the sharding simulator to use the deterministic load model and
route work into elastic partitions.

Recommended location:

- extend `x/sharding/sim`
- import or mirror pure types from `x/load/types` and `x/routing/types`

Implementation tasks:

- Add load state per workchain/zone.
- Add active shard count per zone.
- Add shard activation policy:
  - low load -> `1` active shard;
  - medium load -> configured partial shard count;
  - high load -> configured max shard count.
- Add cooldown policy for deactivation:
  - deactivation only after `CooldownBlocks` below lower threshold.
- Add deterministic shard split:
  - parent shard final header;
  - child shard ids;
  - inherited validator subset rule;
  - inherited message queue partition by routing key.
- Add deterministic shard merge:
  - children must be below load threshold;
  - queues merged by canonical message order;
  - receipt roots recomputed deterministically.
- Add data availability flag propagation.
- Add validator reassignment triggered by routing epoch.

Tests:

- Low load keeps one shard.
- Medium load activates partial sharding.
- High load activates full sharding.
- Spike load is capped by `MAX_DELTA`.
- Shard split keeps every message exactly once.
- Shard merge preserves message ordering.
- Validator reassignment is deterministic.
- Data unavailable shard cannot be finalized for routing.
- Export/import preserves load windows and shard activation state.

Benchmarks:

- routing table lookup;
- shard split;
- shard merge;
- state export/import;
- validator reassignment.

Acceptance:

- `x/sharding/sim` proves that load-driven sharding can be deterministic before
  any production keeper is written.

## Phase 5: Aether Mesh Simulator

Goal: implement cross-zone and cross-shard messaging with proof, receipt,
replay, bounce, and refund behavior in a simulator first.

Recommended package:

- `x/mesh/types` or extension to `x/sharding/sim`

Core types:

- `MeshMessage`
- `MeshProof`
- `MeshReceipt`
- `ReplayMarker`
- `BounceReceipt`
- `RefundReceipt`
- `MeshState`

Implementation tasks:

- Define message fields:
  - source zone;
  - source shard;
  - destination zone;
  - destination shard;
  - message id;
  - sender;
  - recipient;
  - asset commitment;
  - payload hash;
  - timeout height;
  - finality reference;
  - proof;
  - sequence.
- Implement deterministic message id:
  - `hash(source_zone, source_shard, destination_zone, destination_shard, nonce, payload_hash, source_logical_time)`.
- Implement source proof validation against finalized source commitment.
- Implement single-use receipt validation.
- Implement replay marker writes.
- Implement bounce/refund:
  - invalid destination -> bounce/refund;
  - expired message -> bounce/refund;
  - failed execution result -> bounce/refund;
  - bounced/refund message cannot create infinite refund loop.
- Implement deterministic message ordering:
  - source finality height;
  - source zone;
  - source shard;
  - message id;
  - destination zone;
  - destination shard;
  - sequence.

Tests:

- Valid cross-zone message succeeds.
- Duplicate message rejected.
- Duplicate receipt rejected.
- Missing source proof rejected.
- Stale source finality rejected.
- Wrong destination shard rejected.
- Expired message bounces.
- Failed execution refunds.
- Bounce/refund cannot double spend.
- Export/import preserves replay markers.

Acceptance:

- Mesh proves no-double-spend and deterministic ordering before keeper wiring.

## Phase 6: Identity Zone Executable Spec

Goal: implement `.aet` naming semantics as a pure executable spec before it is
made a production module.

Recommended package:

- extend `x/identity/types`

Core types:

- `Domain`
- `DomainOwner`
- `ResolverRecord`
- `ReverseRecord`
- `SubdomainRecord`
- `DomainNFT`
- `Auction`
- `IdentityState`

Implementation tasks:

- Implement deterministic domain normalization:
  - lowercase;
  - trim whitespace;
  - reject empty labels;
  - reject duplicate normalized names;
  - define max label and full domain length.
- Implement lifecycle:
  - available;
  - commit;
  - reveal/register;
  - active;
  - renewal window;
  - expired;
  - available.
- Implement resolver records:
  - domain -> address;
  - domain -> contract;
  - domain -> zone endpoint.
- Implement reverse lookup:
  - address -> primary domain;
  - update requires address owner authorization.
- Implement subdomains:
  - parent owner controls issuance;
  - child owner controls child resolver unless parent policy says otherwise.
- Implement NFT ownership model:
  - domain ownership represented by deterministic NFT id;
  - transfer updates owner and invalidates unauthorized pending resolver
    updates.
- Implement optional auction model:
  - sealed commit phase;
  - reveal phase;
  - deterministic tie-breaker by bid, reveal height, commitment hash;
  - losing bids refunded via deterministic receipts.

Tests:

- Register `.aet` domain.
- Duplicate normalized name rejected.
- Expired domain cannot resolve.
- Renewal preserves ownership.
- Resolver update requires owner.
- Reverse lookup requires address owner.
- Subdomain issuance requires parent owner.
- NFT transfer changes domain owner.
- Auction tie-breaker deterministic.
- Export/import preserves domain lifecycle.

Acceptance:

- Identity Zone can be simulated without keeper wiring and without spoofing or
  duplicate names.

## Phase 7: Contract Zone Readiness For AVM And Gated CosmWasm

Goal: prepare smart-contract execution without enabling production state
mutation prematurely.

Recommended packages:

- `x/vm/types`
- `x/aetherisvm/avm`
- `x/aetherisvm/async`
- `app/wasmconfig`

Implementation tasks:

- Keep AVM as the native Aetheris-defined VM direction.
- Keep CosmWasm behind explicit config/governance gate.
- Define contract upload policy:
  - governance-only by default;
  - allowlist optional only for testnet.
- Define instantiate policy:
  - code-owner-only by default;
  - public instantiate only when explicitly enabled.
- Define migration policy:
  - contract admin required;
  - admin cannot be zero address;
  - governance can disable migrations globally.
- Define gas model:
  - deploy gas;
  - execute gas;
  - query gas;
  - storage write gas;
  - message forwarding gas.
- Define contract limits:
  - max code size;
  - max state size;
  - max query response bytes;
  - max query depth;
  - max emitted messages;
  - max messages per block.
- Define storage ABI:
  - contract namespace;
  - deterministic key ordering;
  - bounded iteration;
  - export/import format.
- Define host functions:
  - no local time;
  - no randomness unless consensus-approved;
  - no external APIs;
  - no direct mutation of another contract's state.

Tests:

- AVM deploy/execute/query executable spec.
- Unauthorized upload rejected.
- Unauthorized instantiate rejected.
- Unauthorized migrate rejected.
- Zero admin rejected.
- Oversized code rejected.
- Oversized state rejected.
- Query response limit enforced.
- Gas limit enforced.
- Export/import preserves contract state and queues.
- Fuzz AVM parser and message handling.

Acceptance:

- Contract Zone can run as executable spec and testnet-only prototype without
  bypassing base-chain signer, fee, zero-address, staking, slashing, or
  governance validation.

## Phase 8: Keeper Prototype Behind Feature Gates

Goal: convert approved pure specs into SDK keepers without enabling production
claims.

Prototype modules:

- `x/load`
- `x/routing`
- `x/zones`
- `x/mesh`
- future `x/identity` keeper when Identity Zone spec is complete.

Implementation tasks:

- Add each keeper only after its executable spec has:
  - unit tests;
  - adversarial tests;
  - export/import tests;
  - benchmarks;
  - security review notes.
- Add module account permissions only where needed.
- Add KVStore keys only when persistence is required.
- Add params with governance authority.
- Add genesis validation and default genesis.
- Add migration skeletons.
- Add Msg and Query services only when the state transition requires public
  access.
- Add feature gate params:
  - disabled by default;
  - testnet profile can enable;
  - production enablement requires governance and software version gate.
- Ensure all keeper iteration is prefix-bounded and sorted.

Tests:

- Keeper default genesis validates.
- Corrupted genesis rejected.
- Export/import round trip deterministic.
- Feature disabled rejects mutating messages.
- Unauthorized authority rejected.
- Query pagination bounded.
- No panic from nil/malformed request.
- Migration from version `1` to next no-op path tested.

Acceptance:

- Prototype keepers exist but cannot alter production consensus behavior unless
  explicit feature gate is enabled.

## Phase 9: Aether Core Wiring Gate

Goal: wire mature prototype modules into Aether Core only after determinism and
security gates pass.

Preconditions:

- Phase 1 through Phase 8 accepted.
- No untriaged critical/high security findings.
- No consensus nondeterminism findings.
- No failing export/import tests.
- No unbounded query or block execution paths.
- Public docs still say experimental sharding until production gate passes.

Implementation tasks:

- Add module manager wiring for accepted modules.
- Add store keys.
- Add keepers.
- Set BeginBlocker/EndBlocker order explicitly.
- Define whether routing happens in:
  - ante only;
  - PrepareProposal/ProcessProposal;
  - FinalizeBlock;
  - module message server.
- Use one deterministic routing decision per tx and persist only if required.
- Keep Aether Core free from contract execution.
- Commit only roots/receipts/messages to Core.

Tests:

- Same genesis + same tx sequence -> same AppHash.
- Export/import after routed txs is deterministic.
- Restart preserves routing/load/zone/shard state.
- Feature-disabled mainnet profile has no active production sharding behavior.
- 3-validator and 5-validator localnet smoke tests pass.

Acceptance:

- Aether Core coordinates execution without directly executing contract logic.

## Phase 10: Localnet, Testnet, And Operator Tooling

Goal: make modular execution testable by operators without manual guessing.

Implementation tasks:

- Add localnet profiles:
  - `base`: current chain only;
  - `execution-os-sim`: pure simulator commands;
  - `zones-prototype`: feature-gated zones;
  - `mesh-prototype`: feature-gated Mesh;
  - `identity-prototype`: feature-gated `.aet`.
- Add smoke tests:
  - load score update;
  - route tx to zone;
  - activate shard under load;
  - send cross-zone message;
  - process receipt;
  - register `.aet` domain;
  - restart and query state.
- Add diagnostics:
  - current load score;
  - active zones;
  - active shards;
  - pending Mesh messages;
  - replay marker count;
  - zone commitment roots.
- Add log redaction:
  - no mnemonics;
  - no private validator keys;
  - no raw secrets;
  - no local `.localnet` key material.

Tests:

- `tests/e2e/execution_os_smoke.ps1`
- `tests/e2e/zones_smoke.ps1`
- `tests/e2e/mesh_smoke.ps1`
- `tests/e2e/identity_smoke.ps1`
- restart persistence smoke for every enabled prototype.

Acceptance:

- An operator can run a feature-gated prototype locally and inspect state
  without reading source code.

## Phase 11: Security, Fuzzing, And Invariants

Goal: prove the modular execution model resists consensus failures, DoS, and
fund loss before public testnet.

Implementation tasks:

- Add attacker model per subsystem:
  - load model;
  - routing;
  - zones;
  - shards;
  - Mesh;
  - identity;
  - VM;
  - fees.
- Add invariants:
  - no duplicate messages;
  - no duplicate receipts;
  - no double spend across zones;
  - no zero address ownership;
  - no non-`naet` protocol fees;
  - zone roots match exported zone state;
  - shard queues preserve every message exactly once;
  - load score cannot jump more than `MAX_DELTA`;
  - routing is deterministic for same state and tx.
- Add fuzz tests:
  - malformed Mesh messages;
  - malformed domain labels;
  - malformed zone commitments;
  - shard split/merge sequences;
  - AVM message parser;
  - export/import corruption.
- Add benchmarks:
  - load score update;
  - routing decision;
  - zone commitment validation;
  - shard split/merge;
  - Mesh proof verification;
  - identity lookup;
  - export/import large state.

Acceptance:

- No high-risk subsystem proceeds to keeper wiring without adversarial,
  invariant, fuzz, and benchmark coverage.

## Phase 12: Public Testnet Production Gate

Goal: define the minimum conditions before public testnet can advertise modular
execution features.

Required gates:

- Base chain hardening complete.
- Determinism gate passes.
- Export/import gate passes.
- Genesis migration gate passes.
- 3-validator localnet long-run passes.
- 5-validator localnet long-run passes.
- 10-validator stress profile passes.
- State-sync and snapshot restore pass.
- Load score and routing simulator pass.
- Sharding simulator pass.
- Mesh simulator pass.
- Identity executable spec pass.
- VM readiness tests pass.
- Security scans pass:
  - govulncheck;
  - gosec;
  - CodeQL;
  - gitleaks;
  - dependency review.
- Independent audit findings are triaged.
- Public docs do not overclaim production sharding or production smart-contract
  execution.

Acceptance:

- Public testnet can advertise only the features that passed their gate.
- Any feature still behind R&D remains documented as experimental.

## Suggested Branch Order

1. `feature/load-score-spec`
2. `feature/routing-engine-spec`
3. `feature/zone-registry-sim`
4. `feature/load-driven-sharding-sim`
5. `feature/aether-mesh-sim`
6. `feature/identity-zone-spec`
7. `feature/contract-zone-readiness`
8. `prototype/execution-os-keepers`
9. `prototype/aether-core-routing-wiring`
10. `tooling/execution-os-localnet`
11. `security/execution-os-invariants`
12. `testnet/modular-execution-gate`

Each branch must end with:

```powershell
go test ./...
go vet ./...
buf lint
powershell -NoProfile -ExecutionPolicy Bypass -File tests\scripts\determinism_gate_test.ps1
```

Run `buf generate` only when protobuf files are changed.

## Traceability Matrix

| Design requirement | Implementation phase |
| --- | --- |
| Aether Core as control plane | Phase 9 |
| Execution Zones | Phase 3, Phase 8 |
| Compute Shards | Phase 4 |
| Deterministic LOAD_SCORE | Phase 1 |
| Load spike resistance | Phase 1, Phase 4 |
| Deterministic routing | Phase 2 |
| Aether Mesh | Phase 5 |
| `.aet` Identity Layer | Phase 6 |
| Economic security | Phase 0, Phase 9, Phase 12 |
| Low-fee congestion model | Phase 1, Phase 2, Phase 11 |
| Trilemma claim support | Phase 12 only after accepted gates |

## Final Rule

Do not wire production Execution Zones, production Compute Shards, production
Aether Mesh, or production contract execution into Aether Core until the
corresponding executable spec and simulator have passed deterministic tests,
adversarial tests, export/import tests, benchmarks, long-run localnet tests,
and independent audit review.
