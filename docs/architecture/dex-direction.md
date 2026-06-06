> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# DEX Direction

Phase 12 keeps the current `x/dex` native module in the blockchain repository
while Aetra async contract execution and VM selection mature.

The production target for the native DEX is documented in
[x/dex Production Specification](dex-production-spec.md).
The current native DEX audit is documented in
[x/dex Security Audit Report](../security/dex-audit-report.md).

## Current Native Module

`x/dex` remains a native module. It is the current reference implementation for
constant-product pool accounting, reserve custody, LP share mint/burn, slippage
checks, and DEX observability.

Current safety requirements:

- DEX tests use native `naet` for native-side liquidity and swap paths.
- Pool creator must be a valid non-zero user address.
- Liquidity provider must be a valid non-zero user address.
- Liquidity withdrawer must be a valid non-zero user address.
- Swap trader is also the current swap recipient and must be a valid non-zero
  user address.
- Pool asset denoms must not spoof native AET metadata through display aliases
  such as `AET` or `Aetra`.
- Factory denoms whose subdenom spoofs `naet`, `AET`, or `Aetra` are
  rejected.
- Factory denoms whose subdenom spoofs `naet`, `AET`, or `Aetra` are rejected.
- LP denoms are native module-controlled `lp/{pool_id}` denoms and must not be
  accepted from user input as arbitrary native-token aliases.

## Native Invariants

The native DEX must keep these invariants after create, add liquidity, remove
liquidity, swap, genesis import, and export:

- recorded reserves match the `dex` module account balances
- LP supply matches `pool.total_shares`
- swaps preserve constant-product constraints within fee and integer rounding
  policy
- slippage bounds are enforced before state mutation
- malformed or corrupted pool state returns errors, not panics
- active pools cannot be reduced to zero reserves or zero LP supply through
  ordinary liquidity removal

The DEX module account remains the reserve custodian. LP supply is minted and
burned only by the DEX module.

## Future Contract DEX

Aetra should move toward contract-based pools and routers only after async
contract execution is safe and audited. Until then, the native DEX remains the
reference implementation or migration bridge.

Future contract model:

- `pool_contract`: owns one asset pair, reserves, fee policy, LP accounting, and
  pool-local swap math.
- `router_contract`: routes user swaps across one or more pools, validates
  paths and deadlines, enforces max hops, and manages async settlement.
- LP representation:
  - AFT-44 LP token master/wallet contracts, or
  - native LP representation during migration, but not both without explicit
    compatibility boundaries.
- `async_swap_settlement`: swaps settle through deterministic async messages
  with query_id tracking, deadline checks, and bounded hop count.
- Failed swap path:
  - failed pool call emits deterministic bounce
  - router finalizes failed path
  - user receives deterministic refund/excess in `naet` or the original input
    asset according to the route state

## Migration Boundary

Native DEX pools should not be silently replaced by contracts. A migration path
must define:

- pool state snapshot format
- reserve custody transfer
- LP supply migration or LP token master deployment
- router path compatibility
- fee policy equivalence
- invariant checks before and after migration
- rollback or pause procedure
- adversarial audit before public migration

## Required Tests

```powershell
go test ./x/dex/types ./x/dex/keeper
go test ./...
```

The current regression suite covers:

- native `naet` pool paths
- zero-address actor rejection
- native spoof rejection through display/factory denoms
- reserve/module balance invariant
- LP supply/share invariant
- constant-product swap invariant
- slippage rejection
- corrupted pool state error paths
- duplicate unordered pair rejection
