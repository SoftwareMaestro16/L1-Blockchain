> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Aetra Modular L1 Execution OS

This document defines the target architecture for Aetra as a scalable,
secure, low-fee blockchain operating system. It is a system design, not an
implementation claim. Current production-facing code remains the Cosmos SDK
base chain and explicitly gated executable specifications until their
implementation, simulator, long-run testnet, consensus-safety proof, and audit
gates pass.

## Architecture Diagram

```text
Users / wallets / contracts
  |
  v
CLI / API / mempool
  |
  v
Aether Core
  |-- PoS consensus
  |-- validator set management
  |-- staking and delegation
  |-- slashing and evidence
  |-- governance parameters
  |-- deterministic routing
  |
  | commits finalized roots
  v
Zone Commitments
  |-- state roots
  |-- receipt roots
  |-- message roots
  |
  v
Execution Zones
  |-- Financial Zone: AET, bank, fees, token settlement
  |-- Identity Zone: .aet registry, resolver, domains
  |-- Application Zones: DEX, workflow, future apps
  |-- Contract Zones: AVM and explicitly gated CosmWasm
  |
  v
Compute Shards
  |-- elastic deterministic execution partitions
  |-- activated by LOAD_SCORE thresholds
  |-- not independent blockchains
  |
  v
Aether Mesh
  |-- cross-zone messages
  |-- asset transfer receipts
  |-- asynchronous contract calls
  |-- proof-based verification
```

## Layer-By-Layer Design

### Aether Core

Aether Core is the control plane of the network. It provides shared security
and deterministic coordination for the whole system.

Responsibilities:

- Proof-of-Stake consensus and deterministic finality.
- Validator set management, validator power updates, staking, delegation,
  unbonding, redelegation, and validator lifecycle transitions.
- Slashing and evidence handling for equivocation, downtime, invalid zone
  commitments, invalid receipt roots, and data-unavailable shard work once
  those subsystems are implemented.
- Governance over bounded protocol parameters.
- Deterministic routing of transactions to Execution Zones and Compute Shards.
- Commitment of zone state roots, receipt roots, and message roots.

Constraints:

- Aether Core must not execute smart contracts.
- Aether Core must not process application-specific business logic.
- Aether Core must remain deterministic, minimal, and auditable.
- Aether Core acts only as the security, finality, and coordination layer.

State variables:

- `validator_set`
- `validator_power`
- `delegations`
- `unbonding_entries`
- `redelegations`
- `jailed_validators`
- `tombstoned_validators`
- `slashing_evidence`
- `governance_params`
- `fee_params`
- `routing_table`
- `zone_registry`
- `shard_registry`
- `zone_state_roots`
- `zone_receipt_roots`
- `zone_message_roots`
- `load_windows`
- `load_score`

Validator lifecycle:

```text
candidate
  -> bonded validator
  -> active validator
  -> jailed validator
  -> unbonding validator
  -> removed validator
```

Tombstoned validators are permanently excluded from re-entering the active set
for the same operator identity. All validator status changes are derived from
staking transactions, unbonding windows, and cryptographically verifiable
evidence.

### Execution Zones

Execution Zones are deterministic execution environments that share Aether Core
security while isolating application workloads.

Each zone has:

- a zone id;
- a VM or runtime policy;
- a local mempool policy;
- local fee market parameters;
- local state root;
- local receipt root;
- local message root;
- zone-specific transaction classifier;
- zone-specific execution result format;
- bounded export/import state.

Default zone classes:

- `Financial Zone`: native AET settlement, bank transfers, fees, tokenfactory
  assets, and protocol accounting.
- `Identity Zone`: `.aet` domains, resolver records, reverse records,
  subdomains, expiry, renewal, and NFT-based ownership.
- `Application Zone`: application modules such as DEX, workflow, scheduler, or
  future app-specific state machines.
- `Contract Zone`: AVM smart contracts and explicitly gated CosmWasm contracts.

Zone creation is governance-controlled. A zone proposal must include the zone
id, VM policy, fee policy, genesis state hash, state transition identifier,
upgrade policy, data availability policy, and audit status. Aether Core only
activates a zone after the proposal is finalized and the activation height is
reached.

State is committed back to Core through deterministic commitments:

```text
zone_commitment = hash(
  zone_id,
  zone_height,
  state_root,
  receipt_root,
  message_root,
  execution_result_root,
  previous_zone_commitment
)
```

### Compute Shards

Compute Shards are dynamic scaling units inside or across Execution Zones. They
increase throughput under load without becoming independent blockchains.

Rules:

- Shards are elastic execution partitions, not sovereign chains.
- Shard activation and deactivation are deterministic.
- Shards inherit security and finality through Aether Core commitments.
- Shard state transitions must be replayable from finalized inputs.
- Shard routing must not depend on local node timing, local mempool order, or
  validator preference.

Shard activation:

- Low load uses one shard per zone.
- Medium load enables partial sharding for hot zones.
- High load enables full sharding for all overloaded zones.
- Deactivation requires the zone LOAD_SCORE EMA to remain below the lower
  threshold for the configured cooldown window.

Shard assignment:

```text
shard_id = hash(zone_id || primary_actor || routing_epoch) % active_shards(zone_id)
```

`primary_actor` is the deterministic account, contract, domain, or asset key
that owns the state touched by the transaction. Transactions that touch multiple
primary actors must use a deterministic lock ordering or be routed through an
async message flow.

## Deterministic Load Detection

The Load Detection System is consensus-critical. It must be fully
deterministic, mathematically defined, resistant to manipulation, identical
across all nodes, and independent of node-specific behavior.

No load metric may use local wall-clock observations, local node latency,
machine learning, adaptive weights, random weights, or validator-specific
preferences.

Protocol constants:

```text
N = 60 blocks
alpha = 2 / (N + 1)
MAX_DELTA = 0.05

a = 0.20
b = 0.30
c = 0.20
d = 0.10
e = 0.20

a + b + c + d + e = 1.00
```

All inputs are normalized to `[0,1]`:

```text
mempool_size_score      = min(1, canonical_mempool_size / target_mempool_size)
block_utilization_score = min(1, used_block_gas / target_block_gas)
tx_latency_score        = min(1, avg_inclusion_delay_blocks / target_latency_blocks)
failure_rate_score      = failed_tx_count / max(1, total_tx_count)
execution_time_score    = min(1, execution_step_count / target_execution_steps)
```

`canonical_mempool_size` is not a local node mempool count. It is the
deterministic count of valid pending transaction commitments admitted into the
protocol load window. `execution_step_count` is a deterministic gas or step
counter from execution, not wall-clock milliseconds.

Sliding window:

```text
EMA_t(metric) = alpha * value_t + (1 - alpha) * EMA_{t-1}(metric)
```

Exact formula:

```text
LOAD_SCORE = 0.20*mempool_size_score + 0.30*block_utilization_score + 0.20*tx_latency_score + 0.10*failure_rate_score + 0.20*execution_time_score
```

Consensus calculation:

```text
raw_load_t =
  0.20 * EMA_t(mempool_size_score) +
  0.30 * EMA_t(block_utilization_score) +
  0.20 * EMA_t(tx_latency_score) +
  0.10 * EMA_t(failure_rate_score) +
  0.20 * EMA_t(execution_time_score)

LOAD_SCORE_t = clamp(
  raw_load_t,
  LOAD_SCORE_{t-1} - MAX_DELTA,
  LOAD_SCORE_{t-1} + MAX_DELTA
)
```

This creates three stability layers:

- EMA smoothing prevents one-block manipulation from changing routing abruptly.
- Per-metric normalization caps every metric at `1`.
- `MAX_DELTA` caps the per-block change of the final score.

## Load Lifecycle

### Low Load: `0.0 <= LOAD_SCORE < 0.3`

Behavior:

- normal execution;
- low base fees;
- single shard execution per zone;
- full valid transaction inclusion subject to block limits;
- relaxed but deterministic priority ordering.

### Medium Load: `0.3 <= LOAD_SCORE < 0.7`

Behavior:

- moderate deterministic fee increase;
- partial compute sharding for zones above threshold;
- priority-based mempool ordering;
- low-priority transactions may be delayed;
- per-sender and per-contract queue limits are enforced more strictly.

### High Load: `0.7 <= LOAD_SCORE <= 1.0`

Behavior:

- full compute sharding for overloaded zones;
- parallel execution across shards;
- strict transaction prioritization;
- aggressive spam filtering mode;
- low-value transactions are deferred or batched;
- cross-zone messages are queued by deterministic finality and message order.

## Routing Algorithm Specification

Routing is an Aether Core responsibility. Every validator must compute the same
route for the same transaction and committed state.

Inputs:

- transaction type;
- signer and primary actor;
- fee class;
- sender reputation class, if available;
- contract locality;
- domain locality;
- current zone and shard registry;
- current `LOAD_SCORE`;
- routing epoch.

Outputs:

- target Execution Zone;
- target Compute Shard;
- deterministic priority key.

Algorithm:

```text
1. Reject malformed tx, invalid signer, invalid fee denom, invalid memo,
   zero address, or unauthorized message.
2. Classify tx:
   critical system
   staking/gov/security
   financial
   identity
   contract
   application
   async message
3. Resolve locality:
   account key
   contract address
   .aet domain key
   asset denom key
   message destination key
4. Select zone from tx class and locality.
5. Select shard:
   shard_id = hash(zone_id || primary_actor || routing_epoch) % active_shards(zone_id)
6. Build deterministic priority key:
   priority_class
   effective_fee_class
   reputation_class
   admission_height
   tx_hash
7. Return route decision.
```

Overload priority order:

```text
1. critical system transactions
2. high fee transactions
3. normal user transactions
4. low fee or spam-like transactions
```

Tie-breakers are deterministic:

- higher bounded fee class before lower fee class;
- higher bounded reputation class before lower reputation class;
- older valid admission height before newer admission height;
- lexicographic transaction hash as the final tie-breaker.

Routing must prevent:

- hot-spot attacks through shard activation, per-actor limits, and deterministic
  actor-key routing;
- manipulation by fee overpayment through bounded fee classes;
- non-determinism by forbidding wall-clock time, local latency, randomness,
  machine learning, and validator-local preferences.

## Aether Mesh

Aether Mesh is the interoperability layer between zones and shards.

Capabilities:

- cross-zone messaging;
- asset transfer receipts;
- asynchronous contract calls;
- proof-based verification of cross-zone actions;
- bounce and refund semantics for failed async delivery.

Message fields:

```text
source_zone
source_shard
destination_zone
destination_shard
message_id
sender
recipient
asset_commitment
payload_hash
timeout_height
finality_reference
proof
sequence
```

Ordering:

```text
source_finality_height,
source_zone,
source_shard,
message_id,
destination_zone,
destination_shard,
sequence
```

No double spend is guaranteed by:

- finalized source roots before destination execution;
- single-use receipt markers;
- replay markers keyed by message id;
- receipt roots committed back to Aether Core;
- deterministic bounce/refund for failed delivery.

## Identity Layer

The Identity Layer provides global `.aet` names.

Features:

- human-readable `.aet` domains;
- resolver records from name to account, contract, or zone endpoint;
- reverse lookup from address to primary name;
- subdomain support;
- NFT-based ownership model;
- deterministic expiry and renewal;
- optional auction mechanism for contested names.

Domain lifecycle:

```text
available -> commit -> reveal/register -> active -> renewal window -> expired -> available
```

Rules:

- one active owner per fully qualified domain;
- domain labels are normalized before storage;
- duplicate normalized names are rejected;
- resolver updates require current owner authorization;
- reverse records require address owner authorization;
- subdomain issuance requires parent owner authorization;
- expired names cannot resolve unless renewed before expiration.

Auction mechanism, if enabled, must be deterministic and parameter bounded:

- sealed commit phase;
- reveal phase;
- deterministic tie-breaker by bid amount, reveal height, and commitment hash;
- losing bids are refundable through deterministic receipts.

## Security Model Summary

Economic security:

- Validators stake `naet`.
- Delegators delegate `naet`.
- Voting power derives from bonded stake.
- Rewards come from staking inflation and protocol fee distribution.
- Slashing penalties are deterministic and evidence-based.

Slashing:

- all penalties are derived from deterministic evidence;
- double-sign evidence tombstones the validator;
- downtime jails the validator according to configured parameters;
- invalid zone commitment evidence is slashable once zone execution is
  production-wired;
- invalid receipt root evidence is slashable once Aether Mesh is
  production-wired;
- no penalty depends on human discretion.

Fee model:

- protocol fees use native `naet`;
- normal load keeps low base fees;
- congestion raises fees through deterministic parameters;
- spam resistance uses rate limits, queue limits, bounded priority classes, and
  dynamic load response rather than permanently high base fees.

Reward and supply model:

- validator and delegator rewards are paid from staking issuance and protocol
  fee distribution;
- burn logic may be enabled by governance-bounded parameters;
- inflation and burn settings must remain deterministic and auditable.

## Failure Scenarios And Mitigations

- Mempool spam burst: EMA smoothing, per-block metric caps, per-sender limits,
  and `MAX_DELTA` prevent abrupt routing changes.
- Gas usage spike: block utilization is normalized and smoothed before changing
  shard activation.
- Hot contract or domain: primary-actor routing keeps state ownership
  deterministic while shard count scales around the overloaded zone.
- Shard execution failure: shard work is replayed from deterministic inputs and
  only finalized through committed roots.
- Cross-zone message failure: Aether Mesh creates deterministic bounce or refund
  receipts and rejects replayed receipts.
- Missing or spoofed domain: resolver checks fail before funds or messages move.
- Validator equivocation: cryptographic evidence triggers deterministic
  slashing and tombstoning.
- Data unavailable shard work: commitments are not final for routing until data
  availability conditions are met.
- Governance abuse: parameter changes are bounded, delayed, and auditable.
- Non-deterministic implementation attempt: any dependence on local wall-clock
  time, local latency, randomness, adaptive ML weights, or validator-specific
  preferences is a protocol failure.

## Trilemma Model

- Scalability comes from Execution Zones, Compute Shards, bounded async queues,
  and deterministic routing.
- Security comes from Aether Core, PoS staking, slashing, finalized
  commitments, and proof-verified cross-zone actions.
- Decentralization comes from a distributed validator set and independently
  verifiable execution zones coordinated by a minimal Core.

Aetra is a modular Layer 1 execution operating system where consensus, execution, and scaling are separated into independent but cryptographically connected layers to achieve secure and scalable decentralized computation.
