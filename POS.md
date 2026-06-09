# Aetheris Proof-of-Stake Upgrade Layer

Status: Internal design document
Scope: Epoch-based validator economy, workload-aware consensus security, evidence, slashing, delegation markets
Visibility: Private, not for public repository inclusion

## 1. Goal

Design a next-generation Proof-of-Stake system on top of Cosmos SDK and CometBFT.

The target system upgrades the current staking model into a multi-layer economic security engine with:

- Epoch-based validator lifecycle.
- Performance-aware validator selection.
- Delegated capital markets.
- Task-based validator assignments.
- Multi-role consensus security model.
- Verifiable slashing and accountability.
- Structured evidence marketplace.
- Workload-aware validator economics.

The target outcome is a dynamic validator economy where stake, performance, workload assignment, and fault accountability jointly determine validator participation and rewards.

## 2. Baseline System

### 2.1 Current Modules

Current baseline:

- Cosmos SDK `x/staking`.
- Cosmos SDK `x/slashing`.
- Cosmos SDK `x/distribution`.
- Cosmos SDK `x/mint`.
- CometBFT consensus.

### 2.2 Current Limitations

- Static validator set per height or simple epoch boundary.
- Validator selection primarily stake-weighted.
- No workload segmentation.
- No validator specialization.
- No native performance-weighted selection.
- No delegation risk markets.
- No multi-role validator economy.
- No task assignment layer.
- No structured evidence marketplace.
- Limited linkage between validator performance and future economic weight.

## 3. Target Architecture

### 3.1 Layered PoS Model

```text
+----------------------------------------+
| ECONOMIC CONSENSUS LAYER               |
| validator scoring + incentives         |
+----------------------------------------+
| TASK ASSIGNMENT LAYER                  |
| shard/workload validator groups        |
+----------------------------------------+
| VALIDATOR EXECUTION LAYER              |
| block production + verification        |
+----------------------------------------+
| STAKING AND CAPITAL LAYER              |
| delegators, delegation markets         |
+----------------------------------------+
| BASE COMETBFT CONSENSUS                |
+----------------------------------------+
```

### 3.2 Layer Responsibilities

Base CometBFT Consensus:

- Finality.
- Proposal and vote protocol.
- Validator public key set.
- Consensus safety and liveness.

Staking and Capital Layer:

- Validators.
- Delegators.
- Bonded stake.
- Unbonding.
- Redelegation.
- Capital risk preferences.
- Commission and delegation market metadata.

Validator Execution Layer:

- Block production.
- State transition verification.
- Cross-domain proof verification.
- Signature production.
- Fault rejection.

Task Assignment Layer:

- Workload grouping.
- Shard validator groups.
- Zone validator groups.
- Evidence verification subsets.
- Collator and verifier assignments.

Economic Consensus Layer:

- Validator scoring.
- Performance incentives.
- Stake saturation.
- Role-specific reward weights.
- Slashing severity.
- Reporter incentives.
- Treasury, burn, and stabilization routing.

## 4. Epoch-Based Validator Lifecycle

### 4.1 Epoch Definition

Epoch parameters:

- `epoch_duration`: configurable, target `12-24h`.
- `delegation_phase_duration`.
- `election_phase_duration`.
- `assignment_phase_duration`.
- `active_validation_duration`.
- `settlement_phase_duration`.
- `epoch_seed_source`.
- `max_validator_set_change_rate`.

### 4.2 Lifecycle

```text
DELEGATION PHASE
  |
  v
VALIDATOR ELECTION
  |
  v
TASK GROUP ASSIGNMENT
  |
  v
ACTIVE VALIDATION
  |
  v
SETTLEMENT + REWARD + SLASH FINALITY
```

### 4.3 Epoch State

`EpochRecord` fields:

- `epoch_id`
- `start_height`
- `end_height`
- `phase`
- `seed`
- `validator_set_hash`
- `task_group_root`
- `performance_root`
- `reward_root`
- `slash_root`
- `settlement_status`

Phase values:

- `delegation`
- `election`
- `assignment`
- `active`
- `settlement`
- `closed`

### 4.4 Epoch Rules

- Validator set changes activate only at epoch boundary.
- Delegation changes affect election only after configured activation delay.
- Evidence from prior epochs remains valid through slashable window.
- Task assignments are computed from committed validator set and epoch seed.
- Settlement must finalize rewards, penalties, and performance records before epoch closes.

### 4.5 Implementation Tasks

- Implement `x/epoch`.
- Define `EpochRecord`.
- Add epoch phase transition keeper.
- Add epoch begin and end hooks.
- Add query for current and historical epochs.
- Add deterministic epoch seed derivation.
- Add tests for phase boundaries and delayed delegation activation.

## 5. Validator Election Upgrade

### 5.1 Validator Score Model

Replace pure stake-weighted selection with composite scoring:

```text
validator_score =
  stake_weight
  * performance_factor
  * uptime_factor
  * latency_factor
  * reliability_index
```

All factors must use deterministic fixed-point integer math.

### 5.2 Score Components

`stake_weight`:

- Derived from effective stake after saturation.
- Includes self-bond and delegated stake.
- Applies delegation activation delay.

`performance_factor`:

- Based on prior epoch task completion.
- Penalizes missed assignments.
- Rewards correct verification and availability.

`uptime_factor`:

- Based on signed block participation.
- Uses CometBFT vote data and task-specific participation.

`latency_factor`:

- Based only on committed measurement windows.
- Advisory at first activation.
- Must not depend on live local node observations.

`reliability_index`:

- Based on historical slashing, downtime, missed tasks, and evidence outcomes.
- Decays slowly to allow recovery.

### 5.3 Stake Saturation

Effective stake:

```text
effective_stake = min(stake, cap_factor * threshold_stake)
```

Purpose:

- Reduce marginal advantage of excessive concentration.
- Preserve delegator freedom while reducing centralization pressure.
- Encourage stake distribution across reliable validators.

Rules:

- Saturated stake can still earn base delegation rewards if configured.
- Saturated stake receives reduced election weight.
- Saturation cap must be queryable before delegation.
- Saturation does not change actual bonded balance.

### 5.4 Voting Power Soft Cap

Parameter:

- `max_voting_power_per_validator = 15%` default target.

If exceeded:

- Marginal stake has reduced effective weight.
- Delegation UI and queries expose saturation warning.
- Reward curve applies dampening to excess effective power.

### 5.5 ValidatorScoreRecord

Fields:

- `epoch_id`
- `validator_address`
- `raw_stake`
- `effective_stake`
- `stake_weight`
- `performance_factor`
- `uptime_factor`
- `latency_factor`
- `reliability_index`
- `validator_score`
- `saturation_status`
- `score_version`

### 5.6 Implementation Tasks

- Implement `x/validator-economy`.
- Define deterministic fixed-point score math.
- Add score component state.
- Add stake saturation calculation.
- Add election ranking query.
- Add validator set transition limits.
- Add score simulation tests.
- Add centralization and stake-splitting tests.

## 6. Delegation System Upgrade

### 6.1 Capital Layer Model

Delegators are treated as capital providers in a validator marketplace.

Delegation attributes:

- Validator choice.
- Risk appetite.
- Commission tolerance.
- Lock duration preference.
- Reward preference.
- Auto-redelegation preference where supported.
- First-loss exposure preference where supported.

### 6.2 DelegationRecord Extension

Fields:

- `delegator`
- `validator`
- `amount`
- `activation_epoch`
- `risk_appetite`
- `commission_tolerance`
- `lock_duration_preference`
- `reward_strategy`
- `risk_tranche_optional`
- `created_height`
- `updated_height`

### 6.3 Risk Propagation

Rules:

- Slashing applies proportionally to delegators.
- Validator self-bond absorbs first-loss tranche where configured.
- Delegators inherit validator risk history.
- Redelegation does not erase historical slash exposure.
- Delegator risk profile is advisory unless used in explicit delegation product.

### 6.4 Delegation Market Queries

Required queries:

- `QueryValidatorRisk`
- `QueryValidatorEffectiveYield`
- `QueryValidatorSaturation`
- `QueryDelegationRiskExposure`
- `QueryDelegationActivationEpoch`
- `QueryValidatorCommissionHistory`
- `QueryValidatorSlashHistory`
- `QueryValidatorPerformanceHistory`

### 6.5 Commission Tolerance

Rules:

- Delegator can define maximum acceptable commission.
- Validator commission increase above tolerance marks delegation as `commission_exceeded`.
- Optional redelegation alert event is emitted.
- Automatic movement of stake is out of scope unless explicitly implemented as separate delegation strategy.

### 6.6 Lock Duration Preference

Purpose:

- Allow longer lock preference to signal stronger capital commitment.
- Support future reward differentiation.
- Preserve baseline unbonding safety.

Rules:

- Lock preference cannot reduce protocol unbonding period below minimum.
- Longer lock can qualify for reward multiplier only if slashable window is extended accordingly.
- Redelegation remains subject to risk history.

### 6.7 Implementation Tasks

- Extend delegation metadata without breaking `x/staking` compatibility.
- Add risk profile storage.
- Add commission tolerance checks.
- Add yield and risk query endpoints.
- Add first-loss self-bond accounting design.
- Add tests for slash propagation and redelegation risk retention.

## 7. Task Group System

### 7.1 Purpose

Task groups assign validators to workload domains such as:

- Consensus block production.
- Zone execution verification.
- Shard verification.
- Cross-zone proof verification.
- Evidence verification.
- Data availability sampling.
- State sync witness duties.

### 7.2 TaskGroup

Fields:

- `epoch_id`
- `task_group_id`
- `workload_id`
- `workload_type`
- `validator_members`
- `proposer_order`
- `verifier_set`
- `minimum_group_size`
- `stake_weight_root`
- `assignment_seed`
- `activation_height`
- `expiry_height`

Workload types:

- `global_consensus`
- `zone_execution`
- `shard_execution`
- `proof_verification`
- `evidence_verification`
- `data_availability`
- `service_validation`

### 7.3 Assignment Function

```text
task_group = f(
  validator_set,
  workload_id,
  epoch_seed
)
```

Inputs:

- Active validator set.
- Validator scores.
- Workload ID.
- Epoch seed.
- Minimum group size.
- Role requirements.
- Exclusion rules.

Rules:

- Assignment is deterministic.
- Assignment is reproducible from committed state.
- Stake-weighted randomness uses epoch seed.
- Minimum group size is enforced.
- Validators can be assigned to multiple roles within capacity bounds.
- Assignment is known before active validation phase.

### 7.4 Validator Capacity

Capacity fields:

- `max_task_groups`
- `supported_workloads`
- `zone_support`
- `hardware_class_optional`
- `network_class_optional`
- `availability_commitment`

Rules:

- Validator cannot be assigned beyond capacity.
- Unsupported workload assignment is invalid.
- Capacity declaration is slashable if proven false and used for assignment.

### 7.5 Implementation Tasks

- Implement `x/taskgroups`.
- Define workload registry.
- Define deterministic assignment function.
- Add task group root.
- Add capacity declaration.
- Add assignment proof query.
- Add tests for deterministic reproducibility.
- Add tests for minimum group size and overload prevention.

## 8. Block Producer Rotation

### 8.1 Proposer Selection

Within each task group:

- Validators are ordered by priority score.
- Highest priority validator proposes for the slot.
- Fallback rotation is allowed when proposer is unavailable.

Rule:

- Only one canonical proposer exists per slot.
- Other assigned validators act as verifiers.

### 8.2 ProposerPriority

Fields:

- `epoch_id`
- `slot`
- `task_group_id`
- `validator_address`
- `priority_score`
- `fallback_order`
- `proposer_status`

Priority inputs:

- validator score.
- prior proposer performance.
- missed proposal count.
- task-specific reliability.
- stake saturation dampening.

### 8.3 Fallback Rules

- Fallback order is deterministic.
- Fallback activates after missed proposal timeout.
- Fallback proposer must include proof of slot eligibility.
- Repeated missed proposals reduce future priority.

### 8.4 Implementation Tasks

- Add proposer priority calculation.
- Add slot assignment records.
- Add fallback order query.
- Add missed proposer tracking.
- Add tests for proposer conflicts and deterministic fallback.

## 9. Verification Model

### 9.1 Validator Duties

Each validator assigned to a task group must:

- Re-execute state transition where required.
- Validate cross-domain proofs.
- Verify task group correctness.
- Validate consensus ordering.
- Verify message inclusion and receipts.
- Sign valid output.
- Reject invalid output.
- Submit evidence when invalidity is provable.

### 9.2 VerificationReceipt

Fields:

- `epoch_id`
- `task_group_id`
- `workload_id`
- `validator_address`
- `verified_object_hash`
- `result`
- `signature`
- `gas_or_cost_optional`
- `created_height`

Result values:

- `valid`
- `invalid`
- `abstain`
- `unavailable`

### 9.3 Proof Verification

Validators must verify:

- Zone roots.
- Shard roots.
- Message roots.
- Receipt roots.
- Identity proofs.
- Payment settlement proofs.
- Contract execution proofs where configured.

### 9.4 Implementation Tasks

- Define verification receipt schema.
- Add task verification keeper.
- Add receipt aggregation.
- Add verifier participation tracking.
- Add invalid result evidence path.
- Add cross-domain proof verification tests.

## 10. Evidence and Slashing System

### 10.1 Structured Evidence Protocol

Evidence types:

- `double_sign_proof`
- `invalid_state_transition_proof`
- `equivocation_proof`
- `downtime_proof`
- `invalid_task_execution_proof`
- `invalid_proof_acceptance`
- `false_capacity_declaration`
- `invalid_evidence_submission`

### 10.2 Evidence Lifecycle

```text
submit evidence
  |
  v
verification by consensus subset
  |
  v
finality vote
  |
  v
slashing execution
```

### 10.3 EvidenceRecord

Fields:

- `evidence_id`
- `evidence_type`
- `accused_validator`
- `reporter`
- `epoch_id`
- `task_group_id_optional`
- `object_hash`
- `proof_payload_hash`
- `submitted_height`
- `status`
- `verification_group_id`
- `decision_height`
- `penalty_id_optional`

Status values:

- `submitted`
- `in_verification`
- `accepted`
- `rejected`
- `expired`
- `slashed`

### 10.4 Evidence Verification Group

Purpose:

- Assign deterministic validator subset to verify evidence.
- Prevent one validator from unilaterally forcing slashing.
- Bound evidence verification workload.

Rules:

- Verification group is selected from active validators using epoch seed and evidence ID.
- Accused validator is excluded.
- Reporter is excluded if validator.
- Minimum verification group size is enforced.
- Decision threshold is parameterized.

### 10.5 Evidence Marketplace

Reporter model:

- Reporter submits deposit.
- Deposit prevents evidence spam.
- Valid evidence returns deposit and pays reward.
- Invalid evidence burns or redirects part of deposit.
- Reporter reward is capped by penalty amount.

### 10.6 Fishermen Model

Fishermen are external verification nodes that submit fraud proofs.

Rules:

- Small deposit is required.
- Valid proof earns reward.
- Invalid proof loses deposit.
- Fishermen do not need to be active validators.
- Fishermen cannot decide evidence outcome.

Purpose:

- Externalize correctness verification.
- Reduce validator-only trust assumptions.
- Improve detection of invalid task execution.

### 10.7 Implementation Tasks

- Implement `x/evidence`.
- Define evidence schemas.
- Add evidence deposit accounting.
- Add deterministic verification group assignment.
- Add evidence decision voting.
- Add reporter reward calculation.
- Add invalid evidence penalty.
- Add tests for duplicate evidence and forged evidence.

## 11. Slashing Model Upgrade

### 11.1 Penalty Components

Penalty components:

- Validator stake slash.
- Delegator proportional slash.
- Reward confiscation.
- Temporary jail.
- Permanent tombstone.
- Identity invalidation where configured.
- Role suspension.
- Future election score penalty.

### 11.2 Penalty Scaling

```text
penalty = severity * stake_exposure * role_weight
```

Inputs:

- Severity level.
- Stake exposure.
- Role weight.
- Repeat-offense multiplier.
- Task impact.
- Safety impact.
- Liveness impact.

### 11.3 Severity Levels

Severity classes:

- `minor_liveness_fault`
- `major_liveness_fault`
- `repeated_liveness_fault`
- `invalid_task_execution`
- `invalid_state_transition`
- `equivocation`
- `double_sign`
- `evidence_fraud`

### 11.4 Penalty Routing

Reward redistribution:

- Burned portion to inflation control.
- Reporter reward to evidence incentive pool.
- Protocol treasury to stabilization fund.
- Counterparty or affected pool compensation where applicable.

### 11.5 SlashingRecord

Fields:

- `penalty_id`
- `validator_address`
- `evidence_id`
- `severity`
- `stake_exposure`
- `role_weight`
- `slash_amount`
- `delegator_slash_amount`
- `reward_confiscation`
- `jail_until_epoch_optional`
- `tombstone`
- `routing`
- `executed_height`

### 11.6 Implementation Tasks

- Extend slashing keeper with severity matrix.
- Add penalty routing.
- Add reward confiscation.
- Add role suspension.
- Add score penalty integration.
- Add delegator slash propagation.
- Add invariant tests for non-negative balances and exact routing.

## 12. Validator Role Expansion

### 12.1 Roles

Roles:

- Validator.
- Proposer.
- Verifier.
- Evidence Reporter.
- Delegation Operator.
- Collator.
- Fisherman.

Roles can overlap but are logically separable.

### 12.2 RoleRecord

Fields:

- `validator_address`
- `role`
- `epoch_id`
- `status`
- `eligibility_score`
- `capacity`
- `assigned_task_count`
- `performance_score`

### 12.3 Role Rules

Validator:

- Participates in consensus security.
- Must meet minimum stake and performance requirements.

Proposer:

- Produces canonical block or task output for slot.
- Receives proposer-specific rewards.

Verifier:

- Re-executes and signs verification receipts.
- Receives verifier rewards for correct participation.

Evidence Reporter:

- Detects and submits faults.
- May be validator or non-validator with deposit.

Delegation Operator:

- Manages delegated capital strategy where explicitly authorized.
- Must disclose fees and risk policy.

Collator:

- Assembles transactions, state transitions, and proof bundles.
- Does not finalize without validator verification.

Fisherman:

- External fault detector.
- Submits fraud proofs with deposit.

### 12.4 Implementation Tasks

- Define role registry.
- Add role eligibility checks.
- Add role-specific performance metrics.
- Add role-specific rewards.
- Add role suspension on faults.
- Add tests for role overlap and role-specific penalties.

## 13. Collator Model

### 13.1 Purpose

Collators are optional block assembly agents.

They can:

- Aggregate transactions.
- Prepare state transitions.
- Generate proof bundles.
- Build candidate outputs for task groups.

Validators remain responsible for verification and signing.

### 13.2 CollatorRecord

Fields:

- `collator_id`
- `operator_address`
- `supported_workloads`
- `bond_optional`
- `reputation`
- `status`
- `registered_epoch`

### 13.3 Collator Rules

- Collator output is advisory until verified.
- Collator cannot finalize blocks alone.
- Invalid collator output can be penalized if bonded.
- Validators must independently verify collator output.

### 13.4 Implementation Tasks

- Add collator registry.
- Add candidate output schema.
- Add collator bond option.
- Add validator verification path.
- Add invalid collator output evidence type.

## 14. Performance-Based Economics

### 14.1 Reward Formula

```text
reward =
  base_emission
  * uptime_score
  * latency_score
  * correctness_score
  * task_completion_rate
```

All components use deterministic fixed-point math.

### 14.2 Reward Components

`base_emission`:

- Epoch reward budget from mint and fee distribution.

`uptime_score`:

- Vote participation and assigned task availability.

`latency_score`:

- Based on committed performance reports where enabled.

`correctness_score`:

- Penalizes invalid signatures, invalid task outputs, and accepted evidence.

`task_completion_rate`:

- Measures completed assigned tasks over expected tasks.

### 14.3 Reward Dampening

Bad performance reduces:

- Current epoch rewards.
- Future election probability.
- Delegation attractiveness.
- Role eligibility.
- Collator assignment probability if applicable.

### 14.4 PerformanceRecord

Fields:

- `epoch_id`
- `operator_address`
- `role`
- `assigned_tasks`
- `completed_tasks`
- `missed_tasks`
- `invalid_tasks`
- `uptime_score`
- `latency_score`
- `correctness_score`
- `task_completion_rate`
- `reward_multiplier`

### 14.5 Implementation Tasks

- Implement `x/performance`.
- Add role-specific metric collection.
- Add reward multiplier calculation.
- Add performance query endpoints.
- Integrate with distribution rewards.
- Add tests for performance reward bounds and score manipulation.

## 15. Unbonding and Risk Window

### 15.1 Parameters

Target unbonding period:

- `7-21 days`, configurable.

Slashable window:

- Extends beyond unbonding.
- Covers delayed evidence submission.
- Covers prior epoch task assignments.

### 15.2 Rules

- Unbonding stake remains slashable for historical faults.
- Redelegation does not reset risk history.
- Delegation activation and unbonding are epoch-aware.
- Validator self-bond changes have activation delay.
- Pending unbonding stake participates in slash exposure where fault occurred before exit.

### 15.3 RiskWindowRecord

Fields:

- `stake_owner`
- `validator_address`
- `amount`
- `start_epoch`
- `end_epoch`
- `slashable_until_epoch`
- `risk_history_root`
- `status`

### 15.4 Implementation Tasks

- Extend unbonding records with epoch metadata.
- Add slashable window tracking.
- Add redelegation risk history.
- Add query for slash exposure.
- Add tests for delayed evidence against unbonding stake.

## 16. Economic Security Model

### 16.1 Security Formula

```text
security =
  total_stake_at_risk
  * participation_rate
  * slashing_efficiency
```

Goals:

- Maximize stake at risk.
- Maximize participation distribution.
- Minimize centralization pressure.
- Increase detection and slashing efficiency.
- Preserve validator operational viability.

### 16.2 Metrics

Required metrics:

- Total bonded stake.
- Effective stake.
- Stake saturation ratio.
- Top-N voting power concentration.
- Participation rate.
- Slashing efficiency.
- Evidence acceptance rate.
- Average validator score.
- Delegation risk distribution.
- Task completion rate.

### 16.3 Centralization Controls

Controls:

- Effective stake saturation.
- Reward dampening above soft cap.
- Delegation risk warnings.
- Bootstrap path for new reliable validators.
- Task assignment diversity constraints.
- Concentration metrics.

### 16.4 Implementation Tasks

- Add security metric queries.
- Add centralization dashboard data.
- Add concentration invariant alerts.
- Add simulations for stake concentration and stake splitting.

## 17. Cosmos SDK Compatibility

### 17.1 Implementation Strategy

Initial implementation must extend existing modules rather than replacing them.

Core upgrades:

- Extend `x/staking`.
- Extend `x/slashing`.
- Extend `x/distribution`.
- Extend `x/mint` reward inputs.
- Add middleware for scoring, epoch management, task assignment, and performance.

### 17.2 New Modules

Required modules:

- `x/epoch`
- `x/validator-economy`
- `x/taskgroups`
- `x/evidence`
- `x/performance`

Optional modules:

- `x/delegation-market`
- `x/collators`
- `x/fishermen`
- `x/security-metrics`

### 17.3 Module Boundaries

`x/epoch`:

- Epoch lifecycle and phase transitions.
- Epoch seed.
- Epoch queries.

`x/validator-economy`:

- Validator score.
- Effective stake.
- Stake saturation.
- Election ranking.
- Role eligibility.

`x/taskgroups`:

- Workload registry.
- Task group assignment.
- Proposer rotation.
- Verification groups.

`x/evidence`:

- Structured evidence records.
- Evidence deposits.
- Verification group decisions.
- Reporter rewards.

`x/performance`:

- Uptime.
- Latency.
- Correctness.
- Task completion.
- Reward multipliers.

### 17.4 Keeper Integration

Integration points:

- Staking keeper for validator and delegation state.
- Slashing keeper for jail, tombstone, and slash execution.
- Distribution keeper for reward allocation.
- Mint keeper for epoch reward budget.
- Bank keeper for deposits, reporter rewards, and penalty routing.
- Governance keeper for parameter updates.

### 17.5 Implementation Tasks

- Define keeper interfaces.
- Add hooks from staking lifecycle to epoch logic.
- Add hooks from slashing to performance and validator economy.
- Add distribution reward multiplier integration.
- Add migration handlers preserving existing staking state.
- Add export/import support for new modules.

## 18. State Model

### 18.1 Key Prefixes

Epoch:

- `epoch/current`
- `epoch/records/{epoch_id}`
- `epoch/phase/{epoch_id}`
- `epoch/seed/{epoch_id}`

Validator economy:

- `valecon/scores/{epoch_id}/{validator}`
- `valecon/effective_stake/{epoch_id}/{validator}`
- `valecon/saturation/{epoch_id}/{validator}`
- `valecon/roles/{epoch_id}/{validator}/{role}`

Task groups:

- `taskgroups/groups/{epoch_id}/{task_group_id}`
- `taskgroups/workloads/{workload_id}`
- `taskgroups/assignments/{epoch_id}/{validator}/{task_group_id}`
- `taskgroups/proposer/{epoch_id}/{slot}/{task_group_id}`

Evidence:

- `evidence/records/{evidence_id}`
- `evidence/by_accused/{validator}/{evidence_id}`
- `evidence/by_reporter/{reporter}/{evidence_id}`
- `evidence/verification_groups/{evidence_id}`
- `evidence/deposits/{evidence_id}`

Performance:

- `performance/records/{epoch_id}/{operator}/{role}`
- `performance/uptime/{epoch_id}/{validator}`
- `performance/correctness/{epoch_id}/{validator}`
- `performance/tasks/{epoch_id}/{validator}`

Risk windows:

- `risk/unbonding/{delegator}/{validator}/{creation_height}`
- `risk/redelegation/{delegator}/{src_validator}/{dst_validator}/{epoch_id}`
- `risk/exposure/{epoch_id}/{validator}/{delegator}`

### 18.2 Root Commitments

Required roots:

- `epoch_root`
- `validator_score_root`
- `task_group_root`
- `evidence_root`
- `performance_root`
- `slashing_root`
- `risk_window_root`

### 18.3 State Invariants

- Every active validator has score record for active epoch.
- Effective stake is never greater than raw stake.
- Task group members are active validators.
- Task group size meets minimum group size.
- Accused validator is excluded from evidence verification group.
- Every accepted evidence record maps to exactly one penalty decision.
- Slash routing sums exactly to penalty amount.
- Unbonding stake remains slashable for faults within risk window.

## 19. Messages and Queries

### 19.1 Messages

Epoch:

- `MsgStartEpoch`
- `MsgAdvanceEpochPhase`
- `MsgFinalizeEpochSettlement`

Validator economy:

- `MsgDeclareValidatorCapacity`
- `MsgUpdateValidatorMetadata`
- `MsgSetDelegationRiskProfile`
- `MsgUpdateCommissionTolerance`

Task groups:

- `MsgRegisterWorkload`
- `MsgAssignTaskGroups`
- `MsgSubmitVerificationReceipt`
- `MsgReportMissedTask`

Evidence:

- `MsgSubmitEvidence`
- `MsgVoteEvidenceDecision`
- `MsgFinalizeEvidence`
- `MsgClaimReporterReward`

Performance:

- `MsgSubmitPerformanceReport`
- `MsgFinalizePerformanceEpoch`

Collators and fishermen:

- `MsgRegisterCollator`
- `MsgSubmitCollatorOutput`
- `MsgRegisterFisherman`
- `MsgSubmitFraudProof`

### 19.2 Queries

- `QueryCurrentEpoch`
- `QueryEpoch`
- `QueryValidatorScore`
- `QueryValidatorEffectiveStake`
- `QueryValidatorSaturation`
- `QueryElectionRanking`
- `QueryTaskGroup`
- `QueryTaskGroupsByValidator`
- `QueryProposerForSlot`
- `QueryEvidence`
- `QueryEvidenceByValidator`
- `QueryPerformanceRecord`
- `QueryDelegationRiskExposure`
- `QuerySlashableWindow`
- `QuerySecurityMetrics`
- `QueryValidatorRoleEligibility`

## 20. Migration Strategy

### 20.1 Phase 1: Scoring and Epoch Simulation

Tasks:

- Keep Cosmos staking unchanged.
- Add `x/epoch`.
- Add read-only validator scoring.
- Add effective stake simulation.
- Add performance metric collection.
- Add score and saturation queries.

Exit criteria:

- Existing staking behavior is unchanged.
- Validator scores can be computed and compared.
- Epoch simulation matches deterministic replay.

### 20.2 Phase 2: Task Groups and Roles

Tasks:

- Add `x/taskgroups`.
- Add workload registry.
- Add task group assignment.
- Add role records.
- Add proposer priority records.
- Add verification receipts.

Exit criteria:

- Validators can be deterministically assigned to workload groups.
- Roles and task groups are queryable.
- Assignment roots are reproducible.

### 20.3 Phase 3: Performance-Based Rewards

Tasks:

- Activate reward multipliers.
- Integrate `x/performance` with distribution.
- Add task completion rewards.
- Add missed task penalties to future score.
- Add delegation risk queries.

Exit criteria:

- Rewards reflect deterministic performance metrics.
- Reward changes are bounded.
- Performance impact is visible before delegation.

### 20.4 Phase 4: Full Economic Consensus Activation

Tasks:

- Activate stake saturation in validator election.
- Activate performance-weighted selection.
- Activate structured evidence verification.
- Activate severity-based slashing.
- Activate reporter rewards and penalty routing.
- Activate collator and fisherman roles where configured.

Exit criteria:

- Validator selection uses stake and performance.
- Evidence and slashing are structured and test-covered.
- Task-based validator economy is live.

## 21. Required Test Coverage

### 21.1 Unit Tests

- Epoch phase transition.
- Epoch seed derivation.
- Validator score calculation.
- Effective stake saturation.
- Reward multiplier calculation.
- Task assignment function.
- Proposer priority calculation.
- Evidence ID derivation.
- Penalty scaling.
- Slash routing.
- Risk window calculation.

### 21.2 Integration Tests

- Existing staking state migrates into epoch system.
- Delegation affects future epoch only after activation delay.
- Validator election ranking is deterministic.
- Task groups are reproducible from epoch seed.
- Proposer fallback works.
- Verification receipts aggregate correctly.
- Valid evidence triggers penalty.
- Invalid evidence burns reporter deposit.
- Unbonding stake is slashed for historical fault.
- Distribution rewards use performance multiplier.

### 21.3 Invariant Tests

- Effective stake <= raw stake.
- Total active task group membership references active validators.
- Penalty routing exactly equals slashed amount.
- Reporter rewards do not exceed configured cap.
- Validator cannot escape slash exposure by redelegation.
- Evidence cannot be finalized twice.
- Task group root matches assignments.
- Performance root matches records.

### 21.4 Simulation Tests

- Stake concentration above soft cap.
- Stake splitting across validators.
- Low participation epoch.
- Repeated downtime.
- Invalid task execution.
- Collator invalid output.
- Fisherman valid and invalid proof submissions.
- High evidence spam.
- Validator churn at epoch boundary.
- Delegation market response to commission increase.

### 21.5 Performance Tests

- Validator score calculation for 400 validators.
- Task group assignment for many workloads.
- Evidence verification group assignment.
- Epoch settlement runtime.
- Reward distribution with performance multipliers.
- Query latency for validator score and risk data.

## 22. Observability

### 22.1 Metrics

- Current epoch.
- Current epoch phase.
- Active validator count.
- Average validator score.
- Effective stake total.
- Raw stake total.
- Saturated stake amount.
- Top-N voting power concentration.
- Task groups active.
- Average task completion rate.
- Evidence submitted.
- Evidence accepted.
- Evidence rejected.
- Slashed amount by severity.
- Reporter rewards paid.
- Performance reward multiplier distribution.
- Unbonding slash exposure.

### 22.2 Events

- `epoch_started`
- `epoch_phase_advanced`
- `epoch_settled`
- `validator_score_updated`
- `validator_saturated`
- `task_group_assigned`
- `proposer_selected`
- `verification_receipt_submitted`
- `evidence_submitted`
- `evidence_accepted`
- `evidence_rejected`
- `validator_slashed`
- `reporter_reward_paid`
- `performance_record_finalized`
- `delegation_risk_profile_updated`

### 22.3 Alerts

- Validator concentration above threshold.
- Task group below minimum size.
- Evidence spam spike.
- Performance score collapse.
- Epoch settlement delayed.
- Slash routing invariant failure.
- Low participation rate.
- Repeated proposer failure.
- Unbonding exposure backlog.

## 23. Governance Parameters

Epoch:

- `epoch_duration`
- `phase_durations`
- `epoch_seed_source`
- `settlement_work_limit`

Validator economy:

- `score_component_weights`
- `stake_saturation_cap_factor`
- `threshold_stake`
- `max_voting_power_per_validator`
- `score_decay_rate`
- `minimum_score_for_active_set`

Task groups:

- `minimum_group_size`
- `maximum_task_groups_per_validator`
- `workload_role_requirements`
- `assignment_seed_domain`

Evidence:

- `evidence_deposit`
- `evidence_submission_window`
- `verification_group_size`
- `decision_threshold`
- `reporter_reward_cap`

Slashing:

- `severity_matrix`
- `role_weight_matrix`
- `repeat_offense_multiplier`
- `burn_allocation`
- `treasury_allocation`
- `reporter_allocation`

Performance:

- `reward_multiplier_bounds`
- `uptime_window`
- `latency_window`
- `task_completion_window`
- `performance_decay_rate`

Unbonding:

- `unbonding_period`
- `slashable_window`
- `redelegation_risk_retention`

## 24. Acceptance Criteria

The PoS upgrade layer is ready for implementation planning when:

- Epoch lifecycle is deterministic and queryable.
- Validator scoring is deterministic and fixed-point.
- Effective stake saturation reduces marginal centralization pressure.
- Delegation risk metadata is stored and queryable.
- Task groups are reproducible from committed validator set and epoch seed.
- Proposer rotation and fallback are deterministic.
- Validators submit verification receipts for assigned workloads.
- Structured evidence has deposit, verification group, decision, and finalization paths.
- Slashing supports severity, role weight, reward confiscation, jailing, tombstone, and routing.
- Fishermen and collators are defined as optional roles with bounded authority.
- Performance-based rewards are integrated with distribution using bounded multipliers.
- Unbonding and redelegation retain slash exposure through configured risk window.
- Migration can start with read-only scoring while preserving existing `x/staking` behavior.
- Tests cover determinism, invariants, simulations, and performance.
