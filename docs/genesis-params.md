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
- `denom_creation_fee`
- `max_denom_name_len`
- `max_metadata_len`
- `minting_enabled`
- `burning_enabled`

Genesis state:
- `params`
- `denoms`
- `admins`
- `metadata`

Validation:
- Denom IDs must be canonical and collision-free.
- Admin addresses must decode with the chain address codec.
- Existing supply must match bank state when imported.

## `x/dex`

Params:
- `swap_fee_bps`
- `protocol_fee_bps`
- `max_fee_bps`
- `min_initial_liquidity`
- `max_pools`
- `pool_creation_fee`

Genesis state:
- `params`
- `pools`
- `lp_positions`
- `next_pool_id`

Validation:
- Pool IDs must be unique and monotonic.
- Asset pairs must be canonical sorted pairs.
- Reserves and LP supply must be positive for active pools.
- Fee values must be bounded by `max_fee_bps`.

## `x/fees`

Params:
- `collector_module_account`
- `distribution_weights`
- `min_distribution_amount`
- `enabled`

Genesis state:
- `params`
- `accrued_fees`

Validation:
- Distribution weights must sum exactly to the configured denominator.
- Collector account must be a valid module account.
- Disabled fees must not leave partially active distribution routes.

## Upgrade Policy

Every module with persistent state must include a migrations package once implementation begins. Initial migrations may be no-ops, but the upgrade path must exist before testnet launch.
