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

## Governance Genesis Params

The governance parameter catalog is represented by `DefaultGovernanceParameterSpecs` and `DefaultGovernanceGenesisParams` in `app/params/governance_parameters.go`. Genesis validation must reject missing, unknown, duplicate, or out-of-bounds governance params before launch.

| Param key | Module/category | Default | Validation rule |
| --- | --- | ---: | --- |
| `validator_set_size` | staking policy | `100` | Must stay within `100-300` unless a local/testnet override explicitly validates a smaller count outside normal genesis governance params. |
| `validator_entry_stake_naet` | staking policy | `1_000_000 AET` | Must equal `1_000_000 AET` at genesis. |
| `pool_backed_validator_self_stake_naet` | staking policy | `400_000 AET` | Pool-backed validator self-stake side of the `400_000/600_000` split. |
| `pool_backed_validator_pool_stake_naet` | staking policy | `600_000 AET` | Pool-backed pool/nominator side of the `400_000/600_000` split; split must sum to validator entry stake. |
| `liquid_staking_pool_min_deposit_naet` | staking policy | `10 AET` | Small-user pool deposit floor; users do not need validator-sized stake. |
| `direct_user_validator_delegation` | staking policy | `disabled` | Normal user delegation to validators is disabled; staking flows through official liquid staking pools. |
| `staking_unbonding_blocks` | staking policy | `18 days` | Must equal the 18-day unbonding period in blocks. |
| `commission_floor_bps` | staking policy | `500` | Bounded commission floor. |
| `commission_max_bps` | staking policy | `2000` | Bounded validator commission ceiling. |
| `commission_max_change_bps` | staking policy | `100` | Bounded daily commission change. |
| `min_tx_fee_naet` | fees | `0.003 AET` | Must equal `3_000_000 naet` at genesis. |
| `fee_burn_share_bps` | economics | `5000` | Fee split part; burn + rewards + treasury must sum to `10000 bps`. |
| `fee_reward_share_bps` | economics | `3500` | Fee split part; burn + rewards + treasury must sum to `10000 bps`. |
| `fee_treasury_share_bps` | economics | `1500` | Fee split part; burn + rewards + treasury must sum to `10000 bps`. |
| `storage_rent_rate_per_byte_second_naet` | storage rent | `1` | Active persistent state rent baseline. |
| `system_storage_reserve_min_runway_days` | storage rent | `365` | Must be greater than or equal to warning runway. |
| `system_storage_reserve_warning_runway_days` | storage rent | `180` | Must be between minimum and critical runway. |
| `system_storage_reserve_critical_runway_days` | storage rent | `90` | Must be positive and no greater than warning runway. |

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

The tracked local profile is `aetra-local-1`. Its operator-facing genesis, account, validator, endpoint, and audit contract is defined in [Aetra Local Bootstrap Profile](bootstrap-profile.md).

## Upgrade Policy

Every module with persistent state must include a migrations package once implementation begins. Initial migrations may be no-ops, but the upgrade path must exist before testnet launch. See [upgrade-migrations.md](upgrade-migrations.md) for the dry-run and future migration checklist.
