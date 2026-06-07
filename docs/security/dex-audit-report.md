> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# x/dex Security Audit Report

Date: 2026-06-06

Scope:

- `x/dex/types`
- `x/dex/keeper`
- `proto/l1/dex/v1`
- DEX-related app genesis validation
- fee ante address policy that can affect DEX and standard SDK messages
- prototype keeper wiring that can affect export/import and future runtime
  integration

## Executive Summary

The current `x/dex` module is a deterministic constant-product AMM prototype. It
is not yet the production-grade DEX described in the Aetra target
architecture: there is no multi-hop execution, exact-out swap, LP locks,
farming, TWAP oracle, protocol-owned liquidity, or DEX-specific protocol-fee
distribution.

The audit found and fixed four correctness issues that could affect export/import
safety, address policy, cross-module accounting, and prototype runtime state.

## Fixed Findings

### P1: Full LP Withdrawal Could Leave an Invalid Active Pool

Status: fixed.

Impact:

`MsgRemoveLiquidity` previously allowed `shares == total_shares`. That wrote
zero reserves and zero LP supply into an active pool. DEX genesis/export
validation requires positive reserves and positive shares, so a user owning all
LP shares could make the exported state fail import.

Fix:

`MsgRemoveLiquidity` now rejects full active-pool withdrawal. Pool destruction
must be implemented as an explicit future operation that atomically deletes pool
state and pair indexes.

Coverage:

- `TestRemoveLiquidityRejectsFullPoolWithdrawal`
- `go test -count=1 ./x/dex/types ./x/dex/keeper`

### P2: Zero-Address Policy Missed SDK Non-Signer Address Fields

Status: fixed for the identified SDK path.

Impact:

Aetra policy rejects zero addresses for signers, admins, recipients, and
authorities. The fee ante decorator already covered fee payer, tx signers, bank
send, and bank multisend. Standard SDK messages can also contain non-signer
address fields. The concrete audited gap was
`cosmos.distribution.v1beta1.MsgSetWithdrawAddress.WithdrawAddress`.

Fix:

The fee ante address policy now validates both `DelegatorAddress` and
`WithdrawAddress` in `MsgSetWithdrawAddress` with Aetra zero-address
rejection before fee admission and before SDK handler execution.

Coverage:

- `TestAnteHandlerDecoratorFeePolicy`
- `go test -count=1 ./x/fees/keeper`

Remaining work:

Continue enumerating newly enabled SDK messages as modules are added. Any
message with non-signer address fields must be explicitly covered or routed
through a zero-rejecting codec/verifier.

### P2: DEX Genesis Did Not Reconcile Pools With Bank State

Status: fixed at app-level genesis validation.

Impact:

DEX-local validation checked pool structure, positive reserves, LP denom shape,
and unique pairs, but did not prove that the `dex` module account actually held
the declared reserves or that bank supply for each LP denom matched
`pool.total_shares`.

Fix:

`validateAetraGenesis` now performs app-level DEX/bank reconciliation:

- sums expected reserves for all DEX pools by asset denom
- compares those sums with the `dex` module account bank balances
- compares each `lp/{pool_id}` bank supply with `pool.total_shares`

This is app-level rather than DEX-local because only the app has both DEX
genesis and bank genesis at validation time.

Coverage:

- `dex reserve module balance mismatch`
- `dex LP supply mismatch`
- `go test -count=1 ./app`

### P3: Prototype Persistent Keepers Could Diverge From Runtime Memory

Status: mitigated.

Impact:

Several disabled prototype modules registered persistent stores, but their
operational methods read and mutated `k.genesis` in memory. The module manager
held copies of those keepers, so genesis import could populate a module copy
while `app.*Keeper` runtime methods still saw default memory state.

Fix:

Prototype `AppModule` wrappers now hold pointers to the actual app keeper
instances. Their `InitGenesisState` methods synchronize in-memory state and KV
snapshots. Their export path returns in-memory runtime state when it differs
from default, while still supporting restart export from the KV snapshot.

Coverage:

- `TestPrototypeGenesisInitializesRuntimeKeeperState`
- `TestAetraCorePrototypeStateSurvivesRestartWhenDisabled`
- prototype keeper package tests

Residual limitation:

These modules are still prototype/non-runtime surfaces. Most runtime methods do
not accept `ctx`, so a fully production KV-backed state machine requires a
larger keeper redesign before enabling them for live transactions.

## DEX-Specific Audit Results

### Determinism

Result: pass for current DEX keeper state transitions.

No consensus-critical use of wall-clock time, randomness, goroutines, channels,
float math, or unordered map iteration was found in current DEX create/add/remove
liquidity or swap paths.

### Authorization And Address Policy

Result: mostly pass for current DEX messages.

Current DEX actors are validated with Aetra non-zero user address parsing:

- pool creator
- liquidity depositor
- liquidity withdrawer
- swap trader
- params authority

`MsgUpdateParams` also checks the configured authority.

### Denom Safety

Result: pass for current pool assets and LP denoms.

Pool assets use SDK denom validation, reject zero factory denom admins, and
reject native AET/naet spoofing through display aliases or factory subdenoms.
LP denoms are module-controlled as `lp/{pool_id}` and are not accepted from user
input during pool creation.

### Bank Accounting

Result: improved.

Runtime operations use cache-context atomicity for multi-step bank and pool state
updates. Tests cover insufficient funds, corrupted pool state, reserve/module
balance invariant, LP supply invariant, and swap constant-product invariant.

Genesis/import now reconciles DEX pools with bank balances and LP supply at app
validation.

### Integer Math And Rounding

Result: acceptable for prototype, incomplete for production.

Swap math uses integer arithmetic only. Exact-in output floors the result and
tiny inputs that round to zero are rejected by slippage/output positivity checks.
Current coverage checks fee application, monotonicity, tiny-input rounding, and
constant-product non-decrease.

Production DEX still needs exact-out ceil math, multi-hop rounding policy, route
slippage accounting, and fuzz/property tests across reserves, fees, and path
lengths.

### Panic / Chain-Halt Paths

Result: improved.

`InitGenesis` no longer panics if pair-index write fails; it returns an error.
Module-level genesis functions still panic on invalid genesis as required by the
Cosmos SDK module interface after validation has failed.

### Query And Pagination

Result: acceptable for prototype.

Pool queries use bounded pagination helpers. There is no unbounded transaction
route search because routing is not implemented yet.

## Remaining Production Gaps

These are not regressions in the current prototype, but they are blockers for a
production-grade DEX:

- no multi-hop swaps or route execution
- no exact-out swaps
- no deadline field on swaps
- no TWAP oracle or manipulation checks
- no LP lock positions
- no farming gauges, epochs, boosts, or early unlock penalties
- no DEX protocol-fee split engine
- no protocol-owned liquidity
- no explicit pool destruction message
- no IBC asset interface
- no concentrated-liquidity interface
- no DEX invariant registration for production diagnostics
- no fuzz suite for swap and liquidity math

The target design and migration plan are documented in
`docs/architecture/dex-production-spec.md`.

## Recommended Next Steps

1. Add explicit `MsgDestroyPool` or keep full withdrawal permanently forbidden.
2. Add exact-out swap math with ceil input calculation and property tests.
3. Add bounded multi-hop route execution with deadline and route-level slippage.
4. Register runtime DEX invariants for reserve and LP supply reconciliation.
5. Add TWAP observations before exposing oracle-dependent routing or pricing.
6. Add lock/farming state only after LP transfer/escrow semantics are finalized.
