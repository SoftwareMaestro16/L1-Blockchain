# Aetheris DNS and Identity Specification

Status: Internal design document
Scope: Next-generation `.aet` identity, naming, resolver, and execution-routing layer
Visibility: Private, not for public repository inclusion

## 1. Design Scope

### 1.1 Objective

Define `x/identity v2` as a fully on-chain Aetheris identity module that supports:

- Trustless `.aet` name registration and lifecycle management.
- Proof-verifiable domain and resolver lookup.
- Recursive hierarchical resolution.
- NFT-backed domain ownership consistency.
- Unified identity to wallet, contract, service, interface, and routing metadata resolution.
- Execution-aware transaction routing.
- High-frequency resolver updates without global state bottlenecks.
- Store v2 optimized registry and resolver access.
- BlockSTM-compatible parallel execution for independent identity updates.
- AdaptiveSync-compatible state recovery for identity-heavy nodes.

### 1.2 Non-Goals

- Do not introduce non-`.aet` top-level identity roots in the initial module.
- Do not make off-chain metadata authoritative for ownership or primary address resolution.
- Do not allow resolver records to override registry ownership.
- Do not rely on non-deterministic validation rules.
- Do not make wallet UI rendering consensus-critical.
- Do not require on-chain writes for ordinary read-only resolution.
- Do not couple identity module safety to staking, bank, or execution modules beyond explicit integration points.

### 1.3 Current System Baseline

Existing constraints to preserve:

- Domain names must be lowercase.
- Domain names must end with `.aet`.
- Maximum labels: `8`.
- Maximum encoded name length: `253` bytes.
- ASCII-only validation.
- Empty labels rejected.
- Invalid spoofing patterns rejected.
- NFT ownership must match registry ownership.
- Reverse resolution must verify forward resolution consistency.
- Subdomains are hierarchical and can be delegated.
- Full identity state must be canonical and exportable.

Current gaps to close:

- `x/identity` is not production-activated as a complete on-chain module.
- No standard proof-based light-client resolution path.
- No unified resolver interface for service, contract, interface, and routing metadata.
- No execution-level send-by-name or invoke-by-name path.
- No optimized high-throughput identity update design for BlockSTM.
- No Store v2 state layout specifically designed for millions of domains.

## 2. Architecture Overview

### 2.1 Module Decomposition

```text
                 +-----------------------------+
                 |       x/identity v2         |
                 +-----------------------------+
                   |        |        |       |
                   |        |        |       |
       +-----------+        |        |       +----------------+
       |                    |        |                        |
+-------------+     +---------------+     +-------------+     +------------------+
| identity    |     | resolver      |     | subdomain   |     | auction          |
| core        |     | module        |     | module      |     | module           |
+-------------+     +---------------+     +-------------+     +------------------+
       |                    |                    |                    |
       +--------------------+--------------------+--------------------+
                            |
                    +----------------+
                    | proof module   |
                    +----------------+
                            |
                    +---------------------+
                    | routing integration |
                    +---------------------+
```

### 2.2 Core Data Flow

```text
register/renew/transfer/update
        |
        v
deterministic validation
        |
        v
registry state + NFT ownership check
        |
        v
resolver state update
        |
        v
proof-indexed Store v2 writes
        |
        v
events + queryable proofs + execution routing hooks
```

### 2.3 Trust Boundaries

- Registry ownership is authoritative only when registry and NFT state agree.
- Resolver records are authoritative only when controlled by the active domain owner or authorized delegate.
- Service endpoints are discoverable metadata, not consensus-authoritative execution results.
- Interface descriptors define client behavior, not chain execution validity.
- Light clients trust only verified state proofs against trusted headers.
- Cached resolution paths are advisory until proof-validated or freshness-validated.

## 3. On-Chain Identity Module: x/identity v2

### 3.1 Module State Schema

#### 3.1.1 DomainRecord

Fields:

- `name`
- `name_hash`
- `normalized_name`
- `parent_name_hash`
- `tld`
- `owner`
- `resolver`
- `expiry_height`
- `expiry_time`
- `renewal_start_height`
- `nft_class_id`
- `nft_item_id`
- `status`
- `lifecycle_epoch`
- `created_at_height`
- `updated_at_height`
- `version`
- `flags`

Status values:

- `available`
- `committed`
- `auction`
- `active`
- `renewal_window`
- `expired`
- `grace_locked`
- `released`
- `revoked`

Invariants:

- `name == normalized_name`.
- `tld == ".aet"`.
- `name_hash` is derived from canonical normalized name.
- `owner` must match NFT owner when status is active or renewal-window.
- `resolver` must be empty or controlled by owner or authorized delegate.
- `expiry_height` must be greater than current height for active domains.
- `parent_name_hash` must match immediate parent for subdomains.

#### 3.1.2 DomainCommitment

Fields:

- `commitment_hash`
- `committer`
- `created_at_height`
- `expires_at_height`
- `deposit`
- `commitment_version`
- `salt_hash_optional`

Rules:

- Commitment hash binds normalized name, committer, salt, chain ID, module version, and registration intent.
- Commitment expires after configured reveal window.
- Commitment deposit is refundable only for valid reveal or clean expiry where configured.
- Same commitment hash cannot be replayed after reveal.

#### 3.1.3 DomainNFTBinding

Fields:

- `name_hash`
- `nft_class_id`
- `nft_item_id`
- `owner`
- `last_verified_height`
- `binding_version`

Rules:

- Binding is valid only if NFT module owner equals registry owner.
- Transfer must update registry owner and NFT owner atomically.
- Broken binding puts domain into restricted state until repaired by authorized flow.

#### 3.1.4 UnifiedResolutionRecord

Fields:

- `name_hash`
- `primary_address`
- `contract_targets`
- `service_endpoints`
- `interface_descriptors`
- `routing_metadata`
- `execution_hints`
- `record_version`
- `record_ttl`
- `updated_at_height`
- `owner_signature_optional`

#### 3.1.5 ReverseResolutionRecord

Fields:

- `address`
- `name_hash`
- `name`
- `verified`
- `updated_at_height`
- `expiry_height`

Rules:

- Reverse record is valid only when forward resolution of `name` returns `address` as primary or authorized alias.
- Unverified reverse records must not be displayed as canonical.

#### 3.1.6 DelegationRecord

Fields:

- `name_hash`
- `delegate`
- `scope`
- `permissions`
- `expires_at_height`
- `subtree_limit`
- `record_prefix_limit`
- `created_at_height`

Scopes:

- `resolver_update`
- `subdomain_create`
- `subdomain_transfer`
- `service_record_update`
- `interface_record_update`
- `routing_record_update`
- `zone_admin`

#### 3.1.7 AuctionRecord

Fields:

- `auction_id`
- `name_hash`
- `status`
- `commit_start_height`
- `commit_end_height`
- `reveal_start_height`
- `reveal_end_height`
- `min_bid`
- `winning_bid`
- `winner`
- `sealed_commitments_root`
- `revealed_bids_count`
- `fee_split_id`

#### 3.1.8 ResolutionCacheRecord

Fields:

- `name_hash`
- `resolution_path_hash`
- `resolved_record_hash`
- `valid_until_height`
- `source_version`
- `parent_epoch`
- `child_epoch`

Rules:

- Cache records are optimization hints.
- Cache records must not bypass proof verification for light clients.
- Cache invalidates on domain transfer, resolver update, delegation update, expiry, or parent epoch change.

### 3.2 Store v2 Key Layout

Primary keys:

- `domains/{name_hash}` -> `DomainRecord`
- `domain_names/{normalized_name}` -> `name_hash`
- `commitments/{commitment_hash}` -> `DomainCommitment`
- `nft_bindings/{nft_class_id}/{nft_item_id}` -> `name_hash`
- `nft_bindings_by_name/{name_hash}` -> `DomainNFTBinding`
- `resolvers/{name_hash}` -> `UnifiedResolutionRecord`
- `reverse/{address}` -> `ReverseResolutionRecord`
- `delegations/{name_hash}/{delegate}/{scope}` -> `DelegationRecord`
- `subdomains/{parent_name_hash}/{child_label_hash}` -> `child_name_hash`
- `auctions/{auction_id}` -> `AuctionRecord`
- `auctions_by_name/{name_hash}/{auction_id}` -> `auction_id`
- `resolution_cache/{name_hash}/{path_hash}` -> `ResolutionCacheRecord`
- `expiry_index/{expiry_height}/{name_hash}` -> `name_hash`
- `owner_index/{owner}/{name_hash}` -> `name_hash`
- `resolver_index/{resolver}/{name_hash}` -> `name_hash`

Performance requirements:

- Hot update paths write only domain-specific keys where possible.
- Global indexes must be append-only or partitioned by height to reduce BlockSTM conflicts.
- Expiry processing must be bounded per block.
- Resolver reads must require one primary key lookup for direct domains.
- Recursive resolution must use deterministic parent path keys.

### 3.3 Keeper Logic

Required keepers:

- `IdentityKeeper`
  - Domain lifecycle transitions.
  - Ownership and NFT consistency checks.
  - Registration, renewal, transfer, expiry.
- `ResolverKeeper`
  - Unified resolver record validation and updates.
  - Reverse resolution verification.
  - Record TTL and version control.
- `DelegationKeeper`
  - Scoped delegate authorization.
  - Subdomain permission enforcement.
  - Zone-level execution control.
- `AuctionKeeper`
  - Commit/reveal auction lifecycle.
  - Deterministic winner selection.
  - Fee split execution.
- `ProofKeeper`
  - Proof path construction.
  - Resolution proof query assembly.
  - Verification helper logic.
- `RoutingIntegrationKeeper`
  - Identity to transaction target mapping.
  - Contract invocation target resolution.
  - Service and interface metadata routing.

Keeper invariants:

- No active domain can exist without valid NFT binding.
- No resolver update can be accepted from unauthorized signer.
- No reverse record can be verified without forward consistency.
- No subdomain can outlive parent unless explicitly configured as detached and paid.
- No auction result can activate a domain without deterministic winner proof.
- No cached resolution can remain valid after source record version changes.

### 3.4 Transaction Interface

Core messages:

- `MsgCommitRegistration`
- `MsgRevealRegistration`
- `MsgRegisterDirect`
- `MsgRenewDomain`
- `MsgTransferDomain`
- `MsgSetResolver`
- `MsgUpdateResolverRecord`
- `MsgSetReverseRecord`
- `MsgVerifyReverseRecord`
- `MsgCreateSubdomain`
- `MsgDelegateSubdomain`
- `MsgRevokeDelegation`
- `MsgStartAuction`
- `MsgCommitBid`
- `MsgRevealBid`
- `MsgFinalizeAuction`
- `MsgExpireDomain`
- `MsgBatchUpdateResolvers`
- `MsgBatchRenewDomains`
- `MsgInvalidateResolutionCache`

Message requirements:

- Include `chain_id` or sign mode domain separation where applicable.
- Include signer authorization scope.
- Include normalized name or name hash.
- Include expected record version for update messages.
- Include fee and storage-cost payment.
- Include replay-resistant nonce or commitment for commit/reveal flows.

### 3.5 Query Interface

Core queries:

- `QueryDomain`
- `QueryDomainByName`
- `QueryDomainsByOwner`
- `QueryDomainNFTBinding`
- `QueryResolver`
- `QueryResolvePrimary`
- `QueryResolveTarget`
- `QueryResolveService`
- `QueryResolveInterface`
- `QueryResolveRoute`
- `QueryReverse`
- `QueryVerifiedReverse`
- `QuerySubdomains`
- `QueryDelegations`
- `QueryAuction`
- `QueryResolutionProof`
- `QueryRecursiveResolutionProof`
- `QueryDomainLifecycle`
- `QueryIdentityParams`

Query requirements:

- Support pagination for owner, resolver, and subdomain indexes.
- Return record version and height for cache validation.
- Return proof paths when requested.
- Return explicit failure codes instead of ambiguous empty responses.

### 3.6 Domain Lifecycle State Machine

```text
available
    |
    | MsgCommitRegistration
    v
committed
    |
    | MsgRevealRegistration or MsgRegisterDirect
    v
active
    |
    | renewal window begins
    v
renewal_window
    |
    | no renewal by expiry
    v
expired
    |
    | grace period begins
    v
grace_locked
    |
    | grace period ends
    v
released
    |
    | public registration or auction
    v
available or auction
```

Auction alternative:

```text
available
    |
    | MsgStartAuction
    v
auction
    |
    | commit/reveal/finalize
    v
active
```

Transition rules:

- `available -> committed` requires valid commitment and deposit.
- `committed -> active` requires valid reveal, valid payment, and NFT mint or binding.
- `available -> active` is allowed only for direct registration class where enabled.
- `active -> renewal_window` is automatic based on height or time.
- `renewal_window -> active` requires renewal payment.
- `renewal_window -> expired` is automatic after expiry.
- `expired -> grace_locked` preserves prior owner recovery rights where configured.
- `grace_locked -> released` removes owner recovery rights.
- `released -> available` makes the name registrable again.
- `auction -> active` requires deterministic finalization.

## 4. Deterministic Validation Rules

### 4.1 Name Normalization

Validation rules:

- Input name must be ASCII.
- Name must be lowercase.
- Name must end with `.aet`.
- Total encoded length must be `<= 253` bytes.
- Label count must be `<= 8`.
- Empty labels are invalid.
- Leading and trailing separators are invalid.
- Reserved labels are rejected.
- Invalid spoofing patterns are rejected.
- Normalization must be deterministic and versioned.

Required tasks:

- Define `NameNormalizationVersion`.
- Add test vectors for valid and invalid names.
- Add migration rule for future normalization versions.
- Reject updates using unsupported normalization versions.

### 4.2 Ownership Validation

Validation rules:

- Registry owner must equal NFT owner for active records.
- Transfer must update NFT ownership and registry owner in one atomic operation.
- Resolver updates require owner or delegated permission.
- Subdomain creation requires parent owner or scoped delegate.
- Expired domain owner cannot update resolver except recovery metadata where explicitly allowed.

Required tasks:

- Add pre-transfer and post-transfer invariant checks.
- Add NFT binding repair flow for module-internal failures.
- Add tests for transfer race conditions.

### 4.3 Resolver Validation

Validation rules:

- Primary address must be syntactically valid.
- Contract targets must reference valid contract addresses or registered code identifiers where applicable.
- Service endpoints must match allowed endpoint schemas.
- Interface descriptors must have deterministic schema hashes.
- Routing metadata must remain bounded in size.
- Execution hints must be advisory unless explicitly supported by execution router.

Required tasks:

- Define resolver field size limits.
- Define allowed service endpoint schemes.
- Define interface descriptor hash format.
- Add record version checks for update conflicts.

### 4.4 Reverse Resolution Validation

Validation rules:

- Reverse record must reference an existing active domain.
- Forward resolution must return the reverse address as primary or approved alias.
- Reverse record expires no later than the domain.
- Reverse verification must be rechecked after transfer, resolver update, or expiry.

Required tasks:

- Add reverse verification transaction.
- Add automatic reverse invalidation on relevant updates.
- Add query distinction between claimed and verified reverse records.

## 5. Proof-Based Resolution Layer

### 5.1 Proof Objectives

Proof-based resolution must allow a light client to verify:

- Domain existence.
- Domain non-existence.
- Domain status and expiry.
- NFT ownership binding.
- Resolver record contents.
- Reverse resolution consistency.
- Subdomain delegation chain.
- Recursive resolution path.
- Record version and freshness at a trusted height.

### 5.2 Proof Format

`IdentityResolutionProof` fields:

- `proof_version`
- `chain_id`
- `height`
- `app_hash`
- `name`
- `name_hash`
- `query_type`
- `normalized_name_proof`
- `domain_record`
- `domain_record_proof`
- `nft_binding`
- `nft_binding_proof`
- `resolver_record`
- `resolver_record_proof`
- `reverse_record_optional`
- `reverse_record_proof_optional`
- `delegation_chain`
- `delegation_chain_proofs`
- `subdomain_path`
- `subdomain_path_proofs`
- `non_existence_proof_optional`
- `record_version`
- `proof_commitment_hash`

`RecursiveResolutionProof` fields:

- `proof_version`
- `chain_id`
- `height`
- `root_name`
- `target_name`
- `path_labels`
- `path_hashes`
- `path_domain_records`
- `path_resolver_records`
- `path_delegation_records`
- `path_proofs`
- `final_resolution_record`
- `final_record_proof`
- `cache_record_optional`
- `cache_record_proof_optional`

Proof encoding requirements:

- Deterministic binary encoding.
- Explicit field order.
- Versioned proof schema.
- Domain-separated proof commitment hash.
- No reliance on off-chain ordering.

### 5.3 Deterministic Resolution Path

For `service.api.alice.aet`, traversal order:

- `alice.aet`
- `api.alice.aet`
- `service.api.alice.aet`

Resolution steps:

- Normalize target name.
- Split labels.
- Compute canonical parent path from right to left, excluding `.aet`.
- Verify each parent exists or has valid delegation rule.
- Verify child ownership or delegation authorization.
- Verify final resolver record.
- Verify requested target type exists.
- Verify record freshness and expiry.

### 5.4 Light Client Verification Algorithm

Algorithm:

1. Verify trusted header for proof height.
2. Verify proof `chain_id` matches expected chain.
3. Normalize requested name locally.
4. Compute `name_hash`.
5. Verify domain record proof against `app_hash`.
6. Verify domain status is active or renewal-window where allowed.
7. Verify expiry has not passed at proof height.
8. Verify NFT binding proof.
9. Verify registry owner equals NFT owner.
10. Verify resolver record proof.
11. Verify resolver record version and TTL.
12. If recursive, verify each path component and delegation proof.
13. If reverse resolution is requested, verify forward consistency.
14. Return requested target only if all proof checks pass.

Failure codes:

- `ERR_INVALID_NAME`
- `ERR_UNSUPPORTED_NORMALIZATION_VERSION`
- `ERR_PROOF_HEIGHT_UNTRUSTED`
- `ERR_DOMAIN_NOT_FOUND`
- `ERR_DOMAIN_EXPIRED`
- `ERR_DOMAIN_NOT_ACTIVE`
- `ERR_NFT_BINDING_MISMATCH`
- `ERR_RESOLVER_NOT_FOUND`
- `ERR_RESOLVER_UNAUTHORIZED`
- `ERR_TARGET_NOT_FOUND`
- `ERR_DELEGATION_MISSING`
- `ERR_DELEGATION_EXPIRED`
- `ERR_REVERSE_NOT_VERIFIED`
- `ERR_PROOF_INVALID`
- `ERR_RECORD_STALE`

### 5.5 Failure Handling

On-chain query failure:

- Return typed error and optional non-existence proof.
- Do not return partial target as valid.
- Include latest known record version where safe.

Light client failure:

- Reject target use.
- Request fresh proof from another node.
- Fall back to direct address entry only with explicit user confirmation.
- Mark cached resolution as invalid.

Wallet failure:

- Display verified status only after proof success.
- Do not auto-fill transaction target from unverified resolution.
- Warn when proof height is older than local freshness threshold.

## 6. Unified Resolver System

### 6.1 UnifiedResolutionRecord

Canonical schema:

- `primary_address`
  - Main account address for payment and default send-by-name.
- `contract_targets`
  - Named contract targets for invocation.
- `service_endpoints`
  - Network endpoints and service descriptors.
- `interface_descriptors`
  - UI and interaction schema descriptors.
- `routing_metadata`
  - Transaction, execution, and service routing metadata.
- `execution_hints`
  - Optional advisory execution preferences.

Record-level fields:

- `name_hash`
- `owner`
- `record_version`
- `record_ttl`
- `updated_at_height`
- `max_payload_bytes`
- `schema_version`

### 6.2 Contract Targets

Fields:

- `target_id`
- `contract_address`
- `entrypoint`
- `interface_hash`
- `required_funds_policy`
- `gas_hint`
- `enabled`
- `updated_at_height`

Validation:

- `target_id` must be unique within resolver.
- Contract address must be valid.
- Entrypoint must be bounded and ASCII.
- Interface hash must match descriptor where referenced.
- Gas hint is advisory and must not bypass gas metering.

### 6.3 Service Endpoints

Fields:

- `service_id`
- `service_type`
- `endpoint`
- `transport`
- `auth_policy`
- `health_path_optional`
- `priority`
- `weight`
- `ttl`
- `schema_hash_optional`

Validation:

- Endpoint scheme must be allowed by governance parameters.
- Payload size must remain under resolver record limit.
- Service type must be versioned.
- TTL must not exceed domain expiry.

### 6.4 Interface Descriptors

Fields:

- `interface_id`
- `schema_hash`
- `schema_uri_optional`
- `schema_inline_optional`
- `version`
- `render_policy`
- `permissions_required`
- `contract_target_id_optional`
- `service_id_optional`

Rules:

- Inline schemas must be size-bounded.
- URI schemas are advisory unless hash is verified.
- Wallets must verify `schema_hash` before using external schema content.
- Interface descriptors cannot grant execution permission by themselves.

### 6.5 Routing Metadata

Fields:

- `route_id`
- `target_type`
- `preferred_target`
- `fallback_targets`
- `chain_context`
- `fee_hint`
- `timeout_hint`
- `memo_policy`
- `capability_requirements`

Rules:

- Routing metadata is advisory for clients and transaction builders.
- Execution validity remains determined by target module validation.
- Fallback targets must be explicitly listed.
- Fee hints cannot override fee module requirements.

### 6.6 Execution Hints

Fields:

- `default_gas_limit_hint`
- `preferred_fee_mode`
- `message_type`
- `async_allowed`
- `requires_memo`
- `requires_interface_confirmation`
- `simulation_required`

Rules:

- Execution hints must not bypass ante-handler checks.
- Hints are ignored when unsupported.
- Hints must be included in proof when used by light clients.

## 7. Execution Integration Layer

### 7.1 Identity to Transaction Mapping

Send-by-name flow:

1. User enters `alice.aet`.
2. Client normalizes name.
3. Client queries `QueryResolvePrimary` with proof request.
4. Client verifies proof or relies on full-node local state.
5. Client maps primary address to bank send target.
6. Client builds transaction with resolved address.
7. Transaction memo may include original name and resolution height for auditability.

Rules:

- On-chain bank send receives address, not name.
- Name resolution must happen before transaction signing.
- Wallet must display resolved address, proof status, and proof height.
- If resolution changes before inclusion, signed transaction still targets resolved address.

Required tasks:

- Define wallet resolution API.
- Add proof-aware send-by-name SDK helper.
- Add memo format for optional identity audit trail.
- Add stale proof warning policy.

### 7.2 Identity to Contract Invocation Mapping

Invoke-by-name flow:

1. User selects domain and contract target ID.
2. Client resolves `contract_targets[target_id]`.
3. Client verifies interface descriptor if used.
4. Client simulates transaction with resolved contract address and entrypoint.
5. Client displays verified target and interface hash.
6. Client signs transaction using resolved address and message type.

Rules:

- Contract target resolution is pre-signing client behavior.
- Contract execution validates normal message authorization.
- Interface metadata must not create implicit approvals.
- Target ID must be stable across resolver updates unless version changes.

Required tasks:

- Define contract target query.
- Define interface descriptor verification.
- Add SDK helper for invoke-by-name.
- Add stale interface descriptor warnings.

### 7.3 Identity to RPC and Service Mapping

Service discovery flow:

1. Client resolves `service_endpoints[service_id]`.
2. Client verifies proof for service record.
3. Client filters endpoint by supported transport and policy.
4. Client verifies external metadata hash where applicable.
5. Client connects using endpoint priority and weight.

Rules:

- Service endpoints are discoverable metadata.
- Endpoint availability is not guaranteed by consensus.
- Clients must fail closed when endpoint proof is invalid.
- Service record TTL must be respected.

Required tasks:

- Define service endpoint type registry.
- Add service record query.
- Add endpoint freshness policy.
- Add endpoint fallback behavior.

### 7.4 Identity to Interface Schema Mapping

UI generation flow:

1. Client resolves interface descriptor.
2. Client verifies descriptor proof and schema hash.
3. Client loads inline schema or fetches external schema.
4. Client verifies external schema hash.
5. Client renders UI based on local wallet policy.

Rules:

- Interface descriptors are declarative and untrusted until hash-verified.
- Wallet policy may reject unsupported render policies.
- UI generation must require explicit user confirmation for signing.
- Interface schema cannot alter resolved execution target.

Required tasks:

- Define descriptor schema.
- Define render policy values.
- Add schema hash verification.
- Add wallet policy compatibility matrix.

## 8. Hierarchical DNS Improvement

### 8.1 Subdomain Delegation Model

Delegation types:

- Owner-controlled subdomain.
- Delegate-controlled subdomain.
- Zone-managed subdomain.
- Detached paid subdomain.
- Ephemeral service subdomain.

Rules:

- Parent owner can create child domains.
- Parent owner can delegate scoped subdomain permissions.
- Child expiry cannot exceed parent expiry unless detached mode is enabled.
- Detached mode requires independent registration payment and explicit parent authorization.
- Delegation must be revocable unless configured as time-locked.

Required tasks:

- Define `DelegationRecord` scope bits.
- Add subdomain creation validation.
- Add detached subdomain lifecycle.
- Add delegation revocation flow.

### 8.2 Resolution Hop Reduction

Strategies:

- Store direct parent hash in each domain record.
- Store child index under parent prefix.
- Store optional path commitment for recursive resolution.
- Store zone policy at parent level.
- Cache resolved path commitments with version dependencies.

Rules:

- Direct domain lookup uses one `domains/{name_hash}` read.
- Recursive proof lookup fetches only path labels required for target.
- Cache records are invalidated by parent or child version changes.
- Cache cannot override actual registry or resolver state.

Required tasks:

- Define path commitment format.
- Add cache invalidation on parent epoch changes.
- Add recursive proof query optimization.
- Add tests for stale path cache rejection.

### 8.3 Partial Delegation

Permission scopes:

- Create specific child label.
- Create any child under prefix.
- Update resolver fields only.
- Update service records only.
- Update interface descriptors only.
- Update routing metadata only.
- Manage zone expiry renewals.

Rules:

- Delegation must specify scope, expiry, and label constraints.
- Delegate cannot transfer parent domain unless explicitly authorized.
- Delegate cannot extend permissions beyond its own scope.
- Delegate updates must include expected delegation version.

Required tasks:

- Add scoped authorization checks.
- Add delegation versioning.
- Add prefix-bound delegation tests.
- Add permission escalation tests.

### 8.4 Zone-Level Execution Control

Zone policy fields:

- `allowed_record_types`
- `allowed_service_types`
- `subdomain_creation_policy`
- `resolver_update_policy`
- `interface_policy`
- `routing_policy`
- `max_child_depth`
- `max_child_records`

Rules:

- Zone policy applies to child domains unless overridden by valid child policy.
- Parent policy cannot override child ownership for detached subdomains.
- Zone policy changes increment parent lifecycle epoch.
- Policy changes invalidate dependent cache records.

Required tasks:

- Define zone policy schema.
- Add policy inheritance rules.
- Add cache invalidation tests.
- Add recursive policy proof output.

## 9. Performance Optimization

### 9.1 Store v2 Optimization

Goals:

- Support millions of domains with bounded lookup cost.
- Keep direct resolution read path compact.
- Minimize write amplification on resolver updates.
- Support efficient export and state sync.

State layout rules:

- Use `name_hash` as primary lookup key.
- Keep resolver record separate from domain ownership record.
- Keep reverse records under address prefix.
- Keep owner indexes separate from primary records.
- Keep expiry indexes partitioned by height.
- Store large interface metadata by hash reference where possible.

Read path targets:

- Direct primary resolution:
  - Read `domains/{name_hash}`.
  - Read `resolvers/{name_hash}`.
  - Optional read `nft_bindings_by_name/{name_hash}` when proof requires explicit binding.
- Reverse resolution:
  - Read `reverse/{address}`.
  - Read `domains/{name_hash}`.
  - Read `resolvers/{name_hash}`.
- Recursive resolution:
  - Read one domain record per path label.
  - Read final resolver record.
  - Read delegation records only when child ownership depends on delegation.

Required tasks:

- Define key prefix constants.
- Add store benchmark for direct and recursive resolution.
- Add state export/import tests.
- Add proof query benchmark.

### 9.2 BlockSTM Parallel Execution

Parallel-safe transaction classes:

- Register different names.
- Renew different names.
- Transfer different names.
- Update resolver for different names.
- Update reverse records for different addresses.
- Create subdomains under different parents.
- Reveal bids for different auctions.
- Finalize different auctions.

Conflict-prone transaction classes:

- Multiple updates to same name.
- Parent policy update concurrent with child creation.
- Transfer concurrent with resolver update for same name.
- Reverse verification concurrent with primary resolver update.
- Auction finalize concurrent with late reveal.

Design rules:

- Use domain-specific keys for hot writes.
- Avoid global counters in transaction hot path.
- Partition fee accounting by block or module accumulator.
- Use expected record versions to detect same-domain update conflicts.
- Batch resolver updates by disjoint name hashes.

Required tasks:

- Add BlockSTM conflict tests for identity messages.
- Add benchmarks for batch resolver updates.
- Add versioned update rejection tests.
- Add parent-child conflict classification.

### 9.3 Batched Resolver Updates

Batch message:

- `MsgBatchUpdateResolvers`

Rules:

- Batch must contain unique name hashes.
- Each update must include expected record version.
- Each update must be independently authorized.
- Failure mode is configurable:
  - atomic all-or-nothing.
  - partial success with per-update results.
- Partial success must emit deterministic result order.

Use cases:

- Service endpoint rotation.
- Interface descriptor update.
- Contract target migration.
- Routing metadata refresh.
- Zone-level child resolver maintenance.

Required tasks:

- Define batch update limits.
- Add per-update authorization.
- Add gas metering per update.
- Add partial success result schema if enabled.

### 9.4 Caching Strategy

Cache layers:

- Full node in-memory direct resolution cache.
- Full node recursive path cache.
- Wallet verified proof cache.
- Service endpoint TTL cache.
- Reverse resolution verified cache.

Invalidation triggers:

- Domain transfer.
- Resolver update.
- NFT binding update.
- Domain expiry.
- Renewal changing lifecycle epoch.
- Delegation update.
- Zone policy update.
- Reverse record update.

Rules:

- Caches must be keyed by name hash, record version, and height.
- Light-client cache entries must include proof height and trusted header reference.
- TTL must not exceed resolver TTL or domain expiry.
- Execution-critical use should refresh proof if freshness threshold is exceeded.

Required tasks:

- Define cache key format.
- Add event-driven invalidation.
- Add verified cache metadata format.
- Add stale cache tests.

### 9.5 Minimize State Writes Per Resolution Cycle

Rules:

- Read-only resolution performs no writes.
- Cache warming must not write consensus state.
- Resolver access counters are off-chain metrics unless explicitly enabled.
- Proof queries do not mutate state.
- Reverse verification writes only when requested.

Required tasks:

- Add read-only query audit.
- Add tests that resolution queries do not modify state.
- Add event-only observability for lookup volume where needed.

### 9.6 ABCI++ Lifecycle Integration

Use cases:

- Pre-proposal filtering for malformed high-volume identity updates.
- Proposal preparation grouping disjoint domain updates for better parallel execution.
- Vote extension data for non-consensus identity telemetry where enabled.
- Finalization hooks for expiry and cache invalidation events.

Rules:

- Proposal optimization must not change transaction validity.
- Identity expiration processing must be deterministic.
- Vote extension telemetry must not affect identity state unless committed through normal messages.

Required tasks:

- Define malformed identity transaction precheck.
- Add proposal grouping by name hash.
- Add bounded expiry processing in finalize path.
- Add deterministic event emission tests.

### 9.7 AdaptiveSync Integration

Requirements:

- Snapshot must include domains, resolvers, delegations, reverse records, NFT bindings, auctions, commitments, and expiry indexes.
- Recovering nodes must verify identity invariants after state sync.
- Proof queries must work immediately after successful sync.
- Watcher and wallet caches must resync from events or queries.

Required tasks:

- Add identity module snapshot support.
- Add post-sync invariant verification.
- Add recovery tests with active auctions and expiring domains.
- Add state export compatibility tests.

## 10. Security Model

### 10.1 Registration Commit Replay Protection

Commitment preimage fields:

- `chain_id`
- `module_name`
- `module_version`
- `normalized_name`
- `committer`
- `salt`
- `registration_class`
- `max_price`
- `expiry_height`

Rules:

- Commitment hash is single-use.
- Reveal must occur within reveal window.
- Expired commitments cannot be revealed.
- Commitments from another chain domain are invalid.
- Reusing a revealed commitment is rejected.

Required tasks:

- Add used commitment tombstones.
- Add commit/reveal fuzz tests.
- Add reveal window boundary tests.

### 10.2 Anti-Squatting Constraints

Mechanisms:

- Length-based pricing.
- Renewal fees.
- Commitment deposits.
- Auction for contested names.
- Storage fees for resolver payload size.
- Grace period recovery cost.
- Limits on bulk registration per account per window where enabled.

Rules:

- Anti-squatting controls must be deterministic.
- Bulk registration limits must not depend on non-consensus identity.
- Pricing must be queryable before registration.
- Expired names cannot be held indefinitely without renewal payment.

Required tasks:

- Define pricing formula.
- Add bulk registration simulations.
- Add renewal incentive tests.
- Add expired-domain release tests.

### 10.3 Resolver Spoofing Prevention

Rules:

- Resolver updates require owner or scoped delegate authorization.
- Resolver record owner must match active domain owner.
- Service endpoint metadata must not be displayed as verified ownership.
- Interface descriptors must be hash-verified before use.
- Reverse records are unverified until forward resolution confirms them.
- Similar-looking invalid spoofing patterns are rejected at normalization stage where deterministic rules apply.

Required tasks:

- Add unauthorized resolver update tests.
- Add interface hash mismatch tests.
- Add reverse spoofing tests.
- Add service endpoint display policy.

### 10.4 NFT Ownership Consistency Enforcement

Rules:

- Domain activation mints or binds NFT atomically.
- Domain transfer updates registry and NFT owner atomically.
- NFT transfer hook must update registry owner or reject transfer where configured.
- Registry transfer hook must update NFT owner.
- Broken binding blocks resolver changes until repaired.

Required tasks:

- Add transfer hook integration.
- Add atomicity tests.
- Add broken-binding repair flow.
- Add invariant check at end-block or periodic audit.

### 10.5 State Consistency Validation

Invariants:

- No duplicate `name_hash` records for different names.
- No active domain without active NFT binding.
- No resolver record without existing domain unless reserved system record.
- No verified reverse record without forward consistency.
- No subdomain without parent or detached authorization.
- No active auction for a currently active name unless renewal auction mode is explicitly enabled.
- No delegation extending beyond domain expiry unless detached and paid.

Required tasks:

- Add invariant tests.
- Add export validation.
- Add migration validation.
- Add periodic consistency audit query.

### 10.6 Reverse Resolution Safety

Rules:

- Reverse record is a claim until verified.
- Verified reverse record invalidates on primary resolver change.
- Verified reverse record invalidates on domain transfer.
- Verified reverse record invalidates on domain expiry.
- Wallets must display unverified reverse records separately from verified records.

Required tasks:

- Add reverse invalidation hooks.
- Add verified reverse proof format.
- Add wallet display state definitions.

### 10.7 Resolver Payload Safety

Risks:

- Oversized resolver records.
- Excessive interface schema payload.
- Endpoint spam.
- Record churn for high-frequency updates.
- Malformed metadata intended to break clients.

Controls:

- Maximum record byte size.
- Per-field byte limits.
- Storage fee per byte.
- Inline schema size cap.
- Endpoint count limit.
- Update fee multiplier for high churn.
- Client-side schema validation.

Required tasks:

- Define field limits.
- Add storage fee calculation.
- Add malformed metadata tests.
- Add payload size fuzz tests.

## 11. Economic Layer Improvements

### 11.1 Domain Pricing Model

Inputs:

- Name length.
- Label count.
- Registration duration.
- Demand class.
- Renewal status.
- Auction status.
- Resolver payload size.
- Subdomain mode.

Pricing components:

- Base registration fee.
- Scarcity multiplier for short names.
- Label-depth fee.
- Storage footprint fee.
- Commitment deposit.
- Renewal fee.
- Auction settlement fee where applicable.

Rules:

- Price must be deterministic and queryable.
- Price must be paid in `naet`.
- Storage-heavy resolver records pay higher fees.
- Renewal should be cheaper than post-expiry reacquisition where configured.

Required tasks:

- Define pricing function.
- Add `QueryRegistrationPrice`.
- Add `QueryRenewalPrice`.
- Add pricing boundary tests.

### 11.2 Renewal Incentives

Mechanisms:

- Renewal window before expiry.
- Grace period after expiry.
- Higher recovery fee during grace period.
- Loss of resolver update rights after expiry.
- Resolver soft-freeze during grace period.
- Discount for multi-period renewal where enabled.

Rules:

- Renewal cannot exceed maximum duration.
- Renewal cannot bypass ownership consistency checks.
- Expired domain cannot create new subdomains.
- Grace recovery must preserve NFT binding or repair it atomically.

Required tasks:

- Add renewal window state transition.
- Add grace recovery flow.
- Add resolver freeze tests.
- Add maximum duration checks.

### 11.3 Spam Resistance

Attack surfaces:

- Bulk registration.
- Commit spam.
- Resolver update spam.
- Subdomain creation spam.
- Auction commitment spam.
- Service endpoint spam.
- Proof query load.

Controls:

- Commitment deposit.
- Channel-like storage fee per active record.
- Per-byte resolver cost.
- Batch update gas metering.
- Auction bid deposit.
- Subdomain creation fee.
- Endpoint count limits.
- Query-level node rate limits outside consensus.

Required tasks:

- Add spam cost model.
- Add high-volume registration simulation.
- Add resolver update gas benchmarks.
- Add subdomain creation stress test.

### 11.4 Auction Fairness Improvements

Auction rules:

- Commit/reveal phases must be height-bounded.
- Commitments bind bidder, name, bid amount, salt, and chain domain.
- Reveals outside window are rejected.
- Winner selection is deterministic.
- Ties resolve by deterministic rule configured in parameters.
- Unrevealed bids forfeit configured deposit portion.
- Finalization can be called by any account after reveal window.

Fee split buckets:

- Burn.
- Treasury.
- Rewards.
- Community pool.

Required tasks:

- Define auction state machine.
- Add deterministic tie-break rule.
- Add unrevealed bid handling.
- Add auction finalization tests.

### 11.5 Resolver Update Cost Model

Cost inputs:

- Base update fee.
- Updated field count.
- Added bytes.
- Removed bytes.
- Record churn frequency.
- Batch size.
- Proof index impact.

Rules:

- Resolver updates pay for net storage growth.
- Large inline interface descriptors pay higher storage fees.
- Frequent updates may pay churn surcharge where enabled.
- Batch updates must not be cheaper than equivalent individual updates below cost floor.

Required tasks:

- Define resolver update fee formula.
- Add storage delta calculation.
- Add churn surcharge parameter.
- Add batch fee accounting tests.

### 11.6 Subdomain Creation Cost Model

Cost inputs:

- Parent domain status.
- Child label length.
- Delegation mode.
- Detached mode.
- Resolver payload size.
- Expiry duration.
- Zone policy complexity.

Rules:

- Parent-controlled child creation pays minimal child record cost.
- Detached subdomains pay independent registration and renewal costs.
- Delegated subdomain creation charges delegate or parent according to signed policy.
- Child expiry cannot exceed parent expiry unless detached.

Required tasks:

- Define subdomain fee formula.
- Add detached mode pricing.
- Add delegated billing policy.
- Add expiry constraint tests.

## 12. Light Client Experience

### 12.1 Lightweight Resolution Flow

Steps:

1. Client obtains trusted header at height `H`.
2. Client normalizes requested `.aet` name.
3. Client requests `QueryResolutionProof` or `QueryRecursiveResolutionProof`.
4. Client verifies proof against `app_hash` from trusted header.
5. Client verifies domain status, expiry, NFT binding, and resolver authorization.
6. Client extracts requested target.
7. Client caches target with proof height, record version, and TTL.

### 12.2 Proof Verification Steps

Required checks:

- Header trust.
- Chain ID match.
- Proof height match.
- Name normalization match.
- Name hash match.
- Domain record proof.
- Domain lifecycle validity.
- NFT binding proof.
- Ownership consistency.
- Resolver proof.
- Requested target field existence.
- Record TTL and expiry.
- Delegation proof for subdomains.
- Reverse forward-consistency proof when applicable.

### 12.3 Fallback Strategies

Fallback order:

- Request fresh proof from another full node.
- Request proof at a newer trusted height.
- Use previously verified cache only if still within TTL and user policy allows.
- Require explicit address entry when proof cannot be verified.
- Disable auto-generated contract or service UI when interface proof fails.

Rules:

- Fallback must not silently downgrade verified identity to unverified identity.
- Wallet must distinguish stale, missing, invalid, and untrusted proofs.
- Cached proof cannot outlive domain expiry or resolver TTL.

### 12.4 Trustless Wallet UX Model

Wallet display states:

- `verified_current`
- `verified_stale`
- `unverified_claim`
- `proof_failed`
- `expired`
- `ownership_mismatch`
- `resolver_missing`

Rules:

- Send-by-name requires `verified_current` unless user explicitly overrides.
- Invoke-by-name requires verified contract target and verified interface hash where used.
- Reverse names display as canonical only in `verified_current`.
- Service endpoint discovery must display proof height and TTL in advanced details.

Required tasks:

- Define wallet resolution state enum.
- Add proof cache metadata format.
- Add stale proof warning thresholds.
- Add explicit override audit trail.

## 13. Module Breakdown

### 13.1 identity-core-module

Purpose:

- Own canonical domain lifecycle, registry state, NFT binding, ownership transfer, renewal, and expiry.

State model:

- `DomainRecord`
- `DomainCommitment`
- `DomainNFTBinding`
- `DomainLifecycleParams`
- `ExpiryIndex`
- `OwnerIndex`

Message types:

- `MsgCommitRegistration`
- `MsgRevealRegistration`
- `MsgRegisterDirect`
- `MsgRenewDomain`
- `MsgTransferDomain`
- `MsgExpireDomain`
- `MsgRepairNFTBinding`

Query methods:

- `QueryDomain`
- `QueryDomainByName`
- `QueryDomainsByOwner`
- `QueryDomainNFTBinding`
- `QueryDomainLifecycle`
- `QueryRegistrationPrice`
- `QueryRenewalPrice`

Failure modes:

- Registry and NFT owner mismatch.
- Expiry index missing domain.
- Commitment replay.
- Transfer race with resolver update.
- Incorrect lifecycle transition.

Integration points:

- NFT module for domain token mint, transfer, and ownership checks.
- Bank module for registration and renewal fees.
- Fee module for spam resistance.
- Resolver module for ownership authorization.
- Store v2 for domain and index storage.
- BlockSTM for parallel domain updates.

Implementation tasks:

- Define protobuf state and messages.
- Implement lifecycle keeper.
- Add NFT binding hooks.
- Add invariants for ownership consistency.
- Add Store v2 key layout tests.

### 13.2 resolver-module

Purpose:

- Store and validate unified resolution records for addresses, contracts, services, interfaces, routing metadata, and execution hints.

State model:

- `UnifiedResolutionRecord`
- `ContractTarget`
- `ServiceEndpoint`
- `InterfaceDescriptor`
- `RoutingMetadata`
- `ExecutionHints`
- `ReverseResolutionRecord`
- `ResolverParams`

Message types:

- `MsgSetResolver`
- `MsgUpdateResolverRecord`
- `MsgBatchUpdateResolvers`
- `MsgSetReverseRecord`
- `MsgVerifyReverseRecord`
- `MsgClearResolverRecord`

Query methods:

- `QueryResolver`
- `QueryResolvePrimary`
- `QueryResolveTarget`
- `QueryResolveService`
- `QueryResolveInterface`
- `QueryResolveRoute`
- `QueryReverse`
- `QueryVerifiedReverse`

Failure modes:

- Unauthorized resolver update.
- Oversized record payload.
- Stale expected record version.
- Reverse record forward mismatch.
- Interface descriptor hash mismatch.
- Resolver TTL exceeds domain expiry.

Integration points:

- Identity core for owner and lifecycle checks.
- Subdomain module for delegated resolver updates.
- Fee module for update and storage fees.
- Routing integration module for execution target mapping.
- Store v2 for resolver state.

Implementation tasks:

- Define unified record schema.
- Add field-level validation.
- Add reverse verification logic.
- Add batch update handler.
- Add resolver proof query support.

### 13.3 subdomain-module

Purpose:

- Manage hierarchical subdomain creation, delegation, partial permissions, detached mode, and zone policies.

State model:

- `SubdomainRecord`
- `DelegationRecord`
- `ZonePolicy`
- `SubdomainIndex`
- `PathCommitment`

Message types:

- `MsgCreateSubdomain`
- `MsgDelegateSubdomain`
- `MsgRevokeDelegation`
- `MsgUpdateZonePolicy`
- `MsgDetachSubdomain`
- `MsgRenewSubdomain`

Query methods:

- `QuerySubdomains`
- `QueryDelegations`
- `QueryZonePolicy`
- `QueryRecursivePath`
- `QuerySubdomainAuthorization`

Failure modes:

- Child expiry exceeds parent expiry.
- Delegate escalates permissions.
- Parent transfer leaves child cache stale.
- Zone policy creates inconsistent child rules.
- Detached subdomain lacks independent payment.

Integration points:

- Identity core for domain lifecycle.
- Resolver module for child resolver control.
- Fee module for subdomain creation costs.
- Proof module for recursive proof path.
- Store v2 for parent-child indexes.

Implementation tasks:

- Define delegation permission bits.
- Add zone policy validation.
- Add detached lifecycle.
- Add recursive proof integration.
- Add parent-child conflict tests.

### 13.4 auction-module

Purpose:

- Provide deterministic contested-name allocation through sealed and timestamp-based auction flows.

State model:

- `AuctionRecord`
- `BidCommitment`
- `RevealedBid`
- `AuctionParams`
- `AuctionFeeSplit`

Message types:

- `MsgStartAuction`
- `MsgCommitBid`
- `MsgRevealBid`
- `MsgFinalizeAuction`
- `MsgCancelExpiredAuction`
- `MsgClaimAuctionRefund`

Query methods:

- `QueryAuction`
- `QueryAuctionByName`
- `QueryBidCommitment`
- `QueryAuctionParams`
- `QueryAuctionPriceFloor`

Failure modes:

- Bid replay.
- Reveal outside window.
- Non-deterministic tie handling.
- Finalization before reveal end.
- Incorrect refund accounting.

Integration points:

- Identity core for domain activation.
- Bank module for bid deposits and refunds.
- Fee module for auction fees.
- Treasury, burn, rewards, and community pool accounting.
- Store v2 for auction indexes.

Implementation tasks:

- Define auction state machine.
- Add deterministic winner selection.
- Add deposit and refund accounting.
- Add finalization tests.
- Add bid reveal fuzz tests.

### 13.5 proof-verification-module

Purpose:

- Assemble and verify identity resolution proofs for full nodes, light clients, and wallet tooling.

State model:

- `ProofParams`
- `ProofSchemaVersion`
- `ResolutionCacheRecord`
- `ProofPathDescriptor`

Message types:

- No state-changing messages required for basic proof verification.
- Optional `MsgInvalidateResolutionCache` if consensus cache records are enabled.

Query methods:

- `QueryResolutionProof`
- `QueryRecursiveResolutionProof`
- `QueryReverseResolutionProof`
- `QueryNonExistenceProof`
- `QueryProofSchema`

Failure modes:

- Proof assembled from inconsistent record versions.
- Recursive path proof omits delegation constraint.
- Cache proof remains after invalidation.
- Non-existence proof returned for malformed name.
- Proof height unavailable after pruning.

Integration points:

- Identity core for domain records.
- Resolver module for resolution records.
- Subdomain module for delegation path.
- Store v2 proof APIs.
- AdaptiveSync state snapshots.

Implementation tasks:

- Define proof schema.
- Add proof assembly logic.
- Add recursive proof tests.
- Add light-client verifier test vectors.
- Add stale proof rejection tests.

### 13.6 routing-integration-module

Purpose:

- Map identity records into transaction, contract, service, interface, and routing targets for clients and execution tooling.

State model:

- `RoutingPolicy`
- `IdentityTransactionMapping`
- `ContractInvocationMapping`
- `ServiceMapping`
- `InterfaceMapping`
- `ExecutionHintPolicy`

Message types:

- `MsgSetRoutingPolicy`
- `MsgUpdateExecutionHints`
- `MsgRegisterInterfaceDescriptor`
- `MsgRegisterServiceEndpoint`
- `MsgClearRoutingMetadata`

Query methods:

- `QueryTransactionMapping`
- `QueryContractInvocationMapping`
- `QueryServiceMapping`
- `QueryInterfaceMapping`
- `QueryExecutionHints`
- `QueryResolvedExecutionTarget`

Failure modes:

- Stale routing metadata after resolver update.
- Interface descriptor points to wrong target.
- Service endpoint proof succeeds but endpoint unavailable.
- Execution hint conflicts with target module rules.
- Client treats advisory metadata as authorization.

Integration points:

- Resolver module for unified records.
- Fee module for metadata update costs.
- Contract execution layer for target address mapping.
- Wallet SDK for UI generation.
- Service clients for endpoint discovery.

Implementation tasks:

- Define mapping schemas.
- Add execution target query.
- Add interface descriptor hash checks.
- Add wallet SDK helper formats.
- Add stale routing metadata tests.

## 14. API and SDK Requirements

### 14.1 Node API

Required endpoints:

- Resolve primary address with optional proof.
- Resolve contract target with optional proof.
- Resolve service endpoint with optional proof.
- Resolve interface descriptor with optional proof.
- Resolve routing metadata with optional proof.
- Resolve reverse record with forward consistency proof.
- Query domain lifecycle state.
- Query registration and renewal price.
- Query delegation authorization.

### 14.2 Wallet SDK

Required helpers:

- `NormalizeName`
- `ValidateName`
- `ResolvePrimaryVerified`
- `ResolveContractTargetVerified`
- `ResolveServiceVerified`
- `ResolveInterfaceVerified`
- `VerifyResolutionProof`
- `BuildSendByNameTx`
- `BuildInvokeByNameTx`
- `RenderVerifiedInterface`
- `CheckReverseResolution`

### 14.3 Indexer Compatibility

Rules:

- Indexers may serve convenience queries.
- Indexer responses must include proof or proof reference for trustless mode.
- Indexer metadata must not be treated as canonical without verification.
- Events must include enough keys to maintain domain, owner, resolver, and expiry indexes.

Required tasks:

- Define event schema.
- Add indexer replay test.
- Add proof passthrough format.

## 15. Events and Observability

### 15.1 Events

Required events:

- `identity_domain_committed`
- `identity_domain_registered`
- `identity_domain_renewed`
- `identity_domain_transferred`
- `identity_domain_expired`
- `identity_domain_released`
- `identity_nft_binding_updated`
- `identity_resolver_updated`
- `identity_reverse_set`
- `identity_reverse_verified`
- `identity_reverse_invalidated`
- `identity_subdomain_created`
- `identity_delegation_created`
- `identity_delegation_revoked`
- `identity_zone_policy_updated`
- `identity_auction_started`
- `identity_bid_committed`
- `identity_bid_revealed`
- `identity_auction_finalized`
- `identity_cache_invalidated`

### 15.2 Metrics

Required metrics:

- Active domains.
- Expired domains.
- Domains in renewal window.
- Domains in grace period.
- Resolver record count.
- Average resolver payload size.
- Reverse records verified.
- Reverse records invalidated.
- Subdomains by depth.
- Delegation records active.
- Auctions active.
- Commitments active.
- Batch resolver update size.
- BlockSTM conflict rate for identity messages.
- Store v2 read latency for direct resolution.
- Store v2 read latency for recursive resolution.
- Store v2 write latency for resolver updates.
- Proof query latency.
- Proof verification failure count.
- Expiry processing backlog.

### 15.3 Alerts

Required alerts:

- NFT binding mismatch detected.
- Resolver payload size near maximum.
- Expiry processing backlog above threshold.
- Proof query failure spike.
- Registration spam spike.
- Resolver update spam spike.
- Auction finalization backlog.
- BlockSTM conflict rate above threshold.
- Store v2 identity read latency above threshold.
- Reverse verification mismatch spike.

## 16. Governance Parameters

### 16.1 Name Parameters

- Maximum labels.
- Maximum name bytes.
- Supported normalization version.
- Reserved labels.
- Invalid spoofing pattern set version.
- Maximum subdomain depth.

### 16.2 Lifecycle Parameters

- Minimum registration duration.
- Maximum registration duration.
- Renewal window duration.
- Grace period duration.
- Commitment reveal window.
- Commitment tombstone retention.
- Expiry processing limit per block.

### 16.3 Pricing Parameters

- Base registration fee.
- Short-name multiplier.
- Label-depth fee.
- Renewal fee multiplier.
- Grace recovery fee multiplier.
- Resolver byte fee.
- Subdomain creation fee.
- Detached subdomain fee.
- Auction minimum bid.

### 16.4 Resolver Parameters

- Maximum resolver record bytes.
- Maximum contract targets.
- Maximum service endpoints.
- Maximum interface descriptors.
- Maximum routing metadata bytes.
- Maximum inline schema bytes.
- Minimum resolver TTL.
- Maximum resolver TTL.
- Allowed endpoint schemes.

### 16.5 Delegation Parameters

- Maximum delegation duration.
- Maximum scoped delegates per domain.
- Maximum zone policy size.
- Detached subdomain allowed.
- Time-locked delegation allowed.

### 16.6 Auction Parameters

- Commit phase duration.
- Reveal phase duration.
- Bid deposit minimum.
- Unrevealed bid penalty.
- Tie-break rule.
- Auction finalization delay.
- Fee split weights.

### 16.7 Performance Parameters

- Batch resolver update maximum size.
- Batch renewal maximum size.
- Recursive proof maximum depth.
- Cache record maximum lifetime.
- Store pruning horizon for proof availability.
- ABCI++ expiry work limit.

## 17. Required Test Coverage

### 17.1 Unit Tests

- Name normalization.
- Name hash generation.
- Commitment hash generation.
- Commitment reveal validation.
- Domain lifecycle transitions.
- NFT binding checks.
- Resolver field validation.
- Reverse forward-consistency validation.
- Delegation scope checks.
- Zone policy validation.
- Pricing function.
- Auction winner selection.
- Proof encoding.

### 17.2 Integration Tests

- Register domain and mint NFT.
- Transfer domain and update NFT ownership atomically.
- Update resolver as owner.
- Reject resolver update by unauthorized account.
- Create delegated subdomain.
- Revoke delegation and reject further delegate updates.
- Renew domain before expiry.
- Expire and release domain.
- Verify reverse resolution.
- Invalidate reverse on resolver update.
- Run commit/reveal auction.
- Batch update disjoint resolvers.
- Generate and verify recursive proof.

### 17.3 Invariant Tests

- Active registry owner equals NFT owner.
- Resolver record requires active domain.
- Verified reverse record has matching forward resolution.
- Child domain expiry does not exceed parent expiry unless detached.
- Delegation expiry does not exceed authorized domain expiry unless detached.
- Auction cannot activate already active domain.
- Resolver record byte size never exceeds parameter limit.
- Expiry index contains every expiring domain.
- Owner index matches domain ownership.
- Cache record invalidates on source version change.

### 17.4 Fuzz Tests

- Malformed names.
- Boundary-length names.
- Spoofing-pattern candidates.
- Commitment preimages.
- Auction bid reveal ordering.
- Resolver payloads.
- Interface descriptor schemas.
- Delegation permission combinations.
- Recursive proof paths.
- Reverse resolution mismatches.
- Batch update ordering.

### 17.5 Performance Tests

- Direct resolution read latency.
- Recursive resolution read latency by depth.
- Resolver update write latency.
- Batch resolver updates per block.
- Batch renewals per block.
- Domain registrations per block.
- BlockSTM conflict rate under mixed identity workload.
- Store v2 proof generation latency.
- AdaptiveSync recovery time with large identity state.
- Export/import time for identity state.

## 18. Implementation Roadmap

### Phase 0: Specification and Test Vectors

Tasks:

- Define canonical name normalization.
- Define domain hash and proof hash formats.
- Define protobuf state schemas.
- Define Store v2 key layout.
- Define governance parameter set.
- Produce resolution proof test vectors.
- Produce lifecycle transition test vectors.

Exit criteria:

- All signable and hashable identity objects have test vectors.
- All lifecycle states have deterministic transition tests.
- Store key prefixes are finalized.

### Phase 1: Core Registry Activation

Tasks:

- Implement `identity-core-module`.
- Implement registration, renewal, transfer, expiry.
- Implement NFT binding.
- Implement owner and expiry indexes.
- Implement core queries.
- Add invariant checks.

Exit criteria:

- `.aet` domain ownership is fully on-chain.
- NFT and registry ownership remain atomic.
- Export/import preserves registry state.

### Phase 2: Unified Resolver

Tasks:

- Implement `resolver-module`.
- Implement primary address resolution.
- Implement contract targets.
- Implement service endpoints.
- Implement interface descriptors.
- Implement routing metadata.
- Implement reverse resolution.
- Implement batch resolver updates.

Exit criteria:

- Unified resolver supports wallet, contract, service, interface, and routing targets.
- Reverse resolution verifies forward consistency.
- Resolver updates are versioned and size-bounded.

### Phase 3: Subdomains and Zone Control

Tasks:

- Implement `subdomain-module`.
- Implement delegated subdomain creation.
- Implement partial delegation.
- Implement detached subdomains.
- Implement zone policies.
- Implement recursive path queries.

Exit criteria:

- Recursive hierarchy supports scoped delegation.
- Parent and child expiry rules are enforced.
- Zone policy is proof-queryable.

### Phase 4: Proof-Based Resolution

Tasks:

- Implement `proof-verification-module`.
- Implement direct resolution proof query.
- Implement recursive resolution proof query.
- Implement reverse proof query.
- Implement non-existence proof query.
- Add light-client verification SDK.

Exit criteria:

- Light clients can verify address, contract, service, interface, and reverse resolution.
- Proof failure modes are explicit.
- Proof test vectors pass across module versions.

### Phase 5: Execution Integration

Tasks:

- Implement `routing-integration-module`.
- Add send-by-name SDK helper.
- Add invoke-by-name SDK helper.
- Add service discovery helper.
- Add interface descriptor verification helper.
- Add wallet display state definitions.

Exit criteria:

- Identity records can drive pre-signing transaction target resolution.
- Wallets can verify proof-backed identity targets.
- Interface and service metadata are hash-verified before use.

### Phase 6: Performance Hardening

Tasks:

- Add Store v2 resolution benchmarks.
- Add BlockSTM batch update benchmarks.
- Add ABCI++ proposal grouping for identity transactions.
- Add bounded expiry processing.
- Add AdaptiveSync recovery tests.
- Add cache invalidation event tests.

Exit criteria:

- Independent identity updates parallelize without avoidable conflicts.
- Direct resolution remains bounded with large state.
- Recovering nodes can serve proof queries after sync.

## 19. Acceptance Criteria

The next-generation `.aet` identity system is acceptable for production hardening when:

- `x/identity v2` owns domain lifecycle on-chain.
- Registry ownership and NFT ownership are atomically consistent.
- Domain registration, renewal, transfer, resolver update, and subdomain delegation are supported.
- Direct and recursive resolution are proof-verifiable.
- Light clients can verify domain status, ownership, resolver records, and reverse consistency.
- Unified resolver supports primary address, contract target, service endpoint, interface descriptor, routing metadata, and execution hints.
- Send-by-name and invoke-by-name flows resolve before transaction signing.
- Store v2 state layout supports efficient direct and recursive reads.
- BlockSTM can parallelize independent identity updates.
- Batched resolver updates are versioned, bounded, and conflict-safe.
- Caches are invalidated by all ownership, resolver, delegation, expiry, and zone-policy changes.
- Anti-squatting, spam resistance, auction fairness, resolver cost, and subdomain cost models are parameterized.
- Reverse resolution is never treated as verified without forward consistency.
- Proof, lifecycle, ownership, resolver, delegation, auction, and performance tests are implemented.
