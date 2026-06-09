# Aetheris Next-Gen Blockchain Architecture Tasks

Status: Internal task plan
Scope: Cosmos SDK v0.54+ multi-zone, sharded, message-driven Aetheris architecture
Visibility: Private, not for public repository inclusion

## 1. Architecture Design Tasks

### 1.1 Aether Core

Responsibilities:

- Validator set and consensus boundary.
- Global state root.
- Zone commitments.
- Message root commitments.
- Cross-zone finality ordering.
- Proof root registry.

Tasks:

- [ ] Define `x/aethercore` module boundaries.
- [ ] Define `AetherCoreParams`.
- [ ] Define `ZoneDescriptor`.
- [ ] Define `ZoneCommitment`.
- [ ] Define `GlobalStateRoot`.
- [ ] Define `GlobalMessageRoot`.
- [ ] Define `ExecutionReceiptRoot`.
- [ ] Implement zone registry keeper.
- [ ] Implement root aggregation keeper.
- [ ] Implement ABCI++ hooks for zone execution scheduling.
- [ ] Add deterministic proposal grouping by zone and shard.
- [ ] Add root query APIs.
- [ ] Add export/import support for core state.
- [ ] Add replay tests proving identical roots across nodes.

Text diagram:

```text
Aether Core
  |
  +-- Zone Registry
  +-- Global State Root
  +-- Message Root
  +-- Receipt Root
  +-- Proof Registry
  +-- Cross-Zone Ordering
```

### 1.2 Execution Flow

Target flow:

```text
tx/message enters mempool
  |
  v
classify by zone_id + shard_id
  |
  v
PrepareProposal groups independent workloads
  |
  v
FinalizeBlock executes zone batches
  |
  v
zones emit state roots and output messages
  |
  v
Aether Core commits zone roots + message roots
  |
  v
next block delivers eligible cross-zone messages
```

Tasks:

- [ ] Define transaction classification by `zone_id`.
- [ ] Define transaction classification by `shard_id`.
- [ ] Define sync execution path.
- [ ] Define async execution path.
- [ ] Define cross-zone delivery lifecycle.
- [ ] Define receipt creation rules.
- [ ] Define failure and bounce handling.
- [ ] Add integration tests for same-zone execution.
- [ ] Add integration tests for cross-zone execution.

## 2. Execution Zone Tasks

### 2.1 Shared Zone Interface

Every zone must expose:

- Independent state machine.
- Independent mempool lane.
- Independent execution pipeline.
- `zone_state_root` per block.
- Message inbox.
- Message outbox.
- Receipt root.

Tasks:

- [ ] Define `ZoneStateMachine` interface.
- [ ] Define `ExecuteZoneBatch(ctx, batch)`.
- [ ] Define `ApplyInboundMessage(ctx, msg)`.
- [ ] Define `ComputeZoneRoot(ctx)`.
- [ ] Define `ExportZone(ctx)`.
- [ ] Define `ImportZone(ctx, export)`.
- [ ] Add zone-local gas accounting.
- [ ] Add zone-local message queues.
- [ ] Add zone-local proof query surface.

### 2.2 Financial Zone

Scope:

- Bank.
- Fees.
- DEX.
- Tokenfactory.
- Payment settlement.

Tasks:

- [ ] Move financial state under `zone/financial/*` prefixes.
- [ ] Define financial zone keeper boundary.
- [ ] Add bank transfer routing by account key.
- [ ] Add fee bucket state.
- [ ] Add DEX pool state routing by pool ID.
- [ ] Add tokenfactory denom authority state.
- [ ] Add payment settlement state.
- [ ] Add financial zone message handlers.
- [ ] Add financial zone root.
- [ ] Add balance and pool proof queries.

### 2.3 Identity Zone

Scope:

- `.aet` registry.
- Resolver.
- Reverse lookup.
- NFT ownership.
- Delegation and grants.
- Auctions.

Tasks:

- [ ] Move identity state under `zone/identity/*` prefixes.
- [ ] Activate `x/identity` as zone-local state machine.
- [ ] Add name-hash based sharding.
- [ ] Add resolver proof queries.
- [ ] Add reverse lookup proof queries.
- [ ] Add NFT binding checks.
- [ ] Add cross-zone identity lookup message.
- [ ] Add identity response receipt.
- [ ] Add auction finalization receipt.

### 2.4 Application Zone

Scope:

- Apps.
- Workflows.
- Schedulers.
- Automation modules.

Tasks:

- [ ] Define application runtime interface.
- [ ] Define workflow state model.
- [ ] Define scheduler queue state.
- [ ] Add deterministic scheduled task execution.
- [ ] Add bounded work per block.
- [ ] Add app message inbox and outbox.
- [ ] Add app execution receipts.
- [ ] Add app state root.

### 2.5 Contract Zone

Scope:

- AVM runtime.
- Smart contracts.
- Contract storage.
- Contract ABI.
- Contract async messages.

Tasks:

- [ ] Define `x/avm` module boundary.
- [ ] Define code registry.
- [ ] Define contract instance registry.
- [ ] Define contract storage prefixes.
- [ ] Define contract message inbox.
- [ ] Define AVM execution receipts.
- [ ] Add contract state proof queries.
- [ ] Add ABI descriptor registry.
- [ ] Add contract zone root.

## 3. Compute Shard Tasks

### 3.1 Shard Model

Tasks:

- [ ] Define `ShardDescriptor`.
- [ ] Define `ShardLayout`.
- [ ] Define `ShardLayoutEpoch`.
- [ ] Define shard assignment modes:
  - deterministic key prefixing.
  - consistent hashing.
  - explicit object placement.
- [ ] Define `RouteKeyToShard(zone_id, key, layout_epoch)`.
- [ ] Add shard-local inbox.
- [ ] Add shard-local outbox.
- [ ] Add shard-local receipt root.
- [ ] Add shard-local state root.
- [ ] Add shard root aggregation into zone root.

### 3.2 Dynamic Split and Merge

Tasks:

- [ ] Define split trigger metrics:
  - gas utilization.
  - state size.
  - queue backlog.
  - write conflict rate.
  - proof generation latency.
- [ ] Define merge trigger metrics.
- [ ] Define split scheduler.
- [ ] Define merge scheduler.
- [ ] Define deterministic state migration plan.
- [ ] Add migration root for shard movement.
- [ ] Add in-flight message handling during split.
- [ ] Add in-flight message handling during merge.
- [ ] Add tests for routing stability across layout epochs.

## 4. Message Layer Tasks

### 4.1 AetherMessage Struct

Required fields:

```text
AetherMessage {
  msg_id
  sender
  receiver
  source_zone_id
  destination_zone_id
  source_shard_id
  destination_shard_id
  value_naet
  payload
  payload_type
  gas_limit
  forwarding_fee
  expiry_height
  bounce
  execution_mode
  nonce
}
```

Tasks:

- [ ] Define canonical binary encoding.
- [ ] Define `msg_id` derivation.
- [ ] Define replay protection.
- [ ] Define sender nonce scope.
- [ ] Define payload size limits.
- [ ] Define gas reservation rules.
- [ ] Define forwarding fee escrow.
- [ ] Define bounce payload format.
- [ ] Define message receipt format.

### 4.2 Queue Model

Tasks:

- [ ] Define output queue per shard.
- [ ] Define input queue per shard.
- [ ] Define retry queue.
- [ ] Define expiry queue.
- [ ] Define dead letter queue.
- [ ] Define deterministic queue ordering.
- [ ] Define queue root.
- [ ] Add queue proof query.
- [ ] Add bounded queue draining.

### 4.3 Routing Logic

Routing algorithm requirements:

- Hybrid hypercube routing.
- Shortest-path fallback.
- Capacity-aware scoring.
- Congestion-aware routing.
- Deterministic tie-breaks.

Tasks:

- [ ] Define `RoutingTable`.
- [ ] Define `RoutingEpoch`.
- [ ] Define route cost function.
- [ ] Define committed congestion metrics.
- [ ] Define committed capacity metrics.
- [ ] Implement deterministic path selection.
- [ ] Implement route commitment hash.
- [ ] Add retry logic.
- [ ] Add expiry logic.
- [ ] Add bounce logic.
- [ ] Add route proof query.

## 5. AVM 2.0 Tasks

### 5.1 VM Core

Requirements:

- Stack-based execution.
- Deterministic instruction semantics.
- Gas metering per instruction.
- Store v2 KV abstraction.
- Async message creation.
- Proof verification.
- Conditional execution.
- ABI introspection.

Tasks:

- [ ] Define AVM bytecode format.
- [ ] Define instruction set versioning.
- [ ] Define stack frame model.
- [ ] Define memory model.
- [ ] Define deterministic error model.
- [ ] Define gas table.
- [ ] Define Store v2 KV adapter.
- [ ] Define contract account model.
- [ ] Define contract storage root.
- [ ] Define contract event root.

### 5.2 Instruction Set Overview

Tasks:

- [ ] Define stack instructions:
  - `PUSH`
  - `POP`
  - `DUP`
  - `SWAP`
- [ ] Define arithmetic instructions:
  - `ADD`
  - `SUB`
  - `MUL`
  - `DIV`
  - `MOD`
- [ ] Define control instructions:
  - `JMP`
  - `JMP_IF`
  - `CALL`
  - `RET`
  - `ABORT`
- [ ] Define storage instructions:
  - `KV_GET`
  - `KV_SET`
  - `KV_DELETE`
  - `KV_EXISTS`
- [ ] Define proof instructions:
  - `VERIFY_MERKLE_PROOF`
  - `VERIFY_ZONE_ROOT`
  - `VERIFY_MESSAGE_PROOF`
- [ ] Define message instructions:
  - `MSG_NEW`
  - `MSG_SEND`
  - `MSG_BOUNCE`
- [ ] Define promise instructions:
  - `PROMISE_NEW`
  - `PROMISE_RESOLVE`
  - `PROMISE_REJECT`
  - `PROMISE_TIMEOUT`
- [ ] Define ABI instructions:
  - `ABI_EXPORT`
  - `ABI_METHOD`
  - `ABI_EVENT`

### 5.3 Contract Runtime

Tasks:

- [ ] Define code upload flow.
- [ ] Define contract instantiate flow.
- [ ] Define sync contract call flow.
- [ ] Define async contract call flow.
- [ ] Define promise continuation flow.
- [ ] Define gas refund rules.
- [ ] Define contract migration flow.
- [ ] Add deterministic VM replay tests.
- [ ] Add gas fuzz tests.

## 6. Light-Client Proof System Tasks

### 6.1 Proof Root Hierarchy

Tasks:

- [ ] Define `app_hash` root hierarchy.
- [ ] Define `global_zone_root`.
- [ ] Define per-zone state roots.
- [ ] Define per-shard roots.
- [ ] Define `message_root`.
- [ ] Define `receipt_root`.
- [ ] Define `identity_root`.
- [ ] Define `resolver_root`.
- [ ] Define `contract_root`.
- [ ] Define `payment_root`.

### 6.2 Proof Types

Tasks:

- [ ] Define account state proof.
- [ ] Define balance proof.
- [ ] Define message inclusion proof.
- [ ] Define message receipt proof.
- [ ] Define zone state root proof.
- [ ] Define shard root proof.
- [ ] Define domain ownership proof.
- [ ] Define resolver record proof.
- [ ] Define contract state proof.
- [ ] Define payment settlement proof.
- [ ] Define non-existence proof.

### 6.3 Verification Flow

Tasks:

- [ ] Verify trusted header.
- [ ] Verify chain ID.
- [ ] Verify app hash.
- [ ] Verify zone root inclusion.
- [ ] Verify shard root inclusion.
- [ ] Verify Store v2 Merkle proof.
- [ ] Verify object-specific validity.
- [ ] Return typed proof failure codes.
- [ ] Add proof verifier SDK.
- [ ] Add proof test vectors.

## 7. Identity Integration Tasks

### 7.1 Cross-Zone Identity

Tasks:

- [ ] Define `MsgResolveIdentity`.
- [ ] Define `MsgIdentityResolutionResult`.
- [ ] Make identity lookup read-only in Identity Zone.
- [ ] Add proof option to identity lookup.
- [ ] Add resolver record version.
- [ ] Add resolver TTL handling.
- [ ] Add reverse lookup proof.
- [ ] Add delegation proof.
- [ ] Add NFT ownership proof.
- [ ] Add expiry and auction proof.

### 7.2 VM-Native Resolver

Tasks:

- [ ] Define native resolver interface.
- [ ] Define VM resolver adapter.
- [ ] Bound resolver contract gas.
- [ ] Bound resolver output size.
- [ ] Add proof format for resolver contract output.
- [ ] Add fallback to native resolver.

## 8. Payment Routing Tasks

### 8.1 Native Micropayments

Tasks:

- [ ] Define `PaymentChannel`.
- [ ] Define `VirtualPaymentChannel`.
- [ ] Define `ConditionalPayment`.
- [ ] Define `PaymentRoute`.
- [ ] Define `LiquidityReservation`.
- [ ] Define `SettlementProof`.
- [ ] Define `PaymentReceipt`.

### 8.2 Settlement Rules

Tasks:

- [ ] Lock collateral in Financial Zone.
- [ ] Support signed off-chain state updates.
- [ ] Support unilateral close.
- [ ] Support stale-state dispute.
- [ ] Support hash-locked payments.
- [ ] Support time-locked payments.
- [ ] Support route fee accounting.
- [ ] Support cross-zone payment messages.
- [ ] Add proof-queryable final settlement.

### 8.3 Payment Safety

Tasks:

- [ ] Define fraud proof format.
- [ ] Define challenge period.
- [ ] Define timeout ordering for routed payments.
- [ ] Define bounce and refund behavior.
- [ ] Define penalty routing.
- [ ] Add invariant tests for collateral conservation.

## 9. Performance Strategy Tasks

### 9.1 BlockSTM and Zones

Tasks:

- [ ] Group transactions by zone.
- [ ] Group transactions by shard.
- [ ] Group transactions by object key where available.
- [ ] Avoid global counters in hot paths.
- [ ] Use per-zone fee accumulators.
- [ ] Use per-shard message queues.
- [ ] Add expected versions for conflicting updates.
- [ ] Add BlockSTM conflict profiling.
- [ ] Add benchmarks for independent zone execution.
- [ ] Add benchmarks for independent shard execution.

### 9.2 Mempool Separation

Tasks:

- [ ] Define zone-specific mempool lanes.
- [ ] Define shard-specific sublanes.
- [ ] Define message-class priority lanes.
- [ ] Define fee-aware ordering per zone.
- [ ] Define expiry-aware ordering for messages.
- [ ] Add sender and target rate limits.
- [ ] Add proposal grouping validation.

### 9.3 Store v2 Strategy

Tasks:

- [ ] Define object-store records for accounts.
- [ ] Define object-store records for domains.
- [ ] Define object-store records for contracts.
- [ ] Define object-store records for channels.
- [ ] Define KV hybrid layout for contract storage.
- [ ] Define prefix proofs for shard state.
- [ ] Add bounded range scan rules.
- [ ] Add proof generation benchmarks.

## 10. Safety Tasks

### 10.1 Determinism

Tasks:

- [ ] Ban external API calls in consensus execution.
- [ ] Ban local wall-clock use in state transitions.
- [ ] Ban nondeterministic map iteration.
- [ ] Ban floating point math in consensus.
- [ ] Require canonical encoding for all commitments.
- [ ] Add deterministic replay test suite.

### 10.2 Routing Safety

Tasks:

- [ ] Commit routing table before use.
- [ ] Commit congestion metrics before use.
- [ ] Define deterministic path tie-breaks.
- [ ] Prove failed routes cannot burn value without receipt.
- [ ] Prove bounce cannot create value.

### 10.3 Shard Safety

Tasks:

- [ ] Allow shard layout changes only at epoch boundary.
- [ ] Store old shard layout through proof horizon.
- [ ] Include delivery epoch in in-flight messages.
- [ ] Emit migration roots.
- [ ] Add split and merge replay tests.

## 11. Cosmos SDK Module Map

### 11.1 New Modules

Tasks:

- [ ] Implement `x/aethercore`.
- [ ] Implement `x/zones`.
- [ ] Implement `x/shards`.
- [ ] Implement `x/msgbus`.
- [ ] Implement `x/proofregistry`.
- [ ] Implement `x/avm`.
- [ ] Implement `x/identity`.
- [ ] Implement `x/payments`.
- [ ] Implement `x/zonefees`.
- [ ] Implement `x/zonemempool`.

### 11.2 Existing Module Modifications

Bank:

- [ ] Add zone-aware account routing.
- [ ] Add cross-shard escrow transfer flow.
- [ ] Add message-driven transfer settlement.

Staking:

- [ ] Expose validator set commitment to Aether Core.
- [ ] Preserve CometBFT validator update semantics.
- [ ] Add validator metadata for zone service support.

Slashing:

- [ ] Keep consensus slashing in core.
- [ ] Separate payment fraud penalties from validator slashing.

Mint and distribution:

- [ ] Aggregate zone fees into distribution flow.
- [ ] Preserve deterministic reward accounting.

Fees:

- [ ] Add zone-local fee policies.
- [ ] Add forwarding fee escrow.
- [ ] Add congestion metrics by zone and shard.

Tokenfactory:

- [ ] Add zone-aware denom authority.
- [ ] Add token routing rules for cross-zone messages.

DEX:

- [ ] Add pool shard placement.
- [ ] Add async swap flow for cross-shard routes.
- [ ] Add pool proof queries.

## 12. State Model Tasks

### 12.1 Global Keys

Tasks:

- [ ] Define `core/zones/{zone_id}`.
- [ ] Define `core/zone_roots/{height}/{zone_id}`.
- [ ] Define `core/message_roots/{height}`.
- [ ] Define `core/proof_roots/{height}/{root_type}`.
- [ ] Define `zone/{zone_id}/state/*`.
- [ ] Define `zone/{zone_id}/inbox/*`.
- [ ] Define `zone/{zone_id}/outbox/*`.
- [ ] Define `zone/{zone_id}/receipts/*`.
- [ ] Define `zone/{zone_id}/shard/{shard_id}/state/*`.

### 12.2 Zone Keys

Financial:

- [ ] Define `financial/accounts/{address}`.
- [ ] Define `financial/balances/{address}/{denom}`.
- [ ] Define `financial/dex/pools/{pool_id}`.
- [ ] Define `financial/payments/channels/{channel_id}`.

Identity:

- [ ] Define `identity/domains/{name_hash}`.
- [ ] Define `identity/resolvers/{name_hash}`.
- [ ] Define `identity/reverse/{address}`.
- [ ] Define `identity/nft_bindings/{name_hash}`.

Application:

- [ ] Define `apps/instances/{app_id}`.
- [ ] Define `apps/workflows/{workflow_id}`.
- [ ] Define `apps/scheduler/{bucket}/{job_id}`.

Contract:

- [ ] Define `contract/code/{code_id}`.
- [ ] Define `contract/instance/{contract_addr}`.
- [ ] Define `contract/storage/{contract_addr}/{key}`.
- [ ] Define `contract/abi/{code_id}`.

## 13. Migration Path

### Phase 0: Baseline Hardening

Tasks:

- [ ] Finalize current module boundary documentation.
- [ ] Add state export validation.
- [ ] Add deterministic genesis import tests.
- [ ] Add bank, staking, slashing, distribution, and fee invariants.
- [ ] Add Store v2 compatibility audit.

### Phase 1: Core Commitments

Tasks:

- [ ] Implement `x/aethercore` skeleton.
- [ ] Register current chain as default zone.
- [ ] Commit default zone root.
- [ ] Commit empty message root.
- [ ] Add proof root registry.
- [ ] Add root query APIs.

### Phase 2: Message Bus

Tasks:

- [ ] Implement `x/msgbus`.
- [ ] Add message encoding and IDs.
- [ ] Add inbox and outbox stores.
- [ ] Add receipt stores.
- [ ] Add expiry and bounce logic.
- [ ] Add message inclusion proofs.

### Phase 3: Zone Extraction

Tasks:

- [ ] Extract Financial Zone.
- [ ] Extract Identity Zone.
- [ ] Extract Application Zone.
- [ ] Add Contract Zone skeleton.
- [ ] Add zone-specific keepers.
- [ ] Add zone execution summaries.

### Phase 4: Sharding Runtime

Tasks:

- [ ] Implement `x/shards`.
- [ ] Add shard layout descriptors.
- [ ] Add shard route-key calculation.
- [ ] Add shard inbox and outbox.
- [ ] Add shard root aggregation.
- [ ] Add split and merge scheduler.

### Phase 5: AVM 2.0

Tasks:

- [ ] Implement AVM bytecode format.
- [ ] Implement deterministic interpreter.
- [ ] Implement gas table.
- [ ] Implement contract storage adapter.
- [ ] Implement message syscalls.
- [ ] Implement proof verification syscalls.
- [ ] Implement ABI registry.

### Phase 6: Identity and Payment Integration

Tasks:

- [ ] Activate `.aet` identity proofs.
- [ ] Add cross-zone identity lookup.
- [ ] Add payment channel settlement.
- [ ] Add conditional payment routing.
- [ ] Add wallet SDK helpers.

### Phase 7: Performance Hardening

Tasks:

- [ ] Enable BlockSTM workloads for zone and shard batches.
- [ ] Add conflict profiling.
- [ ] Add Store v2 benchmarks.
- [ ] Add mempool lanes.
- [ ] Add congestion-aware routing.
- [ ] Add AdaptiveSync recovery tests.
- [ ] Add multi-zone load simulation.

## 14. Required Tests

Determinism:

- [ ] Same block produces identical zone roots across nodes.
- [ ] Same block produces identical message roots across nodes.
- [ ] Same routing table produces identical paths.
- [ ] Same shard layout produces identical shard IDs.
- [ ] Same AVM bytecode produces identical output.

Invariants:

- [ ] Zone root includes all shard roots.
- [ ] Message outbox inclusion has matching receipt or pending status.
- [ ] Cross-zone value transfer conserves `naet`.
- [ ] Payment settlement cannot overpay collateral.
- [ ] Identity resolver proof matches Identity Zone root.
- [ ] Contract state proof matches Contract Zone root.
- [ ] Shard split preserves all state keys.
- [ ] Shard merge preserves all state keys.

Simulations:

- [ ] High-volume bank transfers across shards.
- [ ] Identity resolver update bursts.
- [ ] Contract async call chains.
- [ ] Payment route timeout and bounce.
- [ ] DEX pool updates under shard conflict.
- [ ] Cross-zone congestion.
- [ ] Shard split under sustained load.
- [ ] Node recovery with AdaptiveSync during active message queues.

Performance:

- [ ] Local zone TPS.
- [ ] Cross-shard message throughput.
- [ ] Cross-zone message throughput.
- [ ] AVM instruction throughput.
- [ ] Store v2 proof generation latency.
- [ ] BlockSTM conflict rate by workload.
- [ ] Mempool grouping effectiveness.
- [ ] State sync time with multiple zones.

## 15. Acceptance Criteria

- [ ] Aether Core commits zone roots and message roots.
- [ ] At least one zone executes through the zone adapter.
- [ ] Messages have deterministic routing, inclusion proofs, and receipts.
- [ ] Store v2 key layout supports zone and shard proof generation.
- [ ] BlockSTM tests show independent shard execution can run in parallel.
- [ ] Shard split and merge rules are deterministic from committed state.
- [ ] AVM 2.0 instruction set and gas table are specified.
- [ ] Identity resolution is proof-backed and cross-zone callable.
- [ ] Payment settlement is trustless and proof-verifiable.
- [ ] Migration preserves current Aetheris module state and invariants.
