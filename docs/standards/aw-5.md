# AW-5 Wallet Contract Standard

Working standard name: Aetra Wallet Standard, `AW-5`.

`AW-5` defines replay-safe Aetra contract wallets while keeping consensus
rules deterministic. A relayer may submit a wallet command, but on-chain
protocol fee payment remains native-only in `naet`.

## Wallet Storage

- `signature_allowed` flag
- `seqno`
- `wallet_id`
- public key
- owner address if separate from public key
- extensions dictionary/map
- recovery policy with enabled flag, authority, and bounded delay
- optional subscription or standing-payment policy

## Wallet Messages

- external signed command
- internal extension command
- update `signature_allowed`
- install extension
- remove extension
- multi-send message batch
- query wallet state

## Deterministic Rules

- `seqno` prevents replay.
- `wallet_id` prevents cross-wallet replay.
- `valid_until` prevents old signed commands.
- Signature validation happens before protocol fee acceptance for external
  commands.
- Extension auth is explicit and revocable.
- Multi-send is bounded.
- Recovery policy is explicit and bounded.
- Wallet cannot silently pay protocol fees in non-`naet`.
- A relayer can submit a command, but on-chain fee payment is still `naet`.

## VM-Independent Contract Requirements

- Explicit storage schema: wallet flags, sequence, wallet id, public key,
  owner, extensions, recovery policy, and subscription policy are listed above.
- Explicit inbound messages: external signed commands, internal extension
  commands, signature policy updates, extension admin messages, and multi-send
  batches are listed above.
- Explicit outbound messages: bounded send batches emitted by signed commands
  or authorized extensions.
- Explicit getters: wallet state, sequence, wallet id, signature policy, public
  key, owner, and installed extensions.
- Explicit unknown-message policy: reject unknown command kind or malformed
  command before state mutation; external commands must fail before fee
  acceptance when signature, `seqno`, `wallet_id`, or `valid_until` is invalid.
- Explicit bounce behavior: failed outbound contract messages are handled by
  the async semantics layer; wallet state is not silently mutated by bounce
  unless a versioned wallet handler explicitly accepts that message.
- Explicit fee behavior: all protocol fees are paid in `naet`; relayers can
  submit commands but on-chain fee payment remains native-only.
- Explicit deployment behavior: wallet deployment validates non-zero wallet
  address, positive wallet id, public key length, optional owner, and extension
  map bounds.

## Executable Specification

The pure Go package `x/aetravm/standards/aw` is the current executable
specification for AW-5. It does not wire SDK stores or keepers into the app,
but it does expose an async/AVM-compatible contract handler and deterministic
message codec. It defines:

- wallet state validation
- deterministic external command signing bytes
- ed25519 signature verification
- replay checks for `seqno`
- cross-wallet replay checks for `wallet_id`
- expiry checks for `valid_until`
- signature_allowed updates
- explicit extension install/remove and internal extension command auth
- explicit bounded recovery policy validation
- bounded multi-send batches
- native-only protocol fee validation for direct and relayed commands
- async conformance execution through `AsyncHandler`
- deterministic bounce handling for failed outbound sends

When the AVM keeper is wired into the app, bytecode entrypoints must match this
package's message ABI or explicitly migrate the standard version.

## Required Tests

```powershell
go test ./x/aetravm/standards/aw
go test ./...
```

The package tests cover:

- deploy wallet
- signed send
- replayed seqno rejected
- wrong wallet_id rejected
- expired valid_until rejected
- invalid signature rejected
- extension install/remove
- extension authorized send
- unauthorized extension rejected
- recovery policy bounded
- multi-send bounded
- relayer flow pays `naet`
- async contract execution with deterministic bounce for a failed send
