# Module Boundaries

## `x/tokenfactory`

Purpose: create and manage custom denoms without EVM dependency.

State:
- Denom registry keyed by full denom.
- Admin record per denom.
- Bank metadata is written through `x/bank` when a denom is created.

Minimal Msg surface:
- `MsgCreateDenom`
- `MsgMint`
- `MsgBurn`
- `MsgChangeAdmin`

Keeper dependencies:
- Bank keeper interface for mint, burn, send, and metadata operations.

Security invariants:
- Only authorized admins can mint, burn, or transfer admin rights.
- Total supply changes must match bank keeper mint/burn results.
- Denom names must pass factory denom validation.

## `x/dex`

Purpose: deterministic constant-product AMM.

State:
- Pool registry keyed by pool ID.
- Next pool ID counter.
- LP share accounting through minted bank coins.

Minimal Msg surface:
- `MsgCreatePool`
- `MsgAddLiquidity`
- `MsgRemoveLiquidity`
- `MsgSwapExactAmountIn`

Keeper dependencies:
- Bank keeper interface for escrow, pool balances, LP share mint/burn, and account transfers.

Security invariants:
- Integer math only.
- No pool operation can create value.
- LP shares must remain backed by pool reserves.
- User-provided min-out and min-share fields protect against slippage.

## `x/fees`

Purpose: centralize protocol fee policy and distribution.

State:
- Module params.
- Allowed fee denoms.

Minimal Msg surface:
- `MsgUpdateParams`
- Future fee claim/distribution messages only if they cannot be handled by hooks.

Keeper dependencies:
- Governance authority string for parameter updates.

Security invariants:
- Governance authority controls params.
- Fee denom policy must remain explicit and deterministic.

## `x/bridge`

Purpose: future interoperability. It remains out of scope for the first scaffold.

Activation requires a separate design covering light-client verification, replay domains, validator or relayer trust assumptions, finality, rate limits, and emergency controls.
