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

Genesis state:
- `denoms`

Validation:
- Denom IDs must be canonical and collision-free.
- Admin addresses must decode with the chain address codec.
- Denoms must use the `factory/<admin>/<subdenom>` prefix.
- Subdenoms must not directly spoof native names: `norb`, `ORB`, or `Orbitalis`.
- Prototype default genesis starts with no factory denoms.

## `x/dex`

Genesis state:
- `next_pool_id`
- `pools`

Validation:
- Pool IDs must be unique and monotonic.
- Asset pairs must be canonical sorted pairs.
- Reserves and LP supply must be positive for active pools.
- LP denom must match `lp/<pool_id>`.
- Prototype default genesis starts with `next_pool_id = 1` and no pools.

## `x/fees`

Genesis state:
- `params`

Validation:
- The only prototype fee denom is `norb`.
- Validator rewards ratio and community pool ratio must be valid decimals between `0` and `1`.
- Fee split ratios must sum exactly to `1`.
- Prototype default params are `allowed_fee_denoms = ["norb"]`, `validator_rewards_ratio = "0.98"`, and `community_pool_ratio = "0.02"`.

## Local Bootstrap Profile

The tracked local profile is `orbitalis-local-1`. Its operator-facing genesis, account, validator, endpoint, and audit contract is defined in [Orbitalis Local Bootstrap Profile](bootstrap-profile.md).

## Upgrade Policy

Every module with persistent state must include a migrations package once implementation begins. Initial migrations may be no-ops, but the upgrade path must exist before testnet launch.
