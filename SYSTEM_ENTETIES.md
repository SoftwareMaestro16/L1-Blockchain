# AET Native System Entities Implementation Plan

This document defines native system entities that should be implemented as built-in `x/` modules for the AET blockchain. These entities are not user smart contracts. They are protocol-level components compiled into the chain binary, covered by genesis/export/import, migrations, invariants, and adversarial tests.

Application tokens, NFTs, DEX pools, wallets, DAOs, and dApps may later run as AVM contracts, but the entities below are protocol infrastructure and must remain native unless a separate architecture decision explicitly changes that boundary.

## Core Rule

Native system modules own protocol safety:

- validator selection;
- staking security;
- slashing and evidence;
- validator key separation;
- minting and emissions;
- fee collection and distribution;
- protocol configuration;
- scheduling;
- storage rent for contracts;
- cross-chain trust registry;
- identity root;
- sharding coordination;
- protocol treasury.

AVM contracts own application logic:

- custom tokens;
- NFTs and SBTs;
- DEX logic;
- vaults;
- launchpads;
- games;
- dApp governance;
- application registries.

## Common Requirements For Every Native System Entity

Every module below should include:

- `proto/<module>/v1/genesis.proto`
- `proto/<module>/v1/tx.proto`
- `proto/<module>/v1/query.proto`
- `x/<module>/types`
- `x/<module>/keeper`
- `x/<module>/module.go`
- CLI tx/query commands where operationally useful
- genesis validation
- export/import tests
- migration versioning
- keeper unit tests
- integration tests through `FinalizeBlock`
- adversarial tests for malformed messages
- invariant tests
- telemetry/events
- clear authority model
- zero-address rejection for all address fields
- deterministic iteration over all state collections
- bounded loops and explicit limits in params

## Implementation Phases

### Phase 1: Safety Backbone

Implement first:

- `x/config`
- `x/constitution`
- `x/validator-registry`
- `x/validator-election`
- `x/evidence`
- `x/reporter`
- `x/fee-collector`
- `x/mint-authority`
- `x/emissions`
- `x/treasury`

Goal: make validator lifecycle, fees, minting, and protocol params explicit and auditable.

### Phase 2: Staking Extensions

Implement:

- `x/nominator-pool`
- `x/single-nominator-pool`
- `x/validator-insurance`
- `x/delegator-protection`
- `x/reputation`
- `x/performance-oracle`
- `x/stake-concentration`
- `x/dynamic-commission`

Goal: stronger validator economics and delegator protection.

### Phase 3: Execution Infrastructure

Implement:

- `x/scheduler`
- `x/actor-registry`
- `x/storage-rent`
- `x/avm-scheduler`
- `x/system-registry`

Goal: make AVM contracts safe to run long-term without unbounded state growth.

### Phase 4: Ecosystem Infrastructure

Implement:

- `x/identity-root`
- `x/dex-factory`
- `x/bridge-hub`
- `x/cross-chain-registry`
- `x/sharding-coordinator`

Goal: native coordination layer for app-level contracts and cross-chain infrastructure.

## x/config

### Purpose

Global protocol configuration module. It stores critical chain parameters and exposes a controlled update path.

### State

- consensus parameter references;
- AVM gas schedule;
- fee parameters;
- forwarding fee parameters;
- storage rent parameters;
- validator election parameters;
- validator set transition parameters;
- slashing bounds;
- inflation bounds;
- treasury distribution proportions;
- sharding parameters;
- scheduler parameters;
- module authority addresses;
- system account addresses.

### Messages

- `MsgSubmitConfigChange`
- `MsgApproveConfigChange`
- `MsgRejectConfigChange`
- `MsgExecuteConfigChange`
- `MsgCancelConfigChange`

### Queries

- `Params`
- `ConfigValue`
- `PendingConfigChanges`
- `ConfigChange`
- `Authority`

### Invariants

- all numeric limits are non-negative and bounded;
- gas schedule cannot contain zero-cost execution paths;
- storage rent cannot be disabled accidentally;
- fee denom must match the base denom policy;
- validator election params must be compatible with staking params;
- config changes cannot violate `x/constitution`.

### Tests

- invalid genesis rejects bad params;
- export/import preserves config exactly;
- config update cannot bypass authority;
- invalid config change is rejected before execution;
- config cannot set unlimited block gas;
- config cannot set zero storage rent for non-empty contract state unless explicitly allowed by constitutional rule;
- config cannot remove required system account addresses;
- deterministic ordering of pending changes.

## x/constitution

### Purpose

Hard safety bounds for the network. Normal governance/config updates may change parameters only inside these bounds.

### State

- maximum inflation;
- minimum slash fraction;
- maximum slash fraction;
- maximum validator voting power;
- maximum block gas;
- maximum AVM code size;
- maximum contract state size;
- minimum storage rent rate;
- treasury spending limits;
- upgrade delay requirements;
- emergency pause limits;
- governance quorum floors;
- protected module list.

### Messages

- `MsgProposeConstitutionAmendment`
- `MsgVoteConstitutionAmendment`
- `MsgExecuteConstitutionAmendment`
- `MsgCancelConstitutionAmendment`

### Queries

- `Constitution`
- `PendingAmendments`
- `Amendment`
- `ProtectedLimits`

### Invariants

- normal config updates cannot exceed constitutional limits;
- emergency powers expire;
- protected modules cannot be disabled through ordinary parameter updates;
- constitutional amendment requires stricter delay/quorum than ordinary config.

### Tests

- ordinary config change fails when outside constitutional bounds;
- constitutional update requires special flow;
- emergency changes expire automatically;
- export/import preserves amendment queue;
- malicious authority cannot bypass protected module list.

## x/system-registry

### Purpose

Registry of native system entities and their canonical module accounts. This module gives the chain a single source of truth for system addresses and authorities.

### State

- module name to system account address;
- module name to authority address;
- module name to status: active, paused, deprecated;
- module capabilities;
- module version;
- module dependency graph.

### Messages

- `MsgRegisterSystemEntity`
- `MsgUpdateSystemEntity`
- `MsgPauseSystemEntity`
- `MsgResumeSystemEntity`
- `MsgDeprecateSystemEntity`

### Queries

- `SystemEntity`
- `SystemEntities`
- `ModuleAccount`
- `Capabilities`
- `DependencyGraph`

### Invariants

- no duplicate module account address;
- required system entities must be active;
- paused modules cannot receive privileged calls unless explicitly allowed;
- dependency graph must be acyclic.

### Tests

- genesis rejects duplicate system accounts;
- required modules cannot be removed;
- pause/resume events are deterministic;
- dependency graph cycle is rejected.

## x/validator-registry

### Purpose

Canonical registry of validator identities, operational metadata, key separation, performance history, and security status.

### State

- validator operator address;
- consensus public key;
- treasury address;
- withdrawal address;
- emergency address;
- metadata;
- commission policy;
- uptime history;
- latency history;
- missed block counters;
- slashing history;
- reputation score;
- performance score;
- status: candidate, active, jailed, tombstoned, retired;
- validator capabilities;
- validator self-bond;
- external audit flags if used.

### Messages

- `MsgRegisterValidator`
- `MsgUpdateValidatorMetadata`
- `MsgRotateConsensusKey`
- `MsgUpdateWithdrawalAddress`
- `MsgUpdateTreasuryAddress`
- `MsgRetireValidator`
- `MsgSetValidatorCapabilities`

### Queries

- `Validator`
- `Validators`
- `ValidatorKeys`
- `ValidatorPerformance`
- `ValidatorSecurityStatus`
- `ValidatorHistory`

### Invariants

- operator key and consensus key are distinct roles;
- withdrawal address cannot be zero address;
- consensus key rotation has a delay;
- jailed or tombstoned validators cannot enter the active set;
- validator metadata length is bounded;
- validator status transitions are valid.

### Tests

- register validator with valid key separation;
- reject duplicate consensus keys;
- reject zero withdrawal address;
- key rotation delay is enforced;
- jailed validator cannot become active;
- tombstoned validator cannot re-register with same consensus key;
- export/import preserves history.

## x/validator-election

### Purpose

Native validator election engine. It computes previous, current, and next validator sets and controls epoch transitions.

### State

- previous validator set;
- current validator set;
- next validator set;
- election epoch;
- election window;
- candidate applications;
- frozen stakes;
- pending exits;
- validator power caps;
- election results;
- reward distribution snapshots;
- validator set transition history.

### Messages

- `MsgApplyForValidatorSet`
- `MsgWithdrawApplication`
- `MsgCommitElection`
- `MsgFinalizeElection`
- `MsgRequestValidatorExit`
- `MsgCancelValidatorExit`

### Queries

- `PreviousValidatorSet`
- `CurrentValidatorSet`
- `NextValidatorSet`
- `Election`
- `ElectionCandidates`
- `FrozenStake`
- `ValidatorSetTransition`

### Invariants

- previous/current/next sets are internally consistent;
- validator power is positive and bounded;
- total voting power cannot exceed configured max;
- frozen stake cannot be withdrawn before unlock height;
- next set cannot include jailed validators;
- set transition is deterministic.

### Tests

- validator set transition across epochs;
- export/import during active election;
- candidate withdrawal before deadline;
- withdrawal after deadline rejected;
- frozen stake unlock timing;
- deterministic tie-breaker by address;
- max voting power cap enforced;
- invalid next set rejected at genesis.

## x/evidence

### Purpose

Native evidence module for validator faults. It should process consensus evidence, missed block evidence, performance evidence, and fraud reports.

### State

- submitted evidence;
- evidence status: pending, accepted, rejected, expired;
- evidence type;
- accused validator;
- reporter;
- proof payload hash;
- voting state if validator review is required;
- slash decision;
- reward decision;
- expiration height;

### Messages

- `MsgSubmitEvidence`
- `MsgVoteEvidence`
- `MsgFinalizeEvidence`
- `MsgCancelExpiredEvidence`

### Queries

- `Evidence`
- `EvidenceByValidator`
- `EvidenceByReporter`
- `PendingEvidence`
- `EvidenceParams`

### Invariants

- evidence cannot be accepted twice;
- expired evidence cannot slash;
- reporter reward cannot exceed configured limits;
- slash fraction must be inside constitutional bounds;
- tombstoning is irreversible for critical faults.

### Tests

- valid evidence accepted;
- malformed evidence rejected;
- duplicate evidence rejected;
- expired evidence ignored;
- slash event updates staking and registry;
- reporter reward paid once;
- export/import preserves pending evidence;
- evidence processing cannot panic on invalid payload.

## x/reporter

### Purpose

Reporter and fisherman registry. It tracks actors that submit faults, proofs, latency reports, and availability reports.

### State

- reporter address;
- bonded amount;
- reporter score;
- accepted reports;
- rejected reports;
- slashed reporter bond;
- reporter status;
- reward history.

### Messages

- `MsgRegisterReporter`
- `MsgBondReporter`
- `MsgUnbondReporter`
- `MsgSubmitReport`
- `MsgClaimReporterReward`

### Queries

- `Reporter`
- `Reporters`
- `ReporterReports`
- `ReporterRewards`

### Invariants

- reporter must be bonded to submit slashable reports;
- rejected malicious report can slash reporter bond;
- reporter reward cannot be claimed twice;
- unbonding waits through challenge period.

### Tests

- register and bond reporter;
- submit valid report;
- malicious report slashes reporter;
- reward paid once;
- unbonding delay enforced;
- export/import preserves unclaimed rewards.

## x/nominator-pool

### Purpose

Shared staking pool for delegators. It issues pool shares, tracks delegator ownership, routes stake to validators, and distributes rewards/losses.

### State

- pool id;
- pool operator;
- validator target;
- total shares;
- total bonded stake;
- pending deposits;
- pending withdrawals;
- delegator shares;
- reward index;
- slash index;
- pool commission;
- pool status;
- unbonding queue.

### Messages

- `MsgCreateNominatorPool`
- `MsgDepositToPool`
- `MsgRequestPoolWithdrawal`
- `MsgCancelPoolWithdrawal`
- `MsgClaimPoolRewards`
- `MsgUpdatePoolCommission`
- `MsgChangePoolValidator`

### Queries

- `NominatorPool`
- `NominatorPools`
- `PoolDelegator`
- `PoolRewards`
- `PoolUnbondingQueue`

### Invariants

- pool shares match total stake accounting;
- no withdrawal before unbonding completion;
- slashes reduce pool value proportionally;
- pool commission is bounded;
- pool cannot delegate to jailed validator;
- no negative delegator balance.

### Tests

- deposit mints shares;
- withdrawal burns shares;
- rewards distribute proportionally;
- slash applies proportionally;
- pool cannot withdraw more than total stake;
- export/import preserves reward index;
- pool validator change delay enforced.

## x/single-nominator-pool

### Purpose

Minimal staking pool for one large nominator and one validator. This should have lower complexity and a smaller attack surface than shared pools.

### State

- pool address;
- owner;
- validator;
- bonded stake;
- pending withdrawal;
- reward balance;
- emergency lock;
- status.

### Messages

- `MsgCreateSingleNominatorPool`
- `MsgDepositSingleNominator`
- `MsgWithdrawSingleNominator`
- `MsgClaimSingleNominatorRewards`
- `MsgEmergencyLockSingleNominator`
- `MsgChangeSingleNominatorValidator`

### Queries

- `SingleNominatorPool`
- `SingleNominatorPools`
- `SingleNominatorRewards`

### Invariants

- only owner can manage stake;
- only one nominator per pool;
- emergency lock blocks withdrawals but not slashing;
- no delegation to jailed validator.

### Tests

- create pool;
- deposit and delegate;
- owner-only withdrawal;
- emergency lock behavior;
- slash behavior;
- export/import during pending withdrawal.

## x/validator-insurance

### Purpose

Validator-level insurance stake that absorbs some penalties before delegator stake is affected.

### State

- validator insurance balance;
- minimum insurance requirement;
- insurance lock period;
- claims;
- claim status;
- slash coverage rules.

### Messages

- `MsgFundValidatorInsurance`
- `MsgWithdrawValidatorInsurance`
- `MsgSubmitInsuranceClaim`
- `MsgResolveInsuranceClaim`

### Queries

- `ValidatorInsurance`
- `InsuranceClaims`
- `InsuranceParams`

### Invariants

- active validator must meet minimum insurance requirement if enabled;
- insurance withdrawal cannot violate active requirement;
- claim payout cannot exceed insurance balance;
- claim cannot be paid twice.

### Tests

- fund insurance;
- reject validator activation without required insurance;
- slash drains insurance first according to params;
- claim payout capped;
- withdrawal delay enforced;
- export/import preserves pending claims.

## x/delegator-protection

### Purpose

Network-level delegator protection fund. A configurable portion of protocol fees can be routed into this fund.

### State

- protection fund balance;
- incoming fee share;
- active compensation claims;
- claim eligibility rules;
- payout history;
- reserve floor;
- max payout per epoch.

### Messages

- `MsgSubmitDelegatorProtectionClaim`
- `MsgApproveDelegatorProtectionClaim`
- `MsgRejectDelegatorProtectionClaim`
- `MsgClaimDelegatorCompensation`
- `MsgUpdateProtectionParams`

### Queries

- `ProtectionFund`
- `ProtectionClaims`
- `DelegatorCompensation`
- `ProtectionParams`

### Invariants

- fund cannot go negative;
- payout cannot exceed max per epoch;
- claim cannot be paid twice;
- fee share must be compatible with treasury distribution proportions.

### Tests

- protocol fee share enters fund;
- valid claim approved and paid;
- duplicate claim rejected;
- max payout enforced;
- insufficient fund handled cleanly;
- export/import preserves claim queue.

## x/reputation

### Purpose

Protocol reputation score for validators and optionally reporters. This score influences election, commission modifiers, warnings, and public ranking.

### State

- validator reputation score;
- reporter reputation score;
- score components;
- epoch snapshots;
- decay parameters;
- penalty events;
- recovery events.

### Messages

- `MsgUpdateReputationParams`
- `MsgApplyReputationPenalty`
- `MsgApplyReputationReward`
- `MsgRecomputeReputation`

### Queries

- `ValidatorReputation`
- `ReporterReputation`
- `ReputationHistory`
- `ReputationParams`

### Invariants

- score stays inside configured range;
- score update is deterministic;
- score cannot be directly modified by unauthorized actors;
- slashing always reduces score if configured.

### Tests

- missed block penalty;
- uptime reward;
- slashing penalty;
- score floor/ceiling;
- deterministic recomputation;
- export/import preserves snapshots.

## x/performance-oracle

### Purpose

Native performance measurement and publication module for validators.

### State

- uptime windows;
- latency samples;
- response-time reports;
- missed block windows;
- peer score reports;
- performance score;
- report source;
- aggregation epoch.

### Messages

- `MsgSubmitPerformanceReport`
- `MsgFinalizePerformanceEpoch`
- `MsgChallengePerformanceReport`

### Queries

- `ValidatorPerformance`
- `PerformanceEpoch`
- `PerformanceReports`
- `PerformanceParams`

### Invariants

- report source must be authorized or slashable;
- aggregation is deterministic;
- outlier handling is deterministic;
- score cannot exceed configured bounds.

### Tests

- aggregate uptime;
- aggregate latency;
- reject malformed report;
- challenge report;
- deterministic order for equal reports;
- export/import during aggregation.

## x/stake-concentration

### Purpose

Controls excessive validator concentration by capping voting power or changing reward/acceptance rules when concentration is too high.

### State

- max voting power percent;
- validator concentration metrics;
- delegation acceptance policy;
- reward reduction policy;
- concentration warnings.

### Messages

- `MsgUpdateConcentrationParams`
- `MsgRecomputeConcentration`

### Queries

- `ValidatorConcentration`
- `NetworkConcentration`
- `ConcentrationParams`

### Invariants

- validator power cannot exceed hard constitutional cap;
- new delegation rejection must be deterministic;
- reward reduction cannot create negative rewards;
- concentration calculation uses canonical validator set.

### Tests

- validator above cap rejects new delegation;
- reward modifier applies;
- power cap enforced across epoch transition;
- export/import preserves concentration metrics.

## x/dynamic-commission

### Purpose

Validator commission modifier based on performance, reputation, and policy bounds.

### State

- base commission;
- effective commission;
- performance modifier;
- reputation modifier;
- commission floor;
- commission ceiling;
- commission history.

### Messages

- `MsgSetBaseCommission`
- `MsgUpdateCommissionParams`
- `MsgRecomputeEffectiveCommission`

### Queries

- `ValidatorCommission`
- `CommissionHistory`
- `CommissionParams`

### Invariants

- effective commission stays inside floor/ceiling;
- commission changes respect rate limits;
- jailed validators cannot receive performance bonuses;
- deterministic rounding.

### Tests

- high performance bonus;
- low performance penalty;
- floor and ceiling enforced;
- rate limit enforced;
- export/import preserves effective commission.

## x/emissions

### Purpose

Dynamic emissions controller. It computes mint amounts based on staking ratio, inflation bounds, validator rewards, treasury targets, and burn policy.

### State

- current inflation;
- target staking ratio;
- current staking ratio;
- annual emission bounds;
- epoch emission amount;
- reward distribution weights;
- burn share;
- treasury share;
- protection fund share;
- validator reward share;
- ecosystem share.

### Messages

- `MsgUpdateEmissionsParams`
- `MsgFinalizeEmissionEpoch`

### Queries

- `EmissionsParams`
- `CurrentInflation`
- `EmissionEpoch`
- `DistributionWeights`

### Invariants

- distribution weights sum to 100%;
- inflation cannot exceed constitutional maximum;
- emission amount cannot be negative;
- minted supply must match `x/mint-authority` accounting;
- rounding remainder must be assigned deterministically.

### Tests

- staking ratio below target increases rewards within bounds;
- staking ratio above target decreases rewards within bounds;
- distribution weights sum validation;
- deterministic rounding;
- export/import preserves emission epoch.

## x/mint-authority

### Purpose

Single native authority for creating new base-denom coins.

This module owns the canonical emission minter account. No other module or user account should be able to mint the base denom directly.

Recommended canonical account:

- module name: `mint-authority`
- system account alias: `AETMint`
- purpose: base-denom emission only
- registered in: `x/system-registry`
- allowed emission caller: `x/emissions`
- optional emergency caller: disabled by default, enabled only through constitutional emergency flow

### State

- mint module account;
- canonical minter account alias;
- total minted by epoch;
- total minted lifetime;
- allowed callers;
- allowed mint denoms;
- mint caps;
- mint events.

### Messages

- `MsgMintProtocolCoins`
- `MsgUpdateMintAuthorityParams`

### Queries

- `MintAuthority`
- `MintedByEpoch`
- `MintedLifetime`
- `MintCaps`

### Invariants

- only authorized modules can mint;
- mint amount must match emissions decision;
- minted amount cannot exceed cap;
- minted denom must be base denom;
- canonical minter account must be registered in `x/system-registry`;
- no user-controlled key may control the minter account;
- `x/emissions` is the only normal module allowed to request scheduled emission minting;
- emergency minting, if enabled, must be bounded by `x/constitution`.

### Tests

- unauthorized mint rejected;
- authorized emission mint succeeds;
- wrong denom rejected;
- cap enforced;
- direct user mint rejected;
- non-emissions module mint rejected unless explicitly allowlisted by constitutional emergency path;
- minter account registration required at genesis;
- export/import preserves lifetime minted counter.

## x/fee-collector

### Purpose

Native fee collection and routing module.

### State

- fee collector module account;
- gas fee balance;
- forwarding fee balance;
- protocol fee balance;
- distribution proportions;
- pending distributions;
- fee history by epoch.

### Messages

- `MsgDistributeFees`
- `MsgUpdateFeeDistributionParams`

### Queries

- `FeeCollector`
- `FeeBalances`
- `FeeDistribution`
- `FeeHistory`

### Invariants

- distribution proportions sum to 100%;
- collected fees cannot disappear;
- distribution cannot create coins;
- base-denom fee policy enforced;
- module account balance equals accounting state.

### Tests

- collect fees;
- route fees to treasury/protection/validators/burn in configured proportions;
- rounding remainder deterministic;
- wrong denom rejected;
- export/import preserves balances;
- invariant compares bank balance and module accounting.

## x/burn

### Purpose

Native irreversible burn module for base-denom and approved assets.

### State

- burn module account;
- burned by denom;
- burned by epoch;
- burn permissions;
- burn reasons.

### Messages

- `MsgBurnProtocolCoins`
- `MsgBurnUserCoins`
- `MsgUpdateBurnParams`

### Queries

- `BurnedByDenom`
- `BurnedByEpoch`
- `BurnParams`

### Invariants

- burned coins are removed from bank supply;
- burn amount must be positive;
- burn cannot be reversed;
- unauthorized protocol burn rejected.

### Tests

- user burn reduces balance and supply;
- protocol burn reduces module balance and supply;
- zero/negative burn rejected;
- unauthorized burn rejected;
- export/import preserves burn counters.

## x/treasury

### Purpose

Protocol treasury with configurable distribution, grants, reserves, and spending controls.

### State

- treasury account;
- reserve balance;
- ecosystem balance;
- grant allocations;
- validator incentive allocations;
- burn allocation;
- spending proposals;
- vesting schedules;
- per-epoch spend cap;
- recipient allowlist if enabled.

### Messages

- `MsgSubmitTreasurySpend`
- `MsgApproveTreasurySpend`
- `MsgRejectTreasurySpend`
- `MsgExecuteTreasurySpend`
- `MsgCancelTreasurySpend`
- `MsgUpdateTreasuryParams`

### Queries

- `TreasuryBalance`
- `TreasuryAllocations`
- `TreasurySpend`
- `TreasurySpends`
- `TreasuryParams`

### Invariants

- treasury balance equals bank module account balance;
- spending cannot exceed available balance;
- per-epoch cap enforced;
- distribution proportions sum to 100%;
- vesting cannot release early;
- unauthorized recipient rejected if allowlist enabled.

### Tests

- fee distribution enters treasury;
- spend proposal lifecycle;
- cap enforcement;
- vesting release schedule;
- insufficient funds rejected;
- export/import preserves pending spends.

## x/scheduler

### Purpose

Native protocol scheduler for periodic jobs, delayed jobs, epoch jobs, and controlled automatic execution.

### State

- scheduled jobs;
- job owner module;
- job type;
- next execution height;
- execution interval;
- max gas;
- retry policy;
- failure count;
- paused jobs;
- job execution history.

### Messages

- `MsgRegisterScheduledJob`
- `MsgPauseScheduledJob`
- `MsgResumeScheduledJob`
- `MsgCancelScheduledJob`
- `MsgExecuteDueJobs`

### Queries

- `ScheduledJob`
- `ScheduledJobs`
- `DueJobs`
- `JobHistory`
- `SchedulerParams`

### Invariants

- job order is deterministic;
- per-block job count is bounded;
- failed job cannot halt block production;
- gas per job and total scheduler gas are bounded;
- only authorized modules can register protocol jobs.

### Tests

- periodic job executes at correct height;
- delayed job executes once;
- job failure increments failure count;
- failed job does not panic block;
- job order deterministic;
- block gas limit respected;
- export/import preserves job queue.

## x/avm-scheduler

### Purpose

Execution coordination for AVM contracts. This module should manage read/write sets, dependency graphs, mailbox scheduling, and bounded parallel execution.

### State

- AVM execution queue;
- dependency graph;
- contract read sets;
- contract write sets;
- execution receipts;
- conflict counters;
- parallelism limits;
- fallback serial execution markers.

### Messages

- `MsgSubmitAVMExecutionBatch`
- `MsgFinalizeAVMExecutionBatch`
- `MsgUpdateAVMSchedulerParams`

### Queries

- `AVMExecutionQueue`
- `AVMExecutionReceipt`
- `AVMDependencyGraph`
- `AVMSchedulerParams`

### Invariants

- conflicting write sets cannot execute in parallel;
- receipt order deterministic;
- fallback execution produces same state root;
- per-block AVM execution bounded;
- failed contract execution cannot corrupt unrelated state.

### Tests

- non-conflicting contracts execute in parallel;
- conflicting contracts serialize;
- deterministic receipt order;
- parallel and serial state roots match;
- malformed read/write set rejected;
- export/import preserves queue and receipts.

## x/actor-registry

### Purpose

Native registry of AVM actors and contract accounts.

### State

- actor id;
- contract address;
- owner;
- code hash;
- storage root;
- mailbox root;
- balance;
- logical time;
- status: active, frozen, deleted, migrated;
- rent status;
- last active height;
- capabilities.

### Messages

- `MsgRegisterActor`
- `MsgUpdateActorCode`
- `MsgFreezeActor`
- `MsgUnfreezeActor`
- `MsgDeleteActor`
- `MsgMigrateActor`

### Queries

- `Actor`
- `ActorsByOwner`
- `ActorsByCodeHash`
- `ActorStatus`
- `ActorMailbox`
- `ActorStorageRoot`

### Invariants

- actor id/address derivation deterministic;
- code hash exists in AVM code store;
- frozen actor cannot execute normal messages;
- deleted actor cannot receive value unless policy explicitly redirects/refunds;
- logical time monotonically increases.

### Tests

- register actor;
- duplicate actor rejected;
- freeze/unfreeze lifecycle;
- delete lifecycle;
- migration updates code hash;
- export/import preserves logical time.

## x/storage-rent

### Purpose

Native storage rent controller for AVM contracts. Contracts with persistent state must pay rent. If rent is not paid, the contract becomes frozen. After a longer retention period, it can be deleted or archived according to policy.

### State

- rent rate per byte per block;
- free storage allowance if any;
- contract storage usage;
- prepaid rent balance;
- rent debt;
- frozen contracts;
- freeze height;
- deletion eligibility height;
- archival proof root;
- rent exemptions for approved system accounts if any.

### Messages

- `MsgPayStorageRent`
- `MsgWithdrawExcessRent`
- `MsgFreezeExpiredContract`
- `MsgUnfreezeContract`
- `MsgDeleteExpiredContract`
- `MsgUpdateStorageRentParams`

### Queries

- `ContractRent`
- `RentDebt`
- `FrozenContracts`
- `DeletionQueue`
- `StorageRentParams`

### Invariants

- rent debt cannot be negative;
- frozen contract cannot execute user messages;
- unfreeze requires full debt payment plus configured buffer;
- deletion cannot happen before retention period;
- storage usage must match actor registry/storage module accounting;
- rent collection cannot create or destroy coins except through configured burn/treasury route.

### Tests

- contract with paid rent remains active;
- contract with unpaid rent freezes;
- frozen contract rejects external/internal execution;
- paying debt unfreezes contract;
- contract deletion after retention period;
- storage usage accounting exactness;
- rent distribution to fee collector/treasury/burn in configured proportions;
- export/import preserves freeze/deletion queues.

## x/identity-root

### Purpose

Native root for `.aet` identity and name infrastructure. It should coordinate ownership, resolver roots, reverse records, and NFT binding if enabled.

### State

- root namespace;
- domain records;
- resolver records;
- reverse records;
- identity NFT binding references;
- expiry heights;
- renewal parameters;
- auction parameters;
- root authorities;
- reserved names;

### Messages

- `MsgRegisterName`
- `MsgRenewName`
- `MsgTransferName`
- `MsgSetResolver`
- `MsgSetReverseRecord`
- `MsgCreateSubdomain`
- `MsgReserveName`
- `MsgReleaseReservedName`

### Queries

- `NameRecord`
- `ResolveName`
- `ReverseRecord`
- `Subdomains`
- `IdentityRootParams`

### Invariants

- name normalization deterministic;
- owner and NFT binding agree if binding is enabled;
- expired name cannot resolve as active;
- reserved names cannot be registered by normal users;
- subdomain ownership follows parent policy.

### Tests

- register name;
- transfer name;
- resolver update;
- reverse record;
- expiry and renewal;
- reserved name rejection;
- export/import preserves sorted records;
- ownership binding invariant.

## x/dex-factory

### Purpose

Native registry/coordinator for approved DEX instances and liquidity venues. This module should not implement every app-level DEX strategy. It should govern creation, registration, safety limits, protocol fee capture, and indexability.

### State

- approved DEX implementations;
- pool registry;
- pair registry;
- fee tiers;
- protocol fee recipient;
- pool risk flags;
- pause flags;
- factory authority;
- AVM contract template references if contract-based DEX pools are enabled later.

### Messages

- `MsgRegisterDEXImplementation`
- `MsgCreatePool`
- `MsgPausePool`
- `MsgResumePool`
- `MsgUpdateDEXFeeParams`
- `MsgSetPoolRiskFlag`

### Queries

- `DEXImplementation`
- `Pool`
- `Pools`
- `PoolsByPair`
- `DEXFactoryParams`

### Invariants

- no duplicate active pair for same implementation and fee tier unless explicitly allowed;
- pool reserves must match bank/contract balance accounting;
- protocol fee routing deterministic;
- paused pool cannot process swaps/add/remove liquidity;
- pool registry sorted canonically.

### Tests

- create pool;
- duplicate pair rejected;
- pause/resume pool;
- fee routing;
- reserve accounting invariant;
- export/import preserves pool registry;
- malformed pool rejected at genesis.

## x/bridge-hub

### Purpose

Native coordination module for bridge adapters and cross-chain asset/security policy.

### State

- approved bridges;
- bridge operators;
- bridge risk status;
- asset mappings;
- daily limits;
- emergency pause flags;
- proof verification policy;
- bridge fee policy.

### Messages

- `MsgRegisterBridge`
- `MsgPauseBridge`
- `MsgResumeBridge`
- `MsgRegisterAssetMapping`
- `MsgUpdateBridgeLimits`
- `MsgSubmitBridgeEvent`
- `MsgFinalizeBridgeEvent`

### Queries

- `Bridge`
- `Bridges`
- `AssetMapping`
- `BridgeLimits`
- `BridgeEvents`

### Invariants

- paused bridge cannot finalize events;
- asset mapping cannot conflict;
- daily limit enforced;
- bridge event cannot finalize twice;
- proof policy must match registered chain policy.

### Tests

- register bridge;
- pause/resume bridge;
- mapping conflict rejected;
- daily limit enforced;
- duplicate event rejected;
- export/import preserves pending events.

## x/cross-chain-registry

### Purpose

Registry of trusted external chains, channels, clients, bridge routes, and risk policies.

### State

- chain id;
- chain status;
- client type;
- trust level;
- channel ids;
- bridge routes;
- light client references;
- risk score;
- finality assumptions;
- timeout parameters.

### Messages

- `MsgRegisterChain`
- `MsgUpdateChainStatus`
- `MsgRegisterChannel`
- `MsgUpdateRiskPolicy`
- `MsgRemoveChain`

### Queries

- `RegisteredChain`
- `RegisteredChains`
- `Channel`
- `BridgeRoute`
- `RiskPolicy`

### Invariants

- chain id unique;
- active bridge route requires active chain;
- finality parameters bounded;
- risk policy cannot be empty for active chain.

### Tests

- register chain;
- register channel;
- pause chain disables routes;
- duplicate chain rejected;
- export/import preserves registry order.

## x/sharding-coordinator

### Purpose

Native coordinator for future shard or zone allocation. It should define shard metadata, validator assignments, load metrics, and rebalance jobs.

### State

- shard id;
- shard status;
- validator assignment;
- load metrics;
- rebalance proposals;
- cross-shard routing parameters;
- shard security level;
- shard state root references.

### Messages

- `MsgRegisterShard`
- `MsgUpdateShardStatus`
- `MsgAssignValidatorsToShard`
- `MsgSubmitShardLoad`
- `MsgProposeShardRebalance`
- `MsgExecuteShardRebalance`

### Queries

- `Shard`
- `Shards`
- `ShardValidators`
- `ShardLoad`
- `RebalanceProposal`

### Invariants

- shard ids unique;
- active shard must have sufficient validator coverage;
- validator cannot exceed shard assignment limit;
- rebalance deterministic;
- cross-shard route exists for active shard pair.

### Tests

- register shard;
- assign validators;
- load update;
- rebalance proposal lifecycle;
- insufficient validator coverage rejected;
- export/import preserves assignments.

## x/config-voting

### Purpose

Native voting path for critical protocol configuration changes. This can be integrated into `x/config` or implemented as a separate module if separation is cleaner.

### State

- proposals;
- votes;
- quorum;
- threshold;
- voting period;
- execution delay;
- veto rules;
- emergency path.

### Messages

- `MsgSubmitConfigProposal`
- `MsgVoteConfigProposal`
- `MsgExecuteConfigProposal`
- `MsgVetoConfigProposal`

### Queries

- `ConfigProposal`
- `ConfigProposals`
- `ConfigVotes`
- `ConfigVotingParams`

### Invariants

- voting power snapshot fixed at proposal height;
- proposal cannot execute before delay;
- proposal cannot execute if it violates constitution;
- vote counting deterministic.

### Tests

- submit/vote/execute;
- quorum failure;
- threshold failure;
- veto flow;
- execution delay;
- voting power snapshot preserved across validator set change.

## Cross-Module Distribution Proportions

Protocol income should be routed through configurable proportions. The exact parameters live in `x/config`, but accounting should be enforced by `x/fee-collector`, `x/treasury`, `x/delegator-protection`, `x/burn`, and validator reward modules.

Example configurable buckets:

- validator rewards;
- treasury;
- delegator protection fund;
- validator insurance reserve;
- ecosystem grants;
- storage rent reserve;
- burn;
- reporter rewards.

Requirements:

- weights must sum to 100%;
- each bucket must have a module account;
- rounding must be deterministic;
- zero-weight buckets are allowed only when explicitly configured;
- required safety buckets may have constitutional minimums;
- distribution must be testable from collected fees to final balances.

## Global Test Matrix

Each system entity should be covered by:

- unit tests for validation and keeper state transitions;
- genesis validation tests;
- export/import roundtrip tests;
- integration tests through block execution;
- adversarial malformed message tests;
- invariant tests comparing module accounting with bank balances;
- fuzz tests for parser/codec-heavy modules;
- migration tests for versioned state;
- long-run localnet smoke tests;
- pause/resume tests where supported;
- authority bypass tests;
- deterministic ordering tests.

## Public Testnet Readiness Checklist

Before enabling these modules in public testnet:

- all module params documented;
- every module has genesis validation;
- every module has export/import tests;
- every module has authority tests;
- system module accounts are listed in `x/system-registry`;
- module invariants run in test suite;
- fee/mint/burn accounting reconciles with bank supply;
- validator set transitions tested across at least 10 epochs;
- storage rent freeze/unfreeze/delete tested with AVM contracts;
- scheduler cannot exceed per-block gas limits;
- no unbounded iteration over user-controlled state;
- no panics on malformed tx/query/genesis;
- all address fields reject zero address;
- cross-module invariants included in `app` tests;
- localnet 5-validator run passes epoch transition, slashing, evidence, fee distribution, storage rent, and export/import restart.

## Recommended Module Naming

Preferred names:

- `x/config`
- `x/constitution`
- `x/system-registry`
- `x/validator-registry`
- `x/validator-election`
- `x/evidence`
- `x/reporter`
- `x/nominator-pool`
- `x/single-nominator-pool`
- `x/validator-insurance`
- `x/delegator-protection`
- `x/reputation`
- `x/performance-oracle`
- `x/stake-concentration`
- `x/dynamic-commission`
- `x/emissions`
- `x/mint-authority`
- `x/fee-collector`
- `x/burn`
- `x/treasury`
- `x/scheduler`
- `x/avm-scheduler`
- `x/actor-registry`
- `x/storage-rent`
- `x/identity-root`
- `x/dex-factory`
- `x/bridge-hub`
- `x/cross-chain-registry`
- `x/sharding-coordinator`

Existing modules can be reused where they already match the responsibility, but each must be audited against this document before being marked production-ready.
