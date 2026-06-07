# Async Smart Contract Execution

Phase 10 introduces Aetra-native asynchronous contract message semantics
without changing the Cosmos SDK rule that transaction delivery is synchronous.
The current Cosmos SDK rule that transaction delivery is synchronous remains
the baseline for delivered transactions. Aetra can process deterministic
async contract queues inside blocks. Production partitioning or sharding
remains a later R&D track and must not be claimed until a separate consensus
architecture exists.

Production partitioning or sharding remains a later R&D track. The current
R&D boundary is documented in `docs/architecture/sharding-rd.md`; no
production sharding claim is allowed until the simulator, prototype,
adversarial tests, long-run testnet, independent audit, and consensus-safety
proof are complete.

## Core State

- `contract account`: derived account address, code hash, balance in `naet`,
  logical time, and active state bytes.
- `contract state`: bounded byte state owned by one contract account.
- `incoming message`: queued message whose destination is a contract account.
- `outgoing message`: message emitted by a contract execution.
- `message queue`: global deterministic FIFO queue keyed by sequence.
- `inbox storage`: per-destination view of queued messages.
- `outbox storage`: per-source view of emitted or submitted messages.
- `bounce message`: deterministic response when a send fails and bounce is
  enabled.
- `refund/excess message`: deterministic value return when bounce is disabled
  or excess value remains.
- `logical time`: monotonic per-contract counter incremented on successful
  execution attempt.
- `failure result code`: stable execution outcome code written to receipts.

## Contract Address Derivation

Contract address derivation is deterministic:

```text
address = sha256(
  "aetra/async-contract/v1",
  deployer,
  code_hash,
  salt
)
```

Normal validation rejects zero deployers and malformed code hashes. Address
derivation must be stable across export/import and node implementations.

## Message Envelope

Every async message uses a bounded envelope:

- source
- destination
- value in `naet`
- opcode
- query_id
- body
- bounce flag
- bounced flag
- created logical time
- expiration/deadline block, when needed
- gas limit
- forward fee in `naet`
- recursion/depth counter

The executable specification rejects non-`naet` values, zero source or
destination addresses, non-`naet` forward fees, zero gas limits, oversized
bodies, and messages deeper than the configured recursion limit.

## Queue Ordering

Queued messages carry explicit deterministic ordering metadata:

- tx index
- message index
- source logical time
- destination address key
- sequence tie-breaker

Messages are assigned a strictly increasing sequence number when enqueued. The
queue is ordered by tx index, message index, source logical time, destination
address key, and sequence tie-breaker. New outgoing messages receive a later tx
index than already queued transaction messages, so they cannot jump ahead of
previously queued work. Per-block processing stops at `max_messages_per_block`,
leaving the remaining queue unchanged and exportable.

Import validation rejects duplicate sequence numbers, `next_sequence` drift,
`next_tx_index` drift, source logical time drift, destination key drift, and
queues that are not sorted by the deterministic order key.

## Contract Logical Time

Each contract account has a monotonic logical time counter. A successful
execution attempt increments the recipient logical time before the contract
handler emits outgoing messages. Outgoing messages include the sender's current
logical time in `created_logical_time`, and queued-message metadata copies that
value into `source_logical_time`.

Logical time is exported and imported exactly as part of contract account state
and queued-message ordering metadata. Nodes must not recompute logical time from
wall clock, block time, or local execution history during import.

## Bounce And Refund Behavior

Execution runs against a working copy of recipient state. Recipient state is
committed only on success. If the destination is missing, the handler is absent,
the handler returns failure, state exceeds limits, the message expires, or
enqueueing follow-up messages fails, the recipient mutation is not committed.

If the failed message has `bounce = true` and is not already bounced, the queue
gets a deterministic bounce message:

- source becomes the failed destination
- destination becomes the original source
- opcode is the reserved bounce opcode
- query_id and body are preserved
- value is returned in `naet`

If bounce is disabled, value return uses a deterministic refund/excess message.
Bounce/refund messages are ordinary queued messages and execute in sequence.
A bounced message or refund message that fails does not create another refund;
this prevents double-refund loops when the original source is missing.

## Fee And Gas Model

All protocol accounting is in native `naet`:

- execution gas per message
- per-message gas limit
- storage fee per committed state byte
- message forwarding fee
- per-message forward fee in `naet`
- contract deployment cost

User-created tokens, NFT assets, SBT assets, LP tokens, or test denoms cannot
pay async VM protocol fees.

## Processing Limits

The executable specification defines bounded limits for:

- max messages per tx
- max messages per block
- max recursion/depth
- max body size
- max state size
- max emitted messages per execution
- max storage writes per execution
- max contract deploys per tx
- max contract deploys per block

These limits are consensus parameters when the VM is wired into keepers. They
must remain deterministic and governance-bounded.

## Observability

Nodes should expose deterministic counters or gauges for:

- queued messages
- processed messages
- bounced messages
- refund/excess messages
- failed executions
- gas used
- queue lag
- contract state size

## Executable Specification

The pure Go package `x/aetravm/async` is the current executable
specification for Phase 10. It does not wire the VM into the app. It defines:

- contract address derivation
- contract account and bounded contract state validation
- message envelope validation
- inbox, outbox, and global queue storage
- deterministic queue processing by tx index, message index, source logical
  time, destination key, and sequence tie-breaker
- bounce and refund behavior
- gas limit, storage fee, forwarding fee, and deploy cost accounting fields
- deterministic processing limits
- observability counters
- export/import state shape that preserves queue state exactly
- import validation for duplicate contract addresses, malformed contract state,
  malformed queued messages, duplicate queued sequences, inbox/outbox views, and
  `next_sequence`, `next_tx_index`, logical-time, destination-key, and queue
  ordering drift

When Aetra VM execution is introduced, keeper storage, ABCI lifecycle hooks,
and contract bytecode handlers must match this package or explicitly migrate
the semantics version.

## Required Tests

```powershell
go test ./x/aetravm/async
go test ./...
```

The package tests cover:

- contract address derivation
- contract state storage validation
- internal message emission from one contract to another
- recipient execution in deterministic order
- failed send bounce behavior
- refund behavior when bounce is disabled
- refund messages cannot create double-refund loops
- state rollback on failed execution
- max messages per tx
- max messages per block
- max emitted messages per execution
- max storage writes per execution
- max body size
- gas limit and forward fee validation
- queue lag observability
- export/import preserving queue state exactly
- duplicate contract address rejection
- malformed async queue state rejection
