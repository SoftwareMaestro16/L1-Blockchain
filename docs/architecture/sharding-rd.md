# Production Sharding And Partitioning R&D

Aetra must not claim production sharding until the consensus-safety gate is
passed. Public language must use `sharding R&D` or `experimental sharding`
until written spec, simulator, prototype, fuzz tests, adversarial tests,
long-run testnet, independent audit, and consensus-safety proof are complete.

This design uses masterchain, workchain, and shardchain terminology as an
Aetra-native R&D model. In this model, workchains are execution domains with
their own rules, address formats, and VM sets; shardchains are workchain
partitions keyed by a shard prefix; and the masterchain is the coordination
chain for global configuration, validator state, and shard references.
Aetra must define and audit its own consensus-safe architecture before
implementation.

The broader target-system design is defined in
[Aetra Modular L1 Execution OS](aetra-modular-execution-os.md). This
document remains the narrower sharding and partitioning R&D safety gate.

## Terminology

- `masterchain`: global coordination chain for validator set, staking
  snapshots, workchain configs, shardchain headers, finality references,
  cross-chain routing commitments, and config updates.
- `workchain`: execution domain with its own rules, VM set, address space,
  state transition function, fee policy, genesis state, and upgrade policy.
- `shardchain`: partition of a workchain state and message space.
- `cross-shard message`: async message routed through deterministic proof,
  receipt, timeout, bounce, and replay-protection flow.

## Consensus-Safe Spec Requirements

- Finality model: masterchain finality references shard headers only after a
  deterministic finality lag or stronger BFT proof; cross-shard messages can
  only rely on finalized headers.
- Validator assignment: masterchain assigns validator subsets to shardchains
  from a staking snapshot and deterministic randomness beacon.
- Randomness source: assignment randomness must be consensus-approved,
  bias-resistant enough for validator assignment, exported/imported exactly,
  and unavailable to contracts as host entropy unless separately specified.
- Shard header commitments: each shard header commits to shard id, height,
  state root, message queue root, receipt root, validator subset, data
  availability flag, and parent/split/merge references.
- Data availability: a shard block is not final for routing unless its data is
  available or recoverable under the selected DA scheme.
- Fraud/equivocation evidence: conflicting shard roots for the same
  `(shard_id, height, validator)` produce deterministic evidence records.
- Slashing rules: validator equivocation, unavailable shard blocks, invalid
  header signatures, and invalid receipt roots must have explicit slashing
  policy before public testnet.
- Cross-shard message ordering: route by source finality height, source shard,
  message id, destination shard, and queue sequence tie-breaker.
- Cross-shard replay protection: message id is derived from source shard,
  destination shard, nonce, payload hash, and source logical time; receipts are
  single-use.

## Masterchain State

- validator set
- staking snapshot
- workchain registry
- shardchain registry
- shard headers
- cross-shard receipt roots
- config updates
- randomness seed or beacon commitment
- fraud/equivocation evidence records

## Workchain Model

- workchain id
- allowed VMs, including AVM and explicitly gated CosmWasm if enabled
- native-only fee policy in `naet`
- address format
- genesis state hash
- state transition function identifier
- upgrade policy

## Shardchain Model

- shard id as `(workchain_id, shard_prefix)`
- state root
- message queue root
- receipt root
- validator subset
- data availability status
- split/merge parent and child references
- finality height

## Cross-Shard Async Messages

- source workchain and shard
- destination workchain and shard
- message id
- proof against finalized source shard header
- receipt against destination shard receipt root
- timeout height
- bounce/refund semantics
- replay marker

Delivery rules:

- Source shard commits the outbound message in its message queue root.
- Masterchain finalizes the source shard header.
- Destination shard verifies the source proof and destination routing.
- Destination execution emits exactly one receipt.
- Masterchain commits the receipt root.
- Duplicate receipt, missing receipt, invalid proof, stale header, wrong
  destination shard, and replayed message are rejected.

## Simulator Before Implementation

The pure Go package `x/sharding/sim` is the current simulator. It is not a
keeper and does not register SDK stores, routes, module accounts, or genesis
state.

The simulator covers:

- deterministic message routing
- shard split
- shard merge
- validator reassignment
- delayed receipts
- failed shard block via data-unavailable status
- equivocation evidence
- export/import of sharding state

Required simulator tests:

```powershell
go test ./x/sharding/sim
go test ./...
```

## Prototype Only After Simulator

Prototype implementation may begin only after simulator gates are green:

- masterchain header keeper
- workchain registry keeper
- shard routing keeper
- cross-shard queue keeper
- export/import of sharded state

The prototype must remain feature-gated and labelled experimental.

## Adversarial Tests

- duplicate cross-shard receipt
- missing receipt
- invalid shard proof
- stale shard header
- wrong destination shard
- replayed message
- validator equivocation
- data unavailable shard block

## Benchmarks

- routing table lookup
- cross-shard proof verification
- queue processing
- shard split
- shard merge
- state export/import

Run:

```powershell
go test -run '^$' -bench 'Benchmark(Routing|CrossShard|Queue|Shard|ShardedState)' -benchmem ./x/sharding/sim
```

## Production Gate

Before production:

- written spec is complete
- simulator is complete
- prototype is complete
- fuzz tests pass
- adversarial tests pass
- long-run testnet passes
- independent audit is complete
- consensus-safety proof is accepted

Until this gate passes, public wording must say `sharding R&D` or
`experimental sharding`. No production sharding claim is allowed.
