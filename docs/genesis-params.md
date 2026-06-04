# Genesis And Params Model

## Global Defaults

All chain constants that affect state transitions must be represented in genesis or module params. No consensus-critical value should be hardcoded in keeper logic.

Initial Orbitalis chain choices:
- Address prefix: `orb`
- Native base denom: `norb`
- Display denom: `ORB`
- Native token name: `Orbitalis`
- Native token decimals: `9`
- Governance authority: the `x/gov` module account or SDK authority configured at genesis.

## `x/tokenfactory`

Params:
- `min_subdenom_length`
- `max_subdenom_length`
- `denom_creation_enabled`
- `minting_enabled`
- `burning_enabled`

Genesis state:
- `params`
- `denoms`

Validation:
- Denom IDs must be canonical factory denoms and collision-free.
- Admin addresses must decode with the chain address codec.
- Governance can disable creation, mint, or burn as emergency controls.
- Governance cannot transfer user denom admin rights; admin lifecycle stays per denom.

## `x/dex`

Params:
- `swap_fee_bps`
- `max_swap_fee_bps`
- `pool_creation_enabled`
- `swaps_enabled`
- `liquidity_enabled`
- `min_initial_liquidity`

Genesis state:
- `params`
- `pools`
- `next_pool_id`

Validation:
- Pool IDs must be unique and monotonic.
- Asset pairs must be canonical sorted pairs.
- Reserves and LP supply must be positive for active pools.
- Fee values must be bounded by `max_swap_fee_bps`, which itself is capped at the immutable v1 bound.
- Governance can disable pool creation, swaps, or liquidity operations as emergency controls.
- Governance cannot mutate pool reserves, LP denom format, pool ownership, or user balances.

## `x/fees`

Params:
- `allowed_fee_denoms`
- `validator_rewards_ratio`
- `community_pool_ratio`
- `min_fee_amount`
- `fee_collector_module`
- `validator_rewards_target`
- `community_pool_target`

Genesis state:
- `params`
- `protocol_fee_state`

Validation:
- v1 allows only `norb` fees.
- Minimum fee must be positive.
- Split ratios must be decimals between `0` and `1` and sum exactly to `1`.
- Fee collector and target routes are explicit and fixed to safe v1 module routes.
- Protocol fee accounting must satisfy `total_collected == validator_rewards + community_pool`.

## Upgrade Policy

Every module with persistent state must include a migrations package once implementation begins. Initial migrations may be no-ops, but the upgrade path must exist before testnet launch.

## Custom Module Governance Matrix

| Module | Governance-managed params | Immutable in v1 | Emergency controls |
| --- | --- | --- | --- |
| `x/fees` | allowed fee denom set fixed to `norb`, fee split ratios, minimum fee amount | fee collector route and distribution targets | raise/lower minimum fee within bounds |
| `x/dex` | swap fee bps, operation enable flags, minimum initial liquidity | LP denom scheme, pair ordering, pool reserves, user balances, max fee cap | disable pool creation, swaps, or liquidity |
| `x/tokenfactory` | subdenom length bounds and operation enable flags | existing factory denom IDs and per-denom admin ownership | disable denom creation, minting, or burning |

Governance proposals must submit complete params objects. Missing legacy genesis params are normalized to safe defaults during import, but runtime updates preserve explicit false booleans so emergency disables cannot be silently re-enabled.
