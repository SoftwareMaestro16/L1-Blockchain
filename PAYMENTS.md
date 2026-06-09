# Aetheris Payment Layer Specification

Status: Internal design document
Scope: Off-chain payment architecture for Aetheris
Visibility: Private, not for public repository inclusion

## 1. Design Scope

### 1.1 Objective

Define a Cosmos SDK-native payment layer for Aetheris that supports:

- Instant or near-instant off-chain payments.
- High-frequency `naet` transfers.
- Streaming payment flows.
- Trustless channel settlement.
- Multi-hop payment routing.
- Virtual channels over existing liquidity paths.
- Batched on-chain settlement with parallel execution compatibility.
- Deterministic fraud resolution through on-chain arbitration.

### 1.2 Non-Goals

- Do not introduce a second payment asset.
- Do not require custodial intermediaries.
- Do not require on-chain execution for every payment update.
- Do not depend on non-deterministic off-chain data for settlement.
- Do not make routing node reputation consensus-critical.
- Do not allow channel close rules to depend on wall-clock time outside consensus time.
- Do not couple payment finality to external network guarantees.

### 1.3 Base Assumptions

- Native denom: `naet`.
- Display denom: `AET`.
- Chain runtime:
  - Cosmos SDK `v0.54+`.
  - CometBFT `v0.39+`.
  - Store v2.
  - BlockSTM-capable execution.
  - AdaptiveSync-capable node recovery.
- Economic layer:
  - Native-fee-only execution in `naet`.
  - Congestion-aware dynamic fees.
  - Validator and community fee split.
  - Delegation-based staking.
  - Slashing for downtime and double-sign.
- Execution layer:
  - Parallelizable transaction execution.
  - Async message passing available as a design target.
  - Per-byte storage pricing primitives.

## 2. System Components

### 2.1 Component Map

| Component | Responsibility | Execution Location |
| --- | --- | --- |
| Channel participants | Sign off-chain channel states and submit settlement transactions | Off-chain and on-chain |
| Settlement contract | Deterministic arbitrator for opens, closes, disputes, and final balances | On-chain |
| Payment channel module | Native state machine for channel custody, lifecycle, penalties, and settlement | On-chain |
| Conditional payments module | Hash-lock and time-lock promises for atomic multi-channel payments | On-chain and off-chain |
| Routing engine | Path discovery, capacity scoring, and route construction | Off-chain |
| Gossip network | Channel topology and liquidity signal propagation | Off-chain |
| Liquidity optimizer | Rebalancing, fee selection, and liquidity advertisement | Off-chain with optional on-chain commitments |
| Fraud-proof verifier | Validates signed-state fraud proofs and penalty claims | On-chain |
| Watch service | Monitors channels and submits dispute evidence when needed | Off-chain |
| Settlement batcher | Groups independent channel operations for efficient block inclusion | Off-chain and on-chain |

### 2.2 Trust Boundaries

- Off-chain routing messages are advisory.
- Off-chain payment states are valid only with required signatures.
- Gossip topology is untrusted and must be verified before route use.
- Settlement contract state is authoritative for channel custody and final balances.
- Fraud proofs are valid only if derived from signed channel states.
- Validator arbitration is limited to deterministic execution of settlement rules.

### 2.3 Asset Scope

- All base channel collateral is denominated in `naet`.
- Fees for opening, closing, disputing, and settling channels are paid in `naet`.
- Routing fees are denominated in `naet`.
- Penalties are denominated in `naet`.
- Future support for additional denoms requires explicit accounting isolation and is out of scope for the initial design.

## 3. Channel Architecture

### 3.1 Channel Types

#### 3.1.1 Bidirectional Channels

Purpose:

- Allow two parties to exchange signed balance updates without on-chain confirmation per payment.
- Support repeated transfers in both directions.
- Support conditional transfers and virtual channel backing liquidity.

State:

- `channel_id`
- `chain_id`
- `version`
- `participant_a`
- `participant_b`
- `balance_a`
- `balance_b`
- `reserve_a`
- `reserve_b`
- `nonce`
- `epoch`
- `pending_conditions_root`
- `previous_state_hash`
- `state_hash`
- `timeout_height`
- `timeout_timestamp`
- `close_delay`
- `fee_policy_id`
- `signatures`

Rules:

- `balance_a + balance_b + reserve_a + reserve_b` must equal locked collateral minus settled fees and penalties.
- `nonce` must strictly increase for each mutually accepted state.
- `state_hash` must include all balance, nonce, condition, timeout, and domain fields.
- Both participant signatures are required for cooperative updates.
- Either party may submit the latest signed state to close.
- Any party may dispute a stale close with a newer valid state during the dispute window.

Required implementation tasks:

- Define canonical state encoding.
- Define domain-separated signature preimage.
- Add channel state commitment hashing.
- Add cooperative close flow.
- Add unilateral close flow.
- Add dispute flow for stale states.
- Add penalty flow for provable fraud.

#### 3.1.2 Unidirectional Channels

Purpose:

- Optimize high-frequency payments from one sender to one receiver.
- Reduce signature and state-management overhead.
- Support streaming payments where value only moves in one direction until channel reset or close.

State:

- `channel_id`
- `payer`
- `receiver`
- `locked_amount`
- `claimed_amount`
- `nonce`
- `expiration_height`
- `expiration_timestamp`
- `state_hash`
- `payer_signature`
- `receiver_ack_optional`

Rules:

- Payer locks full collateral on open.
- Receiver accepts monotonically increasing signed claims.
- Receiver can close with the highest signed claim.
- Payer can reclaim unclaimed balance after expiration and dispute window.
- Receiver does not need to sign every update unless acknowledgements are enabled.

Required implementation tasks:

- Add single-signer claim verification.
- Add receiver close transaction.
- Add payer reclaim transaction.
- Add optional receiver acknowledgement mode.
- Add streaming payment helper format.

#### 3.1.3 Asynchronous Payment Channels

Purpose:

- Allow participants to issue updates without waiting for per-step acknowledgement when the risk model permits it.
- Support high-latency counterparties and service-metered payments.
- Enable payment queues that settle into signed channel checkpoints.

State:

- `channel_id`
- `participants`
- `checkpoint_nonce`
- `checkpoint_balances`
- `async_update_root`
- `accepted_update_root`
- `send_window`
- `receive_window`
- `max_unacked_amount`
- `expiry_height`
- `signatures`

Rules:

- Async updates are represented as signed deltas.
- Deltas must include unique `update_id`, `channel_id`, direction, amount, nonce range, and expiry.
- Receivers aggregate deltas into checkpoints.
- Risk exposure is capped by `max_unacked_amount`.
- A disputed checkpoint must reveal enough signed deltas to reconstruct the accepted state.
- Expired unacknowledged deltas are not settleable.

Required implementation tasks:

- Define signed delta format.
- Define checkpoint aggregation format.
- Add max-exposure validation.
- Add delta expiry rules.
- Add dispute proof format for async deltas.

### 3.2 Channel Lifecycle

#### 3.2.1 Open

Inputs:

- Participants.
- Initial balances.
- Channel type.
- Collateral amount.
- Close delay.
- Challenge period.
- Fee policy.
- Optional routing advertisement flag.
- Optional conditional payment capability flag.

On-chain effects:

- Lock `naet` collateral in settlement contract custody.
- Create channel record.
- Store opening state commitment.
- Charge channel opening fee.
- Emit channel-open event.

Validation:

- Initial balances sum to collateral.
- Participants are valid addresses.
- Channel ID is unique.
- Opening fee is paid.
- Close delay is within configured bounds.
- Challenge period is within configured bounds.

#### 3.2.2 Update

Inputs:

- Off-chain signed state.
- Optional condition commitments.
- Optional async delta batch.

Off-chain effects:

- Participants verify domain, nonce, balances, conditions, and signatures.
- Participants persist latest valid state.
- Routing nodes update local liquidity availability.

On-chain effects:

- None for ordinary updates.
- Optional checkpoint registration for async or high-value channels.

Validation:

- Nonce increases.
- Balances conserve collateral.
- Conditions are well-formed.
- Signatures match required participants.
- State does not exceed exposure or reserve limits.

#### 3.2.3 Close

Close modes:

- Cooperative close.
- Unilateral close.
- Expired unidirectional close.
- Forced close due to timeout.
- Fraud close after accepted proof.

Cooperative close:

- Requires final state signed by all required parties.
- Settles immediately or after minimal finalization delay.
- Lowest settlement cost.

Unilateral close:

- Requires one valid signed state.
- Starts challenge period.
- Any counterparty can submit a newer valid state.

Expired unidirectional close:

- Receiver claims signed amount before expiration.
- Payer reclaims unclaimed amount after expiration and finalization.

Forced close:

- Used when channel timeouts or liveness conditions expire.
- Must preserve dispute rights for the configured challenge period.

#### 3.2.4 Dispute

Inputs:

- `channel_id`
- Closing state reference.
- Newer signed state.
- Optional fraud proof.
- Optional condition proof.

Rules:

- Dispute must occur before challenge period ends.
- Newer state must have strictly greater nonce or stronger finality marker.
- State signature set must satisfy channel type requirements.
- Fraud proof must demonstrate invalid close, double-signed state, invalid condition resolution, or replay attempt.

On-chain effects:

- Replace pending close state with newer valid state.
- Extend dispute window only if configured and bounded.
- Apply penalty if fraud is proven.
- Emit dispute event.

#### 3.2.5 Final Settlement

Inputs:

- Pending close state.
- Resolved conditional payments.
- Penalty state.
- Fee accounting state.

On-chain effects:

- Unlock final balances.
- Apply routing fee claims where enforceable.
- Apply penalties.
- Apply settlement fees.
- Delete or archive active channel state.
- Persist minimal settlement record for replay protection.

Validation:

- Challenge period expired or cooperative close is complete.
- All conditions are resolved or expired.
- Final balance distribution is non-negative.
- Settlement result conserves locked collateral minus fees and penalties.

## 4. State Representation

### 4.1 Canonical Channel State

Canonical channel state must include:

- Chain domain:
  - `chain_id`
  - `app_version`
  - `module_name`
- Channel domain:
  - `channel_id`
  - `channel_type`
  - `participant_set_hash`
- Balance domain:
  - Participant balances.
  - Reserves.
  - Pending conditional amounts.
  - Accrued fees.
- Progress domain:
  - `nonce`
  - `epoch`
  - `previous_state_hash`
- Time domain:
  - `timeout_height`
  - `timeout_timestamp`
  - `challenge_period`
- Condition domain:
  - `condition_root`
  - `condition_count`
- Signature domain:
  - Required signer bitmap.
  - Signature scheme identifier.
  - Signature preimage hash.

### 4.2 State Hash

Hash requirements:

- Deterministic binary encoding.
- Stable field order.
- Explicit version byte.
- Domain separation for:
  - Channel state.
  - Async delta.
  - Conditional promise.
  - Cooperative close.
  - Dispute proof.
  - Virtual channel state.
- Include `chain_id` and `channel_id` in every signable object.

Required tasks:

- Implement canonical encoding tests.
- Add cross-version hash compatibility tests.
- Reject unknown required fields.
- Add state hash query endpoint for debugging.

### 4.3 Nonce Model

Rules:

- Every mutually signed channel state has a strictly increasing `nonce`.
- Async deltas use unique delta nonces and aggregate into checkpoint nonces.
- Conditional payments use condition IDs tied to channel nonce ranges.
- Virtual channels use independent virtual nonces anchored to parent-channel commitments.
- Settlement contract stores highest finalized or disputed nonce per channel.

Failure handling:

- Same nonce with different state hashes is double-sign evidence.
- Lower nonce submitted during close is stale-close evidence when a higher signed nonce exists.
- Missing nonce gaps are valid only for async channels if checkpoint proof reconstructs the final state.

### 4.4 Commitment Model

Commitment types:

- Opening commitment.
- Balance state commitment.
- Condition root commitment.
- Async delta root commitment.
- Virtual channel anchor commitment.
- Settlement result commitment.

Rules:

- All commitments must include channel domain fields.
- Condition roots must be reproducible from included promises.
- Virtual channel anchors must bind parent channels, participants, capacity, and expiration.
- Settlement results must bind final balances and penalty routes.

## 5. Cryptographic Model

### 5.1 Signature Requirements

Every signed object must include:

- `chain_id`
- `channel_id`
- `object_type`
- `version`
- `nonce` or unique object ID
- `expiration_height` where applicable
- Full commitment hash

Signature validation:

- Verify signer address matches expected participant.
- Reject signatures for a different chain domain.
- Reject signatures for a different channel domain.
- Reject expired signatures where expiration applies.
- Reject duplicate signatures for the same signer slot.

### 5.2 Double-Sign Prevention

Definition:

- A participant double-signs if they sign two conflicting channel states with the same `channel_id`, `epoch`, and `nonce`.

Prevention:

- Local signer must persist highest signed nonce per channel.
- Signer must refuse same-nonce replacement unless state hash is identical.
- Signer must support durable write-ahead logging before signature release.
- Hardware or process-isolated signer should be supported for high-value routing nodes.

On-chain handling:

- Conflicting same-nonce signed states are valid fraud evidence.
- Penalty applies to the double-signing participant.
- If both participants double-sign, penalties apply independently.
- Settlement uses highest valid non-conflicting state after penalty rules are applied.

### 5.3 Replay Attack Prevention

Replay protection fields:

- `chain_id`
- `channel_id`
- `participant_set_hash`
- `nonce`
- `epoch`
- `object_type`
- `expiration_height`

Settlement contract rules:

- Reject states for closed channels.
- Reject states below finalized nonce.
- Reject reused condition claims.
- Reject reused preimage claims after condition resolution.
- Persist minimal closed-channel tombstone until replay horizon expires.

### 5.4 State Rollback Protection

Mechanisms:

- Monotonic nonces.
- Previous-state hash chaining.
- Highest locally signed nonce persistence.
- Optional checkpoint registration for high-value channels.
- Watch service dispute submission.
- On-chain tombstones for finalized channels.

Required tasks:

- Add signer persistence library.
- Add rollback test vectors.
- Add stale-close dispute tests.
- Add previous-hash continuity checks where enabled.

## 6. Trustless Settlement Model

### 6.1 Settlement Contract Responsibilities

The settlement contract acts as deterministic arbitrator for:

- Channel opening.
- Collateral custody.
- Cooperative close.
- Unilateral close.
- Dispute acceptance.
- Fraud proof verification.
- Conditional payment resolution.
- Penalty routing.
- Final balance settlement.
- Replay protection.

The settlement contract must not:

- Select payment routes.
- Trust gossip state.
- Depend on external liquidity reports.
- Accept unsigned balance updates.
- Infer participant intent from off-chain messages that are not part of the signed state.

### 6.2 Unilateral Close

Rules:

- Any participant can close with a valid signed state.
- Close submission stores pending settlement state.
- Challenge period starts at close inclusion height.
- Counterparty can submit newer valid state before finalization.
- Finalization occurs after challenge period if no valid dispute supersedes the close state.

Required fields:

- `channel_id`
- `closing_state`
- `signatures`
- `close_reason`
- `settlement_fee`

Validation:

- Channel exists and is open.
- State signatures meet channel type requirements.
- State balances conserve collateral.
- State nonce is at least the stored opening nonce.
- Submitter is authorized.

### 6.3 Fraud Proofs

Fraud proof categories:

- Stale close:
  - Submit higher-nonce valid state than pending close.
- Same-nonce conflict:
  - Submit two signed states with same nonce and different hashes.
- Invalid balance:
  - Prove submitted state violates collateral conservation.
- Invalid condition:
  - Prove condition was resolved with wrong preimage, expired promise, or invalid proof.
- Replay:
  - Prove state was submitted outside channel domain or after channel finalization.
- Async overexposure:
  - Prove aggregate deltas exceed configured exposure.

Penalty rules:

- Fraud submitter pays dispute processing cost if proof is invalid.
- Fraudulent closer loses penalty amount from channel balance or bonded deposit.
- Double-signing participant loses configured double-sign penalty.
- Accepted reporter receives configured reward capped by penalty amount.
- Remaining penalty routes to burn, security reserve, or community pool according to governance parameters.

### 6.4 Settlement Finality

Finality states:

- `Open`
- `PendingClose`
- `InDispute`
- `PendingConditionResolution`
- `Finalizable`
- `Settled`
- `Penalized`
- `Expired`

Economic finality:

- Cooperative close achieves economic finality after inclusion and settlement execution.
- Unilateral close achieves economic finality after challenge period expires.
- Disputed close achieves economic finality after final accepted state plus challenge period rules.
- Conditional payments achieve finality after condition resolution or expiry.

Required tasks:

- Add explicit channel finality enum.
- Emit events for every finality transition.
- Add query for pending finalization height.
- Add invariant tests for locked collateral during every finality state.

## 7. Conditional Payments

### 7.1 Promise Object

Purpose:

- Represent a conditional transfer that can be resolved by proof or timeout.

Fields:

- `promise_id`
- `channel_id`
- `source`
- `destination`
- `amount`
- `fee`
- `hash_lock`
- `timeout_height`
- `timeout_timestamp`
- `condition_type`
- `route_id_optional`
- `previous_promise_id_optional`
- `next_promise_id_optional`
- `nonce`
- `signature`

Rules:

- Promise ID must be unique within channel.
- Promise amount must be reserved before advertisement.
- Promise timeout must fit within channel close and dispute windows.
- Promise cannot be settled twice.
- Promise cannot exceed available reserve.

### 7.2 Hash-Locked Transfers

Rules:

- Sender creates promise with `hash_lock = H(preimage)`.
- Receiver claims by revealing valid preimage before timeout.
- Intermediaries forward conditional promises with compatible hash locks.
- Preimage revelation resolves all linked promises that are still valid.

Required tasks:

- Add hash-lock verification.
- Add preimage reveal transaction.
- Add used-preimage tracking per condition.
- Add condition root update rules.

### 7.3 Time-Locked Guarantees

Rules:

- Every conditional promise has an expiry.
- Downstream timeout must expire before upstream timeout by a safety margin.
- Timeout margin must cover dispute and settlement latency.
- Expired promises release reserved capacity back to channel participants.

Required tasks:

- Define timeout ordering validation.
- Add expiry resolution transaction.
- Add timeout margin parameter.
- Add tests for timeout races.

### 7.4 Multi-Party Chained Payments

Flow:

- Sender selects route with sufficient capacity.
- Sender constructs chained promises across route.
- Each hop verifies incoming and outgoing amounts, fees, hash lock, and timeout ordering.
- Receiver reveals preimage to claim final promise.
- Preimage propagates backward to settle each prior promise.

Rules:

- Each hop can claim fee only when forwarding obligation is fulfilled.
- Timeout ordering must protect intermediaries.
- Route-level amount conservation must include all hop fees.
- Intermediaries cannot settle unless they can prove downstream resolution or timeout.

### 7.5 Atomic Cross-Channel Settlement

Atomicity rule:

- Linked conditional promises either settle through valid preimage before timeout or expire without value transfer.

Settlement contract behavior:

- Validate promise linkage.
- Validate hash lock and preimage.
- Validate timeout.
- Apply settlement across all submitted linked promises in one transaction when possible.
- Allow partial on-chain proofs when some channels settle off-chain and only disputed links require arbitration.

Required tasks:

- Define condition linkage proof.
- Add batch condition settlement message.
- Add partial dispute handling.
- Add invariant tests for atomic promise resolution.

## 8. Payment Routing Network

### 8.1 Topology Discovery

Gossip message types:

- `ChannelAnnouncement`
- `ChannelUpdate`
- `LiquidityHint`
- `FeePolicyUpdate`
- `NodeAnnouncement`
- `RouteFailure`
- `CapacityProbe`

Rules:

- Gossip messages are signed by advertising node.
- Channel announcements must reference an existing open channel or a verifiable channel commitment.
- Liquidity hints are advisory and must be treated as stale-prone.
- Fee policies include validity window and maximum fee.
- Nodes maintain local reputation only for routing decisions, not consensus.

Required tasks:

- Define signed gossip envelope.
- Add topology store.
- Add message expiry and pruning.
- Add invalid-gossip penalty in local routing score.

### 8.2 Path Selection

Path scoring inputs:

- Available capacity.
- Minimum required amount.
- Hop count.
- Routing fee.
- Timeout margin.
- Historical success rate.
- Liquidity freshness.
- Channel congestion.
- Node availability.
- Local policy constraints.

Algorithm requirements:

- Use Dijkstra-like path search with capacity-aware edge weighting.
- Exclude edges below required capacity.
- Penalize stale liquidity hints.
- Penalize high failure rate.
- Penalize routes with insufficient timeout margin.
- Support multi-path splitting when enabled.
- Support deterministic route scoring for local reproducibility.

Edge weight example:

- `edge_cost = base_fee + proportional_fee + hop_penalty + congestion_penalty + stale_liquidity_penalty + failure_penalty`

Required tasks:

- Implement path scorer.
- Add configurable route policy.
- Add capacity-aware path search.
- Add multi-path splitting policy.
- Add route simulation before payment attempt.

### 8.3 Dynamic Congestion Adaptation

Congestion signals:

- Channel update failure rate.
- Pending conditional payment count.
- Average condition resolution latency.
- Route-level retry count.
- Channel reserve pressure.
- Node queue delay.

Adaptation rules:

- Increase edge weight for congested channels.
- Prefer routes with lower pending condition exposure.
- Reduce maximum payment size through congested edges.
- Retry with alternate route after bounded failure count.
- Decay congestion penalty over time.

Required tasks:

- Add route failure classification.
- Add edge penalty decay.
- Add liquidity freshness timers.
- Add congestion-aware retry policy.

### 8.4 Routing Privacy Constraints

Requirements:

- Routing nodes should only learn information required to forward their hop.
- Route construction should avoid exposing full path to every intermediary where possible.
- Payment identifiers must be unique per hop.
- Reused route identifiers must be rejected.

Required tasks:

- Define per-hop forwarding packet.
- Add route ID derivation.
- Add replay protection for forwarding packets.
- Add logging rules that avoid storing unnecessary payment details.

## 9. Virtual Payment Channels

### 9.1 Purpose

Virtual channels allow two endpoints to transact without opening a direct on-chain channel when an existing route can reserve sufficient capacity.

Properties:

- No on-chain deployment per virtual relationship.
- Backed by one or more existing base channels.
- Uses intermediaries as liquidity connectors.
- Supports off-chain updates between virtual endpoints.
- Settles through parent channels if cooperative.
- Can fall back to on-chain dispute using parent-channel commitments.

### 9.2 Virtual Channel State

Fields:

- `virtual_channel_id`
- `parent_route_id`
- `endpoint_a`
- `endpoint_b`
- `intermediary_set_hash`
- `capacity`
- `balance_a`
- `balance_b`
- `nonce`
- `expiry_height`
- `anchor_commitment`
- `condition_root`
- `signatures`

Rules:

- Parent channels must reserve capacity before virtual channel activation.
- Virtual balances must not exceed reserved capacity.
- Virtual channel expiry must be earlier than parent channel safety timeout.
- Intermediaries receive routing or liquidity fee according to signed terms.
- Virtual channel close must release parent reserves.

### 9.3 Virtual Channel Opening

Flow:

- Endpoint requests virtual channel route.
- Routing engine selects path with sufficient liquidity.
- Each intermediary signs capacity reservation.
- Endpoints sign virtual opening state.
- Parent channels update reserve commitments.
- Virtual channel becomes active after all reservations are acknowledged.

Required tasks:

- Define reservation signature format.
- Add parent-channel reserve accounting.
- Add activation proof.
- Add virtual channel route timeout validation.

### 9.4 Virtual Channel Updates

Rules:

- Endpoints exchange signed virtual states.
- Intermediaries do not sign each endpoint-level update.
- Parent channel state only changes when virtual channel opens, closes, or disputes.
- Latest valid virtual state can be submitted during dispute.

Required tasks:

- Define virtual state signing domain.
- Add update nonce rules.
- Add dispute proof that maps virtual state to parent reserve commitments.

### 9.5 Virtual Channel Close and Dispute

Close modes:

- Cooperative endpoint close.
- Expired virtual close.
- Intermediary-triggered close due to parent channel risk.
- Disputed virtual close.

Rules:

- Cooperative close releases reserved liquidity immediately after required signatures.
- Disputed close uses latest signed virtual state.
- Parent channels remain locked for the required dispute period.
- Intermediaries can recover reserved capacity after virtual expiry and finalization.

Required tasks:

- Add virtual close proof format.
- Add parent reserve release rules.
- Add timeout hierarchy tests.
- Add nested dispute simulation tests.

### 9.6 Liquidity Aggregation

Rules:

- Virtual channel capacity may be aggregated across multiple base channels only if each segment signs explicit reservation commitments.
- Multi-path virtual channels must define deterministic split accounting.
- Settlement must be able to resolve each reserved segment independently if dispute occurs.

Required tasks:

- Define multi-segment reservation schema.
- Add split-balance accounting.
- Add settlement proof for each segment.
- Add failure handling for partial route activation.

## 10. Cosmos SDK Integration Layer

### 10.1 BlockSTM Parallel Execution

Design goals:

- Make independent channel settlement operations parallelizable.
- Minimize write conflicts.
- Batch settlement operations by disjoint channel IDs.
- Avoid global mutable state in hot paths.

State partitioning:

- Primary key: `channel_id`.
- Secondary keys:
  - Participant index.
  - Pending close index.
  - Condition index.
  - Routing advertisement index.
  - Settlement tombstone index.

Parallel-safe transaction classes:

- Open channel with unique `channel_id`.
- Update checkpoint for distinct channels.
- Close distinct channels.
- Dispute distinct channels.
- Settle distinct channels.
- Resolve conditions for distinct condition IDs.

Conflict-prone transaction classes:

- Multiple operations on same channel.
- Batch condition settlement touching shared promises.
- Penalty routing that updates shared module accounts.
- Fee bucket accounting if implemented as one global counter.

Required implementation tasks:

- Use per-channel state keys for hot-path writes.
- Defer global accounting aggregation to end-block or epoch accounting where possible.
- Use per-block temporary accumulators for fees, burns, and penalties.
- Add BlockSTM conflict profiling tests.
- Add settlement batch grouping by channel key.
- Add benchmarks for channel open, close, dispute, and batch settlement.

Acceptance criteria:

- Independent channel settlements execute without avoidable write conflicts.
- Same-channel transactions conflict deterministically.
- Global accounting does not serialize the entire settlement workload.

### 10.2 Store v2 State Layout

Store design:

- `channels/{channel_id}`:
  - Core channel record.
- `channel_states/{channel_id}/{nonce}`:
  - Optional checkpointed state commitments.
- `pending_closes/{channel_id}`:
  - Pending close state.
- `conditions/{condition_id}`:
  - Conditional promise state.
- `virtual_channels/{virtual_channel_id}`:
  - Virtual channel anchor state.
- `participant_channels/{address}/{channel_id}`:
  - Participant index.
- `settlement_tombstones/{channel_id}`:
  - Replay protection record.
- `fee_accumulators/{block_or_epoch}/{bucket}`:
  - Aggregated fee accounting.
- `fraud_proofs/{proof_id}`:
  - Accepted fraud evidence references.

Store requirements:

- Keep active channel records compact.
- Store full signed states only when submitted on-chain.
- Store hashes for off-chain states unless dispute requires full data.
- Use prefix iteration for participant channel queries.
- Use TTL or pruning policy for tombstones after replay horizon where safe.

Required implementation tasks:

- Define protobuf state types.
- Define key prefixes and migration versioning.
- Add store layout tests.
- Add pruning rules for finalized channels and expired conditions.
- Add query pagination for participant indexes.

### 10.3 AdaptiveSync Integration

Purpose:

- Allow recovering nodes to quickly reconstruct payment settlement state.
- Avoid requiring historical off-chain updates for consensus correctness.

Requirements:

- Consensus state must include all active channels, pending closes, unresolved conditions, virtual anchors, and settlement tombstones.
- Off-chain routing topology is not consensus-critical.
- Nodes can recover settlement safety from on-chain state alone.
- Watch services resync from events and channel indexes.

Required implementation tasks:

- Add state sync snapshot support for payment module stores.
- Add event replay compatibility for watchers.
- Add compact indexes for active disputes and pending finalizations.
- Add recovery test for node joining during active disputes.

### 10.4 Staking Module Integration

Use cases:

- Validator-assisted arbitration through deterministic block execution.
- Optional validator-backed watch service marketplace.
- Security reserve funded by penalties and settlement fees.
- Slashing integration for validator misbehavior remains separate from participant channel penalties.

Integration points:

- Validator set queries for arbitration service selection where needed.
- Delegation-based service discovery for validator-operated payment infrastructure.
- Community or security pool funding for critical monitoring infrastructure.
- Optional validator metadata extension for payment routing service endpoints.

Rules:

- Validators do not decide disputes subjectively.
- Arbitration is execution of deterministic settlement messages.
- Validator-operated routing services must not receive consensus privilege.
- Payment participant penalties must not be confused with validator slashing unless validator identity is directly involved in channel fraud.

Required implementation tasks:

- Define optional validator payment-service metadata.
- Add validator-assisted watch service registration if needed.
- Add clear separation between channel penalties and consensus slashing.
- Add tests for validator-submitted dispute messages.

### 10.5 Fee Module Integration

Fee classes:

- Channel open fee.
- Channel checkpoint fee.
- Cooperative close fee.
- Unilateral close fee.
- Dispute fee.
- Fraud proof verification fee.
- Conditional promise settlement fee.
- Virtual channel anchor fee.
- Routing advertisement fee.

Anti-spam requirements:

- Channel open fee must cover storage footprint and prevent mass channel creation.
- Dispute fee must discourage invalid disputes without blocking valid fraud proofs.
- Fraud proof submitter may recover verification fee when proof is accepted.
- Routing advertisement fee or deposit must deter spam topology announcements.
- Storage fee applies to active channel records, conditions, and tombstones.

Required implementation tasks:

- Add fee schedule parameters.
- Add dynamic fee multiplier hooks for congestion.
- Add storage fee integration for channel footprint.
- Add fee refund path for accepted fraud proofs.
- Add tests for fee bypass attempts.

### 10.6 Async Execution Integration

Use cases:

- Batch promise resolution.
- Deferred finalization after challenge period.
- Scheduled cleanup of expired channels and promises.
- Settlement result emission to downstream modules.

Required implementation tasks:

- Add queue for finalizable channels.
- Add queue for expired promises.
- Add async settlement completion events.
- Add idempotent handlers for retried async jobs.
- Add bounded work per block for cleanup queues.

## 11. Economic Model

### 11.1 Channel Opening Cost

Purpose:

- Prevent channel spam.
- Cover storage footprint.
- Price settlement contract state.
- Fund security monitoring and verification.

Fee components:

- Base channel open fee.
- Per-participant fee.
- Per-byte state fee.
- Conditional capability surcharge.
- Virtual-channel anchor surcharge.
- Routing advertisement deposit where applicable.

Rules:

- Open fee is paid in `naet`.
- Fee increases under channel-open congestion.
- Long-lived channels may require storage rent or periodic renewal.
- Advertisement deposit can be returned after clean channel close and gossip expiry.

Required tasks:

- Define channel open fee formula.
- Define storage footprint estimator.
- Add governance parameters for fee bounds.
- Add tests for opening many small channels.

### 11.2 Routing Micro-Fees

Fee types:

- Base hop fee.
- Proportional amount fee.
- Liquidity reservation fee.
- Virtual channel setup fee.
- Congestion surcharge.
- Failure penalty for repeated invalid attempts where locally enforceable.

Rules:

- Routing fees are agreed off-chain in signed promises.
- Intermediaries earn fees only when forwarding conditions are fulfilled.
- Fee maximum must be included in sender route policy.
- Fee policy updates must be signed and time-bounded.
- Fee claims must be enforceable only through signed channel state or conditional promise settlement.

Required tasks:

- Define fee policy message.
- Define hop fee calculation.
- Add route-level fee ceiling.
- Add tests for fee overcharge rejection.

### 11.3 Validator Incentives for Settlement Verification

Incentive sources:

- Transaction fees for settlement messages.
- Dispute fees.
- Fraud proof verification fees.
- Optional security reserve distributions for critical arbitration monitoring.

Rules:

- Validators receive normal transaction fee rewards for settlement transactions.
- No validator receives special payment for choosing dispute outcomes.
- Accepted fraud proofs can refund submitter fees and allocate remaining penalty according to configured buckets.
- Validator-operated watch services may charge off-chain fees but must not receive consensus priority.

Required tasks:

- Define settlement message gas costs.
- Add fraud proof fee refund accounting.
- Add security reserve allocation hooks if enabled.
- Add monitoring for settlement transaction inclusion latency.

### 11.4 Fraud and Invalid State Penalties

Penalty classes:

- Invalid close submission.
- Stale close with newer signed state available.
- Same-nonce double-sign.
- Invalid condition claim.
- Replay attempt.
- Async overexposure attempt.
- Invalid fraud proof submission.

Penalty sources:

- Channel balance.
- Participant channel bond if configured.
- Routing advertisement deposit.
- Fraud proof submission deposit for invalid proofs.

Penalty routing:

- Reporter reward.
- Burn allocation.
- Security reserve.
- Community pool.
- Counterparty compensation.

Rules:

- Penalties must be deterministic and bounded.
- Penalty cannot create negative balances.
- Reporter reward cannot exceed configured cap.
- Invalid fraud proof submitter pays verification cost.
- Proven fraudulent closer pays counterparty compensation before other routing buckets.

Required tasks:

- Define penalty matrix.
- Add penalty route accounting.
- Add invariant tests for non-negative balances.
- Add tests for each fraud proof category.

### 11.5 Liquidity Incentives

Incentive mechanisms:

- Routing fees.
- Liquidity reservation fees.
- Virtual channel setup fees.
- Optional liquidity availability score.
- Optional fee discounts for well-capitalized reliable routes.

Rules:

- Liquidity incentives must be earned through signed channel updates or settled promises.
- Advertised capacity is advisory unless backed by signed reservation.
- False liquidity advertisement reduces local routing score and may forfeit deposit if tied to on-chain advertisement.
- Locked liquidity must earn enough fee opportunity to justify capital cost.

Required tasks:

- Define liquidity advertisement format.
- Define signed reservation format.
- Add local liquidity scoring.
- Add on-chain advertisement deposit mechanism if enabled.

## 12. Security Model

### 12.1 Threat Model

Threats:

- Stale state close.
- Same-nonce double-sign.
- Replay across chain or channel domains.
- Invalid condition resolution.
- Preimage withholding.
- Timeout race.
- Route griefing.
- Liquidity exhaustion.
- Gossip spam.
- Channel open spam.
- State bloat through unresolved promises.
- Watch service downtime.
- Participant key compromise.
- Validator transaction censorship within bounded windows.
- Settlement batch conflict amplification.

### 12.2 Security Guarantees

Guarantees:

- Any participant can unilaterally close.
- Latest valid signed state can supersede stale close during challenge period.
- Same-nonce conflicting signatures are punishable.
- Conditional payments settle only with valid proof before timeout.
- Expired conditional payments release reserves.
- Replay across chains, channels, epochs, or finalized states is rejected.
- Locked collateral cannot be withdrawn outside settlement rules.
- On-chain state is sufficient to resolve disputes.

### 12.3 Economic Finality Guarantees

Finality requirements:

- Cooperative settlement finalizes after on-chain inclusion and successful execution.
- Unilateral settlement finalizes after challenge period.
- Conditional settlement finalizes after proof or timeout.
- Virtual channel settlement finalizes after endpoint state and parent reserve resolution.
- Penalty settlement finalizes with accepted fraud proof execution.

Challenge period sizing:

- Must exceed expected message propagation delay.
- Must allow watch service reaction.
- Must include congestion buffer.
- Must account for timeout ordering across multi-hop conditions.

### 12.4 Validator-Assisted Arbitration

Model:

- Validators include and execute settlement transactions.
- Settlement contract verifies proofs deterministically.
- Validators do not inspect off-chain intent beyond submitted signed evidence.
- Validators may operate watch or routing services outside consensus.

Security requirements:

- Dispute transactions must have fee rules that allow valid fraud proofs to be submitted during congestion.
- Critical dispute messages may use fee prioritization rules only if deterministic and parameterized.
- Validator censorship risk is mitigated by challenge period sizing and multiple submission paths.
- Settlement logic must avoid single global locks that make disputes easy to congest.

Required tasks:

- Define dispute transaction priority policy.
- Add congestion stress tests for dispute inclusion.
- Add monitoring for challenge-period near-expiry disputes.
- Add validator-operated watcher registration format if enabled.

### 12.5 Routing Network Sybil Resistance

Mechanisms:

- Channel collateral requirement.
- Channel opening fee.
- Routing advertisement deposit where enabled.
- Signed gossip messages.
- Local reputation and failure scoring.
- Liquidity proof requirement for high-value routes.
- Rate limits for topology updates.

Rules:

- Sybil resistance is economic and local-policy based.
- Routing identity alone must not grant settlement privilege.
- Nodes may ignore unbacked liquidity advertisements.
- Repeated failed route attempts reduce local score.

Required tasks:

- Add gossip rate limits.
- Add route failure scoring.
- Add liquidity proof verification.
- Add topology spam simulations.

### 12.6 Key Management Requirements

Participant signer:

- Persist highest signed nonce before releasing signature.
- Reject same-nonce divergent state.
- Support channel-specific signing limits.
- Support emergency pause for compromised keys.

Routing node signer:

- Separate routing gossip key from channel funds key where possible.
- Limit automated signing exposure.
- Use per-channel spending limits.
- Log all signed state commitments.

Required tasks:

- Define signer API.
- Add durable nonce store.
- Add key compromise close procedure.
- Add audit log format for signed states.

## 13. Advanced Modules

### 13.1 payment-channel-module

Purpose:

- Provide native channel custody, lifecycle management, dispute handling, and final settlement.

State model:

- `Channel`
- `ChannelParticipant`
- `ChannelConfig`
- `PendingClose`
- `SettlementRecord`
- `SettlementTombstone`
- `ChannelFeeAccumulator`

Message types:

- `MsgOpenChannel`
- `MsgCooperativeClose`
- `MsgUnilateralClose`
- `MsgDisputeClose`
- `MsgFinalizeClose`
- `MsgSubmitCheckpoint`
- `MsgCancelExpiredChannel`
- `MsgRegisterChannelAdvertisement`

Failure modes:

- Stale close submitted and not disputed in time.
- Same-channel transaction conflicts under parallel execution.
- Incorrect collateral accounting.
- Tombstone pruning too early enables replay.
- Challenge period too short under congestion.

Integration points:

- Bank module for collateral lock and release.
- Fee module for open, close, dispute, and storage costs.
- Distribution or fee collector for settlement fees.
- Store v2 for channel state and participant indexes.
- BlockSTM for parallel channel operation execution.
- Async execution queues for finalization and cleanup.

Implementation tasks:

- Define protobuf messages and state.
- Implement keeper methods for lifecycle transitions.
- Add ante validation for channel fees.
- Add invariant tests for collateral conservation.
- Add BlockSTM conflict benchmarks.

### 13.2 routing-engine-module

Purpose:

- Maintain off-chain topology, score routes, select paths, and manage payment retries.

State model:

- `RoutingNode`
- `ChannelEdge`
- `LiquidityHint`
- `FeePolicy`
- `RouteAttempt`
- `RouteFailure`
- `LocalPeerScore`

Message types:

- `GossipNodeAnnouncement`
- `GossipChannelAnnouncement`
- `GossipChannelUpdate`
- `GossipLiquidityHint`
- `GossipFeePolicyUpdate`
- `GossipRouteFailure`
- `CapacityProbeRequest`
- `CapacityProbeResponse`

Failure modes:

- Stale liquidity causes failed payments.
- Malicious nodes advertise false capacity.
- Gossip spam consumes bandwidth.
- Route probing leaks payment intent.
- Local score manipulation biases route selection.

Integration points:

- Payment channel module for open channel proofs.
- Conditional payments module for routed promises.
- Fee module for on-chain advertisement deposits if enabled.
- Node P2P layer for gossip transport.
- Wallet and service APIs for path requests.

Implementation tasks:

- Define gossip envelope.
- Implement topology database.
- Implement path scoring.
- Add capacity-aware Dijkstra-like search.
- Add route retry policy.
- Add gossip spam resistance.

### 13.3 conditional-payments-module

Purpose:

- Provide hash-locked and time-locked promise semantics for atomic routed payments.

State model:

- `ConditionalPromise`
- `ConditionRoot`
- `PreimageClaim`
- `PromiseLink`
- `PromiseTimeout`
- `ConditionSettlementRecord`

Message types:

- `MsgRegisterPromise`
- `MsgResolveWithPreimage`
- `MsgExpirePromise`
- `MsgBatchResolvePromises`
- `MsgDisputeCondition`
- `MsgFinalizeConditionSettlement`

Failure modes:

- Timeout ordering invalid across route.
- Preimage withheld until unsafe boundary.
- Promise resolved twice.
- Invalid condition root accepted.
- Expired promise remains reserved and locks liquidity.

Integration points:

- Payment channel module for channel reserves.
- Fee module for condition settlement fees.
- Store v2 for condition indexing.
- Async execution for expiry cleanup.
- Fraud-proof module for invalid resolution proofs.

Implementation tasks:

- Define promise schema.
- Implement hash-lock and time-lock validation.
- Add batch settlement handler.
- Add timeout hierarchy tests.
- Add invariant tests for reserved balances.

### 13.4 liquidity-optimization-module

Purpose:

- Improve liquidity placement, reservation, rebalancing, and fee policy selection.

State model:

- `LiquidityPosition`
- `Reservation`
- `RebalanceIntent`
- `FeePolicy`
- `CapacityForecast`
- `LiquidityScore`

Message types:

- `MsgAdvertiseLiquidity`
- `MsgReserveLiquidity`
- `MsgReleaseReservation`
- `MsgUpdateFeePolicy`
- `MsgSubmitRebalanceIntent`
- `MsgSetLiquidityLimits`

Failure modes:

- Over-reservation blocks usable channel capacity.
- False liquidity advertisement causes payment failures.
- Rebalancing attempts create excessive on-chain settlement load.
- Fee policy oscillation reduces routing reliability.
- Liquidity score becomes stale or manipulable.

Integration points:

- Routing engine for path scoring.
- Payment channel module for reserve accounting.
- Fee module for advertisement deposits.
- Store v2 for reservation indexes.
- Async execution for reservation expiry.

Implementation tasks:

- Define reservation accounting.
- Add reservation expiry.
- Add capacity forecast API.
- Add fee policy bounds.
- Add liquidity score decay.

### 13.5 fraud-proof-verification-module

Purpose:

- Verify signed-state fraud proofs and route penalties deterministically.

State model:

- `FraudProof`
- `EvidenceRecord`
- `PenaltyRecord`
- `ReporterReward`
- `DoubleSignEvidence`
- `ReplayEvidence`

Message types:

- `MsgSubmitStaleCloseProof`
- `MsgSubmitDoubleSignProof`
- `MsgSubmitInvalidConditionProof`
- `MsgSubmitReplayProof`
- `MsgSubmitAsyncOverexposureProof`
- `MsgClaimReporterReward`

Failure modes:

- Duplicate evidence drains reporter rewards.
- Invalid proof consumes excessive verification gas.
- Penalty exceeds available balance.
- Evidence accepted after finalization horizon.
- Same evidence represented in multiple encodings bypasses duplicate detection.

Integration points:

- Payment channel module for pending close and settlement state.
- Conditional payments module for condition evidence.
- Bank module for penalty movement.
- Fee module for proof deposits and refunds.
- Store v2 for evidence deduplication.
- Security reserve accounting.

Implementation tasks:

- Define canonical evidence hash.
- Add duplicate evidence rejection.
- Add proof gas metering.
- Add penalty accounting.
- Add reporter reward caps.
- Add fuzz tests for malformed evidence.

## 14. API Surface

### 14.1 On-Chain Messages

Required:

- `MsgOpenChannel`
- `MsgCooperativeClose`
- `MsgUnilateralClose`
- `MsgDisputeClose`
- `MsgFinalizeClose`
- `MsgSubmitCheckpoint`
- `MsgRegisterPromise`
- `MsgResolvePromise`
- `MsgExpirePromise`
- `MsgBatchResolvePromises`
- `MsgOpenVirtualChannel`
- `MsgCloseVirtualChannel`
- `MsgDisputeVirtualChannel`
- `MsgSubmitFraudProof`
- `MsgRegisterRoutingAdvertisement`

### 14.2 Queries

Required:

- `QueryChannel`
- `QueryChannelsByParticipant`
- `QueryPendingClose`
- `QueryFinalizationHeight`
- `QueryCondition`
- `QueryConditionsByChannel`
- `QueryVirtualChannel`
- `QueryChannelCapacity`
- `QueryFeeSchedule`
- `QuerySettlementTombstone`
- `QueryFraudProof`
- `QueryActiveDisputes`
- `QueryPendingFinalizations`

### 14.3 Events

Required:

- `channel_opened`
- `channel_checkpointed`
- `channel_close_started`
- `channel_disputed`
- `channel_finalized`
- `channel_settled`
- `channel_penalized`
- `promise_registered`
- `promise_resolved`
- `promise_expired`
- `virtual_channel_opened`
- `virtual_channel_closed`
- `virtual_channel_disputed`
- `fraud_proof_accepted`
- `fraud_proof_rejected`
- `routing_advertisement_registered`
- `settlement_fee_charged`

## 15. Implementation Roadmap

### Phase 0: Specification and Test Vectors

Tasks:

- Define canonical encoding for channel states, promises, deltas, and virtual channels.
- Define signature domains.
- Define settlement lifecycle state machine.
- Define fee schedule.
- Produce fraud proof test vectors.
- Produce timeout ordering test vectors.
- Produce BlockSTM conflict test plan.

Exit criteria:

- All signable objects have canonical test vectors.
- All lifecycle transitions are represented in state-machine tests.
- Collateral conservation invariants are specified.

### Phase 1: Base Channel Settlement

Tasks:

- Implement payment channel state.
- Implement channel open.
- Implement cooperative close.
- Implement unilateral close.
- Implement dispute with higher signed nonce.
- Implement final settlement.
- Implement settlement tombstones.
- Add participant channel queries.

Exit criteria:

- Bidirectional channel lifecycle works end to end.
- Unilateral close can be disputed with a newer valid state.
- Final balances conserve locked collateral.
- Closed channels reject replayed states.

### Phase 2: Fraud Proofs and Penalties

Tasks:

- Implement same-nonce double-sign proof.
- Implement stale close proof.
- Implement invalid balance proof.
- Implement replay proof.
- Implement penalty routing.
- Implement reporter reward caps.
- Add malformed proof fuzz tests.

Exit criteria:

- Fraud proofs are deterministic and gas-bounded.
- Penalty accounting cannot create negative balances.
- Duplicate evidence is rejected.

### Phase 3: Conditional Payments

Tasks:

- Implement promise object.
- Implement hash-lock resolution.
- Implement time-lock expiry.
- Implement reserved balance accounting.
- Implement batch promise resolution.
- Add timeout hierarchy validation.
- Add atomic route settlement tests.

Exit criteria:

- Multi-hop conditional payments settle atomically.
- Expired promises release reserves.
- Preimage replay is rejected.

### Phase 4: Routing Engine

Tasks:

- Implement signed gossip envelope.
- Implement topology database.
- Implement liquidity hints.
- Implement fee policies.
- Implement capacity-aware path search.
- Implement congestion-aware route scoring.
- Implement route retry policy.

Exit criteria:

- Routes are selected using capacity, fee, timeout, and failure signals.
- Stale or false liquidity reduces local route score.
- Route attempts produce structured failure data.

### Phase 5: Virtual Channels

Tasks:

- Implement virtual channel state.
- Implement parent-channel reservation accounting.
- Implement virtual open proof.
- Implement endpoint updates.
- Implement virtual close.
- Implement virtual dispute.
- Implement parent reserve release.

Exit criteria:

- Endpoints transact without direct on-chain channel.
- Parent channel reserves enforce virtual capacity.
- Virtual disputes resolve through parent commitments.

### Phase 6: Performance and Operations

Tasks:

- Add Store v2 layout benchmarks.
- Add BlockSTM settlement batch benchmarks.
- Add AdaptiveSync recovery tests.
- Add watcher event replay.
- Add cleanup queues for expired promises and finalizable channels.
- Add operational metrics and alerts.

Exit criteria:

- Independent channel operations parallelize under BlockSTM.
- Recovering nodes reconstruct active payment state from snapshots.
- Watch services can resync from events and queries.

## 16. Required Test Coverage

### 16.1 Unit Tests

- Channel ID generation.
- State hash encoding.
- Signature domain validation.
- Balance conservation.
- Nonce monotonicity.
- Cooperative close.
- Unilateral close.
- Dispute supersession.
- Final settlement.
- Tombstone replay rejection.
- Hash-lock proof validation.
- Time-lock expiry.
- Penalty routing.
- Fee calculation.

### 16.2 Integration Tests

- Bidirectional channel open, update, close, settle.
- Unidirectional streaming claim and reclaim.
- Async delta checkpoint and dispute.
- Multi-hop conditional payment.
- Virtual channel open, update, close.
- Parent channel dispute while virtual channel is active.
- Fraud proof with reporter reward.
- Fee congestion during dispute.
- Store snapshot recovery during pending close.

### 16.3 Invariant Tests

- Locked collateral equals active balances plus reserves plus pending penalties.
- Settlement never pays more than locked collateral.
- Promise reserve cannot exceed channel balance.
- Expired promises cannot be resolved.
- Same preimage cannot settle the same promise twice.
- Finalized channel cannot be reopened with same ID.
- Penalties cannot produce negative balances.
- Fee buckets sum to collected fees.
- Same-channel writes conflict deterministically.
- Distinct-channel settlements do not share hot write keys.

### 16.4 Fuzz Tests

- Malformed signed states.
- Random nonce ordering.
- Conflicting same-nonce states.
- Invalid promise links.
- Timeout boundary conditions.
- Batch settlement ordering.
- Fraud proof duplicate encodings.
- Route failure classifications.
- Async delta aggregation.

### 16.5 Performance Tests

- Channel opens per block.
- Cooperative closes per block.
- Unilateral closes per block.
- Disputes per block.
- Promise resolutions per block.
- Virtual channel disputes per block.
- BlockSTM conflict rate by transaction mix.
- Store v2 read/write latency for channel indexes.
- Snapshot recovery time with active disputes.

## 17. Observability

### 17.1 Metrics

- Active channels.
- Pending closes.
- Active disputes.
- Finalizable channels.
- Settled channels per block.
- Average channel lifetime.
- Total locked `naet`.
- Locked `naet` by channel type.
- Conditional promises active.
- Conditional promises expired.
- Conditional promises resolved.
- Virtual channels active.
- Routing advertisements active.
- Fraud proofs submitted.
- Fraud proofs accepted.
- Fraud proofs rejected.
- Penalties applied.
- Reporter rewards paid.
- Settlement fees collected.
- Channel open fee average.
- Dispute inclusion latency.
- Challenge period near-expiry count.
- BlockSTM conflict rate for payment messages.
- Store v2 payment module read latency.
- Store v2 payment module write latency.

### 17.2 Alerts

- High pending dispute count.
- Challenge period near expiry without finalization.
- Fraud proof rejection spike.
- Channel open spam spike.
- Promise expiry backlog.
- Settlement queue backlog.
- BlockSTM conflict rate above threshold.
- Payment module store latency above threshold.
- Watch service event replay lag.
- Routing gossip spam rate above threshold.

### 17.3 Reports

- Daily locked liquidity report.
- Daily settlement volume report.
- Daily routing fee report.
- Daily dispute and fraud report.
- Daily state footprint report.
- Weekly channel churn report.
- Weekly liquidity concentration report.
- Weekly performance report.

## 18. Governance Parameters

### 18.1 Channel Parameters

- Minimum channel collateral.
- Maximum channel collateral.
- Minimum challenge period.
- Maximum challenge period.
- Default challenge period.
- Minimum close delay.
- Maximum close delay.
- Channel open base fee.
- Channel storage fee per byte.
- Channel tombstone retention.

### 18.2 Conditional Payment Parameters

- Maximum active promises per channel.
- Maximum promise amount ratio per channel.
- Minimum timeout margin.
- Maximum promise lifetime.
- Batch resolution maximum size.
- Promise storage fee.
- Expired promise cleanup limit per block.

### 18.3 Virtual Channel Parameters

- Maximum virtual channels per parent channel.
- Maximum virtual channel depth.
- Minimum parent timeout margin.
- Virtual channel anchor fee.
- Virtual channel reservation expiry.
- Multi-segment virtual channel maximum segments.

### 18.4 Fraud and Penalty Parameters

- Stale close penalty.
- Same-nonce double-sign penalty.
- Invalid condition penalty.
- Replay attempt penalty.
- Invalid fraud proof deposit.
- Reporter reward percentage.
- Reporter reward cap.
- Penalty burn allocation.
- Security reserve allocation.
- Counterparty compensation priority.

### 18.5 Routing Parameters

- Routing advertisement deposit.
- Gossip message expiry.
- Liquidity hint expiry.
- Maximum topology updates per peer per window.
- Route failure score decay.
- Congestion penalty decay.
- Capacity probe rate limit.

### 18.6 Execution Parameters

- Settlement batch maximum size.
- Finalization queue work limit per block.
- Expired promise cleanup work limit per block.
- Fee multiplier under channel-open congestion.
- Fee multiplier under dispute congestion.
- Store pruning horizon.

## 19. Engineering Backlog

High priority:

- Create `PAYMENTS.md` as locally ignored internal document.
- Define canonical channel state encoding.
- Define signature domains and replay protection fields.
- Define bidirectional and unidirectional channel state machines.
- Implement collateral conservation invariants.
- Implement channel open, unilateral close, dispute, and finalize flows.
- Implement Store v2 key layout.
- Add BlockSTM conflict analysis for settlement messages.
- Add fee schedule for open, close, dispute, and fraud proof messages.

Medium priority:

- Implement hash-locked conditional promises.
- Implement time-lock expiry.
- Implement penalty matrix and reporter reward caps.
- Implement watcher event stream.
- Implement routing gossip envelope.
- Implement capacity-aware path search.
- Implement virtual channel reservation schema.
- Add AdaptiveSync recovery tests.

Lower priority:

- Implement async delta channels.
- Implement liquidity optimization module.
- Implement on-chain routing advertisement deposits.
- Implement virtual channel multi-segment aggregation.
- Implement validator-operated watch service marketplace.
- Implement advanced route privacy packetization.

## 20. Acceptance Criteria

The payment layer is acceptable for initial production hardening when:

- Bidirectional and unidirectional channels can settle trustlessly.
- Any participant can close unilaterally.
- Stale closes can be disputed with newer signed states.
- Same-nonce conflicting signatures are slashable or penalizable.
- Conditional payments support hash-lock and time-lock settlement.
- Multi-hop payments preserve atomic settlement semantics.
- Virtual channels can be backed by parent-channel reservations.
- Store v2 state layout is compact and queryable.
- Independent settlement transactions are parallelizable under BlockSTM.
- AdaptiveSync recovery restores all consensus-critical payment state.
- Fees cover channel storage and dispute verification costs.
- Fraud proofs are deterministic, bounded, and test-covered.
- Replay protection covers chain, channel, nonce, epoch, and finalization domains.
- Observability covers liquidity, settlement, disputes, fees, and performance.
