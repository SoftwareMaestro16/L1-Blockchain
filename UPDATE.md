# Aetra Native Wallet, Staking, Reputation Update Plan

## [ALL CHATS] 0. Purpose

This file is a local implementation backlog for adding native wallet/account support and native staking/reputation behavior to Aetra. It is intentionally not tracked by git, like the root `architecture.md`.

The goal is not to build a wallet smart contract. The goal is to make the wallet/account a native chain identity, auth, staking, storage-rent, and smart-contract endpoint:

- user-facing account, validator, and consensus addresses are `AE...`;
- raw `4:...` addresses remain an internal/system format;
- no `aevaloper` or `aevalcons` user-facing formats;
- private key and seed phrase are never stored on-chain;
- balances remain in the bank/native balance layer;
- tokens, NFTs, DEX pools, and app-specific assets are smart contracts or registries, not new `x/` asset modules;
- user staking is officially exposed as `User -> Liquid Staking Contract -> Pool Contract -> Validators`;
- the user buys a share of the network staking index and must not be forced to pick a concrete validator;
- native wallet/account, validator registry, staking policy, allocation proofs, reputation, auth policy, storage rent, and protocol-critical accounting are native state.

Every task below must ship with tests. A task is incomplete if it has only implementation code and no test proving behavior, security, determinism, or export/import stability.

## [CHAT 1] 1. Repository Baseline And Guardrails

### Task 1.1 - Map Existing Boundaries Before Coding

Implementation:

- Document current ownership boundaries for `app/addressing`, `x/identity`, `x/reputation`, `x/storage-rent`, `x/pos`, `x/nominator-pool`, `x/single-nominator-pool`, `x/validator-*`, `x/stake-concentration`, `x/fees`, `x/burn`, `x/treasury`, and contract/VM modules.
- Decide whether the native account model lands in an existing module or a new bounded module, for example `x/account` or `x/native-account`.
- Keep Cosmos SDK `x/auth`, `x/bank`, `x/staking`, `x/slashing`, and `x/distribution` invariants intact unless a later migration explicitly replaces them with compatibility tests.
- Do not add large catch-all files. Split by types, keys, keeper, messages, queries, genesis, migration, ante/auth, events, and tests.

Tests:

- Architecture test listing the new module boundary and rejected cross-module writes.
- Test that token/NFT/DEX behavior remains contract-routed and is not reintroduced as native asset modules.
- Test that app module order and keeper wiring still include required SDK modules and existing custom modules.

Definition of done:

- A short docs entry or package-level spec says which module owns account state, auth policy, storage-rent integration, staking reputation, and proof metadata.
- `go test ./...` passes before starting deeper feature work.

## [CHAT 1] 2. Addressing And Wallet Identity

### Task 2.1 - Freeze Address Format Compatibility

Implementation:

- Treat `AE...` as the only user-facing account, validator, and consensus address format.
- Keep `4:...` raw format stable for internal/system usage and proof keys.
- Do not change address derivation in migrations or account upgrades.
- Add a typed helper for deriving account addresses from public keys, with stable `AE...` and `4:...` roundtrip conversion.
- Reject `aevaloper`, `aevalcons`, old raw prefixes, mixed-case raw addresses, malformed `AE...`, and foreign Bech32 in user-facing APIs.

Tests:

- `AE...` account address roundtrip.
- `AE...` validator address roundtrip.
- `AE...` consensus address roundtrip.
- `4:...` raw address roundtrip.
- `AE... <-> 4:...` roundtrip is stable.
- `derive(pubkey)` equals requested activation address.
- `aevaloper` and `aevalcons` are rejected at message validation and query parsing boundaries.
- Address derivation golden vectors cannot change between versions.

Definition of done:

- Address derivation has golden tests.
- Existing `app/addressing` tests still pass.
- User-facing CLI/API examples use `AE...` only.

### Task 2.2 - Virtual Account Model

Implementation:

- Define the difference between virtual inactive accounts and active on-chain accounts.
- A virtual account exists only as a derivable address before activation and must not create state or storage rent.
- `MsgActivateAccount` creates the first persistent account state.
- Inactive `AE...` addresses must not be exported in genesis unless they were activated.

Tests:

- Querying an unactivated address returns a virtual/inactive result without persisted state.
- Export genesis does not include unactivated addresses.
- Storage rent does not accrue for unactivated addresses.
- Sending a non-activation tx from an inactive account is rejected.

Definition of done:

- Query behavior is explicit and documented.
- No keeper path creates persistent account state except activation or a controlled migration.

## [CHAT 1] 3. Native Account State

### Task 3.1 - Add Versioned Account State

Implementation:

Create a versioned native account record:

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

Allowed statuses:

- `inactive`: virtual only, no persistent state;
- `active`: persistent account state exists;
- `frozen`: recoverable state, balance, ownership, code, and data are preserved, but normal spending/execution is blocked until policy or storage debt issue is resolved;
- `recovered`: account was recovered through recovery policy;
- `archived`: minimal state retained, optional fields pruned;
- `closed`: state removed only after strict zero-obligation checks.

Do not store:

- private key;
- seed phrase;
- all token balances;
- all NFTs;
- all domains;
- transaction history;
- off-chain profile JSON;
- avatars or large metadata blobs.

Metadata should be minimal and optional:

```text
metadata_hash
display_name_hash
domain_alias
created_height
```

Tests:

- Account validation accepts a complete active account.
- Account validation rejects private key, seed, or secret-like serialized fields.
- Account validation rejects empty `AE...`, malformed raw address, mismatched address pair, unsupported status, and unsupported version.
- Export/import preserves account state exactly.
- Metadata size and fields are bounded.
- Balances are not duplicated in account state.

Definition of done:

- Account state is deterministic, versioned, and exportable.
- Private key/seed cannot appear in JSON, protobuf, events, logs, or genesis export.

### Task 3.2 - Store Keys And Queries

Implementation:

- Add deterministic prefix keys:
  - `account/by_user/{AE... or canonical bytes}`;
  - `account/by_raw/{4:... or canonical raw bytes}`;
  - `account/number/{account_number}`;
  - `account/reputation/{reputation_id}`;
  - `account/storage/{account}`;
- Add paginated queries:
  - `Account`;
  - `Accounts`;
  - `AccountByRawAddress`;
  - `AccountReputation`.
- Queries must not scan all accounts unless explicitly paginated by prefix with a bounded page size.

Tests:

- State keys are deterministic golden values.
- `Account` query finds active account by `AE...`.
- `AccountByRawAddress` query returns the same account as `Account`.
- `Accounts` query is paginated and deterministic.
- Querying unsupported address formats fails safely.

Definition of done:

- Every query has pagination or a single deterministic key lookup.
- Store prefixes are documented and covered by golden tests.

## [CHAT 1] 4. Account Activation

### Task 4.1 - Implement MsgActivateAccount

Implementation:

Add `MsgActivateAccount` with validation:

- address must equal `derive(pubkey)`;
- account must not already be active;
- account number and sequence initialization must be deterministic;
- activation fee must be validated by fee policy;
- address pair `AE...` and `4:...` must be stable;
- emit deterministic `AccountActivated`.

Initial sequence policy:

- choose and document whether sequence starts at `0` or `1`;
- keep compatibility forever or add explicit migration compatibility.

Tests:

- Activation success.
- Duplicate activation rejected.
- `address != derive(pubkey)` rejected.
- Activation with malformed `AE...` rejected.
- Activation with malformed `4:...` rejected.
- Activation fee under minimum rejected.
- Sequence starts deterministic.
- Account number assignment deterministic.
- `AccountActivated` event golden test.
- Export/import preserves activated account.

Definition of done:

- Activation is idempotency-safe.
- The same input produces the same state and event across runs.

### Task 4.2 - External And Internal Message Model

Implementation:

Add typed message model:

```text
ExternalMessage: signed user tx message
InternalMessage: module/contract/system message
```

Rules:

- external messages require account auth policy validation and sequence handling;
- internal messages must be accepted only by versioned feature rules;
- account internal-message rules must be upgradeable by governance/chain upgrade without changing address derivation;
- internal messages/events from contracts and modules must not bypass freeze/recovery restrictions unless explicitly whitelisted.

Tests:

- External tx from active account succeeds with valid auth.
- External tx from inactive/frozen account rejected.
- Internal message accepted by enabled feature rule.
- Internal message rejected when feature rule disabled.
- Internal message rules migration preserves existing accounts.
- Internal message handling does not increment user sequence unless policy explicitly says so.

Definition of done:

- Internal and external message paths are separate in types and tests.
- Future account versions can change internal message rules without changing addresses.

## [CHAT 1] 5. Auth Policy

### Task 5.1 - Implement AuthPolicy Model

Implementation:

Add versioned `AuthPolicy`:

```text
AuthPolicy {
  version
  mode: single_key | multisig | threshold | weighted | two_device
  keys[]
  threshold
  weights
  recovery_policy
  timelock
  spending_limits
}
```

Rules:

- 2FA-like behavior is additional public key/signature policy, never SMS/TOTP secrets on-chain.
- Small transfers can allow one primary key if spending limits allow it.
- Dangerous operations require stronger policy:
  - staking changes can require primary + device key;
  - auth policy update can require primary + recovery key + timelock;
  - large transfer can require threshold signatures.
- Address must not change when auth policy changes.

Messages:

- `MsgUpdateAuthPolicy`;
- `MsgRotateKey`;
- `MsgRecoverAccount`;
- `MsgFreezeAccount`;
- `MsgUnfreezeAccount`;
- `MsgUpdateAccountMetadata`;
- `MsgUpdateAccountParams`.

Tests:

- Single-key policy authorizes a normal tx.
- Multisig threshold policy rejects insufficient signatures.
- Weighted multisig sums weights deterministically.
- Two-device policy requires both configured public keys for protected operations.
- Spending limit allows small transfer and rejects large transfer.
- Timelock prevents early recovery/auth change.
- Recovery policy changes status to `recovered` only after valid authorization.
- Key rotation preserves `AE...` and `4:...` addresses.
- Auth policy update requires authorization.
- Private keys, seed phrases, SMS secrets, and TOTP secrets are rejected from policy serialization.

Definition of done:

- Auth policy is reusable by ante/auth checks.
- Every auth mode has positive and negative tests.

## [CHAT 1] 6. Native Storage Rent

### Task 6.1 - Extend Storage Rent To Wallet And App State

Implementation:

Storage rent is the rent paid for occupying persistent blockchain state. It is computed from stored size and storage duration:

```text
storage_size = code_bytes + data_bytes
storage_rent_delta = storage_size * rent_rate_per_byte_second * elapsed_seconds
effective_fee = gas_fee + storage_rent_delta
```

Rules:

- storage size includes account code, contract code, account data, contract data, indexes, and long-lived state records owned by the account/contract/pool/module;
- rent accrues every second or by deterministic block-time equivalent;
- rent is collected lazily during account/contract/pool actions and explicit maintenance/recovery actions;
- for normal accounts and contracts, rent is paid from the balance of the account/contract that owns the state;
- for long-lived records without their own spendable balance, rent is charged to the owning account, contract, pool reserve, module balance, or governance-configured payer;
- user-facing tx cost is still presented as `effective_fee = gas_fee + storage_rent_delta`.

Storage rent applies to every active account or persistent on-chain state:

- active native wallets;
- smart contracts;
- token/NFT/DEX contract state;
- official liquid staking pool contracts and pool accounting records;
- domain registry records;
- staking/reputation records that create long-lived state;
- active module/protocol accounts and protocol-owned persistent records through protocol accounting.

Storage rent does not apply to:

- unactivated `AE...` addresses;
- empty/no-state accounts;
- deleted/closed accounts after state is removed;
- transient tx/event data that is not retained as persistent state.

Native wallet rent policy:

- very small rent rate;
- computed as `code_bytes + data_bytes` times elapsed seconds;
- store `last_storage_charge_height` or equivalent time field;
- accrue debt lazily on account tx or targeted maintenance;
- freeze if debt exceeds configured threshold;
- never delete accounts automatically;
- allow close only when balance, stake, unbonding, rewards, domains, ownership obligations, and required reputation state are zero/absent.

Recoverable freeze/unfreeze policy:

- `frozen` never wipes balance, account state, sequence, account number, auth policy, ownership, reputation, pool shares, domains, pending rewards, contract code, or contract data;
- `frozen` wallet/contract must still be able to receive incoming top-up transfers;
- `frozen` wallet/contract must allow `MsgPayStorageDebt`;
- `frozen` wallet/contract must allow `MsgUnfreezeAccount` or `MsgUnfreezeContract`;
- `MsgUnfreezeAccount` verifies that storage debt is fully paid and no other freeze reason remains, then changes status back to `active`;
- `MsgUnfreezeContract` verifies that storage debt is fully paid and no other freeze reason remains, then changes status back to `active`;
- unfreeze must preserve the same address, raw address, account number, sequence, code hash, data root, ownership, and proof keys;
- top-up does not by itself have to unfreeze the account/contract, but it must make debt payment possible;
- normal spend, contract execute, staking deposit, and state-growing operations remain blocked while status is `frozen`;
- read-only queries, proof queries, top-up, debt payment, and unfreeze remain available while status is `frozen`;
- `archived` and `closed/deleted` are separate flows and must not be used automatically just because storage debt exists.

Example:

```text
wallet balance = 0 AET
storage_debt = 5 AET
status = frozen

top-up 10 AET
MsgPayStorageDebt 5 AET
MsgUnfreezeAccount

wallet balance = 5 AET
storage_debt = 0
status = active
state = unchanged
```

Contract example:

```text
contract balance = 0 AET
storage_debt = 20 AET
status = frozen
code/data = preserved

top-up 25 AET
MsgPayStorageDebt 20 AET
MsgUnfreezeContract

contract balance = 5 AET
storage_debt = 0
status = active
code/data = unchanged
```

Zero balance and insufficient storage-rent behavior:

Wallet behavior:

```text
effective_fee = gas_fee + storage_rent_delta + unpaid_storage_debt
```

- if wallet balance is zero or insufficient for storage rent, rent debt accumulates;
- on the next wallet tx, the chain computes `effective_fee = gas_fee + storage_rent_delta + unpaid_storage_debt`;
- if wallet balance covers the full effective fee, debt is paid and wallet remains `active`;
- if wallet balance does not cover the full effective fee, wallet becomes `frozen`;
- `frozen` wallet cannot send normal txs;
- `frozen` wallet cannot spend funds;
- `frozen` wallet cannot make staking pool deposits;
- `frozen` wallet can receive incoming transfers/top-up;
- `frozen` wallet can execute `MsgPayStorageDebt`;
- `frozen` wallet can execute `MsgUnfreezeAccount`;
- after debt is paid and no other freeze reason remains, wallet returns to `active`.

Wallet automatic deletion is forbidden because a wallet may still own or reference:

- reputation;
- staking pool shares;
- unbonding;
- pending rewards;
- domain ownership;
- contract ownership/admin rights;
- recovery/auth policy.

Wallet close is voluntary only and requires:

```text
balance = 0
storage_debt = 0
stake/pool shares = 0
unbonding = 0
pending rewards = 0
domains = 0
contract ownership obligations = 0
required reputation state cleared or archived
```

Smart contract behavior:

- if contract balance is zero or insufficient for storage rent, rent debt accumulates;
- on `execute`, `migrate`, `instantiate`-related action, or maintenance action, rent is collected from contract balance or configured payer;
- if the balance/payer covers debt, contract remains `active`;
- if the balance/payer cannot cover debt, normal contract becomes `frozen`;
- critical or official contracts, including the official liquid staking pool, become `frozen_limited` instead of trapping funds.

Normal frozen contract behavior:

- normal `execute` is forbidden;
- new state writes are forbidden;
- read-only query can remain available;
- proof query can remain available;
- top-up is allowed;
- `MsgPayStorageDebt` is allowed;
- `MsgUnfreezeContract` is allowed;
- after debt is paid and no other freeze reason remains, contract returns to `active`;
- contract code, contract data, balance, admin/owner, storage records, and proof keys remain unchanged.

Critical/official contract behavior:

- official or critical contracts must not become fund traps;
- `frozen_limited` blocks new deposits and state growth;
- `frozen_limited` allows claim, unbond, matured withdrawal, top-up, debt payment, unfreeze, and governance recovery;
- official liquid staking pool must preserve user shares, active stake, unbonding, pending rewards, allocation records, and proof metadata while `frozen_limited`.

Exact zero balance cases:

```text
zero balance + no state = free
zero balance + persistent state = debt + freeze, not delete
system/critical state = protocol-paid + no freeze
```

- if address is undeployed, unactivated, empty, or has no state, `rent = 0`, `debt = 0`, and nothing happens;
- if active wallet or contract has persistent state, rent debt grows even when balance is zero;
- state remains while debt grows;
- account or contract may become `frozen`;
- top-up plus debt payment can recover it;
- deletion is allowed only for non-critical account/contract after explicit close/delete policy and only when there are no obligations.

Status semantics:

```text
frozen   = recoverable, state intact, balance intact
archived = reduced state, recoverable only if enough metadata/proofs remain
deleted  = state removed, not normally recoverable
```

Wallets should avoid automatic delete entirely. At most, a wallet can move from `frozen` to `archived` after policy-defined inactivity and only if there is no balance, stake, pool share, domain, pending reward, ownership obligation, or required reputation state.

Protocol/system accounting policy:

- protocol/module accounts are not exempt from storage accounting: their active state must be measured as `code + data` and charged over time;
- protocol/module rent is paid through module balance, protocol treasury, fee collector, or governance-configured protocol accounting;
- protocol must maintain a dedicated system storage reserve with governance-controlled minimum runway;
- reserve runway must be measured as `available_system_rent_funds / projected_system_rent_per_second`;
- if runway falls below the warning threshold, emit alerts and restrict non-critical treasury spending if configured;
- if runway falls below the critical threshold, automatically top up from fee collector/treasury according to deterministic priority;
- system rent top-up must happen before any user-account freeze logic runs;
- protocol-critical system state must not be frozen or deleted by the same user-account rent path, because rent collection must not halt consensus or break governance;
- protocol-critical accounts must not have `frozen`, `archived`, or `deleted` statuses caused by storage rent;
- if protocol accounting cannot cover system rent after deterministic top-up attempts, the chain must raise an invariant/alert and require governance/top-up/upgrade action rather than freezing consensus-critical state;
- consensus-critical modules must continue to execute while the underfunded-system-rent invariant is active;
- system rent debt must be observable and bounded by emergency governance/top-up procedures, not hidden;
- validator operator wallet pays rent like a normal wallet from its own balance;
- validator records are storage-accounted, but they should not be frozen by rent because they are consensus-critical and should be governed by slashing/jailing/commission/reputation instead.

Official liquid staking pool rent policy:

- every official pool contract has a user-facing `AE...` address and internal raw `4:...` address;
- pool contract state, pool share records, allocation records, reward indexes, and unbonding records are persistent state and must be storage-accounted;
- pool storage rent is paid from `PoolProtocolFeeBps`, pool reserve, or governance-configured pool treasury;
- normal users should not see separate per-record pool storage charges in the wallet UI;
- user tx fee pays normal gas, while pool internals pay accumulated pool rent from pool fee/reserve;
- official pool cannot be deleted while any pool shares, active stake, unbonding, pending rewards, allocation records, or required proofs exist;
- if official pool rent debt is too high, pool enters `frozen_limited`, not hard delete;
- `frozen_limited` blocks new deposits and optional rebalances but must allow claims, unbond requests, withdrawals after unbonding, and governance/top-up recovery;
- unauthorized contracts must not receive official pool protocol-payer treatment.

Tests:

- Active wallet accrues rent lazily.
- Unactivated address accrues no rent.
- Empty/no-state address accrues no rent.
- Wallet tx computes `effective_fee = gas_fee + storage_rent_delta + unpaid_storage_debt`.
- Wallet remains `active` when balance covers effective fee and accumulated debt.
- Wallet becomes `frozen` when balance cannot cover effective fee and accumulated debt.
- Zero balance undeployed/no-state address has `rent = 0` and `debt = 0`.
- Zero balance active wallet/contract with persistent state accumulates debt and preserves state.
- Contract record accrues rent using existing storage-rent accounting.
- Official liquid staking pool contract accrues rent.
- Pool share, allocation, reward index, and unbonding records are storage-accounted.
- Pool rent is paid from pool protocol fee or reserve.
- Pool rent is not charged as a surprising separate user-facing wallet fee.
- Domain record contributes persistent storage usage.
- Staking/reputation record contributes persistent storage usage or is explicitly protocol-accounted.
- System/module accounts are storage-accounted through protocol accounting.
- Protocol/system rent is paid from module balance, fee collector, treasury, or governance-configured protocol payer.
- Protocol-critical state cannot be frozen or deleted by user-account rent collection.
- Protocol rent underfunding raises invariant/alert instead of halting consensus-critical modules.
- Validator operator wallet is not exempt by validator status alone.
- Official pool with outstanding shares/stake/rewards cannot be deleted for rent debt.
- Official pool with excessive rent debt enters `frozen_limited`.
- `frozen_limited` pool rejects new deposits but allows claim, unbond, matured withdrawal, and governance/top-up recovery.
- Debt above threshold freezes wallet.
- Frozen wallet can be unfrozen by paying debt.
- Frozen wallet preserves balance, account state, sequence, auth policy, ownership, reputation, pool shares, domains, and pending rewards.
- Frozen wallet can receive top-up transfer.
- Frozen wallet rejects normal spend and staking deposit before unfreeze.
- `MsgPayStorageDebt` reduces storage debt deterministically.
- `MsgUnfreezeAccount` restores `active` after debt is paid and preserves address/account/proof keys.
- Frozen contract preserves balance, code hash, data root, admin/owner, and storage records.
- Frozen contract can receive top-up transfer.
- Frozen contract rejects normal execute before unfreeze.
- `MsgUnfreezeContract` restores `active` after debt is paid and preserves code/data/proof keys.
- Top-up alone does not mutate state except balance and optional debt payment if explicitly configured.
- Normal frozen contract rejects new state writes.
- Normal frozen contract allows read-only query and proof query.
- Official/critical `frozen_limited` contract blocks new state growth but allows user exits and recovery.
- `frozen`, `archived`, and `deleted` status semantics are distinct and validated.
- Wallet cannot move automatically from `frozen` to `deleted`.
- Wallet can move from `frozen` to `archived` only when no balance, stake, pool shares, domains, pending rewards, ownership obligations, or required reputation state remain.
- Account close rejected when balance, stake, unbonding, rewards, domains, ownership obligations, or required reputation remains.
- Account close succeeds only with zero obligations.
- Export/import preserves rent debt and last charge height.

Definition of done:

- Existing `x/storage-rent` tests still pass.
- Wallet rent is collected as part of tx fee accounting and auditable in events/queries.

## [CHAT 2] 7. Official Liquid Staking And Native Validator Layer

### Task 7.1 - Define User Staking UX

Implementation:

The official user staking flow must be:

```text
User -> Liquid Staking Contract -> Pool Contract -> Validators
```

The user-facing meaning is:

- user does not choose a concrete validator;
- user deposits any amount above a small minimum into an official liquid staking contract;
- the contract mints a staking receipt token or share record;
- the pool aggregates many small deposits into validator-sized allocations;
- the allocation engine distributes stake across validators by deterministic weights;
- rewards return to the pool and are distributed by user pool share;
- user owns a share of the network staking index, not a manual bet on one validator.

Example:

```text
step 1: user deposits 100 AET
step 2: pool total becomes 10,000 AET
step 3: allocation engine chooses weights:
  V1 = 30%
  V2 = 25%
  V3 = 45%
step 4: pool injects stake to validators
step 5: validator rewards accrue:
  reward_i is proportional to stake_i * performance_i
step 6: user reward share:
  user_share = user_pool_shares / total_pool_shares
  user_rewards = user_share * total_pool_rewards
```

Rules:

- official wallet UI and normal CLI/API must prefer pool deposit, not direct validator selection;
- direct validator delegation can exist only as an advanced/operator path if explicitly enabled by governance policy;
- small users must be able to participate without meeting validator-sized minimums;
- pool share accounting must be deterministic and export/import stable;
- receipt token must represent claim on pool assets/rewards, not validator ownership;
- pool contract and native staking keeper must agree on accounting, slashing exposure, and unbonding state.

Tests:

- User deposits a small amount into official liquid staking contract successfully.
- User receives deterministic pool shares or receipt token amount.
- User does not supply a validator address in the normal staking deposit message.
- Deposit below configured pool minimum is rejected.
- Pool total stake and total shares update deterministically.
- Receipt token/share export/import roundtrip.
- Direct user delegation to a validator is not supported; user staking goes only through official liquid staking pools.
- Validator funding through a nominator pool is allowed only through pool allocation/accounting rules.

Definition of done:

- The default staking path is pool/index-based.
- Documentation and examples no longer teach users to choose a validator for normal staking.

### Task 7.2 - Parameterize Staking, Pooling, And Allocation

Implementation:

Do not hardcode fixed validator stake or fixed allocation weights. Add governance-controlled params:

```text
StakingParams {
  MinValidatorStake
  SoloValidatorMinSelfStake
  PoolBackedValidatorMinSelfStake
  PoolBackedValidatorMaxNominatorStake
  ValidatorSelfStakeMinRatioBps
  ValidatorNominatorStakeMaxRatioBps
  MinPoolDeposit
  TargetValidatorCount
  MaxValidatorCount
  GovernanceMinValidatorCount
  GovernanceMaxValidatorCount
  UnbondingPeriod
  RewardEpochDuration
  BaseRewardRate
  MaxRewardRate
  ValidatorCommissionFloor
  ValidatorCommissionCeiling
  ReputationStakeWeight
  PoolReceiptDenomOrCodeID
  MaxPoolValidatorAllocationBps
  MinPoolValidatorAllocationBps
  AllocationRebalanceEpochs
  AllocationUptimeWeight
  AllocationCommissionWeight
  AllocationReputationWeight
  AllocationStakeEfficiencyWeight
  AllocationSlashingRiskWeight
  AllocationNetworkLoadWeight
  PoolProtocolFeeBps
  ValidatorOperatorBonusBps
  ValidatorInfrastructureCostModel
}
```

Initial genesis/testnet parameter set:

These values are starting governance parameters, not hardcoded constants. They are calibrated from the architecture target of 100-300 active validators, serious validator entry, 3% initial validator power cap, bounded commission, low user friction through liquid staking, and low/moderate APR economics.

```text
Denom:
  BaseDenom: naet
  DisplayDenom: AET
  DisplayExponent: 9
  1 AET = 1_000_000_000 naet

Validator entry:
  MinValidatorStake: 1_000_000 AET
  SoloValidatorMinSelfStake: 1_000_000 AET
  PoolBackedValidatorMinSelfStake: 400_000 AET
  PoolBackedValidatorMaxNominatorStake: 600_000 AET
  ValidatorSelfStakeMinRatioBps: 4000      # 40% of minimum validator stake
  ValidatorNominatorStakeMaxRatioBps: 6000 # 60% of minimum validator stake

User staking:
  MinPoolDeposit: 10 AET
  DirectUserValidatorDelegationEnabled: false

Validator set:
  GovernanceMinValidatorCount: 100
  TargetValidatorCount: 128
  MaxValidatorCount: 300
  GovernanceMaxValidatorCount: 300

Unbonding and rewards:
  UnbondingPeriod: 18 days
  RewardEpochDuration: 1 day

Commission:
  ValidatorCommissionFloor: 500 bps        # 5%
  DefaultValidatorCommission: 1000 bps     # 10%
  ValidatorCommissionCeiling: 2000 bps     # 20%
  ValidatorCommissionMaxDailyChange: 100 bps

Power cap:
  ValidatorPowerCapBps: 300                # 3% while active validator count <= 150
  ValidatorPowerCapSchedule:
    <= 150 validators: 300 bps
    151-250 validators: 250 bps
    > 250 validators: 200 bps
  OverflowRewardMultiplierBps: 0-3000

Pool allocation:
  MaxPoolValidatorAllocationBps: 300
  MinPoolValidatorAllocationBps: 25
  AllocationRebalanceEpochs: 1
  PoolProtocolFeeBps: 100                  # 1% of rewards, not principal

Economics:
  InflationMinBps: 200
  InitialInflationBps: 350
  InflationMaxBps: 600
  TargetBondedRatioBps: 6000
  AprTargetMinBps: 400
  AprTargetMaxBps: 700
  BurnFeeShareBps: 5000
  RewardFeeShareBps: 3500
  TreasuryFeeShareBps: 1500

Transaction fees:
  MinTxFee: 0.003 AET
  MinTxFeeBaseUnits: 3_000_000 naet
  FeeDenom: naet

Storage rent:
  StorageRentRate: 1 naet per byte-second
  StorageRentCharging: lazy on tx/action/maintenance
  SystemStorageReserveMinRunway: 365 days
  SystemStorageReserveWarningRunway: 180 days
  SystemStorageReserveCriticalRunway: 90 days
```

Parameter rationale:

- `MinValidatorStake = 1_000_000 AET` makes validator entry serious and not symbolic.
- A solo validator must provide the full `1_000_000 AET` as self-stake.
- A pool-backed validator can satisfy the same `1_000_000 AET` entry with at least `400_000 AET` own stake and at most `600_000 AET` from nominators/pool allocation.
- `ValidatorSelfStakeMinRatioBps = 4000` prevents a validator from being mostly rented stake with tiny personal exposure.
- `GovernanceMinValidatorCount = 100` and `GovernanceMaxValidatorCount = 300` encode the architecture's validator-count bounds.
- `TargetValidatorCount = 128` follows the architecture recommendation for genesis/early testnet.
- `MaxValidatorCount = 300` is the protocol upper bound; the chain should still target gradual growth from about 128 toward 300 as operator quality allows.
- `MinPoolDeposit = 10 AET` keeps user liquid staking accessible while tx fee and rent still limit dust spam.
- direct user delegation to a validator is removed; users stake only through official liquid staking pools.
- commission floor/ceiling use the architecture range of 3-5% floor and 15-20% max; genesis starts at the conservative upper end of both ranges.
- fee split uses the architecture example: 50% burn, 35% rewards, 15% treasury.
- inflation and APR params stay within the architecture target of low/moderate inflation and delegator APR estimates.
- storage rent is deliberately tiny per byte-second but always applies to active state.
- all values must be adjustable by governance within bounded validation rules.

Rules:

- `MaxValidatorCount` enforced;
- `GovernanceMinValidatorCount` enforced;
- `GovernanceMaxValidatorCount` enforced;
- `TargetValidatorCount` affects deterministic adaptation/election behavior;
- `MinValidatorStake` enforced;
- validator self-stake ratio enforced;
- `MinPoolDeposit` allows small users while still preventing dust spam;
- direct user validator delegation rejected;
- pool-backed validator funding cannot exceed the 60% nominator share limit for the minimum validator entry;
- validator sorting deterministic;
- allocation weights are deterministic and computed from native registry data;
- allocation score is a function of reputation, uptime, stake efficiency, slashing risk, commission, limits, and current network load;
- pool users receive index/delegator yield, not full validator/operator economics;
- validators receive self-stake rewards plus commission/operator upside for running infrastructure;
- jailed validators do not receive positive bonuses;
- jailed/slashed validators receive zero new pool allocation until policy allows recovery;
- validator power cap must respect constitutional/stake-concentration policy;
- params update requires governance authority.

Allocation formula:

```text
score_i = f(
  reputation_i,
  uptime_i,
  stake_efficiency_i,
  slashing_risk_i,
  commission_i,
  validator_limits_i,
  current_network_load
)

weight_i = score_i / sum(score_j)
stake_i = pool_total_active_stake * weight_i
```

Tests:

- Genesis params use `MinValidatorStake = 1_000_000 AET`.
- Solo validator requires `1_000_000 AET` self-stake.
- Pool-backed validator requires at least `400_000 AET` self-stake.
- Pool-backed validator allows at most `600_000 AET` nominator/pool stake toward the minimum validator entry.
- Genesis params enforce `ValidatorSelfStakeMinRatioBps = 4000`.
- Genesis params use `MinPoolDeposit = 10 AET`.
- Genesis params reject direct user delegation to validators.
- Genesis params use `GovernanceMinValidatorCount = 100`.
- Genesis params use `TargetValidatorCount = 128`.
- Genesis params use `MaxValidatorCount = 300`.
- Governance bounds enforce active validator count within `100-300`.
- Genesis params use `ValidatorPowerCapBps = 300` for the initial validator phase.
- Validator power cap schedule returns `300/250/200 bps` for the architecture validator-count phases.
- Genesis params use commission floor `500 bps`, default commission `1000 bps`, ceiling `2000 bps`, and max daily change `100 bps`.
- Genesis params use fee split `5000/3500/1500 bps` for burn/rewards/treasury.
- Genesis params use `UnbondingPeriod = 18 days`.
- Genesis params use `MinTxFee = 0.003 AET` or `3_000_000 naet`.
- Genesis params use `StorageRentRate = 1 naet per byte-second`.
- Genesis params validate system storage reserve runway thresholds `365/180/90 days`.
- All initial values are governance params and not hardcoded keeper constants.
- Validator below `MinValidatorStake` rejected.
- Solo validator below `1_000_000 AET` self-stake rejected.
- Pool-backed validator below `400_000 AET` self-stake rejected.
- Pool-backed validator above `600_000 AET` nominator share for minimum entry rejected.
- Validator below self-stake ratio rejected.
- Pool deposit below `MinPoolDeposit` rejected.
- Direct user delegation to validator rejected.
- `MaxValidatorCount` enforced.
- Active validator count below `100` rejected for production/mainnet params unless explicit testnet override is active.
- Active validator count above `300` rejected.
- `TargetValidatorCount` deterministic behavior with stable candidate set.
- Commission below floor and above ceiling rejected.
- Allocation excludes validators above cap or under policy limits.
- Allocation order and weights are deterministic golden values.
- Allocation responds deterministically to uptime, commission, reputation, stake efficiency, slashing risk, and network load changes.
- Jailed validator excluded from positive bonus.
- Jailed validator receives no new pool allocation.
- Stake concentration cap limits effective power.
- Unauthorized params update rejected.
- Governance-authorized params update succeeds.

Definition of done:

- No keeper logic contains consensus-critical hardcoded stake constants.
- Staking and allocation params export/import roundtrip.

### Task 7.3 - Native Validator Registry And Pool State Model

Implementation:

Add or adapt state:

- `Validator`;
- `ValidatorPerformanceScore`;
- `ValidatorCommission`;
- `ValidatorSlashingRisk`;
- `ValidatorAllocationLimit`;
- `LiquidStakingPool`;
- `PoolShare`;
- `PoolValidatorAllocation`;
- `PoolUnbondingRequest`;
- `PoolRewardIndex`;
- `RewardClaim`;
- `StakeReputationAccumulator`;
- `EpochStakingSnapshot`;
- `ValidatorSetSnapshot`.

Liquid staking pool:

```text
LiquidStakingPool {
  pool_id
  contract_address_user AE...
  contract_address_raw 4:...
  receipt_token
  total_deposited
  total_active_stake
  total_unbonding
  total_shares
  reward_index
  allocation_epoch
  last_storage_charge_height
  storage_rent_debt
  rent_payer_policy
  status
}
```

Pool share:

```text
PoolShare {
  owner AE...
  pool_id
  shares
  principal_amount
  created_height
  updated_height
  last_reward_index
  pending_rewards
  stake_weighted_seconds
  last_reputation_update
}
```

Pool validator allocation:

```text
PoolValidatorAllocation {
  pool_id
  validator AE...
  target_weight_bps
  active_stake
  pending_stake
  unbonding_stake
  performance_score
  commission_bps
  slashing_risk_bps
  updated_height
}
```

Store keys must be prefix-based and deterministic:

- `staking/validator/{validator}`;
- `staking/validator_score/{validator}/{epoch}`;
- `staking/pool/{pool_id}`;
- `staking/pool_by_contract_user/{contract_address_user}`;
- `staking/pool_by_contract_raw/{contract_address_raw}`;
- `staking/pool_share/{pool_id}/{owner}`;
- `staking/pool_allocation/{pool_id}/{validator}`;
- `staking/pool_unbonding/{pool_id}/{owner}/{request_id}`;
- `staking/pool_reward_index/{pool_id}`;
- `staking/reward_claim/{pool_id}/{owner}/{epoch}`;
- `staking/reputation_accumulator/{account}`;
- `staking/snapshot/epoch/{epoch}`;
- `staking/snapshot/validator_set/{height_or_epoch}`.

Tests:

- State key golden tests.
- Pool validation accepts valid `AE...` user-facing contract address and matching `4:...` raw address.
- Pool validation rejects mismatched `AE...` and `4:...` address pair.
- Pool `AE... <-> 4:...` roundtrip is stable.
- Pool share validation rejects malformed or raw user-facing addresses at message/query boundary.
- Pool allocation validation accepts only active eligible validators.
- Pool shares by owner query paginated.
- Pool allocations by pool query paginated.
- Validator registry query paginated.
- Snapshot export is deterministic.
- Export/import preserves validators, pools, pool addresses, pool rent debt, pool shares, allocations, unbondings, reward indexes, rewards, and reputation accumulators.

Definition of done:

- State is queryable without full scans.
- All store prefixes have tests.

### Task 7.4 - Liquid Staking Messages And Pool Contract Hooks

Implementation:

User-facing messages:

- `MsgDepositToStakingPool`;
- `MsgRequestPoolUnbond`;
- `MsgClaimPoolRewards`;
- `MsgClaimStakeReputation`.

Pool/operator messages:

- `MsgRebalancePoolAllocations`;
- `MsgInjectPoolStake`;
- `MsgWithdrawPoolStake`;
- `MsgSetOfficialLiquidStakingContract`;
- `MsgRegisterValidator`;
- `MsgUpdateValidator`;
- `MsgUpdateStakingParams`.

Rules:

- Any active wallet can deposit AET into the official liquid staking contract.
- Normal users must not choose validator addresses for the default staking path.
- Official liquid staking contract accepts deposits, tracks shares, and mints receipt token/share record.
- Official liquid staking contract and pool contract expose `AE...` to users and keep `4:...` for internal keys/proofs.
- Pool contract aggregates deposits and injects stake into validators according to allocation engine output.
- Allocation engine selects validators by uptime, commission, limits, reputation, stake efficiency, slashing risk, and current network load.
- Direct user validator delegation is not supported; validator funding from others must flow through official pool/nominator allocation accounting.
- Any active wallet can register/update validator if it meets params and auth policy.
- Pool unbonding creates per-user unbonding request and cannot release early.
- Pool rebalancing updates reward debt and reputation accumulator before moving stake between validators.
- Pool in `frozen_limited` must reject new deposits but allow user exits and claims.
- All txs update only touched pool, user share, allocation, validator, and reward records.

Tests:

- User deposits AET into official liquid staking pool successfully.
- User deposit does not include validator address.
- Inactive account pool deposit rejected.
- Frozen account pool deposit rejected unless unfreeze/payment path is executed first.
- Low pool deposit rejected.
- Receipt token/share is minted deterministically.
- Official pool `AE...` address is shown in user-facing messages/events.
- Official pool raw `4:...` address is included only in internal proof/state metadata.
- Pool allocation injects stake to multiple validators by deterministic weights.
- Pool rebalance changes validator allocations without changing user shares incorrectly.
- Validator registration below min stake rejected.
- Duplicate validator rejected.
- Pool unbonding request starts unbonding.
- Early pool unbonding release rejected.
- Pool unbonding releases only after `UnbondingPeriod`.
- Pool in `frozen_limited` rejects new deposits.
- Pool in `frozen_limited` allows claims, unbond requests, and matured withdrawals.
- Allocation change updates source and destination validator records deterministically.
- Deposit/unbond/rebalance/claim update only touched pool/share/allocation keys, not all users.

Definition of done:

- Default staking flow works from active `AE...` wallet to official liquid staking contract and then to pool allocations.
- User never needs to choose a validator in normal staking UX.
- Pool unbonding behavior is deterministic and time/height-bound.

## [CHAT 2] 8. Domains And Identity Registry

### Task 8.1 - Keep Domains Out Of Wallet State

Implementation:

Domains must live in a separate domain/identity registry:

```text
domain.aet -> owner AE...
```

Wallet capabilities:

- register domain;
- renew domain;
- transfer domain;
- set resolver records;
- use domain as display alias.

Rules:

- wallet state may hold only optional `domain_alias`;
- ownership is always `AE...`;
- domain records pay or account for storage rent;
- resolver records live in registry/identity module, not account state.

Tests:

- Domain register stores owner `AE...`.
- Domain transfer changes owner to another `AE...`.
- Wallet metadata can reference `domain_alias` without storing domain record.
- Account export does not include all owned domains.
- Domain registry export/import preserves owner and resolver records.
- Domain storage rent is charged or protocol-accounted according to configured policy.

Definition of done:

- Domain ownership and wallet metadata are separate.
- Queries can resolve domain owner without reading account metadata.

## [CHAT 2] 9. Smart Contracts, Tokens, NFTs, And DEX

### Task 9.1 - Wallet As Contract Caller And Owner

Implementation:

Wallet must be able to:

- instantiate contract;
- execute contract;
- pass funds;
- be owner/admin;
- receive internal messages/events.

Example:

```text
MsgInstantiateContract {
  creator AE...
  code_id
  init_msg
  funds
}
```

Rules:

- tokens are contracts;
- NFTs are contracts;
- DEX pools are contracts;
- official liquid staking user entrypoint is a governance-approved contract;
- staking pool contract is allowed to call native staking hooks only through explicit capability and accounting checks;
- wallet does not store all token/NFT/DEX balances;
- contract ownership/admin fields use `AE...`;
- contract execution must respect account auth/freeze/storage-rent policy.
- frozen contracts keep code/data/balance but cannot perform normal execute until unfreeze;
- frozen contracts must allow top-up, storage debt payment, unfreeze, read-only query, and proof query.

Tests:

- Active wallet instantiates contract.
- Active wallet executes contract and passes funds.
- Frozen wallet cannot instantiate/execute until unfrozen unless explicitly allowed by policy.
- Frozen contract cannot execute normal calls until unfrozen.
- Frozen contract can be topped up and unfrozen without losing code/data/balance.
- Contract owner/admin is `AE...`.
- Token/NFT ownership query reads contract/registry state, not wallet account state.
- Official liquid staking contract can deposit pooled stake into native validator allocations.
- Official liquid staking contract has stable `AE...` and `4:...` addresses.
- Unauthorized contract cannot call native staking injection hooks.
- Contract internal message to account follows internal message rules.
- Storage rent charged to contract state.

Definition of done:

- No new native token/NFT/DEX `x/` module is introduced for application assets.
- Contract paths use account auth policy and address validation.

## [CHAT 3] 10. Pool Rewards And User Distribution

### Task 10.1 - Lazy Pool Reward Index

Implementation:

Do not scan all pool users every block. Use:

- global reward index per pool and per validator allocation;
- per-user pool-share reward debt or `last_reward_index`;
- lazy recompute on deposit, unbond request, rebalance, reward claim, and reputation claim;
- deterministic integer/decimal rounding;
- emissions/fee allocation cap.

Reward model:

```text
validator_reward_i = active_stake_i * performance_i * reward_rate
pool_total_rewards = sum(validator_reward_i - validator_commission_i)
user_share = user_pool_shares / total_pool_shares
user_rewards = user_share * pool_total_rewards
```

Validator versus pool-user economics:

```text
pool_user_income =
  pool_share_of_gross_staking_rewards
  - validator_commission
  - pool_protocol_fee
  - slashing_losses

validator_income =
  self_stake_rewards
  + commission_on_pool_or_delegated_rewards
  + operator_performance_bonus
  - infrastructure_cost
  - slashing_or_jail_losses
```

Pool users must not earn full validator/operator economics. A large user can deposit a validator-sized amount into liquid staking, but that user receives index/delegator yield after validator commission and pool fee. To earn validator upside, the user must become a validator, provide self-stake, run infrastructure, maintain uptime, accept slashing/jail risk, and earn allocation through performance.

Illustrative annual example for math validation only:

```text
gross_staking_yield = 14.4%  # example input, not a hardcoded protocol rate
validator_commission = 10%
pool_protocol_fee = 1% of rewards after validator commission
validator_self_stake = 300,000 AET
pool_user_deposit = 300,000 AET
pool_allocation_to_validator = 300,000 AET

pool_user_gross_rewards = 300,000 * 14.4% = 43,200 AET
validator_commission_paid_by_pool_user = 43,200 * 10% = 4,320 AET
pool_fee = (43,200 - 4,320) * 1% = 388.8 AET
pool_user_net_rewards = 38,491.2 AET
pool_user_net_apr = 12.8304%

validator_self_stake_rewards = 300,000 * 14.4% = 43,200 AET
validator_commission_income = 4,320 AET
validator_gross_income = 47,520 AET
validator_effective_apr_before_costs = 15.84%
```

If validator infrastructure cost is `2,000 AET/year`, the same example becomes:

```text
validator_net_income = 47,520 - 2,000 = 45,520 AET
validator_effective_apr_after_costs = 15.173333%
```

The `14.4%` value is not a target APR, promise, fixed validator yield, or consensus constant. It is a test fixture showing how the formula behaves for one input. Real gross staking yield must be derived from governance-controlled emissions, fee allocation, bonded ratio, validator performance, slashing, pool utilization, and reward epoch accounting.

Rules:

- validator reward contribution is proportional to allocated stake and performance;
- validator commission is applied before user distribution;
- pool protocol fee is applied after validator commission unless governance config says otherwise;
- gross staking yield is an input derived from reward policy and chain state, not hardcoded into pool math;
- validator self-stake rewards are not charged validator commission to the same validator;
- validator operator income includes commission and optional operator performance bonus, not just passive stake yield;
- pool users share slashing losses and validator underperformance through the pool share price/reward index;
- slashed/jailed validators contribute according to slashing and jailing policy, never with positive bonus while jailed;
- pool reward distribution must be by shares, not by manual validator choice;
- the pool must preserve exact accounting across export/import;
- rewards cannot exceed allocated emissions/fees.

Messages:

- `MsgClaimPoolRewards`;
- `MsgSyncPoolRewards`;
- `MsgClaimStakingRewards` only for internal compatibility/migration paths, not user-facing direct delegation.

Queries:

- `PoolRewards`;
- `PoolShare`;
- `PoolAllocations`;
- `StakingRewards` for compatibility/internal migration paths only.

Rules:

- rewards cannot exceed allocated emissions/fees;
- claim updates only caller pool-share state and touched pool reward indexes;
- export/import must preserve reward state exactly.

Tests:

- Pool user receives rewards after epoch progression.
- Reward index deterministic rounding golden test.
- Rewards are proportional to `user shares / total pool shares`.
- Validator reward contribution changes deterministically with performance score.
- Validator commission is deducted deterministically before user distribution.
- Pool protocol fee is deducted deterministically after validator commission.
- Validator self-stake earns base rewards without paying self-commission to itself.
- Validator commission income is credited separately from pool-user rewards.
- Pool user with 300,000 AET at illustrative 14.4% gross yield, 10% validator commission, and 1% pool fee receives `38,491.2 AET` annual net rewards.
- Validator with 300,000 AET self-stake and 300,000 AET pool allocation at illustrative 14.4% gross yield and 10% commission receives `47,520 AET` annual gross income before infrastructure cost.
- Validator net income subtracts configured infrastructure cost model.
- Validator economics exceed pool-user economics only through commission/operator bonus and only with validator risk/responsibility.
- Claim rewards updates only caller pool-share state.
- Claim twice without new rewards returns zero or configured no-op behavior.
- Rewards cannot exceed emissions/fee allocation.
- Export/import preserves reward indexes and pending rewards.
- No full scan path in reward calculation, proven by instrumentation or bounded iterator test.
- Million-user style scalability test builds many pool shares and confirms claim touches bounded keys.

Definition of done:

- Reward accounting is lazy and bounded.
- Rounding behavior has golden tests.

## [CHAT 3] 11. Stake Reputation

### Task 11.1 - Stake-Time Reputation Accumulator

Implementation:

Stake reputation belongs to the account/user, not to token/NFT/domain ownership.

Accumulator:

```text
stake_weighted_seconds += stake_amount * duration
```

Rules:

- update lazily when stake changes, rewards are claimed, or `MsgClaimStakeReputation` is executed;
- for normal users, stake-time comes from pool shares and pool active stake exposure;
- no reputation increase without positive stake-time;
- jailed/slashed behavior reduces or blocks validator reputation bonuses;
- reputation state is export/import stable;
- reputation is not transferable as a token/NFT.

Messages:

- `MsgClaimStakeReputation`.

Queries:

- `StakeReputation`;
- `AccountReputation`.

Tests:

- Stake-time reputation increases with pool share exposure times duration.
- No stake-time means no reputation.
- Reputation claim deterministic golden test.
- Slashed/jailed validator cannot receive positive validator bonus.
- Delegator on jailed validator receives only allowed reputation behavior according to policy.
- Pool user reputation is not tied to choosing a concrete validator.
- Export/import preserves accumulator.
- Reputation cannot be minted by direct metadata/account update.
- Reputation remains owned by `AE...` account and cannot be transferred as token/NFT.

Definition of done:

- Reputation math is deterministic.
- All reputation changes are explained by stake-time and policy.

## [CHAT 3] 12. Proofs, Events, And Receipts

### Task 12.1 - Proof Metadata Model

Implementation:

Users must be able to prove:

- that they deposited into the official staking pool;
- how many pool shares or receipt tokens they held;
- how long they held pool stake exposure;
- how the pool allocated stake across validators for the relevant epoch;
- how many rewards they received;
- what reputation they received.

Use state proof metadata:

- height;
- store key;
- state key;
- app hash/root hash reference;
- proof path metadata.

Merkle/IAVL/SMT proofs prove state by key. They must not require storing every accrual as a separate historical record.

Queries:

- `StakingProof`.

Tests:

- Pool deposit proof metadata stable.
- Pool share proof metadata stable.
- Pool allocation proof metadata stable.
- Reward proof metadata stable.
- Reputation proof metadata stable.
- Proof query returns height, store key, state key, root/app hash metadata, and proof path metadata.
- Proofs do not require scanning all users.
- State key golden tests match query proof metadata.

Definition of done:

- Proof metadata is deterministic and bounded.
- Proof queries align with store key layout.

### Task 12.2 - Deterministic Events And Receipts

Implementation:

Events:

- `AccountActivated`;
- `PoolStakeDeposited`;
- `PoolSharesMinted`;
- `PoolAllocationUpdated`;
- `PoolUnbondingRequested`;
- `PoolUnbondingCompleted`;
- `PoolRewardsClaimed`;
- `StakeReputationClaimed`;
- `ValidatorRegistered`;
- `ValidatorUpdated`;
- `AdvancedStakeDelegated`;
- `AdvancedStakeUndelegated`;
- `AdvancedStakeRedelegated`.

Each event must include:

- actor `AE...`;
- pool contract `AE...` when applicable;
- validator `AE...` only for allocation or advanced/operator paths;
- amount/shares when applicable;
- height/epoch;
- state key/proof metadata when applicable.

Rules:

- no wall-clock time;
- no map iteration order;
- no nondeterministic attributes;
- no private key/seed/auth secrets;
- low-cardinality enough for indexing.

Tests:

- Golden event test for every event type.
- Event order deterministic for multi-message tx.
- Events include expected actor, pool contract, validator when allocation-specific, amount/shares, height/epoch, and state key.
- Events never include private key, seed phrase, or secret fields.
- Reward claim event stable.
- Account activation event stable.

Definition of done:

- Event schema is documented.
- Events are deterministic across repeated runs.

## [CHAT 4] 13. Versioning, Migration, And Upgrades

### Task 13.1 - Account Versioning

Implementation:

Account must have:

- version;
- feature flags;
- deterministic defaults;
- migration handlers.

Add:

- `MigrateAccountIfNeeded(account)`;
- `MigrateAccountV1ToV2`;
- `ValidateAccountInvariant`;
- `ExportGenesis/ImportGenesis` for versions;
- app/module upgrade handler;
- lazy migration for millions of accounts;
- optional batched migration job.

Cannot change without backward compatibility plan:

- address derivation;
- `AE...` format;
- `4:...` raw format;
- ownership keys;
- staking ownership keys;
- sequence semantics;
- signature domain.

Can update:

- auth policy;
- recovery policy;
- internal message rules;
- account features;
- metadata;
- reputation links;
- staking capabilities.

Tests:

- Lazy migration preserves existing account.
- Unsupported version rejected safely.
- V1 to V2 migration deterministic golden test.
- Address unchanged across migration.
- Sequence semantics unchanged or compatibility test proves translation.
- Feature defaults deterministic.
- Export/import handles mixed account versions.
- Batched migration resumes safely and does not skip/duplicate accounts.

Definition of done:

- Upgrade handler and lazy migration path are both tested.
- Migration does not require full account scan during normal block execution.

### Task 13.2 - Genesis Export And Import

Implementation:

- Export accounts, auth policies, validator registry, liquid staking pools, pool shares, allocations, rewards, unbondings, reputation accumulators, storage rent state, and proof metadata in deterministic order.
- Import validates before writing state.
- Reject malformed genesis before partial initialization.
- Preserve reward indexes and reputation accumulators exactly.

Tests:

- Export/import preserves accounts.
- Export/import preserves auth policies.
- Export/import preserves pools/shares/allocations/rewards/unbondings.
- Export/import preserves reputation accumulators.
- Export/import preserves storage rent debt.
- Malformed duplicate account rejected.
- Malformed duplicate pool share or allocation rejected.
- Unsupported account/staking version rejected.
- Export order deterministic across repeated runs.

Definition of done:

- Full-state export/import is deterministic and covered by tests.
- No private key/seed appears in exported genesis.

## [CHAT 4] 14. Scalability And Performance

### Task 14.1 - No Full User Scans In Block Lifecycle

Implementation:

- No full scan of all accounts, pool users, or pool shares in BeginBlock/EndBlock.
- Only bounded validator-set operations are allowed in block lifecycle.
- Rewards are lazy.
- Reputation is lazy.
- Storage rent is lazy or bounded by explicit maintenance queue.
- Queries are paginated.
- Store keys are prefix-based and deterministic.

Tests:

- BeginBlock/EndBlock does not iterate over all pool users.
- Reward claim touches bounded key count.
- Reputation claim touches bounded key count.
- Storage rent charge touches bounded key count.
- Million-user style simulation test confirms deposit/claim/reputation paths remain bounded.
- Paginated queries enforce max page size.

Definition of done:

- A reviewer can identify all loops over account, pool-share, or allocation state and see bounded limits.
- Benchmarks or instrumentation prove no accidental O(N users) hot path.

## [CHAT 4] 15. Accounting And Invariants

### Task 15.1 - Register Invariants

Implementation:

Add invariant checks:

- private key never on-chain;
- seed phrase never on-chain;
- `AE...` address roundtrip stable;
- raw `4:...` roundtrip stable;
- account activation idempotency enforced;
- account cannot be activated twice;
- total pool active stake plus pool unbonding plus validator self-stake plus liquid balances does not create coins;
- rewards cannot exceed emissions/fee allocation;
- `MaxValidatorCount` enforced;
- `MinValidatorStake` enforced;
- validator self-stake ratio enforced;
- `MinPoolDeposit` enforced;
- direct user validator delegation rejected;
- unbonding cannot release early;
- reputation cannot increase without stake-time;
- jailed validators cannot receive positive validator bonus;
- export/import preserves accounts, pools, pool shares, allocations, rewards, unbondings, reputation accumulators, and storage rent;
- module account/bank accounting remains consistent;
- protocol-critical system state is not frozen by storage rent.
- system storage reserve has enough runway or raises an invariant/alert;
- deterministic system rent top-up runs before user freeze processing;
- protocol-critical modules remain executable during system rent underfunding.

Tests:

- Invariant registration test includes every invariant.
- Invariant failure fixtures produce expected errors.
- Bank/module accounting invariant passes after pool deposit, reward claim, pool unbonding, storage rent payment, allocation rebalance, and contract execution.
- Fuzz or table test attempts secret/private field injection into account/auth/genesis/events and is rejected.

Definition of done:

- Invariants are wired into the app invariant registry or tested equivalent.
- Failing fixtures prove invariants are meaningful.

## [CHAT 4] 16. Documentation

### Task 16.1 - Update Docs

Implementation:

Add docs for:

- native wallet/account model;
- activation flow;
- `AE...` only user-facing addresses;
- `4:...` raw internal address model;
- official liquid staking flow: `User -> Liquid Staking Contract -> Pool Contract -> Validators`;
- why normal users buy a share of the network staking index instead of choosing a validator;
- staking, pool, and allocation params;
- pool deposit, receipt token/share accounting, rewards, and lazy reward index math;
- validator versus pool-user economics, including commission, pool fee, operator bonus, infrastructure costs, and slashing losses;
- allocation engine math using reputation, uptime, commission, limits, stake efficiency, slashing risk, and network load;
- stake reputation math;
- proof model;
- storage rent model;
- official liquid staking pool storage rent, pool fee/reserve payer policy, and `frozen_limited` safety mode;
- system/protocol storage-rent accounting and protocol-payer rules;
- auth policy modes, multisig, threshold, weighted, two-device, timelock, recovery, and spending limits;
- upgrade/migration model;
- why token/NFT/DEX are contracts while wallet/account/staking/reputation are native.

Tests:

- Documentation tests or static checks ensure docs mention `AE...`, `4:...`, activation, private-key exclusion, official liquid staking, pool shares, allocation engine, validator versus pool-user economics, lazy rewards, storage rent for all active state, protocol-payer accounting, and contract-only token/NFT/DEX design.
- Examples compile or are validated where they include commands/messages.

Definition of done:

- Docs match implemented behavior and test names.
- Deprecated native token/NFT/DEX module language is removed or explicitly marked historical.

## [CHAT 4] 17. Test Matrix

### Wallet Tests

- Activate account success.
- Activate duplicate rejected.
- Address mismatch from `derive(pubkey)` rejected.
- Private key not serialized/exported.
- Seed phrase not serialized/exported.
- `AE...` roundtrip.
- `4:...` roundtrip.
- `AE... <-> 4:...` roundtrip.
- Sequence starts deterministic.
- Account number starts deterministic.
- Export/import preserves account.
- Lazy migration preserves existing account.
- Unsupported version rejected safely.
- Auth policy update requires authorization.
- Frozen account cannot spend.
- Archived account keeps minimal required state.
- Closed account allowed only with zero obligations.

### Auth Policy Tests

- Single-key success.
- Multisig insufficient signatures rejected.
- Threshold exact threshold accepted.
- Weighted threshold deterministic.
- Two-device policy requires device signature for protected action.
- Spending limit allows small transfer.
- Spending limit rejects large transfer.
- Timelock blocks early recovery.
- Recovery rotates key without changing address.
- Auth policy never serializes private key/seed/2FA secrets.

### Staking Tests

- User deposits AET into official liquid staking contract and receives pool shares.
- Normal user staking deposit does not include validator selection.
- User receives rewards from pool share after validators earn rewards.
- Low pool deposit below `MinPoolDeposit` rejected.
- Validator below `MinValidatorStake` rejected.
- Validator entry below `1_000_000 AET` rejected.
- Solo validator below `1_000_000 AET` self-stake rejected.
- Pool-backed validator below `400_000 AET` self-stake rejected.
- Pool-backed validator with more than `600_000 AET` nominator stake toward minimum entry rejected.
- Direct user delegation to validator rejected.
- `MaxValidatorCount` enforced.
- Validator count below `100` rejected for production/mainnet params unless explicit testnet override is active.
- Validator count above `300` rejected.
- `TargetValidatorCount` deterministic behavior.
- Allocation engine weights validators by reputation, uptime, commission, limits, stake efficiency, slashing risk, and network load.
- Allocation engine excludes jailed/slashed validators from new positive allocation.
- Pool injects aggregated stake into multiple validators.
- Pool rebalance deterministic and bounded.
- Pool unbonding delay of `18 days` enforced.
- Early pool unbonding release rejected.
- Reward index deterministic rounding.
- Pool-user income equals gross staking rewards minus validator commission, pool fee, and slashing losses.
- Validator income equals self-stake rewards plus commission/operator bonus minus infrastructure cost and slashing/jail losses.
- 300,000 AET pool user illustrative annual math fixture at 14.4% gross yield, 10% validator commission, and 1% pool fee returns `38,491.2 AET`.
- 300,000 AET validator illustrative annual math fixture with 300,000 AET pool allocation at 14.4% gross yield and 10% commission returns `47,520 AET` before infrastructure cost.
- Staking economics tests verify the same formulas for multiple configurable gross yield inputs, not only 14.4%.
- Claim rewards updates only caller pool-share state.
- No full scan path in reward calculation.
- Million-user style scalability test.
- Export/import preserves pools/shares/allocations/rewards/unbondings.
- Unauthorized params update rejected.
- Jailed validators receive no positive validator bonus.

### Reputation Tests

- Stake-time reputation increases with pool share exposure times duration.
- No stake-time means no reputation.
- Reputation claim deterministic.
- Slashed/jailed validator cannot receive bonus.
- Export/import preserves accumulator.
- Reputation is account-owned and non-transferable.

### Storage Rent Tests

- Active wallet pays tiny lazy rent.
- Unactivated address pays no rent.
- Empty/no-state address pays no rent.
- Rent is computed from `code_bytes + data_bytes`.
- Rent increases with elapsed seconds of storage duration.
- Rent is collected automatically during account tx or contract/pool action.
- Rent is deducted from the balance of the account/contract/pool/module/protocol payer that owns or funds the state.
- Wallet tx effective fee includes gas fee, storage rent delta, and unpaid storage debt.
- Wallet with enough balance pays debt and remains `active`.
- Wallet with insufficient balance becomes `frozen`.
- Zero balance undeployed/no-state address has zero rent and zero debt.
- Zero balance active wallet/contract with persistent state accrues debt and preserves state.
- System/module accounts are not storage-rent exempt and are charged through protocol accounting.
- System storage reserve runway is computed deterministically.
- System storage reserve warning threshold emits alert.
- System storage reserve critical threshold triggers deterministic top-up from fee collector/treasury.
- System rent top-up executes before user-account freeze logic.
- Protocol-critical system account cannot enter `frozen`, `archived`, or `deleted` status because of storage rent.
- Underfunded system rent raises invariant/alert but consensus-critical modules continue executing.
- Empty or undeployed addresses are the only addresses that do not accrue storage rent.
- Validator operator wallet pays like a normal wallet.
- Validator consensus-critical records are not frozen by rent.
- Official liquid staking pool contract has stable `AE...` and `4:...` addresses.
- Official liquid staking pool pays storage rent from pool fee/reserve/treasury policy.
- Pool share, allocation, reward index, and unbonding records contribute to pool storage accounting.
- Official pool cannot be deleted while shares, active stake, unbonding, pending rewards, allocation records, or required proofs exist.
- Official pool enters `frozen_limited` instead of trapping funds when rent debt exceeds threshold.
- `frozen_limited` pool rejects new deposits but allows claims, unbond requests, matured withdrawals, and governance/top-up recovery.
- Contract storage accrues rent.
- Normal frozen contract rejects execute and new state writes.
- Normal frozen contract allows top-up, debt payment, unfreeze, read-only query, and proof query.
- Frozen wallet preserves balance, state, sequence, auth/recovery policy, ownership, reputation, pool shares, domains, and pending rewards.
- Frozen contract preserves balance, code, data, admin/owner, storage records, and proof keys.
- Domain records accrue rent or are protocol-accounted.
- Debt freezes wallet.
- Payment unfreezes wallet.
- Top-up plus debt payment can recover frozen wallet/contract to `active`.
- Frozen status never wipes state, resets sequence, deletes code/data, or deletes ownership.
- `frozen`, `archived`, and `deleted` have distinct validated semantics.
- Wallet cannot be automatically deleted for storage debt.
- Wallet can be archived only when no balance, stake, pool shares, domains, pending rewards, ownership obligations, or required reputation state remain.
- Close rejected while obligations exist.
- Export/import preserves rent state.

### Proof Tests

- Pool deposit proof metadata stable.
- Pool share proof metadata stable.
- Pool allocation proof metadata stable.
- Reward claim proof metadata stable.
- Reputation proof metadata stable.
- State key deterministic.
- Proof query returns height/key/root/path metadata.
- Proofs do not require scanning all users.

### Event Tests

- `AccountActivated` stable.
- `PoolStakeDeposited` stable.
- `PoolSharesMinted` stable.
- `PoolAllocationUpdated` stable.
- `PoolUnbondingRequested` stable.
- `PoolUnbondingCompleted` stable.
- `PoolRewardsClaimed` stable.
- `StakeReputationClaimed` stable.
- `ValidatorRegistered` stable.
- `ValidatorUpdated` stable.
- `UnbondingStarted` stable.
- `UnbondingCompleted` stable.
- Events never include private key, seed phrase, or 2FA secrets.

## [ALL CHATS] 18. Parallel Codex Workstreams

This section exists so multiple Codex chats can implement the plan in parallel without breaking each other. Each chat must pick exactly one workstream, stay inside its ownership boundary, and expose changes through narrow interfaces. If a workstream needs to change another workstream's owned files, it must first add an interface/test fixture or leave a TODO in this file instead of editing across boundaries.

### 18.1 Parallelization Rules

Global rules:

- Start every chat by reading `UPDATE.md`, `architecture.md`, `docs/cosmos-l1-skills.md`, and the packages listed under the selected workstream.
- Do not start by editing `app.go` or global module wiring unless the workstream explicitly owns integration.
- Do not change address derivation, `AE...` format, `4:...` format, sequence semantics, or signature domains outside the address/account workstream.
- Do not reintroduce user direct delegation to validators.
- Do not add native token/NFT/DEX modules; those remain contracts.
- Every workstream must add tests in the same PR/change set as code.
- Every workstream must keep export/import and genesis validation in mind, even if final wiring is done by another workstream.
- Prefer small files split by `types`, `keys`, `keeper`, `messages`, `queries`, `events`, `genesis`, `migrations`, and tests.
- Shared structs must live in the owning module's `types` package; other modules consume interfaces or query methods.
- Avoid circular keeper dependencies. Cross-module access must use explicit interfaces.
- Run targeted package tests first, then `go test ./...` after integration work.

Branch/worktree rule:

- Use one branch per workstream, for example `codex/native-account`, `codex/storage-rent`, `codex/liquid-staking-pool`.
- Do not commit unrelated dirty files.
- Merge small stable packages as they pass tests; broad app wiring is owned by W14 after W0-W13 APIs stabilize.

### 18.2 Shared Contracts

These contracts are shared and must not be independently redefined by parallel workstreams:

Address contract:

```text
user-facing account/validator/consensus/pool address = AE...
internal raw address = 4:...
AE... <-> 4:... roundtrip must be stable
no aevaloper
no aevalcons
```

Validator entry contract:

```text
active validators: 100-300 outside explicit testnet override
minimum validator entry: 1_000_000 AET
solo validator self-stake: 1_000_000 AET
pool-backed validator self-stake: >= 400_000 AET
pool-backed nominator/pool stake toward minimum entry: <= 600_000 AET
direct user validator delegation: disabled
```

User staking contract:

```text
User -> Liquid Staking Contract -> Pool Contract -> Validators
normal user chooses pool/index, not a validator
MinPoolDeposit = 10 AET
UnbondingPeriod = 18 days
```

Fee/rent contract:

```text
MinTxFee = 0.003 AET = 3_000_000 naet
StorageRentRate = 1 naet per byte-second
storage_size = code_bytes + data_bytes
effective_fee = gas_fee + storage_rent_delta + unpaid_storage_debt
zero balance + no state = free
zero balance + persistent state = debt + freeze, not delete
system/critical state = protocol-paid + no freeze
```

Frozen-state contract:

```text
frozen = recoverable, state intact, balance intact
archived = reduced state, recoverable only if enough metadata/proofs remain
deleted = state removed, not normally recoverable
```

### 18.3 Dependency Graph

Independent chat workstream groups:

```text
CHAT 1 - Repository Baseline And Guardrails:
  W0 Address compatibility
  W1 Governance params schema
  W2 Native account/auth
  W3 Storage rent core

CHAT 2 - Validator Registry And Official Pool Entry:
  W4 Validator registry/policy
  W5 Liquid staking pool state
  W6 Contract capability hooks

CHAT 3 - Allocation, Rewards, Reputation, Proofs:
  W7 Allocation engine
  W8 Pool rewards
  W9 Stake reputation
  W10 Proofs/events

CHAT 4 - Genesis, Invariants, Docs, Final Wiring:
  W11 Genesis/migrations/export-import
  W12 Scalability/invariants
  W13 Docs/CLI/query surface
  W14 Final app wiring
```

Parallel execution rule:

```text
All independent groups can start at once.
```

The rule is:

```text
each workstream owns its packages
each workstream can add temporary local interfaces/fixtures
no workstream edits another workstream's owned files
final app wiring happens after feature package APIs stabilize
```

Integration/merge strategy:

```text
Merge stable leaf packages as they pass tests.
Merge shared interfaces before code that consumes them.
Leave broad app wiring to W14 after W0-W13 APIs are stable.
```

Approximate workload balance:

- CHAT 1 has 4 foundational workstreams and many unit tests.
- CHAT 2 has 3 state/keeper/capability workstreams and heavier staking logic.
- CHAT 3 has 4 math/proof/reward workstreams and heavy golden/scalability tests.
- CHAT 4 has 4 integration/hardening workstreams and broad test/docs work.
- These groups are balanced by expected implementation effort, not only by raw workstream count.

Workstream safety rules:

- A workstream may add local mock interfaces for another workstream, but must name them clearly as temporary fixtures.
- A workstream must not edit another workstream's owned packages to make tests pass.
- Cross-workstream changes should be merged through explicit interfaces in the owner package.
- W14 is the only workstream that should do broad `app` wiring after the feature packages are stable.

### CHAT 1 - Repository Baseline And Guardrails

Goal: define the stable foundation every other workstream consumes: address compatibility, params, native account/auth, and storage rent.

Owned workstreams:

- W0 Address Compatibility
- W1 Governance Params Schema
- W2 Native Account And Auth
- W3 Storage Rent Core

CHAT 1 outputs:

- stable address helpers;
- genesis/governance params structs;
- account/auth interfaces;
- storage-rent interfaces;
- frozen/unfreeze semantics;
- system rent reserve behavior.

CHAT 1 workstreams must not implement:

- pool allocation scoring;
- pool rewards;
- reputation rewards;
- final app wiring.

### 18.4 CHAT 1 / Workstream W0 - Address Compatibility

Ownership:

- `app/addressing`
- address validation tests
- address docs snippets

Tasks:

- Freeze `AE...` and `4:...` golden vectors.
- Add pool address helpers if missing: `FormatPoolAddress`, `ParsePoolAddress`, or reuse existing account codec with explicit tests.
- Reject `aevaloper` and `aevalcons` in user-facing account/validator/consensus/pool APIs.
- Add stable `AE... <-> 4:...` roundtrip tests for accounts, validators, consensus addresses, and pools.

Do not touch:

- staking keeper logic;
- storage rent accounting;
- app module wiring except codec registration if required.

Required tests:

- account `AE...` roundtrip;
- validator `AE...` roundtrip;
- consensus `AE...` roundtrip;
- pool `AE...` roundtrip;
- raw `4:...` roundtrip;
- malformed legacy prefixes rejected.

### 18.5 CHAT 1 / Workstream W1 - Governance Params Schema

Ownership:

- params structs for native account/staking/storage rent/economics modules;
- genesis param validation;
- docs table for genesis params.

Tasks:

- Add governance params for validator entry, pool deposit, validator count, commission, unbonding, min tx fee, storage rent, fee split, and system storage reserve runway.
- Ensure initial values match section 7.2.
- Validate bounds:
  - validators within `100-300` outside explicit testnet override;
  - validator entry `1_000_000 AET`;
  - pool-backed split `400_000/600_000`;
  - direct user validator delegation disabled;
  - unbonding `18 days`;
  - min tx fee `0.003 AET`;
  - fee split sums to `10000 bps`.

Do not touch:

- message handlers;
- allocation engine;
- rewards math.

Required tests:

- default genesis params exact match;
- invalid validator count rejected;
- invalid self-stake/nominator ratio rejected;
- invalid fee split rejected;
- params are export/import stable.

### 18.6 CHAT 1 / Workstream W2 - Native Account And Auth

Ownership:

- native account types/state;
- activation;
- auth policy;
- account queries;
- account genesis.

Tasks:

- Implement versioned account state.
- Implement `MsgActivateAccount`.
- Implement auth policy modes.
- Implement frozen wallet behavior, top-up compatibility, `MsgPayStorageDebt`, and `MsgUnfreezeAccount`.
- Ensure balances remain in bank layer and are not duplicated inside wallet state.

Depends on:

- W0 address helpers;
- W1 account params.

Do not touch:

- pool allocation;
- rewards math;
- validator selection.

Required tests:

- activation success/duplicate rejected;
- private key/seed not serialized;
- frozen wallet preserves balance/state/sequence/auth/reputation/pool refs;
- top-up allowed while frozen;
- unfreeze restores `active` after debt payment;
- account export/import stable.

### 18.7 CHAT 1 / Workstream W3 - Storage Rent Core

Ownership:

- storage-rent types/keeper;
- rent debt accounting;
- system storage reserve;
- freeze/unfreeze state machine for normal accounts/contracts through interfaces.

Tasks:

- Implement `storage_size = code_bytes + data_bytes`.
- Implement `rent_delta = size * rate_per_byte_second * elapsed_seconds`.
- Implement `effective_fee = gas_fee + rent_delta + unpaid_debt`.
- Implement protocol-payer accounting and system storage reserve runway.
- Implement system top-up before user freeze processing.
- Ensure protocol-critical state cannot enter `frozen`, `archived`, or `deleted` because of rent.

Depends on:

- W1 storage rent params.

Do not touch:

- liquid staking allocation logic;
- validator score logic.

Required tests:

- zero balance + no state has zero rent/debt;
- zero balance + persistent state accrues debt;
- normal account/contract freezes on unpaid debt;
- system reserve warning/critical thresholds;
- deterministic top-up from fee collector/treasury;
- protocol-critical modules keep executing during underfunded-system-rent invariant.

### CHAT 2 - Validator Registry And Official Pool Entry

Goal: implement validator admission, pool entry state, and the official contract capability boundary without touching reward math.

Owned workstreams:

- W4 Validator Registry And Staking Policy
- W5 Liquid Staking Pool State
- W6 Contract Capability Hooks

CHAT 2 outputs:

- validator entry enforcement;
- `100-300` validator bounds;
- liquid staking pool state;
- pool share/unbonding state;
- official contract capability hooks.

CHAT 2 workstreams must not implement:

- allocation scoring internals;
- pool reward distribution internals;
- stake reputation accumulator;
- final app wiring.

### 18.8 CHAT 2 / Workstream W4 - Validator Registry And Staking Policy

Ownership:

- validator registry/state;
- validator policy params;
- validator score inputs if not owned by W7;
- commission bounds;
- self-stake/nominator stake validation;
- power cap.

Tasks:

- Enforce validator entry:
  - solo: `1_000_000 AET` self-stake;
  - pool-backed: at least `400_000 AET` self-stake and at most `600_000 AET` nominator/pool stake.
- Enforce active validator count bounds `100-300`.
- Enforce commission floor/default/ceiling and max daily change.
- Enforce validator power cap schedule `300/250/200 bps`.
- Expose validator registry queries for allocation engine.

Depends on:

- W0 address helpers;
- W1 staking params.

Do not touch:

- pool share accounting;
- reward distribution;
- storage rent internals.

Required tests:

- validator below `1_000_000 AET` rejected;
- solo validator below `1_000_000 AET` self-stake rejected;
- pool-backed validator below `400_000 AET` self-stake rejected;
- pool-backed validator above `600_000 AET` nominator contribution rejected;
- validator count below/above bounds rejected;
- commission bounds and daily change enforced.

### 18.9 CHAT 2 / Workstream W5 - Liquid Staking Pool State

Ownership:

- liquid staking pool types/state;
- pool share state;
- pool unbonding state;
- pool contract address pair;
- pool deposit/unbond/withdraw message skeletons.

Tasks:

- Implement `LiquidStakingPool` with `contract_address_user AE...` and `contract_address_raw 4:...`.
- Implement pool shares.
- Implement pool deposit with `MinPoolDeposit = 10 AET`.
- Reject validator address in normal user deposit.
- Implement pool unbond request and matured withdrawal state.
- Implement `frozen_limited` pool status.

Depends on:

- W0 address helpers;
- W1 params;
- W3 storage rent interfaces.

Do not touch:

- allocation scoring;
- reward index;
- validator registry internals.

Required tests:

- deposit mints deterministic shares;
- deposit below `10 AET` rejected;
- user cannot specify validator address;
- pool `AE... <-> 4:...` stable;
- unbonding uses `18 days`;
- `frozen_limited` rejects deposits but allows claim/unbond/matured withdrawal/top-up.

### 18.10 CHAT 2 / Workstream W6 - Contract Capability Hooks

Ownership:

- contract capability checks;
- official liquid staking contract registration;
- native staking injection hooks;
- unauthorized contract rejection.

Tasks:

- Add governance-approved official liquid staking contract capability.
- Allow official pool contract to call native staking injection hooks.
- Reject unauthorized contracts.
- Respect frozen/frozen_limited behavior.
- Ensure contract ownership/admin fields use `AE...`.

Depends on:

- W0 address helpers;
- W3 storage rent status interface;
- W5 pool contract state.

Do not touch:

- allocation scoring;
- reward math.

Required tests:

- official contract can inject pooled stake;
- unauthorized contract cannot call staking hook;
- frozen contract cannot execute normal calls;
- frozen contract can receive top-up/pay debt/unfreeze/query.

### CHAT 3 - Allocation, Rewards, Reputation, Proofs

Goal: implement deterministic allocation math, lazy pool rewards, account-owned stake reputation, and deterministic proof/event surfaces.

Owned workstreams:

- W7 Allocation Engine
- W8 Pool Rewards
- W9 Stake Reputation
- W10 Proofs And Events

CHAT 3 outputs:

- deterministic allocation engine;
- lazy pool reward index;
- validator commission and pool fee reward accounting;
- stake-time reputation from pool share exposure;
- deterministic events;
- state proof metadata.

CHAT 3 workstreams must not implement:

- validator registry policy;
- pool deposit state;
- final app wiring.

### 18.11 CHAT 3 / Workstream W7 - Allocation Engine

Ownership:

- allocation scoring;
- validator weight calculation;
- pool allocation state transitions;
- rebalance logic.

Tasks:

- Implement deterministic score:
  - reputation;
  - uptime;
  - commission;
  - limits;
  - stake efficiency;
  - slashing risk;
  - network load.
- Compute `weight_i = score_i / sum(scores)`.
- Enforce max pool allocation per validator and validator power caps.
- Exclude jailed/slashed validators from new positive allocation.
- Keep rebalancing bounded and deterministic.

Depends on:

- W4 validator registry query interface;
- W5 pool allocation state.

Do not touch:

- user deposit accounting;
- reward distribution;
- account auth.

Required tests:

- deterministic golden allocation weights;
- jailed validator gets zero new allocation;
- over-cap validator receives no extra effective power;
- rebalance touches bounded pool/allocation keys only;
- no map iteration nondeterminism.

### 18.12 CHAT 3 / Workstream W8 - Pool Rewards

Ownership:

- pool reward index;
- reward claim;
- validator commission deduction;
- pool protocol fee deduction;
- reward export/import.

Tasks:

- Implement lazy pool reward index.
- Compute user rewards by pool share.
- Deduct validator commission and pool fee from rewards, not principal.
- Keep rewards capped by emissions/fee allocation.
- Preserve reward state across export/import.

Depends on:

- W5 pool shares;
- W7 allocation records;
- W1 economics params.

Do not touch:

- allocation scoring;
- account activation.

Required tests:

- rewards proportional to shares;
- commission deducted before user rewards;
- pool fee deducted after validator commission;
- claim updates caller only;
- million-user style bounded claim test;
- illustrative 300,000 AET math fixture remains correct as formula test only.

### 18.13 CHAT 3 / Workstream W9 - Stake Reputation

Ownership:

- stake reputation accumulator;
- reputation claim;
- account reputation query.

Tasks:

- Accumulate reputation from pool share exposure and duration.
- Prevent reputation increase without stake-time.
- Block validator bonus while jailed/slashed.
- Keep reputation account-owned and non-transferable.

Depends on:

- W2 account IDs;
- W5 pool shares;
- W8 reward/reputation update touchpoints.

Do not touch:

- validator allocation scoring except through exposed reputation query.

Required tests:

- stake-time increases reputation;
- no stake-time gives no reputation;
- export/import preserves accumulator;
- reputation not token/NFT transferable.

### 18.14 CHAT 3 / Workstream W10 - Proofs And Events

Ownership:

- proof metadata structs;
- event schemas;
- receipt/golden tests.

Tasks:

- Add proof metadata for pool deposit, pool share, allocation, reward claim, and reputation claim.
- Add deterministic events:
  - `PoolStakeDeposited`;
  - `PoolSharesMinted`;
  - `PoolAllocationUpdated`;
  - `PoolUnbondingRequested`;
  - `PoolUnbondingCompleted`;
  - `PoolRewardsClaimed`;
  - `StakeReputationClaimed`.
- Ensure events never expose secrets.

Depends on:

- W5/W7/W8/W9 types.

Do not touch:

- keeper mutation logic except event emission calls through stable helpers.

Required tests:

- golden event for every event;
- proof metadata stable;
- proof query returns height/store key/state key/root/path metadata;
- no scans required for proof query.

### CHAT 4 - Genesis, Invariants, Docs, Final Wiring

Goal: make the independently built pieces production-usable through export/import, migrations, invariants, docs, query/CLI surface, and final app wiring.

Owned workstreams:

- W11 Genesis, Migration, Export/Import
- W12 Scalability And Invariants
- W13 Docs, CLI, Query Surface
- W14 Final App Wiring

CHAT 4 outputs:

- deterministic genesis/export/import;
- versioned migrations;
- scalability checks;
- invariant registry;
- docs and examples;
- final keeper/app/module wiring;
- full test pass.

CHAT 4 workstreams must not rewrite:

- address derivation;
- account auth semantics;
- allocation math;
- reward math;
- storage rent semantics.

### 18.15 CHAT 4 / Workstream W11 - Genesis, Migration, Export/Import

Ownership:

- genesis state for new modules;
- export/import validation;
- versioned migrations;
- lazy migration.

Tasks:

- Add deterministic export/import for accounts, pools, allocations, rewards, reputation, rent, and validator policy.
- Add versioned account/pool migration.
- Reject malformed duplicate state before writes.
- Preserve mixed account versions.

Depends on:

- W2/W3/W4/W5/W7/W8/W9 types.

Do not touch:

- business logic except migration handlers.

Required tests:

- full export/import preserves all new state;
- duplicate account/pool/share/allocation rejected;
- unsupported version rejected safely;
- lazy migration preserves address and sequence.

### 18.16 CHAT 4 / Workstream W12 - Scalability And Invariants

Ownership:

- invariant registration;
- bounded-iteration tests;
- benchmarks/simulations.

Tasks:

- Assert no O(N users) BeginBlock/EndBlock paths.
- Add invariant tests for bank/module accounting, rewards cap, rent, pool shares, validator entry, and direct delegation rejection.
- Add million-user style simulation for pool shares and reward claims.

Depends on:

- all core state modules.

Do not touch:

- feature implementation except small instrumentation hooks.

Required tests:

- BeginBlock/EndBlock bounded;
- reward claim bounded;
- reputation claim bounded;
- rent charge bounded;
- invariant failure fixtures.

### 18.17 CHAT 4 / Workstream W13 - Docs, CLI, Query Surface

Ownership:

- docs;
- CLI examples;
- query docs;
- static doc tests.

Tasks:

- Update docs to say normal users stake only through official liquid staking pools.
- Document validator entry:
  - `1_000_000 AET`;
  - solo full self-stake;
  - pool-backed `400_000/600_000`.
- Document `100-300` validator range.
- Document unbonding `18 days`.
- Document min tx fee `0.003 AET`.
- Document storage rent and recoverable freeze/unfreeze.

Depends on:

- final names from W1/W2/W5.

Do not touch:

- keeper logic.

Required tests:

- static doc tests for required terms;
- command examples compile or are validated.

### 18.18 CHAT 4 / Workstream W14 - Final App Wiring

Ownership:

- app module wiring;
- keeper injection;
- module order;
- integration tests;
- full `go test ./...`.

Tasks:

- Wire modules after workstream APIs are stable.
- Register keepers and interfaces.
- Register migrations.
- Register invariants.
- Run full tests.

Depends on:

- W0-W13 merged or rebased.

Do not touch:

- feature internals except integration fixes.

Required tests:

- app boots with default genesis;
- local integration staking-pool flow;
- export/import restart;
- `go test ./...`.

## [ALL CHATS] 19. Suggested Implementation Order

1. Freeze address compatibility with golden tests.
2. Add native account types, validation, keys, genesis, and query tests.
3. Add `MsgActivateAccount` and account event tests.
4. Add auth policy model and ante/auth integration tests.
5. Extend storage rent to all active accounts/persistent state and define protocol-payer accounting.
6. Add native validator registry, performance score, commission, slashing-risk, and allocation-limit state.
7. Add official liquid staking contract capability and pool state model.
8. Add pool deposit, share minting, unbond request, and pool claim messages.
9. Add deterministic allocation engine and pool stake injection/rebalance hooks.
10. Add lazy pool rewards and reward index tests.
11. Add stake reputation accumulator based on pool share exposure.
12. Add proof metadata queries and event receipts for pool deposits, shares, allocations, rewards, and reputation.
13. Wire contract instantiate/execute account auth behavior and restrict native staking hooks to official contracts.
14. Add migrations, lazy migration, export/import, and upgrade handler.
15. Run scalability tests and bounded-iteration checks.
16. Update docs and static doc coverage tests.
17. Run `gofmt` and `go test ./...`.

## [ALL CHATS] 20. Non-Negotiable Review Checklist

- No private key on-chain.
- No seed phrase on-chain.
- No `aevaloper` or `aevalcons` user-facing address.
- `AE...` remains stable forever.
- `4:...` remains stable forever.
- Address unchanged by wallet upgrade, auth policy update, recovery, multisig changes, metadata changes, or staking changes.
- Activation cannot happen twice.
- Normal users stake through official liquid staking pool, not by choosing a concrete validator.
- Pool deposits support small users above anti-spam minimum.
- Initial validator entry, pool deposit, fees, rent, commission, validator count, inflation, and fee split are governance params, not hardcoded constants.
- Active validator count must stay within `100-300` outside explicit testnet override.
- Initial validator entry requires `1_000_000 AET` total validator stake.
- Solo validator must provide `1_000_000 AET` self-stake.
- Pool-backed validator must provide at least `400_000 AET` own stake and no more than `600_000 AET` from nominators/pool allocation toward the minimum entry.
- Direct user delegation to validators is removed; user staking goes through official pools only.
- Pool shares/receipt tokens are deterministic and export/import stable.
- Allocation engine is deterministic and uses reputation, uptime, commission, limits, stake efficiency, slashing risk, and network load.
- Official pool stake injection cannot be called by unauthorized contracts.
- Rewards are lazy, deterministic, pool-share based, and capped by emissions/fees.
- Reputation cannot increase without stake-time.
- User reputation comes from pool share exposure/stake-time, not from manual validator selection.
- Storage rent applies to every active account and persistent on-chain state.
- Empty/no-state/unactivated addresses are the only addresses that do not accrue storage rent.
- Storage rent is computed from `code + data` size and elapsed storage time.
- Storage rent is collected automatically during transactions/actions from the owning account, contract, pool reserve, module balance, or configured protocol payer.
- Frozen wallet/contract state and balance are recoverable by top-up, debt payment, and unfreeze.
- Frozen status must not wipe state, reset sequence, reset ownership, delete code, or delete data.
- Storage rent cannot freeze protocol-critical system state.
- Account close cannot delete ownership or staking obligations.
- Token/NFT/DEX remain contracts.
- BeginBlock/EndBlock do not scan all users.
- All queries are paginated or direct key lookups.
- Export/import preserves all consensus state.
- Every feature ships with tests.
