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
- None in v1.

Genesis state:
- `denoms`

Validation:
- Denom IDs must be canonical and collision-free.
- Admin addresses must decode with the chain address codec.
- Existing supply must match bank state when imported.

## `x/dex`

Params:
- None in v1. The pool fee constant is part of the module implementation and must not change without a security review and migration plan.

Genesis state:
- `pools`
- `next_pool_id`

Validation:
- Pool IDs must be unique and monotonic.
- Asset pairs must be canonical sorted pairs.
- Reserves and LP supply must be positive for active pools.
- `next_pool_id` must be greater than every imported pool ID.

## `x/fees`

Params:
- `allowed_fee_denoms`
- `validator_rewards_ratio`
- `community_pool_ratio`

Genesis state:
- `params`

Validation:
- v1 allows only the native base denom `norb` in `allowed_fee_denoms`.
- Validator rewards and community pool ratios must be between `0` and `1`.
- Fee split ratios must sum exactly to `1`.

## Upgrade Policy

Every module with persistent state must include a migrations package once implementation begins. Initial migrations may be no-ops, but the upgrade path must exist before testnet launch.
