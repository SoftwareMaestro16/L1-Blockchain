# Aetheris Service and Application Layer Specification

Status: Internal design document
Scope: Service-aware application runtime, registry, discovery, payment, storage, and execution abstraction
Visibility: Private, not for public repository inclusion

## 1. Mission

Design and implement a next-generation Service and Application Layer for Aetheris on Cosmos SDK `v0.54+`.

The system must support:

- On-chain applications.
- Off-chain services.
- Mixed trust services.
- Decentralized fog marketplaces.
- Unified service discovery.
- Formal service interfaces.
- Execution abstraction across chain modules, contracts, and network services.
- Proof-backed or economically constrained off-chain execution.
- Upgrade-safe integration with ABCI++, Store v2, modular state transition functions, and BlockSTM-ready execution.

## Implementation Update: 2026-06-06

Implemented in `x/aethercore/types`:

- Extended `ServiceDescriptor` into a first-class registry object covering owner, service type, status, expiry, formal interface, execution target, discovery metadata, payment policy, storage policy, and verification policy.
- Added runtime enums for on-chain, off-chain, mixed, and fog market services; module, contract, application-zone, external, hybrid, and provider-pool locations; trust, verification, payment, storage, method execution, and failure models.
- Added canonical `ServiceInterfaceDescriptor` and `ServiceMethodDescriptor` hashing so interface hashes commit to methods, schema hashes, auth model, payment model, events, errors, and metadata.
- Extended service descriptor hashing and service root construction to commit to interface, discovery, payment, storage, execution, and verification fields.
- Added validation rules for deterministic on-chain services, signed or proof/economic off-chain services, challengeable mixed services, and provider-collateral-backed fog market services.
- Added clone/canonicalization coverage for nested service interface slices so exported state and lookup results remain deterministic and mutation-safe.

Verification added:

- Unit coverage for registering on-chain, off-chain, mixed, and fog market descriptors.
- Reject coverage for interface hash mismatch, unsafe off-chain descriptors, and mixed services without challenge or fallback.
- Clone isolation coverage for `ServiceByID`.
- Targeted test command: `go test ./x/aethercore/types`.

Implemented for Section 3 execution pipeline:

- Added a deterministic `ServiceConsensusContext`, `ServiceCallEnvelope`, `ServiceProposalPlan`, execution groups, outcomes, service call receipts, finalization roots, and end-block maintenance records.
- Added `PrepareServiceProposal` to classify service calls, enforce active descriptors/interface hashes/payment expectations, prioritize disputes, callbacks/retries, expiring calls, paid calls, and group calls by canonical state-key conflict scope.
- Added `ProcessServiceProposal`/`ValidateServiceProposalPlan` to verify proposal roots, grouping constraints, call envelopes, same-sender nonce ordering, expired-call rejection, and BlockSTM write-conflict separation.
- Added `FinalizeServiceProposal` to validate on-chain, off-chain receipt, mixed dispute, and mixed settlement outcomes, then commit deterministic service receipt roots alongside service registry and interface roots.
- Added `EndBlockServiceMaintenance` for bounded service expiry/cleanup, receipt-derived provider reputation deltas, and deterministic service metrics.
- Added `ServiceStateTransition` validation so service STFs reject external network calls, wall-clock use, unbounded iteration, unmetered proof verification, and direct cross-zone writes.

Additional verification added:

- Unit coverage for Prepare/Process/Finalize/EndBlock flow.
- Reject coverage for expired calls, missing idempotency keys, missing methods, external STF calls, wall-clock STF use, unmetered proof verification, and direct cross-zone writes.

## 2. Core Principle: Service-Aware Blockchain

### 2.1 Service Definition

Every executable or discoverable entity is represented as a service.

```text
Service {
  identity
  execution_target
  interface_schema
  location
  trust_model
  payment_model
  verification_model
}
```

### 2.2 Service Rules

- Services are first-class protocol objects.
- Services must have stable identifiers.
- Services must expose formal interface schemas.
- Services must define execution location.
- Services must define trust and verification model.
- Services must define payment expectations.
- On-chain service execution must be deterministic.
- Off-chain service execution must be signed, proof-anchored, challengeable, or economically constrained.
- Mixed services must define on-chain fallback, dispute, or settlement logic.

### 2.3 Service Object Model

Required fields:

- `service_id`
- `owner`
- `service_type`
- `identity_binding_optional`
- `execution_target`
- `interface_hash`
- `location`
- `trust_model`
- `payment_model`
- `verification_model`
- `status`
- `version`
- `expiry_height`
- `created_height`
- `updated_height`

Service types:

- `on_chain`
- `off_chain`
- `mixed`
- `fog_market`

Service locations:

- `module`
- `contract`
- `application_zone`
- `external_endpoint`
- `hybrid_endpoint`
- `provider_pool`

## 3. Cosmos SDK Execution Model

### 3.1 Base Execution Pipeline

Service-aware execution must align with:

- `PrepareProposal`
- `ProcessProposal`
- `FinalizeBlock`
- `DeliverTx` compatibility path
- `EndBlock`
- Store v2 branching writes
- BlockSTM-compatible conflict handling
- Modular STF separation

### 3.2 Pipeline Responsibilities

`PrepareProposal`:

- Classify service transactions by target service.
- Group transactions by independent state keys.
- Prioritize expiring calls, paid calls, and dispute calls by deterministic policy.
- Include service callbacks and retry messages where eligible.

`ProcessProposal`:

- Verify proposal grouping constraints.
- Verify service call envelopes.
- Verify deterministic ordering for same-sender calls.
- Reject malformed or expired calls.

`FinalizeBlock`:

- Execute on-chain service calls.
- Anchor off-chain service receipts.
- Process mixed service dispute and settlement messages.
- Emit deterministic service receipts.
- Commit service registry and interface roots.

`EndBlock`:

- Process bounded service expiry.
- Process bounded registry cleanup.
- Update reputation from accepted receipts where consensus-safe.
- Aggregate service metrics.

### 3.3 STF Separation

Every service execution path must be expressible as:

```text
next_state = service_stf(current_state, call, consensus_context)
```

Constraints:

- No network calls inside STF.
- No non-deterministic clocks.
- No live service availability checks.
- No unbounded iteration.
- No unmetered proof verification.
- No direct cross-zone writes.

## 4. Service Types

### 4.1 On-Chain Service

Definition:

- Fully executed inside the Aetheris state machine.
- Deterministic.
- Gas metered.
- Stored in Store v2-backed KV or object store layout.
- Produces committed receipts.

Examples:

- DEX module.
- Tokenfactory.
- Governance execution module.
- Payment settlement module.
- Identity resolver module.

Required fields:

- `module_name` or `contract_address`
- `method_set`
- `gas_model`
- `state_root_type`
- `receipt_policy`
- `failure_policy`

Implementation tasks:

- Define service wrapper for existing modules.
- Add service descriptor generation for module MsgServers.
- Add service receipts for module execution.
- Add interface schemas for module methods.
- Add proof query for service state.

### 4.2 Off-Chain Service

Definition:

- Executed outside consensus.
- Accessed through service network, RPC endpoint, overlay, or provider mesh.
- May produce signed results.
- May anchor commitments or receipts on-chain.

Required fields:

- `endpoint`
- `endpoint_type`
- `provider_key`
- `request_signing_policy`
- `response_signing_policy`
- `proof_anchor_policy`
- `availability_policy`
- `result_expiry`

Validation requirements:

- Service descriptor must include interface hash.
- Endpoint metadata must be signed by service owner or authorized provider.
- Requests must include replay-safe nonce or idempotency key.
- Responses must be signed when used for settlement or challenge.
- Optional proof anchor must bind request hash, response hash, provider key, and height.

Implementation tasks:

- Define off-chain service descriptor schema.
- Define request and response signing formats.
- Add optional receipt anchoring message.
- Add endpoint expiry and renewal.
- Add signed advertisement validation.

### 4.3 Mixed Service

Definition:

- On-chain state stores commitments, payment rules, dispute rules, and settlement state.
- Off-chain execution handles computation, storage, indexing, or heavy workflows.
- Results are verifiable, challengeable, or economically constrained.

On-chain responsibilities:

- Commitment anchoring.
- Payment escrow.
- Provider collateral.
- Result receipt.
- Challenge window.
- Fallback execution rule.
- Settlement and penalty.

Off-chain responsibilities:

- Computation.
- Storage retrieval.
- Network serving.
- Result signing.
- Proof production where applicable.

Rules:

- Off-chain result must include request commitment.
- On-chain receipt must include result commitment.
- Challenge period must be explicit.
- Fallback path must be deterministic.
- Provider collateral must cover configured fault class.

Implementation tasks:

- Define mixed service state machine.
- Add result anchoring flow.
- Add challenge and dispute messages.
- Add recompute or proof verification hooks.
- Add settlement and penalty routes.

### 4.4 Fog Market Service

Definition:

- Marketplace-backed service operated by distributed providers.
- Supports compute, storage, routing, indexing, execution, and availability services.

Provider requirements:

- Registered provider identity.
- Pricing model.
- Reputation score.
- Collateral or stake.
- Availability commitment.
- Supported service interfaces.
- Expiry and renewal.

Fog service categories:

- `compute_provider`
- `storage_provider`
- `routing_provider`
- `execution_provider`
- `indexing_provider`
- `availability_provider`

Implementation tasks:

- Define provider registry.
- Define collateral requirements.
- Define provider pricing schema.
- Define availability commitment format.
- Add reputation update rules.
- Add provider selection policy.
- Add dispute and slashing logic for provable faults.

## 5. Service Registry Layer

### 5.1 Registry Purpose

The Service Registry provides deterministic service lookup and lifecycle management.

It must support:

- On-chain registry mode.
- Hybrid registry mode.
- Distributed registry mesh mode.
- Versioned interface storage.
- Proof-backed descriptor lookup.
- Identity binding.
- Payment model discovery.
- Trust and verification model discovery.

### 5.2 ServiceRegistry Record

```text
ServiceRegistry {
  service_id
  owner
  type
  interface_hash
  endpoint
  execution_mode
  payment_model
  stake
  collateral
  reputation
  expiry
}
```

Extended fields:

- `service_name_optional`
- `identity_name_optional`
- `zone_id_optional`
- `contract_address_optional`
- `module_route_optional`
- `provider_set_optional`
- `verification_model`
- `trust_model`
- `status`
- `version`
- `descriptor_hash`
- `metadata_hash`
- `created_height`
- `updated_height`

### 5.3 Registry Modes

#### 5.3.1 On-Chain Registry

Properties:

- Fully stored in Store v2.
- Fully deterministic.
- Proof-verifiable.
- Higher storage cost.
- Used for critical services and settlement-critical descriptors.

Required state:

- Full `ServiceDescriptor`.
- Interface descriptor hash.
- Payment model.
- Verification model.
- Owner authorization.
- Expiry.

#### 5.3.2 Hybrid Registry

Properties:

- Default mode.
- Minimal on-chain anchor.
- Off-chain signed provider advertisements.
- On-chain commitment to descriptor hash and interface hash.
- Lower state footprint.

Required state:

- `service_id`
- `owner`
- `descriptor_hash`
- `interface_hash`
- `provider_root`
- `expiry_height`
- `verification_model`

#### 5.3.3 Distributed Registry Mesh

Properties:

- Gossip-based provider discovery.
- Signed advertisements.
- Optional indexing nodes.
- Proof-attached lookups when on-chain anchors exist.
- Local reputation and cache.

Rules:

- Mesh data is advisory unless anchored or signed.
- Mesh lookup must not directly affect consensus.
- Cached records require expiry height.
- Clients must verify signatures and descriptor hashes.

### 5.4 Registry State

State keys:

- `services/descriptors/{service_id}` -> `ServiceDescriptor`
- `services/anchors/{service_id}` -> `ServiceAnchor`
- `services/interfaces/{interface_hash}` -> `ServiceInterface`
- `services/owners/{owner}/{service_id}` -> `service_id`
- `services/names/{service_name}` -> `service_id`
- `services/identity_bindings/{identity_name}/{service_id}` -> `IdentityServiceBinding`
- `services/providers/{service_id}/{provider_id}` -> `ProviderRecord`
- `services/expiry/{expiry_height}/{service_id}` -> `service_id`
- `services/reputation/{provider_id}` -> `ReputationRecord`
- `services/receipts/{service_id}/{call_id}` -> `ServiceReceipt`

### 5.5 Registry Messages

- `MsgRegisterService`
- `MsgUpdateService`
- `MsgRenewService`
- `MsgDisableService`
- `MsgTransferService`
- `MsgBindServiceIdentity`
- `MsgUnbindServiceIdentity`
- `MsgRegisterProvider`
- `MsgUpdateProvider`
- `MsgStakeProviderCollateral`
- `MsgUnstakeProviderCollateral`
- `MsgAnchorServiceReceipt`
- `MsgSubmitServiceDispute`

### 5.6 Registry Queries

- `QueryService`
- `QueryServiceByName`
- `QueryServicesByOwner`
- `QueryServicesByIdentity`
- `QueryProvidersByService`
- `QueryServiceInterface`
- `QueryServicePaymentModel`
- `QueryServiceVerificationModel`
- `QueryServiceReceipt`
- `QueryServiceProof`
- `QueryServiceParams`

### 5.7 Registry Implementation Tasks

- Implement `x/services`.
- Define descriptor, anchor, provider, reputation, and receipt protobuf types.
- Add MsgServer and QueryServer.
- Add keeper with isolated service store prefix.
- Add registration, update, renewal, and disable flows.
- Add identity binding hooks.
- Add service proof query.
- Add registry export and import.
- Add invariants for descriptor and interface hash consistency.

## 6. Service Interface System

### 6.1 Purpose

Every service exposes a formal interface schema that unifies:

- On-chain module calls.
- Contract calls.
- Off-chain RPC calls.
- Mixed execution workflows.
- Event streams.

Clients use the same interface discovery and call preparation flow across service locations.

### 6.2 ServiceInterface

```text
ServiceInterface {
  methods: [
    {
      name
      input_schema
      output_schema
      execution_type
      gas_model
      verification_model
    }
  ]
}
```

Extended fields:

- `interface_hash`
- `interface_name`
- `version`
- `methods`
- `events`
- `errors`
- `auth_model`
- `payment_requirements`
- `schema_encoding`
- `metadata_hash_optional`
- `created_height`

Execution types:

- `sync`
- `async`
- `evented`

### 6.3 ServiceMethod

Fields:

- `method_id`
- `name`
- `input_schema_hash`
- `output_schema_hash`
- `execution_type`
- `required_payment_model`
- `gas_model_optional`
- `verification_model`
- `timeout_policy`
- `idempotency_required`
- `callback_supported`

### 6.4 Interface Validation

Rules:

- Interface hash must commit to all method, event, error, auth, and payment schema fields.
- Method names must be unique within interface version.
- Schema encoding must be supported.
- On-chain methods must include gas model.
- Off-chain methods must include response verification model.
- Mixed methods must include dispute or fallback model.
- Interface version changes must create a new interface hash.

### 6.5 Interface-Driven UX

Client flow:

```text
user_input
  -> resolve_service()
  -> fetch_interface()
  -> verify_interface_hash()
  -> generate_form()
  -> build_call()
  -> execute_call()
  -> verify_receipt()
```

Rules:

- UI generation is client behavior, not consensus behavior.
- Client must verify interface hash before use.
- Client must display payment and trust model.
- Client must not treat interface metadata as authorization.
- Clients must support explicit user confirmation for signing.

### 6.6 Interface Implementation Tasks

- Define interface schema format.
- Add interface registration and update messages.
- Add interface proof query.
- Add SDK interface verifier.
- Add method-level call builder.
- Add wallet and CLI schema consumption format.
- Add compatibility checks for versioned interfaces.

## 7. Unified Call Model

### 7.1 Call Object

All interactions use a common call envelope:

```text
Call {
  target_service
  method
  payload
  payment
  proof_requirement
  timeout
  signature
}
```

Extended fields:

- `call_id`
- `caller`
- `interface_hash`
- `method_id`
- `execution_location`
- `idempotency_key`
- `callback_target_optional`
- `max_fee`
- `created_height`
- `deadline_height`
- `nonce`

### 7.2 Call Routing

Routing rules:

- On-chain service calls route to `DeliverTx` or `FinalizeBlock` message execution.
- Off-chain service calls route to the service network.
- Mixed service calls split into off-chain execution and on-chain commitment, payment, or dispute flow.
- Calls with payments reserve funds before execution where required.
- Calls with proof requirements must verify result proof before final acceptance.

### 7.3 Call Receipts

`ServiceReceipt` fields:

- `call_id`
- `service_id`
- `method_id`
- `caller`
- `status`
- `request_hash`
- `response_hash_optional`
- `proof_hash_optional`
- `payment_status`
- `gas_used_optional`
- `provider_id_optional`
- `executed_height_optional`
- `anchored_height_optional`
- `error_code_optional`
- `receipt_hash`

Statuses:

- `accepted`
- `executed`
- `failed`
- `expired`
- `challenged`
- `settled`
- `reverted`

### 7.4 Idempotency and Replay Protection

Rules:

- `call_id = H(chain_id, service_id, caller, nonce, idempotency_key, payload_hash)`.
- Same caller and nonce cannot be reused for same target service.
- Idempotency key is required for retries.
- Retried calls must reference original call ID.
- Receipt tombstones must remain through configured proof horizon.

### 7.5 Call Implementation Tasks

- Define canonical call envelope.
- Add call ID derivation.
- Add service call ante validation.
- Add receipt anchoring.
- Add retry rules.
- Add idempotency key validation.
- Add call proof query.
- Add SDK call builder.

## 8. Service Discovery Layer

### 8.1 Service Resolver

Resolver function:

```text
resolve(service_name) -> ServiceDescriptor
```

Resolution sources:

- On-chain registry.
- Off-chain index nodes.
- Cached signed records.
- `.aet` identity records.
- Distributed service mesh.

### 8.2 Resolution Output

Resolution must return:

- Endpoint.
- Interface hash.
- Interface descriptor.
- Trust metadata.
- Payment expectations.
- Verification model.
- Expiry height.
- Proof or signature chain.

### 8.3 ServiceDescriptor

Fields:

- `service_id`
- `service_name`
- `owner`
- `service_type`
- `endpoint`
- `interface_hash`
- `trust_model`
- `payment_model`
- `verification_model`
- `status`
- `expiry_height`
- `proof_optional`
- `signature_optional`

### 8.4 `.aet` Integration

Rules:

- `.aet` identities can bind one or more services.
- Service bindings must include `service_id`, `interface_hash`, and target endpoint.
- Identity proof must verify domain ownership and resolver record.
- Service proof must verify registry descriptor.
- Client must reject identity-service binding if either proof fails.

### 8.5 Discovery Cache

Cache record:

- `service_id`
- `descriptor_hash`
- `interface_hash`
- `source`
- `proof_height_optional`
- `signature_optional`
- `expires_height`
- `fetched_at_height`

Rules:

- Cache cannot outlive registry expiry.
- Cache cannot outlive interface version compatibility.
- Cache invalidates on service update events.
- Unverified cache entries must be marked as advisory.

### 8.6 Discovery Implementation Tasks

- Add service name index.
- Add identity binding index.
- Add discovery query that returns descriptor and proof.
- Define signed service advertisement format.
- Add cache metadata format for SDKs.
- Add resolver fallback order.
- Add tests for stale and forged discovery records.

## 9. Payment Abstraction Layer

### 9.1 PaymentEnvelope

```text
PaymentEnvelope {
  asset
  amount
  settlement_mode
}
```

Extended fields:

- `payer`
- `payee_service`
- `denom`
- `amount`
- `max_amount_optional`
- `pricing_unit`
- `settlement_mode`
- `escrow_id_optional`
- `stream_id_optional`
- `meter_id_optional`
- `expiry_height`

Settlement modes:

- `on_chain`
- `streaming`
- `prepaid`
- `metered`
- `escrow`

### 9.2 Payment Models

Per-call:

- Fixed amount paid per accepted call.
- Payment can settle before or after execution depending on trust model.

Per-byte:

- Amount is calculated from request bytes, response bytes, or storage bytes.
- Requires deterministic metering record.

Per-compute unit:

- Amount is calculated from committed compute units.
- Requires signed or proof-backed usage receipt.

Subscription:

- Access is valid for a height or time range.
- Requires subscription state or proof-backed entitlement.

Escrow-based:

- Funds lock before execution.
- Settlement occurs after receipt, proof, or challenge window.

### 9.3 Payment State

State keys:

- `services/payments/models/{service_id}` -> `PaymentModel`
- `services/payments/escrow/{escrow_id}` -> `ServiceEscrow`
- `services/payments/streams/{stream_id}` -> `PaymentStream`
- `services/payments/meters/{meter_id}` -> `MeteredUsage`
- `services/payments/settlements/{call_id}` -> `PaymentSettlement`

### 9.4 Payment Rules

- Payment must be denominated in supported asset for service.
- `naet` must be supported for protocol-native services.
- Payment model must be known before call signing.
- Escrow settlement must be deterministic.
- Metered usage must be backed by signed or proof-anchored usage receipt.
- Failed calls follow service failure policy.

### 9.5 Payment Implementation Tasks

- Define payment model schema.
- Add service payment model query.
- Add escrow creation and settlement flow.
- Add streaming payment state.
- Add metered usage receipt format.
- Add payment proof query.
- Add integration with Financial Zone or bank keeper.

## 10. Storage Abstraction

### 10.1 StorageModel

Services can declare storage model:

- `ephemeral`
- `persistent_on_chain`
- `distributed_off_chain`
- `hybrid`

### 10.2 Storage Declaration

Fields:

- `storage_model`
- `content_hash_optional`
- `state_root_optional`
- `retrieval_method`
- `verification_method`
- `retention_policy`
- `access_policy`
- `max_payload_bytes`

### 10.3 Hybrid Storage Rule

Hybrid storage stores:

- On-chain commitment.
- Off-chain content location.
- Retrieval method.
- Verification method.
- Optional access receipt.

Consensus state must not require retrieving the off-chain payload to validate unrelated transactions.

### 10.4 Storage Receipts

Fields:

- `receipt_id`
- `service_id`
- `object_id`
- `request_hash`
- `content_hash`
- `provider_id`
- `access_height`
- `signature`
- `proof_optional`

### 10.5 Storage Implementation Tasks

- Define storage model enum.
- Add service storage declaration to descriptor.
- Add content hash and state root fields.
- Add retrieval method schema.
- Add storage receipt anchoring.
- Add proof verification hooks.
- Add storage fee model for on-chain payloads.

## 11. Message and Execution Unification

### 11.1 Unified Interaction Classes

The service layer unifies:

- On-chain transactions.
- Off-chain service calls.
- Hybrid execution flows.
- Async callbacks.
- Retries.
- Evented subscriptions.

### 11.2 Callback Model

Callback fields:

- `callback_target`
- `callback_method`
- `callback_payload_hash`
- `callback_deadline`
- `callback_payment_policy`

Rules:

- Callback must reference original call ID.
- Callback must satisfy target service interface.
- Callback must be replay-safe.
- Callback execution emits receipt.

### 11.3 Retry Model

Rules:

- Retries require idempotency key.
- Retry count is bounded.
- Retry deadline is bounded.
- Retry cannot double-charge unless payment model explicitly charges attempts.
- Retry receipts must reference original call ID.

### 11.4 Deterministic Receipts

Receipt roots:

- `service_receipts_root`
- `call_receipts_root`
- `payment_receipts_root`
- `storage_receipts_root`

Rules:

- Every on-chain service call emits receipt.
- Every anchored off-chain result emits receipt.
- Every mixed-service settlement emits receipt.
- Receipt encoding is canonical.

## 12. Security Model

### 12.1 Trust Models

Service trust model values:

- `fully_trusted`
- `economically_secured`
- `cryptographically_verifiable`
- `consensus_executed`
- `hybrid_challengeable`

Rules:

- `fully_trusted` services cannot be used for consensus-critical state transitions unless result is independently verified.
- `economically_secured` services require collateral and penalty rules.
- `cryptographically_verifiable` services require proof format.
- `consensus_executed` services must be deterministic and gas metered.
- `hybrid_challengeable` services require challenge period and fallback rule.

### 12.2 Failure Behavior

Failure behaviors:

- `revert`
- `retry`
- `fallback_on_chain`
- `challenge`
- `slash_provider`
- `refund`
- `partial_settle`

Every service method must declare failure behavior.

### 12.3 Replay Protection

Controls:

- Caller nonce.
- Service ID.
- Method ID.
- Idempotency key.
- Payload hash.
- Deadline height.
- Receipt tombstone.

### 12.4 Provider Misbehavior

Fault classes:

- Invalid result.
- Missing result.
- Late result.
- Double response.
- Wrong interface version.
- Invalid proof.
- Availability violation.

Penalty sources:

- Provider collateral.
- Service stake.
- Escrowed payment.
- Reputation score.

### 12.5 Security Implementation Tasks

- Add trust model enum.
- Add failure behavior enum.
- Add provider collateral keeper.
- Add challenge message flow.
- Add fault proof schema.
- Add penalty routing.
- Add replay tombstones.
- Add tests for invalid proofs, duplicate calls, and stale receipts.

## 13. Execution Guarantees

### 13.1 Guaranteed Properties

- On-chain paths execute deterministically.
- Hybrid service results are verifiable or challengeable.
- Off-chain service calls are signed and replay-protected.
- Provider misbehavior has economic penalties where declared.
- Service calls have deterministic receipts when anchored.
- Payment rules are known before call signing.
- Interface hash is known before call construction.

### 13.2 Non-Guaranteed Properties

- Off-chain endpoint availability is not guaranteed by consensus unless backed by explicit availability commitment.
- UI generation correctness is a client responsibility.
- Off-chain service result correctness is not assumed unless proof-verified or unchallenged after challenge period.
- Cached discovery data is not authoritative without proof or valid signature.

## 14. Cosmos SDK Integration

### 14.1 Required Modules

- `x/services`
- `x/serviceinterface`
- `x/servicecalls`
- `x/servicepayments`
- `x/serviceproviders`
- `x/servicereceipts`

Optional integration modules:

- `x/identity` for `.aet` service binding.
- `x/storage` for hybrid storage commitments.
- `x/routing` for service network routing.
- `x/payments` for streaming and escrow settlement.
- `x/contracts` or `x/avm` for contract-backed services.

### 14.2 Keeper Architecture

Keeper boundaries:

- `ServicesKeeper`
  - descriptors, anchors, lifecycle.
- `InterfaceKeeper`
  - schemas, methods, interface proofs.
- `CallKeeper`
  - call envelopes, nonces, idempotency, receipts.
- `PaymentKeeper`
  - service payment models, escrow, stream, metering.
- `ProviderKeeper`
  - provider registry, stake, collateral, reputation.
- `ReceiptKeeper`
  - receipt roots, tombstones, proof queries.

Rules:

- Keepers use isolated Store v2 prefixes.
- Cross-zone operations use messages.
- Payment settlement uses bank or Financial Zone integration.
- Identity binding uses explicit `x/identity` authorization.
- Contract service execution uses contract module interface.

### 14.3 Store v2 Layout

Primary prefixes:

- `services/descriptor/*`
- `services/anchor/*`
- `services/interface/*`
- `services/call/*`
- `services/receipt/*`
- `services/provider/*`
- `services/payment/*`
- `services/index/*`
- `services/params`

Performance rules:

- Direct lookup by `service_id` must be one primary read.
- Interface lookup by `interface_hash` must be one primary read.
- Owner, identity, provider, and method indexes must be prefix-queryable.
- Receipts must be height-indexed for pruning.
- Large schemas should be committed by hash and stored only when needed on-chain.

### 14.4 BlockSTM Strategy

Parallel-safe operations:

- Register different services.
- Update different services.
- Anchor receipts for different calls.
- Update providers with disjoint provider IDs.
- Execute on-chain service calls with disjoint state keys.
- Settle payments for different escrow IDs.

Conflict-prone operations:

- Multiple updates to same service descriptor.
- Calls that write same service-local state.
- Provider collateral update concurrent with slashing.
- Payment settlement concurrent with dispute for same call.
- Interface update concurrent with service descriptor update.

Implementation tasks:

- Add expected version fields to descriptor updates.
- Partition receipts by `call_id`.
- Partition provider state by `provider_id`.
- Partition payment state by `escrow_id` or `stream_id`.
- Add conflict tests for high-volume service calls.

### 14.5 Upgrade Strategy

Upgrade requirements:

- Service versioning.
- Interface versioning.
- Backward-compatible schemas.
- Runtime upgrade of registry rules through governance.
- Descriptor migration handlers.
- Interface deprecation markers.
- Provider re-registration rules.

Implementation tasks:

- Add `version` to all descriptor objects.
- Add schema compatibility metadata.
- Add migration handlers for registry versions.
- Add interface deprecation flow.
- Add upgrade simulation tests.

## 15. Module Breakdown

### 15.1 x/services

Purpose:

- Own service descriptors, anchors, lifecycle, identity binding, and registry queries.

State model:

- `ServiceDescriptor`
- `ServiceAnchor`
- `IdentityServiceBinding`
- `ServiceStatus`
- `ServiceParams`

Messages:

- `MsgRegisterService`
- `MsgUpdateService`
- `MsgRenewService`
- `MsgDisableService`
- `MsgTransferService`
- `MsgBindServiceIdentity`
- `MsgUnbindServiceIdentity`

Queries:

- `QueryService`
- `QueryServiceByName`
- `QueryServicesByOwner`
- `QueryServicesByIdentity`
- `QueryServiceProof`

Failure modes:

- Duplicate service ID.
- Unauthorized descriptor update.
- Expired descriptor used for call.
- Interface hash mismatch.
- Identity binding stale after domain transfer.

Integration points:

- `x/identity`
- `x/serviceinterface`
- `x/servicecalls`
- Store v2 proof queries.

### 15.2 x/serviceinterface

Purpose:

- Store and verify formal service interface schemas.

State model:

- `ServiceInterface`
- `ServiceMethod`
- `ServiceEvent`
- `ServiceError`
- `InterfaceVersion`

Messages:

- `MsgRegisterInterface`
- `MsgUpdateInterface`
- `MsgDeprecateInterface`

Queries:

- `QueryInterface`
- `QueryMethod`
- `QueryInterfaceProof`
- `QueryInterfacesByOwner`

Failure modes:

- Schema hash mismatch.
- Method ID collision.
- Unsupported schema encoding.
- Breaking update without new interface hash.

Integration points:

- `x/services`
- wallet SDK
- CLI
- contract adapter

### 15.3 x/servicecalls

Purpose:

- Manage call envelopes, idempotency, callbacks, retries, and execution receipts.

State model:

- `ServiceCall`
- `CallNonce`
- `IdempotencyRecord`
- `CallbackRecord`
- `CallReceipt`

Messages:

- `MsgSubmitServiceCall`
- `MsgAnchorServiceResult`
- `MsgRetryServiceCall`
- `MsgSubmitCallback`
- `MsgExpireServiceCall`

Queries:

- `QueryServiceCall`
- `QueryCallReceipt`
- `QueryCallsByCaller`
- `QueryCallProof`

Failure modes:

- Nonce replay.
- Duplicate idempotency key misuse.
- Callback mismatch.
- Expired call anchored late.
- Result hash mismatch.

Integration points:

- `x/services`
- `x/servicepayments`
- `x/servicereceipts`
- ABCI++ proposal handling

### 15.4 x/servicepayments

Purpose:

- Define service payment models and settle service payments.

State model:

- `PaymentModel`
- `PaymentEnvelope`
- `ServiceEscrow`
- `PaymentStream`
- `MeteredUsage`
- `PaymentSettlement`

Messages:

- `MsgSetServicePaymentModel`
- `MsgCreateServiceEscrow`
- `MsgSettleServiceEscrow`
- `MsgOpenPaymentStream`
- `MsgClosePaymentStream`
- `MsgSubmitMeteredUsage`

Queries:

- `QueryPaymentModel`
- `QueryServiceEscrow`
- `QueryPaymentStream`
- `QueryMeteredUsage`
- `QueryPaymentSettlement`

Failure modes:

- Escrow underfunded.
- Usage receipt invalid.
- Stream settlement exceeds maximum.
- Payment model changed after call signing.

Integration points:

- bank or Financial Zone.
- `x/services`
- `x/servicecalls`
- `x/payments`

### 15.5 x/serviceproviders

Purpose:

- Register providers, collateral, reputation, and availability commitments for fog market services.

State model:

- `ProviderRecord`
- `ProviderCollateral`
- `ProviderReputation`
- `AvailabilityCommitment`
- `ProviderFault`

Messages:

- `MsgRegisterProvider`
- `MsgUpdateProvider`
- `MsgStakeProviderCollateral`
- `MsgUnstakeProviderCollateral`
- `MsgSubmitAvailabilityCommitment`
- `MsgSubmitProviderFault`

Queries:

- `QueryProvider`
- `QueryProvidersByService`
- `QueryProviderCollateral`
- `QueryProviderReputation`
- `QueryAvailabilityCommitment`

Failure modes:

- Collateral insufficient.
- Provider advertises unsupported interface.
- Availability commitment expired.
- Fault proof invalid.
- Reputation update not deterministic.

Integration points:

- `x/services`
- `x/servicepayments`
- slashing-like penalty routing.
- routing and discovery layer.

### 15.6 x/servicereceipts

Purpose:

- Commit service, call, payment, storage, and provider receipts into proof-queryable roots.

State model:

- `ReceiptRecord`
- `ReceiptRoot`
- `ReceiptTombstone`
- `ReceiptParams`

Messages:

- `MsgAnchorReceipt`
- `MsgPruneReceipt`

Queries:

- `QueryReceipt`
- `QueryReceiptsByService`
- `QueryReceiptRoot`
- `QueryReceiptProof`

Failure modes:

- Duplicate receipt.
- Receipt hash mismatch.
- Receipt pruned before proof horizon.
- Missing receipt for executed on-chain call.

Integration points:

- all service modules.
- proof registry.
- Store v2.

## 16. Implementation Roadmap

### Phase 0: Specification and Compatibility

Tasks:

- Finalize service descriptor schema.
- Finalize interface schema format.
- Define call envelope.
- Define receipt format.
- Define payment model enum.
- Define trust and verification model enums.
- Map existing Aetheris modules to on-chain services.

Exit criteria:

- All core objects have protobuf definitions.
- All signable objects have canonical encoding test vectors.
- Existing modules can expose service descriptors.

### Phase 1: Core Registry

Tasks:

- Implement `x/services`.
- Add service registration and update.
- Add service name and owner indexes.
- Add identity binding placeholder.
- Add service proof query.
- Add export and import.

Exit criteria:

- Services are discoverable by ID, owner, and name.
- Descriptors are proof-queryable.
- Registry state is reproducible.

### Phase 2: Interface System

Tasks:

- Implement `x/serviceinterface`.
- Add interface registration.
- Add method schema.
- Add interface hash validation.
- Add interface proof query.
- Add SDK interface verifier.

Exit criteria:

- Clients can fetch and verify formal service interfaces.
- Interface versioning is enforced.
- Existing modules can publish method schemas.

### Phase 3: Unified Calls and Receipts

Tasks:

- Implement `x/servicecalls`.
- Implement `x/servicereceipts`.
- Add call envelope validation.
- Add nonces and idempotency.
- Add callbacks and retries.
- Add deterministic receipts.

Exit criteria:

- On-chain services can be called through unified call envelope.
- Receipts are committed and proof-queryable.
- Replay attempts are rejected.

### Phase 4: Payments

Tasks:

- Implement `x/servicepayments`.
- Add per-call payment.
- Add escrow settlement.
- Add metered usage receipt.
- Add payment model query.
- Integrate with bank or Financial Zone.

Exit criteria:

- Calls can require and settle service payments.
- Payment model is known before signing.
- Escrow and metered usage are test-covered.

### Phase 5: Off-Chain and Mixed Services

Tasks:

- Add signed request and response format.
- Add result anchoring.
- Add mixed service challenge flow.
- Add fallback execution hooks.
- Add provider collateral penalties.

Exit criteria:

- Off-chain service results can be anchored.
- Mixed service results can be challenged.
- Provider faults can be penalized deterministically.

### Phase 6: Fog Market Providers

Tasks:

- Implement `x/serviceproviders`.
- Add provider registry.
- Add collateral staking.
- Add availability commitments.
- Add reputation commitments.
- Add provider selection query.

Exit criteria:

- Providers can advertise services with stake and interface support.
- Clients can query provider set by service.
- Fault and collateral flows are deterministic.

### Phase 7: SDK and UX Tooling

Tasks:

- Add service resolver SDK.
- Add interface-driven call builder.
- Add CLI command generation from interface schema.
- Add wallet metadata format.
- Add proof verification helpers.

Exit criteria:

- Client can resolve service, fetch interface, build call, attach payment, execute, and verify receipt.
- Manual ABI coding is not required for registered services.

### Phase 8: Performance Hardening

Tasks:

- Add BlockSTM conflict benchmarks.
- Add Store v2 read/write benchmarks.
- Add service call throughput tests.
- Add receipt proof generation benchmarks.
- Add registry lookup benchmarks.
- Add mixed service dispute load tests.

Exit criteria:

- Independent services and calls parallelize safely.
- Proof and receipt generation remain bounded.
- Registry and interface lookups remain low-latency.

## 17. Required Test Coverage

### 17.1 Unit Tests

- Service ID validation.
- Descriptor hash calculation.
- Interface hash calculation.
- Method ID validation.
- Call ID derivation.
- Nonce replay rejection.
- Idempotency key behavior.
- Payment model validation.
- Trust model validation.
- Receipt hash calculation.

### 17.2 Integration Tests

- Register on-chain service.
- Register off-chain service anchor.
- Register mixed service.
- Register provider for fog market service.
- Register interface and bind to service.
- Resolve service through `.aet` binding.
- Execute on-chain service call.
- Anchor off-chain service result.
- Challenge mixed service result.
- Settle escrow payment.
- Generate and verify service receipt proof.

### 17.3 Invariant Tests

- Service descriptor hash matches stored descriptor.
- Service interface hash matches registered interface.
- Active service does not reference missing interface.
- Call receipt references existing service and method.
- Payment settlement does not exceed escrow.
- Provider collateral cannot be negative.
- Expired service cannot accept new calls.
- Receipt root includes all committed receipts.

### 17.4 Fuzz Tests

- Malformed descriptors.
- Malformed interface schemas.
- Large payload calls.
- Duplicate nonces.
- Duplicate idempotency keys.
- Forged provider signatures.
- Invalid result anchors.
- Invalid dispute proofs.
- Payment edge cases.

### 17.5 Performance Tests

- Service registry direct lookup latency.
- Interface lookup latency.
- Service call enqueue throughput.
- On-chain service call execution throughput.
- Receipt anchoring throughput.
- Receipt proof generation latency.
- Provider lookup latency.
- BlockSTM conflict rate for independent services.

## 18. Observability

### 18.1 Metrics

- Active services.
- Services by type.
- Services by trust model.
- Registered interfaces.
- Active providers.
- Provider collateral total.
- Calls submitted.
- Calls executed.
- Calls failed.
- Calls expired.
- Receipts anchored.
- Payment escrow total.
- Disputes opened.
- Disputes resolved.
- Average service lookup latency.
- Average interface lookup latency.
- Receipt proof generation latency.
- BlockSTM conflict rate for service calls.

### 18.2 Events

- `service_registered`
- `service_updated`
- `service_renewed`
- `service_disabled`
- `service_identity_bound`
- `interface_registered`
- `interface_deprecated`
- `service_call_submitted`
- `service_call_executed`
- `service_call_failed`
- `service_result_anchored`
- `service_receipt_committed`
- `service_payment_escrowed`
- `service_payment_settled`
- `provider_registered`
- `provider_collateral_staked`
- `provider_fault_submitted`
- `provider_penalized`

### 18.3 Alerts

- Service call failure spike.
- Receipt anchoring backlog.
- Provider fault spike.
- Interface hash mismatch attempt.
- Expired service receiving calls.
- Payment escrow settlement backlog.
- Receipt proof latency above threshold.
- Registry lookup latency above threshold.

## 19. Hard Constraints

- No centralized service backend assumptions.
- No hard dependency on messaging applications.
- No monolithic service execution engine.
- No manual ABI-driven integrations for registered services.
- No external API reliance in consensus execution.
- No nondeterministic state transitions.
- No unbounded registry scans.
- No unmetered proof verification.
- No use of off-chain result as canonical without proof, signature, challenge window, or explicit trust model.

## 20. Acceptance Criteria

The Service and Application Layer is ready for implementation planning when:

- Services are first-class registry objects.
- On-chain, off-chain, mixed, and fog market service types are defined.
- Every service can bind to a formal interface hash.
- Unified calls support payload, payment, proof requirement, timeout, signature, and idempotency.
- Discovery can resolve service descriptors through registry, signed cache, identity binding, or mesh record.
- Payments support on-chain, streaming, prepaid, metered, and escrow modes.
- Storage declarations support ephemeral, on-chain, distributed off-chain, and hybrid models.
- Mixed services have challenge or fallback logic.
- Provider collateral and reputation rules are deterministic.
- Cosmos SDK modules expose MsgServer, QueryServer, Keeper, genesis import/export, invariants, events, typed errors, and root commitments.
- Store v2 layout and BlockSTM conflict strategy are specified.
- SDKs can resolve service, verify interface, build call, attach payment, execute, and verify receipt.
