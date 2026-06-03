# Module Boundaries

## `x/tokenfactory`

Purpose: create and manage custom denoms without EVM dependency.

State:
- Denom registry keyed by full denom.
- Admin record per denom.
- Optional metadata record per denom.
- Module params.

Minimal Msg surface:
- `MsgCreateDenom`
- `MsgMint`
- `MsgBurn`
- `MsgChangeAdmin`
- `MsgUpdateParams`

Keeper dependencies:
- Bank keeper interface for mint, burn, send, and metadata operations.
- Account/address codec where required by the scaffolded SDK version.

Security invariants:
- Only authorized admins can mint, burn, or transfer admin rights.
- Total supply changes must match bank keeper mint/burn results.
- Denom names and metadata lengths must be param-limited.

## `x/dex`

Purpose: deterministic constant-product AMM.

State:
- Pool registry keyed by pool ID.
- Asset pair index.
- LP share accounting.
- Fee accumulator references.
- Module params.

Minimal Msg surface:
- `MsgCreatePool`
- `MsgAddLiquidity`
- `MsgRemoveLiquidity`
- `MsgSwapExactIn`
- `MsgUpdateParams`

Keeper dependencies:
- Bank keeper interface for escrow, pool balances, and LP share movement.
- Fees keeper interface for protocol fee accounting.

Security invariants:
- Integer math only.
- No pool operation can create value.
- LP shares must remain backed by pool reserves.
- User-provided min-out and deadlines protect against stale execution and slippage.

## `x/fees`

Purpose: centralize protocol fee policy and distribution.

State:
- Fee collector module account reference.
- Distribution weights.
- Accrued fee records where needed.
- Module params.

Minimal Msg surface:
- `MsgUpdateParams`
- Future fee claim/distribution messages only if they cannot be handled by hooks.

Keeper dependencies:
- Bank keeper interface for balances and transfers.
- Distribution or auth module interfaces only when explicitly required.

Security invariants:
- Distribution weights must sum to the configured denominator.
- Governance authority controls params.
- Fee collection must be idempotent for repeated block execution inputs.

## `x/bridge`

Purpose: future interoperability. It remains out of scope for the first scaffold.

Activation requires a separate design covering light-client verification, replay domains, validator or relayer trust assumptions, finality, rate limits, and emergency controls.
