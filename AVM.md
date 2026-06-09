# Aether Virtual Machine and Async Execution Engine

Status: Internal design document
Scope: AVM runtime, async execution, actor scheduling, zones, contracts, and interface system
Visibility: Private, not for public repository inclusion

## 1. Overview

Aether Virtual Machine (`AVM`) is a deterministic, modular execution runtime for Aetheris.

AVM supports:

- Synchronous smart contract and module execution.
- Asynchronous actor-based workflows.
- Cross-zone message passing.
- Scheduled and delayed execution.
- Continuation-based long-running workflows.
- Verifiable state transitions.
- Optional WASM-compatible contract backend.
- Interface-driven contract and service interaction.

AVM is not a single VM instance. It is a layered execution system:

```text
Aether Core
  |
  v
AVM Execution Router
  |
  +---------------------+
  | Sync Execution      |
  | Async Engine        |
  | Actor Runtime       |
  | Contract Backends   |
  +---------------------+
  |
  v
Zone State Roots / AppHash
```

## 2. Design Goals

### 2.1 Determinism First

All AVM execution paths must be:

- Deterministic.
- Replayable.
- Consensus-safe.
- Order-controlled.
- Gas-bounded.
- Exportable and importable.

Consensus execution must not use:

- External API calls.
- Local wall-clock time.
- Runtime randomness.
- Non-deterministic map iteration.
- Unbounded queue draining.
- Unmetered storage iteration.

### 2.2 Async-Native Architecture

AVM must support cross-block execution as a first-class primitive:

- Message scheduling across blocks.
- Delayed execution queues.
- Retry queues.
- Actor mailboxes.
- Continuation state.
- Bounce handling.
- Dead letter queues.

Async execution must remain deterministic:

- Queue ordering is canonical.
- Retry policy is bounded.
- Expiry handling is deterministic.
- Failed execution produces committed receipt.

### 2.3 Cosmos SDK Compatibility

AVM must integrate with:

- ABCI++ lifecycle.
- `FinalizeBlock` execution.
- MsgServer and keeper architecture.
- Store v2 state access.
- KVStore-compatible state layout.
- BlockSTM-ready parallel execution.
- Staking-secured execution and CometBFT finality.

### 2.4 Zone Isolation

AVM execution is partitioned into zones:

- Financial Zone.
- Identity Zone.
- Application Zone.
- Contract Zone.

Each zone has:

- Independent state root.
- Independent message queue.
- Execution budget.
- Gas policy.
- Message filters.
- State namespace.
- Proof root contribution.

### 2.5 Developer UX

AVM must expose:

- Formal interfaces.
- Method descriptors.
- Event descriptors.
- Async handler descriptors.
- Get-method descriptors.
- CLI and SDK auto-binding metadata.
- Wallet UI generation metadata.

Interface metadata is advisory for clients but hash-committed for verification.

## 3. Core Components

### 3.1 Component Map

```text
+-------------------------------------------------------------+
|                       Aether Core                           |
| consensus | block lifecycle | global roots | zone registry   |
+----------------------------+--------------------------------+
                             |
                             v
+-------------------------------------------------------------+
|                    AVM Execution Router                     |
| route tx/msg | classify zone | validate budget | dispatch    |
+----------+------------------+------------------+-------------+
           |                  |                  |
           v                  v                  v
+----------------+   +----------------+   +-------------------+
| Sync Engine    |   | Async Engine   |   | Actor Runtime     |
| MsgServer      |   | Queues         |   | Mailboxes         |
| Keeper Calls   |   | Scheduling     |   | Continuations     |
+----------------+   +----------------+   +-------------------+
           |                  |                  |
           +------------------+------------------+
                              |
                              v
+-------------------------------------------------------------+
|                    Contract Backends                        |
| Native Modules | AVM Actors | Optional WASM Adapter          |
+-------------------------------------------------------------+
                              |
                              v
+-------------------------------------------------------------+
|                Zone State / Store v2 / Roots                |
+-------------------------------------------------------------+
```

### 3.2 AVM Router

The AVM Router dispatches transactions and async messages to the correct execution path.

Routing pipeline:

```text
Msg -> Decode -> Validate Envelope -> Classify -> Route -> Execute -> Receipt -> Root
```

Routing inputs:

- Message type.
- Target zone.
- Target module.
- Target actor or contract.
- Gas class.
- Execution priority.
- Scheduling metadata.
- Domain-specific route key.

Routing outputs:

- Execution target.
- Zone ID.
- Queue ID.
- Gas meter.
- Dispatch mode.
- Receipt policy.

Routing rules:

- Financial messages route to Financial Zone.
- Identity and resolver messages route to Identity Zone.
- Scheduler and workflow messages route to Application Zone.
- Contract messages route to Contract Zone.
- Cross-zone writes become async messages.
- Unsupported message types are rejected before state mutation.

### 3.3 Sync Execution Engine

Purpose:

- Execute traditional block-bound transactions.
- Preserve Cosmos SDK module semantics.
- Support existing bank, staking, governance, tokenfactory, and DEX flows.

Execution model:

```text
Tx -> AnteHandler -> MsgServer -> Keeper -> Store v2/KVStore -> Events -> Receipt
```

Properties:

- Atomic.
- Gas metered.
- Executes within current block.
- Produces immediate state transition.
- Fails with deterministic error.
- Emits committed execution receipt where configured.

Use cases:

- Bank transfers.
- Staking operations.
- Governance actions.
- Tokenfactory operations.
- DEX swaps and pool updates.
- Identity registration and resolver updates.
- Payment settlement finalization.

### 3.4 Async Execution Engine

Purpose:

- Support non-block-bound workflows.
- Execute scheduled messages in future blocks.
- Process cross-zone messages.
- Implement actor-style continuation.
- Allow bounded retry and bounce semantics.

Core objects:

- `AsyncMessage`.
- `ZoneQueue`.
- `RetryPolicy`.
- `ExecutionReceipt`.
- `DeadLetterRecord`.
- `Continuation`.

### 3.5 Actor Runtime Layer

Actor model:

```text
Actor {
  actor_id
  code_ref
  state_root
  mailbox
  handler(message)
}
```

Properties:

- Single-threaded per actor.
- No shared mutable state across actors.
- Communication by message only.
- Mailbox ordering is deterministic.
- Actor state writes are isolated by actor key prefix.
- Actor can emit async messages.

### 3.6 Continuation Storage

Continuations represent paused workflows.

```text
Continuation {
  continuation_id
  actor_id
  step_index
  partial_state
  resume_height
  expiry_height
  gas_reserved
  status
}
```

Use cases:

- Multi-step contracts.
- Delayed computation pipelines.
- Payment timeouts.
- Cross-zone workflows.
- Long-running service orchestration.
- Retryable actor execution.

Rules:

- Continuation state must be deterministic and bounded.
- Resume height must be explicit.
- Continuation expiry must be explicit.
- Continuation can resume only through scheduler.
- Expired continuation emits failure receipt.

## 4. State Model

### 4.1 Store Prefixes

AVM state prefixes:

- `avm/params` -> `AVMParams`
- `avm/zones/{zone_id}` -> `ZoneRuntimeConfig`
- `avm/router/routes/{route_key}` -> `RouteDescriptor`
- `avm/async/messages/{message_id}` -> `AsyncMessage`
- `avm/async/queues/{zone_id}/{queue_id}/{sort_key}` -> `message_id`
- `avm/async/retry/{zone_id}/{height}/{message_id}` -> `RetryRecord`
- `avm/async/dead/{zone_id}/{message_id}` -> `DeadLetterRecord`
- `avm/actors/{actor_id}` -> `ActorRecord`
- `avm/actors/mailbox/{actor_id}/{sort_key}` -> `message_id`
- `avm/continuations/{continuation_id}` -> `Continuation`
- `avm/contracts/code/{code_id}` -> `CodeRecord`
- `avm/contracts/instances/{contract_addr}` -> `ContractRecord`
- `avm/contracts/storage/{contract_addr}/{key}` -> `StorageValue`
- `avm/interfaces/{interface_hash}` -> `InterfaceDescriptor`
- `avm/receipts/{receipt_id}` -> `ExecutionReceipt`
- `avm/roots/{height}` -> `AVMRoot`

### 4.2 Zone Runtime Config

Fields:

- `zone_id`
- `enabled`
- `execution_budget_per_block`
- `async_budget_per_block`
- `max_queue_depth`
- `max_message_bytes`
- `gas_policy_id`
- `retry_policy_id`
- `allowed_message_types`
- `state_root_prefix`
- `message_root_prefix`
- `continuation_root_prefix`

### 4.3 AVM Root

Each block computes:

```text
avm_root = hash(
  router_root,
  async_message_root,
  actor_root,
  contract_root,
  continuation_root,
  interface_root,
  receipt_root
)
```

Each zone computes:

```text
zone_root = hash(
  state_root,
  message_root,
  execution_root,
  continuation_root
)
```

### 4.4 State Invariants

- Every queued message has a stored message record.
- Every executed message has exactly one receipt.
- Every expired message has an expired receipt.
- Every bounced message references original message ID.
- Every continuation references an existing actor or workflow.
- Actor mailbox entries must be ordered by deterministic sort key.
- Contract storage keys must be scoped by contract address.
- Zone root must include state, message, execution, and continuation roots.

## 5. Async Message Specification

### 5.1 AsyncMessage

```text
AsyncMessage {
  id
  source
  destination
  payload
  gas_limit
  delay_height
  expiry_height
  retry_policy
  bounce_flag
}
```

Extended fields:

- `source_zone`
- `destination_zone`
- `source_actor_optional`
- `destination_actor_optional`
- `sender_nonce`
- `payload_type`
- `payload_hash`
- `value_naet`
- `forwarding_fee`
- `priority`
- `created_height`
- `route_hint_optional`
- `auth_proof_optional`
- `state_proof_optional`

### 5.2 Message ID

Message ID derivation:

```text
id = H(
  chain_id,
  source_zone,
  source,
  sender_nonce,
  destination_zone,
  destination,
  payload_hash,
  created_height
)
```

Rules:

- Message ID must be globally unique.
- Sender nonce is scoped by source zone and sender.
- Duplicate message ID is rejected.
- Consumed message creates replay tombstone.

### 5.3 RetryPolicy

Fields:

- `mode`
- `max_attempts`
- `retry_delay`
- `backoff_mode`
- `max_retry_height`
- `charge_retry_gas`

Modes:

- `none`
- `fixed`
- `bounded_backoff`

Rules:

- Retry count must be bounded.
- Retry delay must be deterministic.
- Retry cannot exceed message expiry.
- Retry gas must be reserved or explicitly charged.

### 5.4 ExecutionReceipt

Fields:

- `receipt_id`
- `message_id`
- `zone_id`
- `executor`
- `status`
- `gas_used`
- `storage_written`
- `events_hash`
- `output_messages_root`
- `error_code_optional`
- `created_height`
- `receipt_hash`

Status values:

- `submitted`
- `scheduled`
- `executed`
- `failed`
- `retried`
- `expired`
- `bounced`
- `dead_lettered`

### 5.5 DeadLetterRecord

Fields:

- `message_id`
- `zone_id`
- `reason`
- `failed_attempts`
- `last_error_code`
- `final_height`
- `refund_amount_optional`
- `receipt_id`

Rules:

- Dead letter records are terminal.
- Dead letter records must be proof-queryable.
- Dead letter records may trigger bounce if enabled.

## 6. Execution Queues

### 6.1 ZoneQueue

Each zone maintains:

```text
ZoneQueue {
  priority_queue
  delayed_queue
  retry_queue
  failed_queue
}
```

State prefixes:

- `queue/priority/{zone_id}/{sort_key}`.
- `queue/delayed/{zone_id}/{resume_height}/{sort_key}`.
- `queue/retry/{zone_id}/{retry_height}/{sort_key}`.
- `queue/failed/{zone_id}/{sort_key}`.

### 6.2 Deterministic Scheduling Rule

Execution order:

```text
priority -> scheduled_height -> sender_hash -> nonce -> message_id
```

Rules:

- Higher priority executes first.
- Messages below `delay_height` are not eligible.
- Messages above `expiry_height` expire before execution.
- Same sender messages execute in nonce order.
- Ties resolve by `message_id`.
- Per-block work is bounded by zone budget.

### 6.3 Queue Lifecycle

```text
message submitted
  |
  v
stored in zone queue
  |
  v
scheduled for execution
  |
  v
executed in eligible block
  |
  +--> committed receipt
  |
  +--> retry queue
  |
  +--> bounce message
  |
  +--> dead letter queue
```

### 6.4 Queue Implementation Tasks

- Implement queue state prefixes.
- Implement deterministic sort key.
- Implement queue admission validation.
- Implement bounded queue draining.
- Implement delayed queue promotion.
- Implement retry queue promotion.
- Implement failed queue and dead letter records.
- Add queue root computation.
- Add queue proof queries.

## 7. Block Lifecycle Integration

### 7.1 Execution Order

Recommended AVM block flow:

```text
BeginBlock
  |
  v
Execute Sync Tx
  |
  v
Process Async Queue
  |
  v
Execute Scheduled Messages
  |
  v
Process Continuations
  |
  v
Finalize State Roots
```

### 7.2 ABCI++ Mapping

`PrepareProposal`:

- Classify sync and async transactions.
- Group by zone and target actor where possible.
- Include eligible scheduled messages.
- Respect per-zone execution budgets.

`ProcessProposal`:

- Verify deterministic ordering.
- Verify budget bounds.
- Verify message eligibility.
- Verify no expired messages are proposed for execution.

`FinalizeBlock`:

- Execute sync transaction path.
- Drain async queues by deterministic rule.
- Execute actor handlers.
- Resume continuations.
- Emit receipts.
- Commit roots.

`EndBlock`:

- Process bounded cleanup.
- Mark expired messages.
- Prune old tombstones according to proof horizon.
- Emit zone execution summaries.

### 7.3 Block Lifecycle Tasks

- Add AVM hooks in app execution pipeline.
- Add eligible-message selection.
- Add proposal verification for queue ordering.
- Add per-zone budget accounting.
- Add root finalization.
- Add block replay determinism tests.

## 8. Gas Model

### 8.1 Gas Classes

Gas is split into:

- Execution gas.
- Storage gas.
- Scheduling gas.
- Cross-zone routing gas.
- Proof verification gas.
- Continuation gas.
- Interface introspection gas.

### 8.2 Async Gas Reservation

Rules:

- Async message reserves gas upfront.
- Reserved gas is escrowed with the message.
- Execution consumes gas from reserve.
- Retry consumes gas if retry policy charges gas.
- Bounce consumes bounded gas.
- Unused gas may refund according to zone policy.

### 8.3 GasPolicy

Fields:

- `base_message_gas`
- `per_byte_payload_gas`
- `storage_read_gas`
- `storage_write_gas`
- `queue_insert_gas`
- `queue_pop_gas`
- `cross_zone_base_gas`
- `proof_verify_base_gas`
- `continuation_store_gas`
- `bounce_base_gas`

### 8.4 Gas Invariants

- Execution cannot exceed message gas limit.
- Zone cannot exceed per-block async budget.
- Contract cannot create message without routing gas reserve.
- Proof verification must be gas-metered.
- Storage writes must charge storage gas.
- Failed execution consumes gas used before failure.

### 8.5 Gas Implementation Tasks

- Define gas policy parameters.
- Add async gas escrow.
- Add refund accounting.
- Add per-zone budget meter.
- Add proof verification gas costs.
- Add storage byte metering.
- Add gas fuzz tests.

## 9. Cross-Zone Execution

### 9.1 Zone Message Routing

Flow:

```text
Zone A -> AVM Router -> Zone B Queue -> Zone B Execution -> Receipt
```

Each zone defines:

- Gas policy.
- Execution constraints.
- Message filters.
- Allowed opcodes.
- Bounce behavior.
- Proof requirements.

### 9.2 Cross-Zone Rules

- No direct state writes across zones.
- Cross-zone calls are async messages.
- Source zone commits output message.
- Destination zone commits receipt.
- Value transfers use escrow or message value accounting.
- Failed cross-zone execution must either bounce or dead-letter.

### 9.3 Zone Commitment Model

Each zone produces:

```text
zone_root = hash(
  state_root,
  message_root,
  execution_root,
  continuation_root
)
```

Aether Core commits all zone roots each block.

### 9.4 Cross-Zone Implementation Tasks

- Add zone router table.
- Add output message root per zone.
- Add destination inbox root per zone.
- Add cross-zone receipt root.
- Add value escrow and refund handling.
- Add cross-zone proof query.
- Add tests for failed route and bounce behavior.

## 10. Smart Contract Model

### 10.1 Contract Backends

AVM supports multiple execution backends:

- Native modules.
- Optional WASM contracts.
- AVM-native actor contracts.

### 10.2 Native Modules

Properties:

- Fastest execution path.
- Cosmos SDK keeper-backed.
- Fully deterministic.
- Existing MsgServer compatible.
- Best for core protocol functions.

Requirements:

- Must expose service/interface descriptor if used through AVM interface system.
- Must emit execution receipt when called through AVM route.
- Must not directly mutate other zones.

### 10.3 Optional WASM Contracts

Properties:

- Sandboxed execution.
- Portable bytecode.
- Gas metered runtime.
- Store v2-backed key-value adapter.

Requirements:

- Deterministic host functions only.
- Bounded memory.
- Gas conversion table.
- No external network access.
- Cross-zone calls only via async messages.

### 10.4 AVM-Native Actor Contracts

Properties:

- Async-first contracts.
- Stateful actors.
- Continuation-based logic.
- Mailbox-driven execution.

Actor contract state:

- `actor_id`
- `code_id`
- `owner`
- `mailbox_root`
- `state_root`
- `continuation_root`
- `interface_hash`
- `status`

Rules:

- Actor handles one message at a time.
- Actor can emit messages.
- Actor can store continuation.
- Actor cannot read another actor's mutable state directly.
- Actor execution emits receipt.

### 10.5 Contract Implementation Tasks

- Define backend interface.
- Define native module adapter.
- Define WASM adapter boundary.
- Define actor contract runtime.
- Add code registry.
- Add contract instance registry.
- Add contract storage prefixes.
- Add execution receipt emission.
- Add contract proof query.

## 11. Interface System

### 11.1 Interface Descriptor

Contracts and services expose:

```text
Interface {
  methods[]
  events[]
  async_handlers[]
  get_methods[]
}
```

Extended fields:

- `interface_hash`
- `interface_version`
- `owner`
- `target_type`
- `method_descriptors`
- `event_descriptors`
- `async_handler_descriptors`
- `get_method_descriptors`
- `schema_encoding`
- `metadata_hash_optional`

### 11.2 Method Descriptor

Fields:

- `method_id`
- `name`
- `input_schema_hash`
- `output_schema_hash`
- `execution_mode`
- `gas_hint`
- `payment_requirement_optional`
- `proof_requirement_optional`

Execution modes:

- `sync`
- `async`
- `scheduled`
- `get`

### 11.3 Interface Use Cases

Supports:

- Automatic UI generation.
- Wallet-driven forms.
- CLI auto-binding.
- RPC introspection.
- SDK call builders.
- Contract capability discovery.

### 11.4 Interface Rules

- Interface hash commits to all descriptors.
- Interface version changes require new hash.
- Interface metadata cannot grant authorization.
- Get-methods must be read-only.
- Async handlers must specify callback and timeout behavior.

### 11.5 Interface Implementation Tasks

- Define interface schema.
- Add interface registry.
- Add interface hash verification.
- Add query for interface by contract or service.
- Add SDK code generation format.
- Add tests for schema mismatch and version changes.

## 12. Async Failure Model

### 12.1 Failure Classes

Failure classes:

- Invalid payload.
- Insufficient gas.
- Destination not found.
- Destination disabled.
- Expired message.
- Handler failure.
- Storage limit exceeded.
- Proof verification failure.
- Retry exhausted.

### 12.2 Bounce System

If execution fails and `bounce_flag = true`:

- Create bounce message to source.
- Include original message ID.
- Include error code.
- Return remaining value where applicable.
- Consume bounded bounce gas.
- Emit bounce receipt.

Rules:

- Bounce cannot create more value than original remaining value.
- Bounce payload size is bounded.
- Bounce can itself fail and dead-letter.

### 12.3 Dead Letter Queue

Unresolvable messages go to zone-level dead letter queue.

Dead letter triggers:

- Retry exhausted.
- Destination permanently disabled.
- Bounce failed.
- Message expired before execution.
- Invalid message format discovered at execution.

### 12.4 Failure Implementation Tasks

- Define error code enum.
- Implement bounce creation.
- Implement dead letter queue.
- Add failed receipt model.
- Add retry exhaustion handling.
- Add value conservation tests.
- Add failure mode fuzz tests.

## 13. Security Model

### 13.1 Security Assumptions

AVM security relies on:

- CometBFT finality.
- Deterministic execution.
- No external calls during execution.
- Gas bounding.
- Replay protection.
- Zone state isolation.
- Actor state isolation.
- Store root commitments.

### 13.2 Replay Protection

Replay fields:

- `chain_id`.
- `source_zone`.
- `sender`.
- `sender_nonce`.
- `message_id`.
- `created_height`.
- `expiry_height`.

Rules:

- Sender nonce must increase.
- Consumed message creates tombstone.
- Expired message cannot be resubmitted with same nonce.
- Cross-zone messages must bind source and destination zones.

### 13.3 State Isolation

Rules:

- Zone state prefixes are isolated.
- Actor state prefixes are isolated.
- Contract storage prefixes are isolated.
- Cross-zone writes are prohibited.
- Cross-actor mutable state reads are prohibited unless proof-based and read-only.

### 13.4 Scheduler Safety

Rules:

- Scheduling order is deterministic.
- Queue processing is bounded.
- Retry count is bounded.
- Delayed messages cannot execute early.
- Expired messages cannot execute.
- Priority cannot bypass sender nonce ordering where ordering is required.

### 13.5 Security Implementation Tasks

- Add nonce keeper.
- Add replay tombstone store.
- Add zone access capability checks.
- Add actor state isolation tests.
- Add cross-zone write rejection tests.
- Add scheduler determinism tests.
- Add malicious queue load simulations.

## 14. Performance Model

### 14.1 Target Properties

AVM targets:

- Parallel zone execution.
- Pipelined async queue processing.
- Batched message execution.
- Minimized Store v2 writes.
- Lazy state loading.
- Bounded receipt generation.
- Actor-local conflict isolation.

### 14.2 BlockSTM Strategy

Parallel-safe workloads:

- Different zones.
- Different actors.
- Different contracts.
- Different message queues.
- Different continuation IDs.
- Different service calls.

Conflict-prone workloads:

- Same actor mailbox.
- Same contract storage key.
- Same zone queue head.
- Same sender nonce.
- Same payment or escrow record.

Implementation rules:

- Partition queues by zone and actor.
- Use per-zone accumulators.
- Use per-actor mailbox keys.
- Avoid global counters in hot paths.
- Use expected versions for state updates.

### 14.3 Store v2 Strategy

Store layout goals:

- Keep message records compact.
- Store payload by hash where possible.
- Prefix actor state by actor ID.
- Prefix contract state by contract address.
- Use height-bucketed delayed queues.
- Use tombstone pruning by proof horizon.

### 14.4 Performance Implementation Tasks

- Add benchmarks for queue insert and pop.
- Add benchmarks for actor execution.
- Add benchmarks for continuation resume.
- Add benchmarks for cross-zone messages.
- Add BlockSTM conflict tests.
- Add Store v2 read/write latency tests.
- Add root generation benchmarks.

## 15. Upgrade System

### 15.1 Upgradable Components

AVM supports governed upgrades for:

- VM interpreter version.
- Scheduler rules.
- Gas model.
- Zone configuration.
- Backend adapters.
- Interface schema version.
- Retry policies.
- Queue limits.

### 15.2 Upgrade Rules

- Upgrade activation height must be staged.
- Old and new scheduler rules must not overlap in one block.
- Messages created before upgrade use versioned execution policy where required.
- Continuations store runtime version.
- Contract code stores VM version.
- Gas table changes apply at activation height.

### 15.3 Upgrade State

Fields:

- `upgrade_id`
- `component`
- `from_version`
- `to_version`
- `activation_height`
- `migration_required`
- `compatibility_mode`
- `status`

### 15.4 Upgrade Implementation Tasks

- Add runtime version fields.
- Add scheduled upgrade state.
- Add migration handlers for queues and continuations.
- Add versioned gas table.
- Add compatibility tests for pending messages across upgrades.
- Add rollback prevention for activated upgrades.

## 16. Comparison with Existing Execution Models

| Feature | Classic Cosmos SDK | AVM |
| --- | --- | --- |
| Execution | synchronous | sync + async |
| State | KVStore | KVStore + zone roots |
| Messaging | tx-only | message-driven |
| Scheduling | block-bound | cross-block |
| Contracts | module-based | actor + module hybrid |
| UX model | CLI/API | interface-driven |

## 17. Future Extensions

Planned extension areas:

- Speculative execution layer.
- Parallel actor scheduling.
- Zero-knowledge execution attestation.
- Distributed async scheduler.
- Cross-chain message bridge layer.
- Actor state rent.
- Interface package registry.
- Formal VM verification test suite.
- Deterministic replay debugger.

## 18. Cosmos SDK Module Breakdown

### 18.1 x/avm

Purpose:

- Own AVM runtime parameters, routing, roots, and execution receipts.

State:

- `AVMParams`
- `RouteDescriptor`
- `AVMRoot`
- `ExecutionReceipt`
- `RuntimeVersion`

Messages:

- `MsgSubmitAVMMessage`
- `MsgRegisterRoute`
- `MsgUpdateAVMParams`
- `MsgScheduleRuntimeUpgrade`

Queries:

- `QueryAVMParams`
- `QueryAVMRoot`
- `QueryRoute`
- `QueryExecutionReceipt`
- `QueryRuntimeVersion`

### 18.2 x/async

Purpose:

- Own async message queues, retry queues, delayed queues, and dead letter queues.

State:

- `AsyncMessage`
- `ZoneQueue`
- `RetryRecord`
- `DeadLetterRecord`
- `ReplayTombstone`

Messages:

- `MsgSubmitAsyncMessage`
- `MsgCancelAsyncMessage`
- `MsgRetryAsyncMessage`
- `MsgExpireAsyncMessage`

Queries:

- `QueryAsyncMessage`
- `QueryZoneQueue`
- `QueryDeadLetter`
- `QueryReplayTombstone`

### 18.3 x/actors

Purpose:

- Own actor records, mailboxes, actor state, and continuation integration.

State:

- `ActorRecord`
- `ActorMailbox`
- `ActorState`
- `ActorPermission`

Messages:

- `MsgCreateActor`
- `MsgSendActorMessage`
- `MsgUpdateActor`
- `MsgPauseActor`

Queries:

- `QueryActor`
- `QueryActorMailbox`
- `QueryActorState`

### 18.4 x/continuations

Purpose:

- Store and resume long-running async workflow state.

State:

- `Continuation`
- `ContinuationQueue`
- `ContinuationReceipt`

Messages:

- `MsgCreateContinuation`
- `MsgResumeContinuation`
- `MsgCancelContinuation`
- `MsgExpireContinuation`

Queries:

- `QueryContinuation`
- `QueryContinuationsByActor`
- `QueryContinuationReceipt`

### 18.5 x/avmcontracts

Purpose:

- Own AVM-native contract code, instances, storage, and backend adapters.

State:

- `CodeRecord`
- `ContractRecord`
- `StorageValue`
- `BackendConfig`

Messages:

- `MsgStoreCode`
- `MsgInstantiateContract`
- `MsgExecuteContract`
- `MsgMigrateContract`

Queries:

- `QueryCode`
- `QueryContract`
- `QueryContractStorage`
- `QueryContractProof`

### 18.6 x/avminterfaces

Purpose:

- Store interface schemas for contracts, actors, and services.

State:

- `InterfaceDescriptor`
- `MethodDescriptor`
- `EventDescriptor`
- `AsyncHandlerDescriptor`

Messages:

- `MsgRegisterInterface`
- `MsgUpdateInterface`
- `MsgDeprecateInterface`

Queries:

- `QueryInterface`
- `QueryMethod`
- `QueryInterfaceByTarget`

## 19. Implementation Roadmap

### Phase 0: Specification and Test Vectors

Tasks:

- Define canonical `AsyncMessage` encoding.
- Define message ID derivation.
- Define deterministic queue sort key.
- Define execution receipt schema.
- Define AVM root schema.
- Define gas policy schema.
- Define interface descriptor schema.

Exit criteria:

- All signable and hashable objects have test vectors.
- Queue ordering is test-covered.
- Root encoding is fixed.

### Phase 1: Sync Router

Tasks:

- Implement AVM router skeleton.
- Route existing MsgServer calls through sync engine wrapper.
- Add execution receipts for routed sync messages.
- Add zone route descriptors.
- Add AVM root placeholder.

Exit criteria:

- Existing module calls can be represented as AVM-routed sync execution.
- Receipts are emitted deterministically.

### Phase 2: Async Engine

Tasks:

- Implement async message store.
- Implement zone queues.
- Implement delayed queue.
- Implement retry queue.
- Implement dead letter queue.
- Implement replay tombstones.
- Add queue roots.

Exit criteria:

- Async messages can be scheduled and executed in later blocks.
- Expired messages do not execute.
- Retry and dead letter flows are deterministic.

### Phase 3: Cross-Zone Routing

Tasks:

- Add source and destination zone metadata.
- Add zone message filters.
- Add cross-zone inbox and outbox roots.
- Add bounce system.
- Add cross-zone value accounting.
- Add cross-zone proof queries.

Exit criteria:

- Zone A can send async message to Zone B.
- Failed cross-zone messages bounce or dead-letter.
- Cross-zone receipts are proof-queryable.

### Phase 4: Actor Runtime

Tasks:

- Implement actor records.
- Implement actor mailboxes.
- Implement actor handler dispatch.
- Implement actor state isolation.
- Add actor receipt emission.
- Add actor state proof query.

Exit criteria:

- Actors execute mailbox messages one at a time.
- Actors cannot mutate other actor state.
- Actor messages are deterministic and receipt-backed.

### Phase 5: Continuations

Tasks:

- Implement continuation state.
- Implement resume queue.
- Implement expiry handling.
- Implement continuation gas accounting.
- Add continuation proof query.

Exit criteria:

- Long-running workflows can pause and resume.
- Expired continuations produce deterministic receipts.

### Phase 6: Contract Backends

Tasks:

- Implement AVM-native contract backend.
- Define optional WASM adapter boundary.
- Add code registry.
- Add contract instance registry.
- Add Store v2 storage adapter.
- Add gas metering for contract execution.

Exit criteria:

- Contracts can execute sync and async handlers.
- Contracts can emit async messages.
- Contract state is proof-queryable.

### Phase 7: Interface System

Tasks:

- Implement interface registry.
- Add methods, events, async handlers, and get-method descriptors.
- Add interface hash verification.
- Add SDK and CLI binding metadata.
- Add wallet UI generation metadata.

Exit criteria:

- Contracts and services expose verifiable interfaces.
- Clients can build calls from interface descriptors.

### Phase 8: Performance and Hardening

Tasks:

- Add BlockSTM conflict benchmarks.
- Add queue throughput benchmarks.
- Add actor execution benchmarks.
- Add root generation benchmarks.
- Add replay and determinism test suite.
- Add upgrade compatibility tests.

Exit criteria:

- AVM execution is deterministic under replay.
- Independent zones and actors can execute in parallel where supported.
- Queue and root generation costs remain bounded.

## 20. Required Test Coverage

### 20.1 Unit Tests

- Message ID derivation.
- Sender nonce validation.
- Queue sort key ordering.
- Delay height eligibility.
- Expiry handling.
- Retry policy handling.
- Bounce message construction.
- Dead letter transition.
- Gas policy calculation.
- Interface hash calculation.
- Receipt hash calculation.

### 20.2 Integration Tests

- Sync module execution through AVM router.
- Async message submitted and executed in future block.
- Delayed message execution.
- Failed message bounce.
- Retry exhaustion to dead letter queue.
- Cross-zone message execution.
- Actor mailbox execution.
- Continuation resume.
- Contract emits async message.
- Interface descriptor queried by client.

### 20.3 Invariant Tests

- Every executed message has one receipt.
- Every queued message has stored message record.
- Consumed message cannot be replayed.
- Expired message cannot execute.
- Bounce cannot over-refund value.
- Zone root includes queue and continuation roots.
- Actor mailbox order is deterministic.
- Actor state isolation is enforced.
- Contract storage key prefix is isolated.

### 20.4 Fuzz Tests

- Malformed async messages.
- Random nonce ordering.
- Queue priority edge cases.
- Retry and expiry boundary conditions.
- Bounce payload limits.
- Actor handler failures.
- Continuation state payloads.
- Contract storage keys.
- Interface schema payloads.

### 20.5 Performance Tests

- Queue insert throughput.
- Queue drain throughput.
- Async message execution throughput.
- Actor mailbox throughput.
- Continuation resume throughput.
- Cross-zone message throughput.
- Receipt proof generation latency.
- AVM root generation latency.
- BlockSTM conflict rate by workload.

## 21. Observability

### 21.1 Metrics

- Async messages submitted.
- Async messages executed.
- Async messages expired.
- Async messages bounced.
- Dead letter count.
- Retry queue size.
- Delayed queue size.
- Actor count.
- Actor mailbox depth.
- Continuations active.
- Continuations expired.
- Contract executions.
- Contract failures.
- Gas reserved.
- Gas consumed.
- Zone async budget usage.
- Queue drain latency.
- Receipt root generation time.

### 21.2 Events

- `avm_message_submitted`
- `avm_message_scheduled`
- `avm_message_executed`
- `avm_message_failed`
- `avm_message_retried`
- `avm_message_expired`
- `avm_message_bounced`
- `avm_dead_lettered`
- `avm_actor_created`
- `avm_actor_message_handled`
- `avm_continuation_created`
- `avm_continuation_resumed`
- `avm_continuation_expired`
- `avm_contract_executed`
- `avm_interface_registered`
- `avm_runtime_upgrade_scheduled`

### 21.3 Alerts

- Dead letter spike.
- Retry queue backlog.
- Delayed queue backlog.
- Zone async budget saturation.
- Actor mailbox backlog.
- Continuation expiry spike.
- Contract failure spike.
- Receipt generation latency above threshold.
- Queue root generation latency above threshold.

## 22. Acceptance Criteria

AVM is ready for implementation planning when:

- Sync execution can wrap existing Cosmos SDK MsgServer calls.
- Async messages can be scheduled, executed, retried, bounced, expired, and dead-lettered.
- Queue ordering is deterministic and test-covered.
- Cross-zone messages use queue and receipt roots.
- Actor runtime supports isolated mailbox execution.
- Continuations support pause and resume across blocks.
- Contract backends are isolated behind explicit runtime interface.
- Gas model covers execution, storage, scheduling, routing, proof verification, and continuation storage.
- Interface registry supports methods, events, async handlers, and get-methods.
- Root commitments include router, message, actor, contract, continuation, interface, and receipt roots.
- Replay protection prevents duplicate message execution.
- Store v2 layout is prefix-isolated and proof-queryable.
- BlockSTM conflict strategy is defined and benchmarked.
