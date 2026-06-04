# Orbitalis Custom Module Attacker Model

## Tokenfactory

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
- Required invariants: v1 accepts only `norb`, minimum fee applies outside simulation after height 0, accounting totals equal validator plus community targets, unauthorized params updates fail, failed ante checks do not call the next ante handler.

## Cross-Module / App

- Attacker-controlled input: signed tx bytes, malformed protobuf, sequence/account-number reuse, invalid signer pubkeys, repeated tx broadcasts, multi-module tx order, localnet RPC submissions.
- Assets at risk: account balances, nonce replay protection, app liveness, deterministic state execution.
- Main abuse cases: replaying committed tx bytes, signer mismatch, malformed tx decoder panics, mempool spam, governance abuse of custom params, tokenfactory assets used in DEX with broken accounting.
- Required invariants: malformed tx bytes fail safely, invalid signer txs fail before state mutation, replayed tx bytes fail after sequence increment, cross-module happy path keeps bank, DEX, tokenfactory, and fees state queryable and consistent.
