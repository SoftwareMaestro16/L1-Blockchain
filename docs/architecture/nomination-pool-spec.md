# Nomination Pool Detailed Specification

Nomination pools are important for accessibility, but they introduce accounting
and centralization risks.

The implementation gate is `app/params/nomination_pool_spec.go`. A feature is
not complete unless the pool model, delegator model, risk acknowledgements,
queries, genesis validation, export/import safety, and tests are present.

## 26. Nomination Pool Detailed Specification

Purpose: allow users to participate in staking without running validator
infrastructure while preserving deterministic accounting and limiting
centralization risk.

Nomination pools must not become an unbounded operator cartel layer. Pool
operators can coordinate deposits and validator selection, but the protocol must
expose pool accounting, commission, validator target, unbonding state, and
delegator share state.

## 26.1 Pool Model

Each pool should have:

```text
Pool:
  PoolId
  OperatorAddress
  ValidatorAddress
  TotalBonded
  TotalShares
  CommissionBps
  Status
  CreatedHeight
  UnbondingEntries
```

Delegator state:

```text
PoolDelegation:
  DelegatorAddress
  PoolId
  Shares
  PrincipalEstimate
  RewardsAccrued
```

### Model Contract

Required catalog properties:

- `AetraNominationPoolModuleName` must be `x/nominator-pool`;
- `DefaultAetraNominationPoolModelEvidence` must include all required `Pool`
  and `PoolDelegation` fields from section 26.1;
- `BuildAetraNominationPoolModelReport` must reject missing, duplicate, and
  unexpected fields;
- `ValidateAetraNominationPoolModel` must reject wrong module identity and
  incomplete risk acknowledgement;
- model coverage must acknowledge accessibility, deterministic accounting, and
  centralization risks.

### Current Implementation Mapping

Current `x/nominator-pool/types.NominatorPool` uses more explicit names while
preserving the section 26.1 model:

- `PoolId` maps to `PoolID`;
- `OperatorAddress` maps to `PoolOperator`;
- `ValidatorAddress` maps to `ValidatorTarget`;
- `TotalBonded` maps to `TotalBondedStake`;
- `TotalShares` maps to `TotalShares`;
- `CommissionBps` maps to `PoolCommissionBps`;
- `Status` maps to `Status`;
- `CreatedHeight` should be represented by pool creation height and must become
  explicit in stored state before production readiness if it is not already
  persisted by keeper event/state history;
- `UnbondingEntries` maps to `UnbondingQueue`.

Current `x/nominator-pool/types.DelegatorShare` maps to `PoolDelegation`:

- `DelegatorAddress` maps to `Delegator`;
- `PoolId` is implied by the owning pool and should be explicit in query
  responses;
- `Shares` maps to `Shares`;
- `PrincipalEstimate` must be derived deterministically from
  `shares * TotalBondedStake / TotalShares`;
- `RewardsAccrued` maps to `PendingRewards` plus deterministic reward index
  accounting.

### Accounting Requirements

Pool accounting must be deterministic:

- `TotalShares` must equal the sum of all delegator shares;
- deposits mint shares using deterministic integer math;
- withdrawals burn shares and create unbonding entries;
- `PrincipalEstimate` is an estimate derived from current pool accounting, not
  a guaranteed redemption amount;
- `RewardsAccrued` must be derived from reward index checkpoints and pending
  rewards;
- slash losses must reduce pool bonded value without corrupting share supply;
- export/import must preserve sorted pools, delegations, and unbonding entries.

### Centralization Risk Requirements

Pool operators must not hide concentration risk:

- pool operator address must be queryable;
- validator target must be queryable;
- pool commission must be bounded by governance params;
- pool status must be explicit;
- pool delegations must be visible enough for wallets and explorers to warn
  users about overloaded pools or validator targets;
- no mandatory KYC should be embedded into consensus pool admission.

## 26.2 Pool Requirements

Required:

- users deposit native staking denom;
- pool mints shares deterministically;
- pool delegates to validator;
- pool distributes rewards pro-rata;
- pool commission bounded;
- pool withdrawal follows unbonding period;
- pool slashing reduces share value;
- pool operator cannot withdraw user principal;
- pool cannot bypass validator power cap;
- pool must expose risk warnings.

### Requirements Contract

Required catalog properties:

- `AetraNominationPoolRequirementsEvidence` must cover all ten requirements
  from section 26.2;
- `DefaultAetraNominationPoolRequirementsEvidence` must assert native denom
  deposits, deterministic share minting, validator delegation, pro-rata rewards,
  bounded commission, unbonding withdrawals, slash loss accounting, principal
  protection, power-cap compatibility, and risk warnings;
- `BuildAetraNominationPoolRequirementsReport` must reject missing
  requirements;
- `ValidateAetraNominationPoolRequirements` must fail if any requirement is not
  covered.

Implementation expectations:

- deposits must accept only the native staking denom configured for Aetra;
- share minting must use deterministic integer accounting;
- delegation target must be an allowed validator address and cannot be jailed;
- rewards must use share-index accounting so delegators receive pro-rata value;
- pool commission must be bounded by governance params and visible in queries;
- withdrawals must create unbonding entries and respect the chain unbonding
  period;
- slashing must lower pool bonded value and therefore share redemption value;
- operator permissions must not include user principal withdrawal authority;
- validator power-cap rules must apply to stake routed through pools;
- wallet/explorer APIs must expose risk warnings for commission,
  concentration, slashing history, jailed validator target, and unbonding state.

## 26.3 Pool Tests

Required tests:

- first deposit share price;
- subsequent deposit share price;
- reward distribution;
- commission deduction;
- partial withdrawal;
- full withdrawal;
- slashing pool validator;
- jailed validator;
- redelegation if allowed;
- pool operator abuse attempt;
- export/import with active unbonding entries;
- rounding dust handling.

### Testing Contract

Required catalog properties:

- `AetraNominationPoolTestingEvidence` must cover all twelve test categories
  from section 26.3;
- `DefaultAetraNominationPoolTestingEvidence` must assert every required test
  category;
- `BuildAetraNominationPoolTestingReport` must reject missing test coverage;
- `ValidateAetraNominationPoolTesting` must fail if any required test category
  is not covered.

Implementation expectations:

- first deposit must establish a deterministic 1:1 or explicitly configured
  initial share price;
- subsequent deposits must mint shares at current pool exchange rate;
- reward and commission tests must prove pool accounting conserves value;
- partial and full withdrawal tests must burn shares and create correct
  unbonding state;
- slashing tests must reduce bonded value without corrupting `TotalShares`;
- jailed validator tests must prevent new delegation to unsafe validator
  targets;
- redelegation tests must run only if pool redelegation is enabled;
- abuse tests must prove the operator cannot steal or withdraw user principal;
- export/import tests must include active unbonding entries;
- rounding dust must be deterministic and assigned to the configured dust sink.
