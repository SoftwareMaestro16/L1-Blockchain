# AFT-44 Fungible Token Standard

Working standard name: Aetra Fungible Token Standard, `AFT-44`.

`AFT-44` defines user-created fungible tokens through an Aetra-native
master/wallet contract model. It is not the native AET token model: AET remains
the base chain coin for fees, staking, bank balances, and mint inflation.

## Contract Model

- `token_master` is the root contract for one user token.
- `token_wallet` is a separate contract/account for each holder of that token.
- Native `AET`/`naet` is not an `AFT-44` token and cannot be represented by an
  `AFT-44` master contract.

## Token Master Storage

- token id/address
- owner/admin address
- total supply
- mintable flag
- burnable flag
- metadata/content reference
- token name
- token symbol
- decimals
- token wallet code hash or code reference
- admin transfer/renounce state
- optional allowlist/denylist policy, only if added by a later version

## Token Master Messages

- `mint`
- `burn_notification` from token wallet
- `change_admin`
- `accept_admin`
- `renounce_admin`
- `change_metadata`
- `close_minting`
- `get_supply_data`
- `get_wallet_address(owner)`
- `upgrade_wallet_code`, only if explicitly allowed by policy

## Token Wallet Storage

- master address
- owner address
- balance
- wallet code/version
- pending inbound/outbound query ids
- processed query ids for replay rejection

## Token Wallet Messages

- `transfer`
- `internal_transfer`
- `receive_transfer`
- `burn`
- `notify_owner`
- `send_excess/refund`
- `bounce` handling for failed sends

## Deterministic Rules

- Wallet contract address is derived from
  `(token_master, owner, token_wallet_code_hash)`.
- Mint deploys or credits the recipient token wallet.
- Transfer moves balance from sender wallet to recipient wallet through
  deterministic internal messages.
- Burn decreases wallet balance first, then notifies master to decrement total
  supply.
- Master supply and all wallet balances must satisfy
  `master.total_supply == sum(wallet.balance)`.
- Token metadata must not spoof native AET metadata: `Aetra`, `AET`, or
  `naet`.
- Token balances are never accepted as protocol fee payment. Every transaction
  that creates a master, mints, transfers, burns, changes admin, changes
  metadata, or processes bounce/refund messages must pay protocol fees in
  `naet`.
- Replayed wallet query ids are rejected. Bounce/finalize handling must be
  deterministic and must not create supply drift.

## VM-Independent Contract Requirements

- Explicit storage schema: `token_master` and `token_wallet` fields are listed
  above and versioned by the standard package.
- Explicit inbound messages: master admin messages, wallet transfer messages,
  burn messages, and bounce/finalize messages are listed above.
- Explicit outbound messages: internal transfer, burn notification,
  owner notification, excess/refund, and bounce messages.
- Explicit getters: supply data, wallet address for owner, master metadata, and
  wallet state.
- Explicit unknown-message policy: reject unknown opcode or malformed message
  before state mutation; if the envelope is bounceable, emit deterministic
  bounce through the async semantics layer.
- Explicit bounce behavior: replay markers are finalized deterministically and
  supply accounting must remain `master.total_supply == sum(wallet.balance)`.
- Explicit fee behavior: all protocol fees are paid in `naet`; token balances
  never satisfy base-chain fees.
- Explicit deployment behavior: token master deployment validates metadata and
  wallet code hash; token wallets are derived from master, owner, and wallet
  code hash, then deployed or credited on mint/transfer.

## Executable Specification

The pure Go package `x/aetravm/standards/aft` is the current executable
specification for AFT-44. It does not wire SDK stores or keepers into the app,
but it does expose an async/AVM-compatible token master handler and
deterministic message codec. It defines:

- master and wallet state shapes
- deterministic wallet address derivation
- metadata validation and native-spoof rejection
- native-only operation fee validation
- mint, transfer, burn, admin transfer, renounce, close minting, and bounce
  state transitions
- admin-controlled metadata changes
- accounting validation for master supply versus wallet balances
- async conformance execution through `AsyncHandler`
- deterministic bounce/finalize handling for failed wallet delivery

When the AVM keeper is wired into the app, bytecode entrypoints must match this
package's message ABI or explicitly migrate the standard version.

## Required Tests

```powershell
go test ./x/aetravm/standards/aft
go test ./...
```

The package tests cover:

- create token master
- mint to first holder
- derive holder wallet address
- transfer between holders
- deploy missing recipient wallet
- burn and supply decrement
- admin transfer and renounce
- admin controls metadata
- non-admin mint rejected
- native `AET`/`naet` spoof rejected
- non-`naet` fee rejected for token operations
- bounce/finalize state remains deterministic
- malformed message rejected
- replayed wallet message rejected
- async transfer bounce/refund behavior through the queue
