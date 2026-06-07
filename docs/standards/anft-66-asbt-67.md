# ANFT-66 And ASBT-67 Standards

Working standard names:

- Aetra NFT Standard, `ANFT-66`.
- Aetra Soulbound Token Standard, `ASBT-67`.

`ANFT-66` defines NFT collections and NFT item contracts. `ASBT-67` is the
native non-transferable SBT extension for item contracts whose owner is
immutable after mint.

## Contract Model

- `nft_collection` is the source of truth for collection metadata, admin
  authority, item code hash, next item index, and item address derivation.
- `nft_item` is the first-class contract/account for one collectible item.
- `sbt_item` is an `nft_item` extension with immutable owner, revocation
  authority, and transfer disabled by policy.

## NFT Collection Storage

- collection owner/admin
- collection metadata/content
- next item index
- item code hash or code reference
- standard version
- mutable metadata flag
- collection royalty policy with bounded basis points and non-zero recipient
- SBT policy flag if collection is soulbound-only

## NFT Collection Messages

- `mint_item`
- `batch_mint`, with bounded count
- `change_collection_metadata`
- `change_owner/admin`
- `get_collection_data`
- `get_item_address_by_index`
- `prove_item_belongs_to_collection`

## NFT Item Storage

- collection address
- item index
- owner address
- item metadata/content
- initialized flag
- transfer policy

## NFT Item Messages

- `transfer`
- `ownership_notification`
- `send_excess/refund`
- `get_item_data`
- `bounce` handling where needed

## SBT Item Extension Storage

- collection address
- item index
- immutable owner address
- content
- authority address
- revoked_at timestamp or zero
- revoke reason/content if needed

## SBT Item Messages

- `prove_ownership`
- `ownership_proof_response`
- `request_current_owner`
- `revoke_by_authority`
- `destroy`, only if policy allows it
- `transfer` must be rejected

## Deterministic Rules

- Collection is the source of truth for item address derivation.
- Item contract address is derived from
  `(nft_collection, item_index, item_code_hash)`.
- Item membership is proven by comparing the stored item address and index with
  the collection-derived address.
- NFT transfer requires current owner authorization.
- SBT owner is immutable after mint.
- SBT revoke does not transfer ownership.
- Metadata changes must be bounded and authorization-checked.
- Metadata and royalties are bounded.
- Metadata must not spoof native AET metadata: `Aetra`, `AET`, or `naet`.
- Batch minting must have strict limits to prevent block or async queue DoS.

## VM-Independent Contract Requirements

- Explicit storage schema: collection, NFT item, and SBT extension fields are
  listed above and versioned by the standard package.
- Explicit inbound messages: collection mint/admin messages, NFT transfer
  messages, SBT proof/revoke messages, and bounce/refund messages are listed
  above.
- Explicit outbound messages: ownership notification, proof response,
  excess/refund, and bounce messages.
- Explicit getters: collection data, item address by index, item data, item
  membership proof, SBT ownership proof, and current owner query.
- Explicit unknown-message policy: reject unknown opcode or malformed message
  before state mutation; if the envelope is bounceable, emit deterministic
  bounce through the async semantics layer.
- Explicit bounce behavior: failed mint/transfer/revoke paths do not commit
  partial item state; refund/excess behavior is deterministic.
- Explicit fee behavior: all protocol fees are paid in `naet`; NFT and SBT
  assets never satisfy base-chain fees.
- Explicit deployment behavior: collections validate metadata and item code
  hash; items are derived from collection, index, and item code hash; SBT items
  set immutable owner and authority at mint.

## Executable Specification

The pure Go package `x/aetravm/standards/anft` is the current executable
specification for ANFT-66 and ASBT-67. It does not wire SDK stores or keepers
into the app, but it does expose an async/AVM-compatible collection handler
and deterministic message codec. It defines:

- collection, NFT item, and SBT item state shapes
- deterministic item address derivation
- collection membership proof rules
- owner-authorized NFT transfer
- immutable SBT ownership and authority-only revocation
- bounded metadata validation and native-spoof rejection
- bounded royalty policy validation
- bounded, atomic batch minting
- async conformance execution through `AsyncHandler`
- deterministic failure/bounce behavior for rejected SBT transfer attempts

When the AVM keeper is wired into the app, bytecode entrypoints must match this
package's message ABI or explicitly migrate the standard version.

## Required Tests

```powershell
go test ./x/aetravm/standards/anft
go test ./...
```

The package tests cover:

- create collection
- mint NFT item
- verify item address from collection
- transfer NFT item
- unauthorized transfer rejected
- malformed collection address rejected
- SBT mint
- SBT transfer rejected
- SBT prove ownership
- SBT revoke by authority
- unauthorized revoke rejected
- metadata spoofing rejected
- royalty policy bounded
- batch mint bounded
- async NFT transfer, SBT rejection, and bounce behavior through the queue
