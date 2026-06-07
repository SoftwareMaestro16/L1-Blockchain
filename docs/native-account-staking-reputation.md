# Native Account, Staking, Reputation, And Rent Model

This document is the current source of truth for the native wallet/account,
official liquid staking, allocation, reputation, proof, storage-rent, and
migration model.

## Native Wallet And Account Model

A native account is the chain identity record used by auth, account status,
storage-rent references, account number, sequence, metadata bounds, reputation
links, and export/import. Balances stay in the bank/native balance layer. Token
balances, NFT ownership collections, DEX positions, transaction history,
private keys, seed phrases, profile JSON, avatars, and other app data are not
stored in the native account record.

The persistent account record is versioned:

```text
Account {
  version
  address_user
  address_raw
  pubkeys
  account_number
  sequence
  status
  auth_policy
  features
  metadata
  reputation_id
  created_height
  last_active_height
  last_storage_charge_height
  storage_rent_debt
}
```

`inactive` is virtual only. Before activation, a derivable address has no
persistent state, is not exported in genesis, and does not accrue storage rent.
`active`, `frozen`, `recovered`, `archived`, and `closed` are persistent states
with explicit transition rules. `frozen` is recoverable: state, balance,
sequence, ownership links, auth policy, reputation links, pool shares,
unbondings, rewards, contract code, and contract data are preserved.

## Activation Flow

`MsgActivateAccount` is the normal first persistent state creation path.
Activation validates that the requested account address equals the address
derived from the supplied public key, that the account is not already active,
that the `AE...` and `4:...` pair roundtrips, and that account number and
sequence initialization are deterministic.

Activation is idempotency-safe. A duplicate activation for the same account is
rejected before a second persistent account is created.

```text
derive(pubkey) -> AE...
format_raw(pubkey_hash) -> 4:...
MsgActivateAccount(AE..., pubkey)
persist Account(version, address_user, address_raw, account_number, sequence)
emit AccountActivated
```

Private keys and seed phrases are never accepted in account, auth policy,
genesis, export, events, logs, or proof metadata. The docs and tests use the
terms `private-key exclusion` and `seed phrase exclusion` for this invariant.

## Address Model

User-facing account, validator, consensus, pool, contract-owner, and API
addresses are always `AE...`. The raw/internal address format is always
`4:<64 lowercase hex>` for account-like proof keys and internal store keys.

Rules:

- `AE...` is the only user-facing address family.
- `4:...` is raw/internal and proof-oriented.
- `aevaloper` and `aevalcons` are not used in user-facing APIs.
- Address derivation does not change during account migration, auth policy
  update, recovery, multisig changes, metadata changes, or staking changes.
- `AE...` address roundtrip and `4:...` raw roundtrip are invariant checks.

## Official Liquid Staking Flow

Normal users stake only through the official liquid staking path:

```text
User -> Liquid Staking Contract -> Pool Contract -> Validators
```

The user deposits AET into an official liquid staking contract and receives
pool shares or a receipt token representing a claim on the network staking
index. The user does not choose a validator. Direct user validator delegation is
disabled for the normal user path.

Direct user validator delegation is disabled.

Users buy a share of the network staking index instead of selecting a validator
because the chain can then enforce validator-set limits, power caps,
commission bounds, slashing-risk controls, and allocation fairness centrally and
deterministically. This avoids pushing validator research, cartel detection,
commission monitoring, uptime analysis, and slashing-risk modeling onto normal
wallet users.

## Staking, Pool, And Allocation Params

The following values are governance/genesis params with validation bounds, not
forever hardcoded constants:

| Param | Purpose |
| --- | --- |
| `MaxValidatorCount` | Upper bound for active validators, normally in the `100-300` range. |
| `MinValidatorStake` | Minimum validator entry stake, including the `1_000_000 AET` baseline. |
| `MinValidatorSelfStake` | Self-stake floor for solo and pool-backed validators, including the `400_000 AET` pool-backed floor. |
| `MaxPoolBackedStake` | Pool/nominator contribution cap toward validator entry, including the `600_000 AET` baseline. |
| `MinPoolDeposit` | Anti-spam minimum pool deposit, including the `10 AET` baseline. |
| `UnbondingPeriod` | Pool and validator unbonding maturity period, including the `18 days` baseline. |
| `ValidatorCommissionMin/Max` | Validator commission floor and ceiling. |
| `PoolFeeBps` | Protocol or pool fee deducted from rewards, not principal. |
| `OperatorBonusBps` | Bounded bonus for operators who carry infrastructure cost and objective performance risk. |
| `StorageRentRate` | Rent per byte/time unit for active persistent state. |
| `SystemStorageReserveRunway` | Minimum protocol-payer reserve runway before alerts or governance action. |

## Pool Shares, Rewards, And Lazy Reward Index Math

Pool deposits mint deterministic shares. The first deposit normally starts at a
fixed share price; later deposits use current pool value:

```text
minted_shares = deposit_amount * total_shares / pool_value
```

If the pool has no shares, genesis or pool params define the initial share
price. Principal accounting and reward accounting remain separate.

Rewards are lazy. The pool updates a global reward index when rewards arrive or
when a bounded maintenance action runs:

```text
reward_delta = new_rewards - validator_commission - pool_fee
reward_index += reward_delta / total_shares
user_pending = user_shares * (reward_index - user_reward_index)
```

A reward claim touches the caller's share record, the pool reward index, and a
bounded set of accounting keys. It does not scan all pool users.

Unbonding burns or locks the user's shares and creates a maturity record:

```text
unbond_amount = user_shares_to_unbond * pool_value / total_shares
release_height = current_height + UnbondingPeriod
```

Matured withdrawal is allowed only after `release_height`.

## Validator Versus Pool-User Economics

Validators and pool users have different economic exposure.

Validators provide self-stake, hardware, bandwidth, operational monitoring,
sentry architecture, incident response, upgrade execution, and objective
slashing accountability. Their gross rewards include validator commission and
any bounded operator bonus, but their net outcome is reduced by infrastructure
costs, downtime risk, jail risk, slashing losses, and insurance/reserve policy.

Pool users provide capital through pool shares. They receive pro-rata rewards
after validator commission and pool fee deductions. They inherit slashing losses
through share value, but they do not receive validator commission and they do
not operate validator infrastructure. Pool fee and reserve payer policy must be
visible so pool users can compare expected net yield and risk.

Validator economics include infrastructure costs.

Slashing losses reduce pool value before user share redemption:

```text
pool_value_after_slash = pool_value_before_slash - slashed_amount
share_price_after_slash = pool_value_after_slash / total_shares
```

## Allocation Engine Math

The allocation engine assigns pooled stake to validators using deterministic
integer math. Inputs are objective or governance-bounded:

- reputation;
- uptime;
- commission;
- allocation limits;
- stake efficiency;
- slashing risk;
- network load.

Example score shape:

```text
score_i =
  reputation_weight
  + uptime_weight
  + commission_weight
  + limit_weight
  + stake_efficiency_weight
  - slashing_risk_penalty
  - network_load_penalty

weight_i = score_i / sum(score)
allocation_i = pool_available_stake * weight_i
```

Jailed validators receive no new positive allocation. Validators above power
cap or pool-backed contribution limits receive no additional effective power.
Rebalances are bounded and deterministic; BeginBlock/EndBlock must not scan all
pool users.

## Stake Reputation Math

Stake reputation is account-owned and non-transferable. It comes from pool
share exposure over time, not manual validator selection and not token/NFT
transfer.

```text
stake_time_delta = effective_pool_shares * elapsed_epochs
reputation_delta = f(stake_time_delta, risk_adjustment, caps)
```

Reputation cannot increase without stake-time. Jailed or slashed validator
bonus paths cannot add positive validator bonus. Reputation claims are lazy and
touch only bounded account/pool/reputation keys.

## Proof Model

Proof metadata is deterministic and secret-free. It identifies the state object
being proven without exposing private keys or seed phrases:

```text
ProofMetadata {
  height
  module
  store_key
  state_key
  root
  path
  version
}
```

Proof surfaces cover account activation, pool deposit, pool share state,
allocation update, reward claim, unbonding, and stake reputation claim. Proof
queries use direct keys or bounded pagination.

## Storage Rent Model

Storage rent applies to every active account and persistent on-chain state:

- active native wallets;
- smart contracts;
- token/NFT/DEX contract state;
- official liquid staking pool contracts and pool accounting records;
- domain records;
- staking/reputation records that create long-lived state;
- protocol/module state through protocol-payer accounting.

Storage rent does not apply to unactivated `AE...` addresses, empty/no-state
accounts, deleted/closed accounts after state removal, or transient tx/event
data that is not persistent state.

```text
storage_size = code_bytes + data_bytes
storage_rent_delta = storage_size * rent_rate_per_byte_second * elapsed_seconds
effective_fee = gas_fee + storage_rent_delta + unpaid_storage_debt
```

Rent is collected lazily during account, contract, pool, claim, migration, or
explicit maintenance actions. Normal accounts and contracts pay from their own
balance or configured payer. Long-lived records without spendable balance use
the owning account, contract, pool reserve, module balance, or
governance-configured protocol payer.

## Official Pool Rent And `frozen_limited`

The official liquid staking pool has protocol safety obligations. Storage rent
for official pool state is paid from the pool fee/reserve payer policy or a
configured protocol payer before user freeze processing.

If the official pool cannot pay rent, it must not become a fund trap. It enters
`frozen_limited` instead of normal `frozen`:

- new deposits and state growth are blocked;
- claims, unbond requests, matured withdrawals, top-up, debt payment,
  unfreeze, proof queries, and governance recovery remain available;
- user shares, active stake, unbondings, allocation records, pending rewards,
  proof metadata, and pool balances remain intact.

## System And Protocol Storage-Rent Accounting

Protocol-critical state must remain executable during system rent underfunding.
System rent uses protocol-payer accounting:

- deterministic system rent top-up runs before user freeze processing;
- fee collector, treasury, or configured reserve payer funds protocol state;
- system storage reserve runway is checked against governance params;
- low runway raises an invariant/alert before protocol-critical execution is
  disabled;
- protocol-critical modules are not frozen, archived, or deleted because of
  storage rent debt.

## Auth Policy Modes

Auth policy changes do not change `AE...` or `4:...` addresses. Public keys and
signature rules may change; private keys, seed phrases, SMS secrets, and TOTP
secrets are never stored on-chain.

Supported policy modes:

- `single_key`: one configured public key authorizes ordinary actions.
- `multisig`: a set of public keys signs together.
- `threshold`: at least N configured signers are required.
- `weighted`: signer weights must sum to the configured threshold.
- `two_device`: primary key plus device key for protected operations.
- `timelock`: protected updates become executable only after a delay.
- `recovery`: recovery key or policy moves the account to `recovered` after
  required authorization.
- `spending_limits`: small transfers can use weaker authorization while large
  transfers, staking changes, and auth updates require stronger policy.

## Upgrade And Migration Model

Accounts and related state are versioned. Normal block execution uses lazy
migration: loading an old account can migrate only that account and write it
back. It does not require a full account scan.

lazy migration.

Upgrade handlers register migration functions and optional bounded batched
jobs. Batched migration jobs carry a deterministic cursor, can resume safely,
and must not skip or duplicate accounts.

Compatibility cannot change without an explicit backward-compatibility plan:

- address derivation;
- `AE...` user-facing format;
- `4:...` raw format;
- ownership keys;
- staking ownership keys;
- sequence semantics;
- signature domain.

Allowed migration fields include auth policy, recovery policy, internal message
rules, account features, metadata, reputation links, and staking capabilities.

Implemented test names that guard this model include:

- `TestLazyMigrationPreservesExistingAccountAndTouchesSingleKey`;
- `TestMigrateAccountV1ToV2DeterministicGolden`;
- `TestAddressAndSequenceSemanticsUnchangedAcrossMigration`;
- `TestBatchedMigrationResumesSafelyWithoutSkipOrDuplicate`;
- `TestFullGenesisExportOrderDeterministicAcrossRepeatedRuns`;
- `TestNativeAccountInvariantRegistryIncludesEveryRequiredInvariant`.

## Native Versus Contract Boundary

Wallet/account identity, activation, auth policy, storage-rent integration,
validator registry, official liquid staking admission, allocation metadata,
stake reputation, proof metadata, and protocol/system accounting are native
because they protect consensus, bank accounting, slashing/rent safety, and
bounded block lifecycle behavior.

Tokens, NFTs, DEX pools, markets, auctions, and application-specific assets are
contracts or contract standards. They are not new native `x/` asset modules.
Historical native tokenfactory or native DEX prototype docs are migration
references only; production token/NFT/DEX design is contract-only through AVM
standards such as AFT-44 and ANFT-66.

Tokens, NFTs, DEX pools, markets, auctions, and application-specific assets are contracts or contract standards.
Historical native tokenfactory or native DEX prototype docs are migration references only.
