> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Genesis And Params Model

## Global Defaults

All chain constants that affect state transitions must be represented in genesis or module params. No consensus-critical value should be hardcoded in keeper logic.

Initial Aetra chain choices:
- Address prefix: `orb`
- Native base denom: `naet`
- Display denom: `AET`
- Native token name: `Aetra`
- Native token decimals: `9`
- Governance authority: the `x/gov` module account or SDK authority configured at genesis.

## `x/tokenfactory`

Genesis state:
- `denoms`

Validation:
- Denom IDs must be canonical factory denoms and collision-free.
- Admin addresses must decode with the chain address codec.
- Denoms must use the `factory/<admin>/<subdenom>` prefix.
- Subdenoms must not directly spoof native names: `naet`, `AET`, or `Aetra`.
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
- The only prototype fee denom is `naet`.
- V1 fee policy requires exactly one allowed denom; empty, duplicate, or multi-denom lists are invalid.
- Validator rewards ratio and community pool ratio must be valid decimals between `0` and `1`.
- Fee split ratios must sum exactly to `1`.
- Prototype default params are `allowed_fee_denoms = ["naet"]`, `validator_rewards_ratio = "0.98"`, and `community_pool_ratio = "0.02"`.

## Local Bootstrap Profile

The tracked local profile is `aetheris-local-1`. Its operator-facing genesis, account, validator, endpoint, and audit contract is defined in [Aetra Local Bootstrap Profile](bootstrap-profile.md).

## Upgrade Policy

Every module with persistent state must include a migrations package once implementation begins. Initial migrations may be no-ops, but the upgrade path must exist before testnet launch. See [upgrade-migrations.md](upgrade-migrations.md) for the dry-run and future migration checklist.
