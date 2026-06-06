# x/dex Production Specification

Date: 2026-06-06

Status: target architecture for the native Aetheris DEX. The current module is a
constant-product prototype and must migrate toward this specification in
audited increments.

## Goals

`x/dex` is the native liquidity, routing, incentive, and oracle module for
Aetheris. It supports `naet`/custom-token and custom/custom swaps, deterministic
multi-hop execution, LP accounting, locked LP positions, farming rewards,
fee rewards, and future IBC/concentrated-liquidity extensions.

Protocol fees are denominated only in `naet`. No user-created token, LP token,
IBC asset, display denom alias, or factory denom may satisfy a protocol fee.

## Architecture

The module is split into keeper services with a single durable KV store:

- Pool Manager: pool creation, pool status, pair indexes, pool type dispatch.
- Liquidity Engine: add/remove liquidity, reserve custody, LP share math.
- Swap Engine: exact-in/exact-out execution, fee calculation, slippage checks.
- Routing Engine: bounded multi-hop paths and best-execution quote helpers.
- LP Engine: module-controlled LP denoms and supply reconciliation.
- Lock Engine: locked and burnable LP positions with unlock scheduling.
- Farming Engine: epochs, gauges, boost multipliers, and reward accounting.
- Rewards Engine: LP fee accrual, farming rewards, and user reward claims.
- Fee Engine: `naet`-only protocol fee collection and split accounting.
- Oracle/TWAP Engine: deterministic observations and fixed time windows.
- Governance Integration: bounded params, pause flags, and migrations.

Future IBC and concentrated liquidity must enter through explicit interfaces,
not by overloading constant-product state.

## Pool Types

Pool Type A is the production constant-product pool:

```text
x * y = k
amount_out = floor(amount_in_after_fee * reserve_out / (reserve_in + amount_in_after_fee))
```

Pool Type A supports swaps, LP shares, swap fees, protocol fees, and LP fee
rewards. It does not require farming, boosts, or locks.

Pool Type B extends Type A with lock-aware farming. It supports LP locks,
governance-configured boost multipliers, early-unlock penalties, epoch rewards,
and burnable positions. Pool math remains independent from reward math.

Future Pool Type C is concentrated liquidity. It must use a new pool-state
variant with ticks, positions, and tick-level liquidity. It must not mutate Type
A or Type B semantics.

## State Layout

Prefix plan:

```text
0x01 | pool_id                          -> Pool
0x02 | denom0 | 0x00 | denom1           -> uint64 pool_id
0x03 | next_pool_id                     -> uint64
0x04 | params                           -> Params
0x10 | route_index | denom_in | denom_out -> repeated uint64 pool_ids
0x20 | lock_id                          -> LockPosition
0x21 | owner | lock_id                  -> nil
0x22 | pool_id | owner | lock_id        -> nil
0x23 | unlock_time | lock_id            -> nil
0x30 | gauge_id                         -> Gauge
0x31 | pool_id | gauge_id               -> nil
0x32 | epoch_number | gauge_id          -> EpochGaugeState
0x33 | address | reward_denom           -> AccruedReward
0x40 | pool_id | window | timestamp     -> TwapObservation
0x41 | pool_id | window                 -> TwapAccumulator
0x50 | fee_epoch                        -> FeeAccounting
0x51 | fee_recipient_key                -> sdk.Coin amount
0x60 | migration_version                -> uint64
```

Indexes must be deterministic, prefix-scannable, and bounded. Route discovery
may read pool indexes, but transaction execution must use an explicit route
with a max-hop parameter.

## Protobuf Surface

Existing messages remain compatible:

- `MsgCreatePool`
- `MsgAddLiquidity`
- `MsgRemoveLiquidity`
- `MsgSwapExactAmountIn`
- `MsgUpdateParams`

New v1 extension messages:

- `MsgSwapExactAmountOut`
- `MsgMultiHopSwapExactAmountIn`
- `MsgMultiHopSwapExactAmountOut`
- `MsgLockLiquidity`
- `MsgBeginUnlockLiquidity`
- `MsgUnlockLiquidity`
- `MsgBurnLockedLiquidity`
- `MsgCreateGauge`
- `MsgAddGaugeRewards`
- `MsgClaimRewards`
- `MsgSetPoolStatus`
- `MsgCreateProtocolOwnedLiquidity`

Query surface:

- `Pool`, `Pools`, `PoolByPair`
- `SpotPrice`, `QuoteExactIn`, `QuoteExactOut`
- `BestRoute`
- `Twap`
- `LocksByOwner`, `LocksByPool`, `Lock`
- `Gauges`, `Gauge`, `PendingRewards`
- `FeeAccounting`, `Params`

All transaction messages must include signer validation, deadline where price
movement matters, slippage bounds, and bounded repeated fields.

## Params

Core params:

- `base_denom = "naet"`
- `swap_fee_bps`
- `protocol_fee_bps`
- `max_swap_fee_bps`
- `max_protocol_fee_bps`
- `max_route_hops`
- `max_pools_per_route_quote`
- `min_initial_liquidity`
- `pool_creation_enabled`
- `swaps_enabled`
- `liquidity_enabled`
- `locks_enabled`
- `farming_enabled`
- `twap_enabled`

Fee split defaults:

```text
treasury:       40%
validators:     25%
community_pool: 20%
insurance:      10%
burn:            5%
```

The split is governance-configurable, but must always sum to 100% and all
targets must be fixed safe module accounts or explicit protocol addresses.

Lock params:

- allowed lock durations
- boost multiplier per duration
- early unlock penalty per duration
- max locks per account
- min lock amount

TWAP params:

- windows: `1m`, `5m`, `15m`, `1h`, `24h`
- max observations per pool/window
- min liquidity for oracle eligibility

## Keeper Rules

All state transitions must be atomic through cache contexts. Bank movement must
happen before durable pool writes, and failure must leave pool state, balances,
LP supply, locks, gauges, and reward accounting unchanged.

Pool state invariants:

- reserves are positive for active pools
- total shares are positive for active pools
- recorded reserves equal DEX module account balances for each pool asset
- LP total supply equals `pool.total_shares`
- pair indexes are canonical and unique
- constant-product swaps do not reduce `k` after fee-aware rounding
- full pool destruction is explicit and cannot occur through remove liquidity

Routing invariants:

- route hops are bounded
- all pools in the route exist and are active
- adjacent hop denoms match exactly
- exact-in and exact-out reverse calculations use integer ceil/floor rules that
  never overdraw reserves
- deadline is checked before state mutation

Lock and farming invariants:

- locked LP cannot be removed, burned, or transferred by DEX messages except
  through the lock lifecycle
- reward emission per epoch is conserved exactly
- unclaimed rewards are claimable or accounted as protocol-owned remainder
- boost multipliers affect rewards only, not pool reserves or LP supply
- early-unlock penalties are charged only in LP shares or `naet` as configured

TWAP invariants:

- observations are written from committed pool states
- windows are deterministic and derived from block time
- missing or low-liquidity observations return errors, not synthetic prices
- oracle reads never mutate state

## Event Catalog

Required events:

- `dex_create_pool`
- `dex_add_liquidity`
- `dex_remove_liquidity`
- `dex_destroy_pool`
- `dex_swap_exact_amount_in`
- `dex_swap_exact_amount_out`
- `dex_multihop_swap`
- `dex_lock_liquidity`
- `dex_begin_unlock_liquidity`
- `dex_unlock_liquidity`
- `dex_burn_locked_liquidity`
- `dex_create_gauge`
- `dex_add_gauge_rewards`
- `dex_claim_rewards`
- `dex_fee_collected`
- `dex_fee_distributed`
- `dex_twap_observation`
- `dex_params_updated`

Events must include stable IDs, actor address, pool ID when applicable, input
coins, output coins, fees, and route IDs without relying on local ordering that
cannot be reproduced from state.

## Error Catalog

Existing errors remain:

- `ErrInvalidPool`
- `ErrPoolNotFound`
- `ErrInvalidLiquidity`
- `ErrSlippage`
- `ErrInvalidParams`
- `ErrUnauthorized`
- `ErrOperationDisabled`
- `ErrInvalidAddress`

New errors:

- `ErrInvalidRoute`
- `ErrDeadlineExceeded`
- `ErrInsufficientLiquidity`
- `ErrInvalidLock`
- `ErrLockNotMatured`
- `ErrInvalidGauge`
- `ErrInvalidReward`
- `ErrInvalidFeeDenom`
- `ErrOracleUnavailable`
- `ErrInvariantViolation`

## ADRs

ADR-DEX-001: keep `x/dex` native until async contract execution and VM
migration are audited.

ADR-DEX-002: protocol fees are `naet` only; swap assets may be any valid pool
asset, but protocol fee settlement cannot use non-`naet` assets.

ADR-DEX-003: preserve current proto messages and add extension messages rather
than breaking existing clients during the first production migration.

ADR-DEX-004: full pool destruction is explicit. `MsgRemoveLiquidity` cannot
silently leave an active pool with zero reserves or zero LP supply.

ADR-DEX-005: routing execution accepts explicit bounded routes; best-route
discovery is a query/helper surface, not an unbounded transaction scan.

ADR-DEX-006: lock boosts affect only reward distribution, never swap pricing,
pool reserves, or LP share supply.

## Threat Model

Primary risks:

- reserve/accounting divergence between pool state and bank balances
- LP supply divergence from `total_shares`
- integer rounding value creation
- unbounded route search DoS
- stale or manipulated TWAP reads
- governance misconfiguration of fee splits, boosts, penalties, or pause flags
- zero-address actors and native denom spoofing
- reward emission leakage across epochs
- lock bypass through direct LP token transfer assumptions
- migration import/export accepting invalid pool or reward state

Required mitigations:

- invariant tests after every state transition
- bounded params and repeated fields
- deterministic integer math with explicit floor/ceil policy
- cache-context atomicity for every multi-step bank/state operation
- module-account allowlists for fee distribution
- genesis validation and migration validation before writes
- fuzz tests for swap math, route math, locks, rewards, and TWAP windows

## Test Plan

Minimum suites:

- unit tests for every message validation path
- keeper atomicity tests for bank failures and corrupted state
- table tests for swap math exact-in and exact-out
- route tests for valid/invalid multi-hop paths
- invariant tests for reserves, LP supply, fees, and rewards
- lock lifecycle tests including early unlock penalties and maturity
- epoch farming tests with exact reward conservation
- TWAP window tests across block-time boundaries
- genesis import/export round-trip tests
- migration tests from the current prototype state
- CLI and query pagination tests
- adversarial fuzz tests for integer overflows, rounding, and route bounds

Required commands:

```powershell
go test ./x/dex/types ./x/dex/keeper
go test ./tests/adversarial ./tests/integration
go test ./...
```

## Migration Plan

Phase 0: harden current constant-product module. Prevent invalid active-pool
states, verify module balances and LP supply, and document production target.

Phase 1: extend params, fee accounting, and exact-out swap support without
breaking existing messages.

Phase 2: add bounded multi-hop routing and quote queries.

Phase 3: add TWAP observations and read-only oracle queries.

Phase 4: add LP locks, unlock lifecycle, burnable positions, and indexes.

Phase 5: add gauges, epochs, boosts, farming rewards, fee rewards, and claims.

Phase 6: add protocol-owned liquidity and fee distribution integration.

Phase 7: add IBC asset hooks and concentrated-liquidity interfaces.

Each phase requires a migration version, genesis validation update, invariant
coverage, and release notes.

## Readiness Checklist

- [ ] All DEX messages reject zero actors and malformed denoms.
- [ ] All protocol fees are paid only in `naet`.
- [ ] Pool reserve state reconciles with bank module balances.
- [ ] LP total supply reconciles with pool `total_shares`.
- [ ] Full active-pool withdrawal is impossible without explicit destruction.
- [ ] Multi-hop routes are bounded and deterministic.
- [ ] Exact-in and exact-out paths enforce slippage and deadlines.
- [ ] TWAP queries fail closed when observations are unavailable.
- [ ] Lock positions cannot bypass unlock rules.
- [ ] Farming emission is conserved per epoch.
- [ ] Governance params are bounded and safe by construction.
- [ ] Import/export and migrations validate all secondary indexes.
- [ ] Invariants run in tests and can run in production diagnostics.
