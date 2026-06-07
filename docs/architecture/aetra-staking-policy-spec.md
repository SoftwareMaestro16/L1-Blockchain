# x/aetra-staking-policy Module Specification

Purpose: control effective voting power, delegation overflow, commission policy, and anti-concentration incentives.

This module is the central anti-centralization module of Aetra.

## Responsibilities

The module must:

- calculate raw validator stake;
- calculate effective validator stake;
- calculate overflow stake;
- enforce or expose effective voting power cap;
- calculate reward multiplier for overflow stake;
- expose delegation concentration warnings;
- enforce commission floor;
- enforce max commission;
- enforce max commission change rate;
- expose top-N concentration metrics;
- validate governance param changes;
- emit events for cap/overflow/commission policy changes;
- remain deterministic and export/import safe.

## Production Rule

`x/aetra-staking-policy` is not complete when only cap math exists. The production definition is:

```text
staking policy = stake math + cap enforcement/exposure + overflow accounting + commission policy + concentration metrics + governance params + events + export/import safety + tests + docs
```

Every responsibility must be represented in code, genesis/governance parameter validation, query surface, events where state changes, and tests. If a responsibility is temporarily query-only instead of enforcement, the behavior must be explicit and covered by tests so the chain does not silently present fake anti-centralization guarantees.

## Implementation Contract

The implementation gate is `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `AetraStakingPolicyModuleName` must be `x/aetra-staking-policy`;
- `DefaultAetraStakingPolicySpecEvidence` must cover the module purpose and central anti-centralization role;
- `BuildAetraStakingPolicySpecReport` must require all responsibilities from this document;
- `ValidateAetraStakingPolicySpec` must reject incomplete evidence;
- missing stake math, cap, overflow, reward multiplier, concentration warnings, commission controls, top-N metrics, governance validation, policy-change events, or deterministic export/import safety must fail validation;
- module identity must fail validation when missing or not equal to `x/aetra-staking-policy`.

## Required Tests

The module specification tests must prove:

- the default evidence covers all responsibilities;
- raw stake, effective stake, overflow stake, effective voting power cap, and overflow reward multiplier are mandatory;
- delegation concentration warnings, commission floor, max commission, max commission change rate, top-N concentration metrics, governance param validation, policy events, and export/import safety are mandatory;
- purpose and central anti-centralization role are mandatory;
- wrong or missing module identity is rejected.

## 22.2 State

Suggested state:

```text
Params:
  MaxValidatorsSoftTarget
  ValidatorPowerCapBps
  ValidatorPowerCapSchedule
  OverflowRewardMultiplierBps
  CommissionFloorBps
  CommissionMaxBps
  CommissionMaxDailyChangeBps
  Top10TargetBps
  Top20TargetBps
  Top33TargetBps
  MinSelfBond
  MinValidatorBond
  WarningThresholdBps

ValidatorPolicy:
  OperatorAddress
  RawBondedTokens
  EffectiveBondedTokens
  OverflowBondedTokens
  EffectivePowerBps
  IsOverCap
  RewardMultiplierBps
  LastCommissionChangeTime
  LastCommissionRateBps

ConcentrationSnapshot:
  Height
  BondedRatio
  ActiveValidators
  Top10Bps
  Top20Bps
  Top33Bps
  NakamotoCoefficientEstimate
```

All decimal values should use integer basis points or SDK decimal types consistently. Avoid floating point.

### State Requirements

`Params` must contain all governance-controlled and genesis-validated knobs needed for validator set sizing, effective power cap, overflow reward treatment, commission bounds, concentration targets, bond minimums, and warning thresholds.

`ValidatorPolicy` must be derived deterministically from staking state and policy params. It must expose raw bonded tokens, effective bonded tokens, overflow bonded tokens, effective power in basis points, over-cap status, reward multiplier, and commission-change tracking.

`ConcentrationSnapshot` must be a deterministic network-level view for observability, queries, dashboards, governance alerts, and export/import. It must record height, bonded ratio, active validators, top-10/top-20/top-33 concentration, and a Nakamoto coefficient estimate.

### State Implementation Contract

The state gate is `BuildAetraStakingPolicyStateSpecReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyStateSpecEvidence` must list every required `Params`, `ValidatorPolicy`, and `ConcentrationSnapshot` field;
- missing required fields must fail validation;
- duplicate fields must fail validation;
- unexpected fields must fail validation;
- module identity must be `x/aetra-staking-policy`;
- decimal accounting must explicitly use integer basis points or SDK decimal types;
- floating point accounting must fail validation.

## 22.3 Parameter Rules

Parameter validation:

```text
ValidatorPowerCapBps:
  min: 100      # 1%
  max: 500      # 5%
  recommended: 200-300

OverflowRewardMultiplierBps:
  min: 0
  max: 10000
  recommended: 0-3000 for overflow zone

CommissionFloorBps:
  min: 0
  max: 1000
  recommended: 300-500

CommissionMaxBps:
  min: CommissionFloorBps
  max: 3000
  recommended: 1500-2000

CommissionMaxDailyChangeBps:
  min: 1
  max: 500
  recommended: 50-100
```

Governance must not be able to set:

- cap below 1%;
- max commission below floor;
- overflow multiplier above normal reward multiplier;
- invalid top-N targets;
- zero active validator target;
- negative or overflowing math values.

### Parameter Validation Requirements

`ValidatorPowerCapBps` must stay within `100-500` bps. The recommended production operating range is `200-300` bps. Governance proposals below `100` bps must be rejected because they would make the validator set unusable and create liveness/economic distortions.

`OverflowRewardMultiplierBps` must stay within `0-10000` bps. A value above `10000` would pay overflow stake more than normal stake and must be impossible. For Aetra's anti-centralization design, the recommended overflow-zone range is `0-3000` bps.

`CommissionFloorBps` must stay within `0-1000` bps, with `300-500` bps recommended. `CommissionMaxBps` must be greater than or equal to `CommissionFloorBps` and must not exceed `3000` bps, with `1500-2000` bps recommended.

`CommissionMaxDailyChangeBps` must stay within `1-500` bps, with `50-100` bps recommended. A zero daily change would make commission governance unnecessarily rigid, while a high daily change enables commission bait-and-switch behavior.

Top-N concentration targets must be positive basis-point values, must not exceed `10000`, and must preserve ordering:

```text
0 < Top10TargetBps <= Top20TargetBps <= Top33TargetBps <= 10000
```

`MaxValidatorsSoftTarget` must be greater than zero. Negative values, overflowing arithmetic inputs, and floating point accounting must be rejected before proposal execution or genesis acceptance.

### Parameter Implementation Contract

The parameter gate is `BuildAetraStakingPolicyParameterReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyParameterRuleSet` must encode min, max, and recommended ranges for all bounded bps parameters;
- `DefaultAetraStakingPolicyParameterValues` must pass validation;
- `ValidateAetraStakingPolicyParameterValues` must reject unsafe governance values;
- cap below `100` bps must fail validation;
- `CommissionMaxBps < CommissionFloorBps` must fail validation;
- `OverflowRewardMultiplierBps > 10000` must fail validation;
- invalid top-N target ordering or values above `10000` must fail validation;
- `MaxValidatorsSoftTarget <= 0` must fail validation;
- negative math inputs must fail validation.

## 22.4 Effective Power Calculation

The implementation must define whether cap affects:

1. only reward calculation;
2. actual CometBFT voting power;
3. both.

Recommended staged approach:

```text
Stage 1:
  cap affects rewards and delegation warnings
  low consensus risk

Stage 2:
  cap affects effective staking power used for validator updates
  requires deeper integration and heavy tests
```

If Stage 2 is implemented, the staking keeper integration must ensure:

- validator updates sent to CometBFT use capped power;
- total voting power remains consistent;
- no validator can exceed cap;
- delegation and unbonding shares remain correct;
- slashing can still slash the underlying raw stake;
- evidence handling remains correct.

### Effective Power Requirements

Aetra must never leave effective power semantics ambiguous. The module must expose whether the cap is reward-only, CometBFT-voting-power-only, or applies to both rewards and voting power.

Stage 1 is the default safe rollout. In Stage 1, the cap changes reward multipliers and delegation warnings only. It must not mutate validator updates sent to CometBFT. This keeps consensus behavior close to standard staking while still steering delegators away from overloaded validators.

Stage 2 is a consensus-affecting rollout. In Stage 2, the cap affects the effective staking power used to build validator updates for CometBFT. This requires deeper staking keeper integration, migration planning, localnet/load testing, evidence tests, slashing tests, export/import tests, and invariant tests.

Stage 2 must keep two concepts separate:

- raw stake: the underlying slashable and withdrawable stake used for delegation shares, unbonding, and slashing;
- effective power: the capped power exposed to CometBFT for validator set updates.

Slashing must apply to raw stake, not merely to capped effective power. Evidence handling must remain compatible with CometBFT and must not lose accountability for validators whose raw stake exceeds their capped voting power.

### Effective Power Implementation Contract

The effective power gate is `BuildAetraStakingPolicyEffectivePowerReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyEffectivePowerStage1Evidence` must define Stage 1 as rewards plus delegation warnings only;
- Stage 1 must fail if it touches actual CometBFT voting power;
- `DefaultAetraStakingPolicyEffectivePowerStage2Evidence` must define Stage 2 as capped validator-update power;
- Stage 2 must require deeper integration and heavy tests;
- Stage 2 must require validator updates sent to CometBFT to use capped power;
- Stage 2 must require total voting power consistency;
- Stage 2 must require no validator to exceed cap;
- Stage 2 must require delegation and unbonding shares to remain correct;
- Stage 2 must require slashing to slash underlying raw stake;
- Stage 2 must require evidence handling to remain correct;
- missing cap-scope definition, unknown stage, or wrong module identity must fail validation.

## 22.5 Messages

Required governance-only or authority-only messages:

```text
MsgUpdateStakingPolicyParams
MsgUpdateValidatorPowerCapSchedule
MsgSetCommissionPolicy
```

Optional validator messages:

```text
MsgRegisterValidatorIdentity
MsgUpdateValidatorIdentity
MsgAcknowledgeOverCapWarning
```

All messages must:

- validate authority;
- validate signer;
- reject malformed addresses;
- reject invalid params;
- emit events;
- be covered by tests.

### Message Requirements

`MsgUpdateStakingPolicyParams`, `MsgUpdateValidatorPowerCapSchedule`, and `MsgSetCommissionPolicy` must be governance-only or authority-only. They must never be callable by normal user accounts unless the account is the configured governance/authority address.

Validator identity and warning acknowledgement messages are optional because they support operator metadata and UX, not consensus-critical cap enforcement. If implemented, they must validate the validator/operator signer and must not allow one validator to mutate another validator's identity or warning state.

Malformed bech32 addresses, empty authority, wrong signer, invalid params, duplicate identity keys, and invalid over-cap warning acknowledgements must be rejected before state mutation. Successful messages must emit stable events for explorers, indexers, wallets, governance dashboards, and audit trails.

### Message Implementation Contract

The message gate is `BuildAetraStakingPolicyMessageSpecReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyMessageSpecEvidence` must include all required governance/authority messages;
- `DefaultAetraStakingPolicyMessageSpecEvidence` must include all optional validator message names if the optional surface is enabled;
- missing required messages must fail validation;
- duplicate or unexpected message names must fail validation;
- every message must validate authority;
- every message must validate signer;
- every message must reject malformed addresses;
- every message must reject invalid params;
- every message must emit events;
- every message must be covered by tests;
- wrong module identity must fail validation.

## 22.6 Queries

Required queries:

```text
Query/Params
Query/ValidatorPolicy
Query/ValidatorEffectivePower
Query/ValidatorOverflow
Query/TopNConcentration
Query/DelegationWarning
Query/CommissionPolicy
Query/ConcentrationSnapshot
Query/NakamotoCoefficient
```

Query responses must be stable and indexer-friendly.

### Query Requirements

Queries must return deterministic, versionable, indexer-friendly response shapes. Response field names should remain stable across releases unless there is a documented migration. Numeric values must use integer token amounts, basis points, or SDK decimal strings consistently.

`Query/ValidatorPolicy` must expose the full validator policy view for wallets and explorers. `Query/ValidatorEffectivePower` and `Query/ValidatorOverflow` must provide focused views for validator-set and delegation UX. `Query/TopNConcentration`, `Query/ConcentrationSnapshot`, and `Query/NakamotoCoefficient` must support public decentralization dashboards and governance alerts.

`Query/CommissionPolicy` must expose commission floor, max commission, and max daily change. `Query/DelegationWarning` must expose over-cap and near-cap warnings without requiring clients to recompute cap math incorrectly.

### Query Implementation Contract

The query gate is `BuildAetraStakingPolicyQuerySpecReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyQuerySpecEvidence` must include all required query names;
- missing required queries must fail validation;
- duplicate or unexpected query names must fail validation;
- query responses must be stable;
- query responses must be indexer-friendly;
- wrong module identity must fail validation.

## 22.7 Events

Required events:

```text
aetra.staking_policy.params_updated
aetra.staking_policy.validator_over_cap
aetra.staking_policy.validator_back_under_cap
aetra.staking_policy.commission_rejected
aetra.staking_policy.concentration_snapshot
aetra.staking_policy.reward_multiplier_changed
```

### Event Requirements

Events must use stable names and indexer-friendly attributes. Event attributes should include enough context for explorers and monitoring systems to reconstruct policy changes and validator status transitions without replaying custom application logic.

`aetra.staking_policy.params_updated` must be emitted when governance/authority changes staking policy params. `aetra.staking_policy.validator_over_cap` and `aetra.staking_policy.validator_back_under_cap` must mark validator cap status transitions. `aetra.staking_policy.commission_rejected` must be emitted when a commission update is rejected by policy. `aetra.staking_policy.concentration_snapshot` must be emitted when the module records or publishes network concentration metrics. `aetra.staking_policy.reward_multiplier_changed` must be emitted when a validator reward multiplier changes because of cap or concentration policy.

### Event Implementation Contract

The event gate is `BuildAetraStakingPolicyEventSpecReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyEventSpecEvidence` must include all required event names;
- missing required events must fail validation;
- duplicate or unexpected event names must fail validation;
- event names must be stable;
- event attributes must be indexer-friendly;
- wrong module identity must fail validation.

## 22.8 Invariants

Required invariants:

- effective power never exceeds configured cap;
- overflow stake is never negative;
- raw stake = effective stake + overflow stake for capped calculation;
- commission floor <= commission <= commission max;
- commission change <= max daily change;
- top-N calculations do not exceed 100%;
- state export/import preserves policy state.

### Invariant Requirements

The module must treat these invariants as production safety conditions, not observability-only metrics. Violations must be detectable in unit tests, integration tests, simulation/property tests where appropriate, and export/import test coverage.

The capped-power calculation must conserve stake accounting:

```text
raw stake = effective stake + overflow stake
```

`effective power` must never exceed the configured `ValidatorPowerCapBps`. `overflow stake` must never become negative. Commission validation must enforce both absolute bounds and daily-change bounds. Top-N concentration math must clamp or reject impossible results above `10000` bps. Export/import must preserve all policy params, validator policy state, warning acknowledgements, concentration snapshots, and any state needed to enforce commission-change windows.

### Invariant Implementation Contract

The invariant gate is `BuildAetraStakingPolicyInvariantSpecReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- effective power never exceeds configured cap;
- overflow stake is never negative;
- raw stake equals effective stake plus overflow stake;
- commission remains between floor and max;
- commission change remains within max daily change;
- top-N calculations never exceed `10000` bps;
- export/import preserves policy state;
- invariants must be covered by tests;
- wrong module identity must fail validation.

## 22.9 Tests

Required tests:

- cap math for 100 validators;
- cap math for 150 validators;
- cap math for 250 validators;
- cap math for 300 validators;
- validator crossing cap upward;
- validator crossing cap downward;
- delegation to over-cap validator;
- redelegation from over-cap validator;
- unbonding from over-cap validator;
- slashing over-cap validator;
- commission below floor rejected;
- commission above max rejected;
- commission daily jump rejected;
- governance param update accepted within bounds;
- governance param update rejected outside bounds;
- export/import with over-cap validators;
- deterministic concentration snapshot.

### Test Requirements

The test matrix must cover cap math at validator-set sizes that match Aetra's target rollout: 100, 150, 250, and 300 active validators. Cap transition tests must cover validators moving above and back below cap so events, warnings, reward multipliers, and query responses stay correct.

Delegation, redelegation, unbonding, and slashing tests must cover over-cap validators because cap policy must not corrupt staking shares or prevent raw stake from being slashed. Commission policy tests must cover below-floor, above-max, and daily-jump rejection. Governance tests must cover valid parameter updates and rejected unsafe updates. Export/import and deterministic concentration snapshot tests are mandatory launch gates.

### Test Implementation Contract

The test gate is `BuildAetraStakingPolicyTestSpecReport` in `app/params/aetra_staking_policy_spec.go`.

Required catalog properties:

- `DefaultAetraStakingPolicyTestSpecEvidence` must include every required test scenario;
- missing required test scenarios must fail validation;
- duplicate or unexpected test scenario names must fail validation;
- wrong module identity must fail validation.
