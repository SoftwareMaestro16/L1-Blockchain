# Aetra Custom Module Attacker Model

## Contract assets

- Attacker-controlled input: creator/admin Bech32 strings, subdenoms, mint/burn amounts, mint recipient, burn source, admin-change target, query denoms, genesis fixtures.
- Assets at risk: factory token supply, holder balances, denom admin authority, bank metadata integrity.
- Main abuse cases: unauthorized mint/burn, admin takeover, duplicate denom creation, malformed Bech32, malformed denom, metadata spoofing, native/LP-denom confusion, large query response DoS.
- Required invariants: only current admin can mint/burn/change admin, burn source must be the signer/admin account, supply changes exactly by minted/burned amount, metadata must match created factory denom, invalid operations must not mutate supply or authority.

## DEX

- Attacker-controlled input: pool creation pair/amounts, liquidity amounts, LP denom in remove requests, swap input/output denom, slippage limits, pool IDs, query requests, imported pool state.
- Assets at risk: pool reserves, LP token supply, trader funds, module account solvency.
- Main abuse cases: duplicate pair pools, same-denom pools, zero/tiny liquidity, slippage bypass, rounding leakage, corrupted reserve state, reserve-bank desync, LP inflation, large pool query DoS.
- Required invariants: pair index is unique and canonical, reserves match module account balances, LP supply equals pool total shares, constant-product swap math does not make reserves negative, failed operations do not mutate pool or bank state.

## Fees

- Attacker-controlled input: tx fee coins, fee denom, fee amount, fee payer/granter, governance param update payloads, genesis fee accounting state.
- Assets at risk: native fee collection, distribution accounting, validator/community fee split, node mempool resources.
- Main abuse cases: zero fee, non-native fee, below-min fee, duplicate fee denoms, malformed ratios, unsafe fee collector targets, repeated invalid fee spam.
- Required invariants: v1 accepts only `naet`, minimum fee applies outside simulation after height 0, accounting totals equal validator plus community targets, unauthorized params updates fail, failed ante checks do not call the next ante handler.

## Load Model

- Attacker-controlled input: canonical mempool size, block gas utilization, inclusion delay, failed transaction count, total transaction count, execution step estimates, imported load history.
- Assets at risk: shard activation decisions, fee pressure, block execution latency, validator liveness.
- Main abuse cases: spike load manipulation, overflow in score normalization, stale or reordered history import, unbounded history export, validator-local latency influence.
- Required invariants: metrics are normalized to bounded basis points, load score cannot jump more than `MAX_DELTA`, history is canonically ordered, no wall-clock or node-local latency input affects consensus state.

## Routing

- Attacker-controlled input: message type strings, fee class, reputation class, locality keys, tx hash, routing epoch, active shard table.
- Assets at risk: execution zone assignment, shard routing fairness, priority ordering, spam admission.
- Main abuse cases: unknown message classes, zero-address actors, missing primary actor, non-`naet` fee class, fee/reputation class overflow, validator-local routing preference.
- Required invariants: same state plus same tx yields the same route, fee and reputation classes are bounded, system tx priority is deterministic, non-`naet` routing fee is rejected, zero-address actors fail safely.

## Zones

- Attacker-controlled input: zone IDs, VM policy, fee policy, activation height, state roots, commitment roots, imported registry order.
- Assets at risk: zone registry integrity, commitment chain correctness, upgrade safety, root-based settlement.
- Main abuse cases: duplicate zone IDs, non-canonical import ordering, non-`naet` fee policy, unknown VM policy, root spoofing, missing previous commitment.
- Required invariants: zones and commitments are sorted canonically, commitments hash the exported root fields, activation respects height, every registered zone uses `naet` protocol fees.

## Shards

- Attacker-controlled input: routing keys, cross-shard queues, split/merge trigger sequences, data availability flags, validator epoch changes.
- Assets at risk: message delivery, shard liveness, validator assignment, queue correctness.
- Main abuse cases: lost messages during split, duplicated messages during merge, data-unavailable routing, stale headers, nondeterministic validator reassignment.
- Required invariants: shard queues preserve every message exactly once across split/merge, unavailable shards cannot be routed/finalized, validator reassignment is deterministic for an epoch.

## Mesh

- Attacker-controlled input: source/destination zones and shards, message IDs, sender/recipient bytes, asset commitments, payload hash, finality reference, proofs, receipts.
- Assets at risk: cross-zone assets, receipt accounting, replay protection, bounce/refund settlement.
- Main abuse cases: duplicate messages, duplicate receipts, stale source finality, forged proof, wrong destination, infinite bounce/refund loop, replay-based double spend.
- Required invariants: every message has at most one terminal receipt, replay markers are single-use, proof source commitment matches finalized source root, bounce/refund cannot create a second spend.

## Identity

- Attacker-controlled input: domain labels, commit/reveal salts, owner addresses, resolver records, reverse lookup address, subdomain labels, NFT transfer target.
- Assets at risk: `.aet` ownership, resolver integrity, reverse lookup integrity, domain NFT ownership.
- Main abuse cases: duplicate normalized names, spoofed labels, zero-address ownership, expired-domain resolution, unauthorized resolver/reverse updates, parent subdomain takeover.
- Required invariants: names normalize deterministically, duplicate normalized domains fail, zero-address owners fail, expired domains cannot resolve, resolver and reverse updates require the correct owner.

## VM

- Attacker-controlled input: AVM bytecode, entrypoints, imports, opcodes, storage keys, gas limits, emitted messages, exported/imported contract state.
- Assets at risk: contract state, gas accounting, async message queues, deterministic execution.
- Main abuse cases: malformed bytecode panics, forbidden nondeterministic opcodes, oversized code/state/query output, gas bypass, direct mutation of another contract state.
- Required invariants: malformed AVM input never panics, forbidden opcodes are rejected, gas and storage limits are enforced, no local time/randomness/external API is available to contract execution.

## Cross-Module / App

- Attacker-controlled input: signed tx bytes, malformed protobuf, sequence/account-number reuse, invalid signer pubkeys, repeated tx broadcasts, multi-module tx order, localnet RPC submissions.
- Assets at risk: account balances, nonce replay protection, app liveness, deterministic state execution.
- Main abuse cases: replaying committed tx bytes, signer mismatch, malformed tx decoder panics, mempool spam, governance abuse of custom params, contract-assets assets used in DEX with broken accounting.
- Required invariants: malformed tx bytes fail safely, invalid signer txs fail before state mutation, replayed tx bytes fail after sequence increment, cross-module happy path keeps bank, DEX, contract-assets, and fees state queryable and consistent.
