# Execution OS

Track 6 defines the execution operating layer that coordinates transaction
validation, module dispatch, async entrypoints, VM routing, deterministic event
collection, and error handling. The current implementation is a pure executable
specification in `x/execution/types` and `x/vm/types`; it does not register SDK
stores, keepers, module accounts, genesis, CLI, or ABCI hooks.

## x/execution

Responsibilities:

- transaction orchestration;
- execution pipeline;
- module dispatch;
- async entrypoint;
- deterministic ordering;
- event collection;
- error handling.

Pipeline:

```text
CheckTx:
  decode
  validate signatures
  validate fees
  validate memo
  stateless checks

DeliverTx/FinalizeBlock:
  ante
  execution context
  module dispatch
  async enqueue if needed
  event emit
  state write
```

`ExecutionEnvelope` contains tx hash, sender, receiver, route, VM route, gas
limit, fee coins, optional memo metadata, resolver/domain records, reputation
record, stake/load counters, optional async messages, module events, block
height, and timestamp.

The executable pipeline integrates:

- optional memo metadata from `x/memo/types`;
- resolver lookup from `x/identity/types`;
- reputation limits from `x/reputation/types`;
- dynamic fee estimator and admission checks from `x/fees/types`;
- AVM or CosmWasm route validation through `x/vm/types`;
- deterministic execution trace for tests.

Priority, reputation, memo, and resolver data cannot bypass signatures, fee
validation, zero-address rejection, or module authorization.

## x/vm

Responsibilities:

- AVM runtime routing;
- gated CosmWasm runtime routing;
- gas metering policy;
- sandbox policy;
- contract state access boundary;
- host function boundary.

AVM minimum contract:

- counter contract;
- deploy;
- external call;
- internal call;
- bounced call;
- query/getter.

`x/vm/types` is the boundary facade over the existing `x/aetravm/avm`,
`x/aetravm/async`, and `app/wasmconfig` executable specifications. It
validates runtime selection, action-to-entrypoint mapping, code size, gas
limits, query limits, and CosmWasm feature gating.

Required VM tasks remain:

- bytecode/module format;
- ABI;
- storage ABI;
- message ABI;
- deterministic host functions;
- gas schedule;
- fuzz tests;
- adversarial tests.

CosmWasm remains disabled by default. AVM routing is enabled only as an
executable spec and does not mutate production state until keeper wiring is
explicitly added and audited.

## x/messaging

Responsibilities:

- async calls between contracts;
- internal message envelope;
- bounce/refund behavior;
- outgoing message validation.

Message:

```text
Message {
  id
  source
  destination
  value_naet
  opcode
  query_id
  body
  bounce
  deadline
  gas_limit
  created_lt
}
```

`x/messaging/types` maps this message shape onto the existing async
`MessageEnvelope`, validates `naet` value, source/destination addresses, body
size, gas limit, deterministic message id, and forwarding fee. Delivery,
bounce, expiry, and refund/no-double-spend behavior are inherited from
`x/aetravm/async`.

## x/queue

Responsibilities:

- delayed execution;
- scheduled tasks;
- retry and failure handling;
- queue limits;
- queue observability.

Ordering:

```text
priority_key =
  scheduled_height,
  reputation_class,
  tx_height,
  tx_index,
  message_index,
  source_logical_time,
  sequence
```

Rules:

- priority must be deterministic;
- low reputation cannot starve forever;
- max per-block processing limit;
- max per-account queued messages;
- max per-contract queued messages.

`x/queue/types` implements an executable scheduled queue spec with per-account
and per-contract counts, per-block pop limits, retry/failure accounting,
observability, reputation class mapping, and starvation protection that
eventually promotes old ready items into the top effective class.

## x/events

Responsibilities:

- event-driven system;
- protocol events;
- indexer events;
- contract events;
- memo events;
- domain events;
- reputation events.

Event types:

- `EventTransfer`;
- `EventMemoAttached`;
- `EventDomainAuctionStarted`;
- `EventDomainResolved`;
- `EventContractMessageQueued`;
- `EventContractMessageProcessed`;
- `EventReputationUpdated`;
- `EventFeeDistributed`.

`x/events/types` provides deterministic event validation, canonical attribute
ordering, and stable event sorting by height, sequence, type, and transaction
hash.

## x/actors

Responsibilities:

- each contract behaves as an actor;
- actor state isolation;
- actor mailbox;
- actor message processing;
- actor lifecycle.

Actor state:

```text
Actor {
  address
  code_hash
  state_root
  logical_time
  mailbox_stats
  status
}
```

Rules:

- one actor state transition per delivered message;
- actor cannot mutate another actor state directly;
- all cross-actor effects go through messages;
- exported state includes actor state and mailbox.

`x/actors/types` models actor state roots, logical time, mailbox statistics,
lifecycle status, isolated transitions, cross-actor mutation rejection, and
deterministic actor export ordering.

## x/scheduler

Responsibilities:

- parallel execution planning;
- conflict detection;
- deterministic batching;
- safe concurrent state access.

Initial version:

- sequential deterministic execution.

Production version:

- optimistic parallel execution with deterministic conflict resolution;
- DAG scheduler;
- read/write set tracking;
- fallback to sequential on conflict.

`x/scheduler/types` implements an executable planning spec. Sequential mode
sorts tasks by height, tx index, message index, and task id. Optimistic mode
uses declared read/write sets to form deterministic batches and marks full
conflict fallback when no safe parallel batch can be built.

## x/storage

Responsibilities:

- KV state engine;
- versioning;
- snapshots;
- state sync;
- contract storage;
- bounded iteration.

Tasks:

- define contract namespace;
- define storage key format;
- define max state size;
- define storage rent/deposit;
- export/import exact state;
- snapshot/state-sync tests.

`x/storage/types` defines the contract namespace format, bounded storage key
format, max state size validation, storage rent accounting, deterministic
snapshot export/import, state root calculation, and bounded prefix iteration
for future contract storage and state-sync tests.
