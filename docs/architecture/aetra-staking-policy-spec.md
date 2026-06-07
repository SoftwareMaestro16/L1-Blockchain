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
