# DEX Invariant Hardening Audit

Date: 2026-06-04

Scope:

- `x/dex/keeper`
- `x/dex/types`
- `proto/l1/dex/v1`

## Findings

| Risk | Severity | Scenario | Resolution |
| --- | --- | --- | --- |
| Duplicate pair pools | High | A second pool for the same canonical pair fragments liquidity and lets users hit the wrong price surface. | Added pair index keyed by canonical denom pair and reject duplicate `CreatePool`. |
| Reserve desync | Critical | If module balances diverge from stored reserves, swaps can pay more than the module holds or commit incorrect reserves. | Added O(1) per-pool accounting checks against module balances before and after state transitions. |
| LP supply desync | Critical | If LP token supply differs from `pool.total_shares`, remove/add liquidity calculations can inflate or steal value. | Added LP supply checks against bank supply before and after liquidity transitions. |
| Rounding leak on add liquidity | Medium | Unbalanced deposits mint shares from the smaller side while silently donating the larger side, creating ambiguous accounting. | Require deposits to match the pool ratio exactly and reject zero-share rounding. |
| Swap rounding and tiny input | Medium | Very small inputs can round effective input/output to zero and bypass user expectations. | Reject swaps whose fee-adjusted input or output rounds to zero. |
| Constant-product regression | Critical | A malformed math path could lower `x*y` and leak value from the pool. | Added constant-product non-decrease checks for every swap. |
| Zero/insolvent pools | High | Full withdrawal or corrupted reserves can leave a pool with zero reserve/share state. | Reject full pool drain and validate positive reserves/shares before operations. |
| Panic-prone store reads | Medium | Corrupted store bytes could panic via `MustUnmarshal`. | Keeper read paths now return codec errors explicitly. |

## Runtime Complexity

- `CreatePool` duplicate-pair lookup is O(1) through the pair index.
- `AddLiquidity`, `RemoveLiquidity`, and `SwapExactAmountIn` perform O(1) accounting checks for the touched pool only.
- No full pool scan was added to consensus-critical message execution.

## Residual Risk

- `Query/Pools` and `ExportGenesis` still iterate all pools. They are not called during swaps, but public pool queries should receive pagination in a separate RPC-hardening task.
- Existing stores created before the pair index need migration or genesis export/import to populate the pair index before production upgrade.
- Initial LP share minting remains conservative with `min(token0, token1)` for compatibility; a future economics change could switch to integer square root after migration tests.
