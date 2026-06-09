# Aetheris Networking Stack Specification

Status: Internal design document
Scope: Multi-layer execution-aware networking, overlays, routing, discovery, and streaming transport
Visibility: Private, not for public repository inclusion

## 1. Objective

Design the next generation Aetheris networking architecture on top of:

- CometBFT P2P networking as the consensus transport baseline.
- Cosmos SDK execution layer as the state machine boundary.
- ABCI++ block lifecycle with `PrepareProposal`, `ProcessProposal`, and `FinalizeBlock`.
- BlockSTM-ready parallel execution concepts.
- Future execution zones, service layers, storage systems, and async runtimes.

The target is to evolve Aetheris from a validator-focused blockchain P2P network into a universal, modular, service-aware, multi-layer networking stack for:

- Consensus.
- Execution zones.
- Off-chain services.
- Data availability systems.
- Cross-zone messages.
- Streaming state sync.
- Service discovery.
- Future execution engines.

## 2. Architecture Principles

### 2.1 Layer Model

Aetheris networking is split into four layers:

```text
L0 - Physical Transport Layer
     CometBFT P2P baseline, consensus gossip, block propagation

L1 - Secure Node Identity and Session Layer
     cryptographic node identity, session channels, multiplexed streams

L2 - Overlay Routing Layer
     service-aware, zone-aware, shard-aware routing overlays

L3 - Application Networking Layer
     execution messages, services, data streams, queries, receipts
```

Each layer extends the previous one. No layer replaces the safety-critical role of the layer below it.

### 2.2 Hard Rules

- CometBFT remains consensus-critical transport.
- Application service networking must not affect consensus validity unless data is committed through deterministic messages or proofs.
- Routing decisions used by consensus must be deterministic or derived from committed state.
- Live latency, throughput, and peer scoring are advisory until committed.
- No external network calls are allowed inside state transition execution.
- All network messages requiring consensus effects must produce replay-safe identifiers.
- All large payloads must be chunked and commitment-verified.
- All discovery records must be signed, expiring, and optionally proof-attached.

### 2.3 Network Roles

Node roles:

- Validator node.
- Full node.
- Archive node.
- State sync node.
- Zone execution node.
- Service node.
- Storage provider node.
- Routing node.
- Index node.
- Light client gateway.

Rules:

- Validator role is consensus-critical.
- Service, routing, and storage roles are not consensus-critical unless explicitly bonded and committed.
- A node can advertise multiple roles.
- Role advertisements must be signed by node identity.
- Role records must expire and be renewed.

## 3. L0: Base Transport Layer

### 3.1 Role

CometBFT P2P remains responsible for:

- Consensus message gossip.
- Proposal and vote propagation.
- Block propagation.
- Baseline mempool transaction propagation.
- Validator coordination.
- Peer transport primitives.

### 3.2 Aether Networking Adapter

Add Aether Networking Adapter (`ANA`) above CometBFT transport.

ANA responsibilities:

- Peer scoring by latency, reliability, and throughput.
- Connection multiplexing from one physical peer connection into logical channels.
- Adaptive gossip fanout.
- Bandwidth-aware propagation.
- Zone-aware routing hints.
- Message class prioritization.
- Streaming payload negotiation.
- Peer role advertisement validation.

ANA must not:

- Change consensus message validity.
- Hide consensus-critical messages from CometBFT.
- Depend on non-deterministic peer metrics for committed state transitions.
- Replace CometBFT consensus gossip.

### 3.3 L0 Channel Classes

Logical channel classes:

- `CONSENSUS_CHANNEL`
- `MEMPOOL_CHANNEL`
- `BLOCK_CHANNEL`
- `STATE_SYNC_CHANNEL`
- `DATA_CHANNEL`
- `EXECUTION_CHANNEL`
- `SERVICE_CHANNEL`
- `ROUTING_CHANNEL`
- `DISCOVERY_CHANNEL`

Priority order:

1. Consensus votes and proposals.
2. Block headers and availability proofs.
3. State sync and evidence.
4. Execution receipts and cross-zone messages.
5. Mempool transactions.
6. Service and discovery traffic.
7. Bulk data transfer.

### 3.4 L0 Implementation Tasks

- Implement ANA abstraction around peer transport.
- Define logical channel IDs.
- Add peer scoring inputs.
- Add adaptive fanout settings.
- Add channel-level bandwidth accounting.
- Add message class priority policy.
- Add L0 metrics and alerts.
- Add tests proving consensus messages are not delayed by service traffic.

## 4. L1: Secure Node Identity and Sessions

### 4.1 Node Identity

Node identity:

```text
NodeID = H(ValidatorPubKey || NetworkSalt)
```

For non-validator nodes:

```text
NodeID = H(NodePubKey || NetworkSalt)
```

Properties:

- Stable across restarts.
- Cryptographically verifiable.
- Independent of IP and port.
- Bound to advertised roles.
- Rotatable through signed transition record.

### 4.2 NodeRecord

Fields:

- `node_id`
- `node_pub_key`
- `validator_pub_key_optional`
- `operator_address_optional`
- `roles`
- `network_addresses_hash`
- `zones_supported`
- `services_supported`
- `protocol_versions`
- `expires_height`
- `signature`

Rules:

- Node record must be signed by node private key.
- Validator node record must be linkable to validator identity where advertised.
- Network address list may be stored off-chain but must match committed or signed hash.
- Record expires unless renewed.

### 4.3 Session Channels

Instead of raw connections, nodes establish `SessionChannel`.

Session fields:

- `local_node_id`
- `remote_node_id`
- `session_id`
- `handshake_version`
- `cipher_suite`
- `session_keys`
- `opened_at`
- `expires_at`
- `stream_set`
- `qos_policy`

Handshake requirements:

- Verify node identity.
- Negotiate supported protocols.
- Negotiate channel classes.
- Upgrade to encrypted session keys.
- Bind session to node IDs.
- Reject expired or mismatched node records where required.

### 4.4 Multiplexed Streams

Stream fields:

- `stream_id`
- `channel_type`
- `priority`
- `flow_control_window`
- `max_message_bytes`
- `compression_mode`
- `encryption_context`

Stream rules:

- Consensus stream gets reserved capacity.
- Service stream cannot starve consensus stream.
- Bulk data stream must support backpressure.
- Stream reset must not close the entire session unless policy requires it.

### 4.5 QoS Classes

QoS classes:

- `critical_consensus`
- `block_propagation`
- `state_sync`
- `execution_message`
- `service_call`
- `discovery`
- `bulk_data`

Rules:

- Each class has bandwidth floor and ceiling.
- Priority inversion is forbidden for consensus traffic.
- Peers exceeding service quotas can be downgraded without disconnecting consensus traffic.

### 4.6 L1 Implementation Tasks

- Define node identity derivation.
- Define signed node record format.
- Implement session handshake state machine.
- Implement multiplexed stream abstraction.
- Implement QoS policy.
- Add peer role validation.
- Add session key rotation.
- Add tests for replayed handshakes and expired node records.

## 5. L2: Overlay Routing Layer

### 5.1 Overlay Model

An overlay is a virtual network defined by intent:

```text
Overlay = (OverlayID, Policy, MembershipRule)
```

Overlay types:

- Validator Overlay.
- Zone Overlay.
- Execution Overlay.
- Data Overlay.
- Service Overlay.
- Discovery Overlay.
- Storage Overlay.
- Routing Overlay.

### 5.2 OverlayDescriptor

Fields:

- `overlay_id`
- `overlay_type`
- `policy_hash`
- `membership_rule`
- `routing_strategy`
- `min_peers`
- `max_peers`
- `fanout`
- `qos_class`
- `expires_height_optional`
- `version`

### 5.3 Overlay Membership

Nodes join overlays through:

- Deterministic rules.
- Cryptographic authorization.
- Stake-based inclusion.
- Service registry membership.
- Zone assignment.
- Dynamic routing assignment.

Membership proof types:

- Validator set proof.
- Zone assignment proof.
- Service registration proof.
- Provider stake proof.
- Signed authorization record.

### 5.4 Routing Model

Routing pipeline:

```text
message -> classify -> overlay -> routing graph -> target peers
```

Routing strategies:

- Shortest latency path routing.
- Zone-local routing.
- Probabilistic gossip fallback.
- Deterministic shard routing.
- Priority broadcast trees.
- Service-provider routing.
- Storage-provider routing.

Consensus-safe rule:

- Routing that affects committed execution must use committed routing table or deterministic route hints.
- Node-local adaptive routing can optimize delivery but cannot determine state transition order.

### 5.5 Adaptive Overlay Graph

Each node maintains:

- Fast peers with low latency.
- Stable peers with high reliability.
- Random peers for global connectivity.
- Zone peers for execution locality.
- Service peers for endpoint locality.
- Storage peers for data retrieval.

Peer sets:

- `fast_set`
- `stable_set`
- `random_set`
- `zone_set`
- `service_set`
- `fallback_set`

Rules:

- Peer rotation must preserve eclipse resistance.
- Random peer set must remain diverse.
- Zone peers must not fully replace global peers.
- Peer score decay must be bounded.
- Live peer score is advisory unless committed.

### 5.6 RoutingTableCommitment

Fields:

- `routing_epoch`
- `overlay_roots`
- `zone_route_root`
- `service_route_root`
- `peer_class_root`
- `congestion_snapshot_root`
- `policy_hash`

Usage:

- Committed routing table can be used by execution-layer message scheduling.
- Non-committed table can be used for physical packet forwarding only.

### 5.7 L2 Implementation Tasks

- Define overlay descriptor schema.
- Implement overlay membership manager.
- Implement peer set manager.
- Implement routing graph builder.
- Implement deterministic routing table commitment.
- Implement adaptive peer rotation.
- Implement route fallback logic.
- Add tests for overlay partition and peer churn.

## 6. L3: Application Networking Layer

### 6.1 Aether Mesh

Aether Mesh is the service-aware execution network layer.

It handles:

- Execution zone messages.
- Service calls.
- Queries.
- Cross-zone messages.
- Data requests.
- Storage retrieval.
- Execution receipts.
- Async callbacks.

### 6.2 NetworkMessage

```text
NetworkMessage {
  type
  payload
  origin
  destination
  priority
  ttl
}
```

Extended fields:

- `message_id`
- `overlay_id`
- `source_zone_optional`
- `destination_zone_optional`
- `sequence`
- `payload_hash`
- `route_hint_optional`
- `deadline_height_optional`
- `signature`
- `proof_optional`

Message types:

- `consensus`
- `tx`
- `execution`
- `query`
- `service`
- `cross_zone`
- `state_sync`
- `storage`
- `routing`

### 6.3 Execution Zone Messaging

Flow:

```text
tx -> routing class -> zone -> shard -> execution overlay
```

Supports:

- Async execution.
- Deterministic ordering per zone where committed.
- Parallel zone execution.
- Cross-zone message receipts.
- BlockSTM-ready execution groups.

Rules:

- Network delivery order is not consensus order.
- Consensus order is determined by committed transaction and message schedule.
- Cross-zone messages must include sequence ID.
- Message receipts must be rollback-safe and proof-queryable.

### 6.4 Cross-Zone Messaging Protocol

Guarantees:

- Ordered delivery per source zone where required.
- At-least-once delivery at network layer.
- Exactly-once execution through replay protection.
- Deterministic rollback-safe receipts.
- Expiry and bounce semantics at execution layer.

Fields:

- `source_zone`
- `destination_zone`
- `source_sequence`
- `message_hash`
- `expiry_height`
- `receipt_policy`
- `proof_required`

### 6.5 L3 Implementation Tasks

- Define network message envelope.
- Add message signing and verification.
- Add per-overlay message queues.
- Add cross-zone sequence tracking.
- Add receipt delivery protocol.
- Add query-response proof attachment.
- Add L3 metrics and replay tests.

## 7. Reliable Layer 2 Transport: RL2

### 7.1 Purpose

`RL2` is Aetheris reliable transport for large and important payloads.

Use cases:

- Large block propagation.
- Block chunk transfer.
- State sync streams.
- Zone snapshots.
- Execution result delivery.
- Storage object retrieval.
- Proof set transfer.

### 7.2 RL2 Transfer

Transfer fields:

- `transfer_id`
- `source_node`
- `target_node`
- `payload_type`
- `payload_root`
- `chunk_count`
- `chunk_size`
- `fec_policy_optional`
- `priority`
- `deadline_optional`
- `resume_token_optional`

### 7.3 Chunk Descriptor

Fields:

- `transfer_id`
- `chunk_index`
- `chunk_hash`
- `chunk_size`
- `range_start`
- `range_end`
- `proof_path_optional`

Rules:

- Payload root commits to ordered chunk hashes.
- Chunk hash must be verified before acceptance.
- Missing chunks can be requested by index.
- Transfer can resume using verified chunk bitmap.

### 7.4 Features

RL2 supports:

- Chunk-based streaming.
- Erasure coding support.
- Merkle-root reassembly verification.
- Resumable transfers.
- Bandwidth-adaptive streaming.
- Backpressure.
- Parallel stream channels.
- Priority lanes.

### 7.5 RL2 State Machine

```text
offered
  |
  v
accepted
  |
  v
streaming
  |
  +--> paused
  |
  +--> resumed
  |
  v
verified
  |
  v
completed
```

Failure states:

- `timeout`
- `cancelled`
- `invalid_chunk`
- `root_mismatch`
- `peer_disconnected`

### 7.6 RL2 Implementation Tasks

- Define transfer offer protocol.
- Define chunk descriptor encoding.
- Implement chunk verification.
- Implement resumable transfer bitmap.
- Implement adaptive chunk sizing.
- Implement backpressure signals.
- Add Merkle reassembly verification.
- Add tests for interrupted transfers and invalid chunks.

## 8. Discovery Layer

### 8.1 Distributed Routing Table

Aetheris uses a Distributed Routing Table (`DRT`) as a hybrid discovery system:

- Structured routing tables.
- Overlay-native discovery.
- Stake-aware node ranking.
- Zone-aware indexing.
- Service-aware indexing.
- Lease-based advertisements.

### 8.2 Discovery Objects

Indexable objects:

- Nodes.
- Execution zones.
- Service endpoints.
- RPC endpoints.
- Storage providers.
- Routing entry points.
- Overlay membership records.
- Stream providers.

### 8.3 DiscoveryRecord

Fields:

- `record_id`
- `record_type`
- `owner_node_id`
- `target_id`
- `advertisement_hash`
- `zone_id_optional`
- `service_id_optional`
- `overlay_id_optional`
- `expires_height`
- `signature`
- `proof_optional`

### 8.4 Query Model

Operations:

- `Find(Node)`
- `Find(Service)`
- `Find(Zone)`
- `Find(Endpoint)`
- `Find(StorageProvider)`
- `Store(Advertise)`
- `Update(Lease)`
- `Revoke(Advertisement)`

### 8.5 Lease Model

Rules:

- Discovery records expire by height or wall-clock TTL outside consensus.
- Consensus-relevant records must use height.
- Renewals must be signed.
- Expired records are ignored.
- Revoked records must include owner signature.

### 8.6 Proof-Attached Responses

Discovery response includes:

- Matched records.
- Signature chain.
- Optional on-chain proof.
- Expiry height.
- Source node signature.
- Result hash.

Rules:

- Client must verify signature and expiry.
- On-chain service records must match proof where provided.
- Unproofed mesh discovery is advisory.

### 8.7 Discovery Implementation Tasks

- Define discovery record schema.
- Implement DRT store and lookup.
- Implement overlay-native discovery.
- Implement lease renewal.
- Implement signed advertisement verification.
- Add proof-attached response format.
- Add tests for expired, forged, and replayed records.

## 9. Broadcast System

### 9.1 Hybrid Gossip Trees

Aetheris broadcast uses:

- Tree broadcast for fast propagation.
- Gossip fallback for resilience.
- Hash-based deduplication.
- Overlay-specific fanout.
- Priority-aware forwarding.

### 9.2 BroadcastMessage

Fields:

- `broadcast_id`
- `origin_node`
- `overlay_id`
- `payload_hash`
- `payload_type`
- `height_optional`
- `ttl`
- `priority`
- `fanout_policy`
- `signature`

### 9.3 Deduplication

Rules:

- Deduplication key is `broadcast_id`.
- Payload hash must match.
- Duplicate payloads are dropped.
- Conflicting payload for same broadcast ID is peer fault evidence.
- Dedup cache expires after configured horizon.

### 9.4 Block Propagation

Blocks propagate as:

```text
Block = Header + ChunkSet + ProofSet
```

Flow:

1. Receive header first.
2. Verify height and proposer context.
3. Validate availability metadata.
4. Stream chunks in parallel.
5. Verify chunk hashes.
6. Reconstruct block.
7. Verify block root and proof set.

### 9.5 Broadcast Implementation Tasks

- Implement broadcast tree builder.
- Implement gossip fallback.
- Implement dedup cache.
- Implement block header-first propagation.
- Implement block chunk metadata.
- Add tests for duplicate, conflicting, and partial broadcasts.

## 10. Streaming Network Model

### 10.1 Streaming Use Cases

Large data uses streaming abstraction:

- State sync.
- Zone snapshots.
- Block propagation.
- Execution receipts.
- Storage objects.
- Proof bundles.
- Historical query ranges.

### 10.2 StreamSession

Fields:

- `stream_id`
- `session_id`
- `payload_type`
- `priority`
- `flow_control_window`
- `chunk_size`
- `parallelism`
- `bytes_sent`
- `bytes_acknowledged`
- `state`

States:

- `opening`
- `active`
- `paused`
- `draining`
- `closed`
- `failed`

### 10.3 Backpressure

Signals:

- `window_update`
- `pause`
- `resume`
- `slow_down`
- `cancel`

Rules:

- Receiver controls flow window.
- Sender must respect flow control.
- Priority lanes can preempt lower priority streams.
- Bulk stream cannot starve consensus and execution streams.

### 10.4 Adaptive Chunking

Inputs:

- observed throughput.
- loss rate.
- peer score.
- payload priority.
- stream class.

Rules:

- Chunk size changes only at chunk boundaries.
- Chunk hash uses actual chunk bytes.
- Reassembly root remains stable.

### 10.5 Streaming Implementation Tasks

- Define stream session protocol.
- Implement flow control.
- Implement adaptive chunk sizing.
- Implement parallel chunk fetch.
- Implement priority lanes.
- Add metrics for stream throughput and stalls.
- Add tests for backpressure and partial stream recovery.

## 11. Network Security Model

### 11.1 Threat Assumptions

Threats:

- Malicious peers.
- Eclipse attacks.
- Routing manipulation.
- Spam floods.
- Discovery poisoning.
- Service advertisement forgery.
- Chunk corruption.
- Bandwidth exhaustion.
- Sybil peers.
- Delayed or withheld block chunks.
- Cross-zone message replay.

### 11.2 Mitigations

Controls:

- Peer reputation scoring.
- Adaptive peer rotation.
- Cryptographic channel binding.
- Message authentication at every layer.
- Deterministic replay protection.
- Overlay isolation per zone.
- Signed discovery records.
- Expiring advertisements.
- Hash-based deduplication.
- Chunk Merkle verification.
- Per-channel rate limits.
- QoS isolation.

### 11.3 Peer Reputation

Reputation inputs:

- valid message rate.
- invalid message rate.
- latency.
- throughput.
- chunk correctness.
- discovery response validity.
- service response validity.
- timeout rate.
- duplicate or conflicting broadcast rate.

Rules:

- Local reputation is advisory.
- Consensus-affecting reputation must be committed through deterministic evidence.
- Peer penalties are local unless converted into on-chain evidence.
- Reputation decay is bounded.

### 11.4 Eclipse Resistance

Requirements:

- Maintain random peer set.
- Maintain validator peer diversity.
- Maintain zone peer diversity.
- Rotate discovery sources.
- Limit peers per network identity cluster where detectable.
- Prefer proof-backed records for critical routing.

### 11.5 Spam Resistance

Controls:

- Per-peer message rate limits.
- Per-channel byte limits.
- Handshake cost limits.
- Proof-of-resource or stake-backed service advertisements where needed.
- Payload size limits.
- Chunk request limits.
- Duplicate suppression.

### 11.6 Security Implementation Tasks

- Implement peer scoring.
- Implement signed message envelopes.
- Implement replay cache.
- Implement per-channel rate limiting.
- Implement advertisement signature validation.
- Implement chunk root verification.
- Add simulations for eclipse, spam, and routing manipulation.

## 12. Performance Model

### 12.1 Optimization Goals

Networking stack is optimized for:

- Parallel message propagation.
- Multi-overlay concurrency.
- No dependency on global broadcast for all traffic.
- Shard-local execution scaling.
- Zone isolation.
- Streaming large payloads.
- Bounded gossip fanout.
- Predictable latency per zone.

### 12.2 Target Properties

Targets:

- `O(log N)` discovery under normal mesh conditions.
- `O(k)` neighbor propagation for overlay fanout.
- Bounded gossip fanout by overlay policy.
- Header-first block propagation.
- Parallel chunk streaming.
- Zone-local low-latency paths.
- Service-layer traffic isolation from consensus traffic.

### 12.3 Performance Metrics

Metrics:

- peer count by role.
- overlay membership size.
- message propagation latency.
- block header propagation latency.
- block reconstruction time.
- chunk retry rate.
- stream throughput.
- stream stall count.
- discovery query latency.
- route failure rate.
- per-channel bandwidth usage.
- peer score distribution.
- cross-zone message delivery latency.

### 12.4 Performance Implementation Tasks

- Add per-overlay metrics.
- Add channel-level bandwidth metrics.
- Add block propagation benchmarks.
- Add chunk streaming benchmarks.
- Add discovery latency benchmarks.
- Add cross-zone message delivery benchmarks.
- Add service traffic isolation tests.

## 13. Cosmos SDK and CometBFT Compatibility

### 13.1 Compatibility Requirements

Networking must remain compatible with:

- CometBFT P2P transport.
- CometBFT consensus messages.
- Cosmos SDK transaction flow.
- ABCI++ lifecycle:
  - `PrepareProposal`
  - `ProcessProposal`
  - `FinalizeBlock`
- BlockSTM-ready execution grouping.
- gRPC, REST, and RPC external APIs.
- State sync and snapshot mechanisms.

### 13.2 ABCI++ Integration

PrepareProposal:

- Use ANA hints to improve transaction and message proposal grouping.
- Do not use non-deterministic network state as validity input.
- Include eligible execution messages by deterministic state schedule.

ProcessProposal:

- Verify proposal does not depend on unverifiable peer-local data.
- Verify message ordering commitments.

FinalizeBlock:

- Execute committed messages only.
- Emit receipts and roots.
- Ignore live network state.

### 13.3 BlockSTM Integration

Network assists BlockSTM by:

- Propagating zone and shard route hints.
- Grouping transactions by target zone.
- Prioritizing execution overlay traffic.
- Delivering cross-zone messages to correct execution queues.

Rules:

- BlockSTM conflict resolution is execution-layer logic.
- Networking can optimize delivery but cannot decide committed conflicts.

### 13.4 API Integration

External APIs:

- gRPC service discovery queries.
- REST node and overlay queries.
- RPC peer and stream diagnostics.
- Proof-attached discovery responses.
- State sync stream endpoints.

Implementation tasks:

- Add node networking query service.
- Add overlay diagnostics endpoint.
- Add stream diagnostics endpoint.
- Add discovery proof API.
- Add route hint API for clients.

## 14. Module and Component Map

### 14.1 Node-Side Components

Components:

- `ana`
  - Aether Networking Adapter.
- `sessionmgr`
  - Node identity, handshake, session keys, stream multiplexing.
- `overlaymgr`
  - Overlay membership and peer sets.
- `drt`
  - Distributed routing table.
- `rl2`
  - Reliable chunked transport.
- `mesh`
  - Application networking and service message flow.
- `broadcast`
  - Hybrid gossip tree and dedup.

### 14.2 On-Chain Support Modules

Optional support modules:

- `x/network`
  - committed network parameters and node records.
- `x/routing`
  - routing table commitments and overlay descriptors.
- `x/services`
  - service endpoint and provider records.
- `x/storage`
  - storage provider commitments.
- `x/messages`
  - cross-zone message receipts and replay protection.

### 14.3 x/network State

State keys:

- `network/params`
- `network/nodes/{node_id}`
- `network/roles/{role}/{node_id}`
- `network/overlays/{overlay_id}`
- `network/discovery/{record_id}`
- `network/reputation/{node_id}`
- `network/evidence/{evidence_id}`

Messages:

- `MsgRegisterNode`
- `MsgUpdateNode`
- `MsgRenewNode`
- `MsgRevokeNode`
- `MsgSubmitNetworkEvidence`

Queries:

- `QueryNode`
- `QueryNodesByRole`
- `QueryOverlay`
- `QueryDiscoveryRecord`
- `QueryNetworkParams`
- `QueryNetworkEvidence`

## 15. Implementation Roadmap

### Phase 0: Baseline and Instrumentation

Tasks:

- Inventory current CometBFT P2P configuration.
- Add peer metrics collection.
- Add channel bandwidth metrics.
- Add block propagation latency metrics.
- Add mempool propagation metrics.
- Define network parameter schema.

Exit criteria:

- Current network behavior is measurable.
- Baseline propagation and peer quality metrics exist.

### Phase 1: Aether Networking Adapter

Tasks:

- Implement ANA wrapper.
- Add logical channel classes.
- Add peer score model.
- Add adaptive fanout configuration.
- Add QoS policy.
- Add service traffic isolation.

Exit criteria:

- Consensus traffic has protected priority.
- Peer scoring and channel metrics are available.
- Service traffic cannot starve consensus traffic.

### Phase 2: Node Identity and Sessions

Tasks:

- Define NodeID derivation.
- Define NodeRecord.
- Implement signed node advertisements.
- Implement session handshake.
- Implement multiplexed streams.
- Implement session key rotation.

Exit criteria:

- Nodes authenticate each other through cryptographic identity.
- Logical streams can share one peer session.
- Expired or forged node records are rejected.

### Phase 3: Overlay Routing

Tasks:

- Define OverlayDescriptor.
- Implement overlay membership.
- Implement overlay peer sets.
- Implement route graph.
- Implement routing table commitment.
- Add zone and service overlays.

Exit criteria:

- Nodes can join validator, zone, service, data, and discovery overlays.
- Overlay routing decisions are reproducible when committed.
- Peer rotation preserves connectivity.

### Phase 4: RL2 Streaming

Tasks:

- Implement transfer offer protocol.
- Implement chunk descriptors.
- Implement chunk Merkle verification.
- Implement resumable transfer.
- Implement adaptive chunk sizing.
- Implement backpressure.

Exit criteria:

- Blocks, state snapshots, and proof bundles can stream in chunks.
- Interrupted transfers can resume.
- Invalid chunks are rejected.

### Phase 5: Discovery Layer

Tasks:

- Implement DRT.
- Add discovery records.
- Add lease renewal.
- Add proof-attached lookup responses.
- Add service and zone indexes.
- Add signed advertisement validation.

Exit criteria:

- Nodes, zones, services, endpoints, and storage providers are discoverable.
- Discovery records expire and can be verified.
- Forged or expired records are rejected.

### Phase 6: Hybrid Broadcast

Tasks:

- Implement tree broadcast.
- Implement gossip fallback.
- Implement hash deduplication.
- Implement header-first block propagation.
- Implement parallel chunk fetch.

Exit criteria:

- Blocks propagate as header plus chunks and proof set.
- Duplicate and conflicting broadcasts are handled.
- Fallback gossip preserves resilience.

### Phase 7: Aether Mesh

Tasks:

- Implement application network message envelope.
- Add execution zone message routing.
- Add cross-zone sequence handling.
- Add receipt delivery protocol.
- Add query response proof attachment.
- Add service network traffic path.

Exit criteria:

- L3 messages support execution, service, query, storage, and cross-zone classes.
- Cross-zone delivery is at-least-once at network layer and exactly-once at execution layer.
- Receipts are delivery-visible and proof-queryable where committed.

### Phase 8: Security and Load Hardening

Tasks:

- Add peer reputation hardening.
- Add eclipse resistance tests.
- Add spam flood simulations.
- Add routing manipulation simulations.
- Add bandwidth exhaustion tests.
- Add chunk corruption tests.

Exit criteria:

- Malicious peers are isolated locally.
- Critical channels remain available under service flood.
- Discovery poisoning is detected by signature and proof checks.

## 16. Required Test Coverage

### 16.1 Unit Tests

- NodeID derivation.
- NodeRecord signature verification.
- Session handshake validation.
- Stream priority classification.
- Overlay membership validation.
- Route cost calculation.
- NetworkMessage ID derivation.
- DiscoveryRecord expiry.
- Chunk hash verification.
- Broadcast deduplication.

### 16.2 Integration Tests

- CometBFT consensus traffic with ANA enabled.
- Multiplexed streams over one session.
- Zone overlay formation.
- Service overlay formation.
- Cross-zone message delivery.
- RL2 block chunk transfer.
- Resumable state snapshot transfer.
- Discovery lookup with proof-attached response.
- Header-first block propagation.

### 16.3 Security Tests

- Replayed handshake.
- Forged node advertisement.
- Expired discovery record.
- Conflicting broadcast payload.
- Invalid chunk.
- Eclipse peer set simulation.
- Spam flood on service channel.
- Consensus traffic under bulk data load.
- Cross-zone message replay.

### 16.4 Performance Tests

- Block header propagation latency.
- Block reconstruction time.
- Chunk streaming throughput.
- Discovery query latency.
- Overlay join latency.
- Cross-zone message propagation latency.
- Service traffic throughput.
- Consensus traffic latency under mixed load.
- Peer rotation stability.

## 17. Observability

### 17.1 Metrics

- Active peers.
- Peers by role.
- Active sessions.
- Streams by channel type.
- Per-channel bandwidth.
- Peer score.
- Overlay size.
- Overlay churn.
- Discovery query latency.
- Broadcast dedup hit rate.
- RL2 transfer throughput.
- RL2 chunk retry rate.
- Block propagation latency.
- Cross-zone message delivery latency.
- Service traffic volume.
- Routing failure count.

### 17.2 Events

- `network_node_registered`
- `network_session_opened`
- `network_session_closed`
- `network_peer_score_updated`
- `network_overlay_joined`
- `network_overlay_left`
- `network_discovery_record_stored`
- `network_discovery_record_expired`
- `network_rl2_transfer_started`
- `network_rl2_transfer_completed`
- `network_invalid_chunk`
- `network_broadcast_conflict`
- `network_route_failed`

### 17.3 Alerts

- Consensus channel latency above threshold.
- Block propagation latency spike.
- Peer score collapse.
- Overlay partition suspected.
- Discovery poisoning attempt.
- RL2 invalid chunk spike.
- Service traffic exceeding quota.
- Cross-zone message delivery backlog.
- Eclipse risk peer diversity low.

## 18. Non-Goals

This networking layer does not:

- Implement application logic.
- Replace CometBFT consensus.
- Assume centralized routing.
- Rely on external discovery services.
- Introduce a messaging or social network layer.
- Make live network metrics consensus-authoritative.
- Execute off-chain service logic inside consensus.

## 19. Acceptance Criteria

Aetheris networking is ready for implementation planning when:

- L0 CometBFT transport remains consensus-critical and protected.
- ANA provides channel classification, peer scoring, adaptive fanout, and QoS isolation.
- L1 node identity and session channels are defined and testable.
- L2 overlays support validator, zone, execution, data, service, and discovery networks.
- L3 Aether Mesh supports execution, query, service, storage, and cross-zone message classes.
- RL2 supports chunked, verified, resumable, bandwidth-adaptive streaming.
- DRT supports signed, expiring, proof-attached discovery records.
- Hybrid broadcast supports tree propagation, gossip fallback, and hash deduplication.
- Security controls cover peer reputation, eclipse resistance, spam resistance, replay protection, and chunk verification.
- Cosmos SDK and ABCI++ integration rules are explicit.
- Tests cover unit, integration, security, and performance cases.
