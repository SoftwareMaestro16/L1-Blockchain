# x/aetra-validator-score Module Specification

Purpose: public accountability without subjective consensus control.

This module publishes deterministic validator accountability metrics for wallets, explorers, delegators, operators, governance dashboards, and reward-policy modules. It must not become a subjective censorship mechanism and must not override consensus behavior unless a future governance-approved design explicitly adds that behavior with objective inputs and tests.

## 24. Module Specification: `x/aetra-validator-score`

Goal: public accountability without subjective consensus control.

The module should make validator behavior visible without creating a centralized reputation oracle. Scores are allowed to inform delegators and may affect rewards only when based on objective chain data.

## 24.1 Responsibilities

The module must:

- track validator uptime;
- track missed block windows;
- track jail history;
- track slashing history;
- track commission behavior;
- track self-bond ratio;
- track governance participation;
- track concentration status;
- produce public score;
- expose explorer-friendly queries.

Score must not become a subjective censorship mechanism. It should be informational first and reward-affecting only when based on objective chain data.

### Implementation Contract

The implementation gate is `app/params/aetra_validator_score_spec.go`.

Required catalog properties:

- `AetraValidatorScoreModuleName` must be `x/aetra-validator-score`;
- `DefaultAetraValidatorScoreSpecEvidence` must cover public accountability without subjective consensus control;
- `BuildAetraValidatorScoreSpecReport` must reject missing purpose evidence;
- `DefaultAetraValidatorScoreResponsibilitiesEvidence` must cover all ten responsibilities from section 24.1;
- `BuildAetraValidatorScoreResponsibilitiesReport` must reject missing responsibilities;
- `DefaultAetraValidatorScoreSubjectiveControlEvidence` must require no subjective censorship mechanism, informational-first behavior, objective-only reward effects, consensus override disabled by default, and deterministic objective inputs;
- `BuildAetraValidatorScoreSubjectiveControlReport` must reject missing subjective-control guards.

### Module Requirements

The concrete `x/aetra-validator-score` module must:

- keep `ConsensusOverrideEnabled` disabled by default;
- set `ConsensusOverrideAllowed` to false unless params explicitly enable it;
- calculate scores from deterministic chain-derived inputs;
- keep reward effects behind `ObjectiveRewardModifierEnabled`;
- expose `QueryValidatorScore`, `QueryPublicValidatorMetrics`, and `QueryAllValidatorScores`;
- keep genesis export/import deterministic and canonically sorted;
- reject malformed metric inputs and invalid params.

## 24.2 State

Suggested state:

```text
Params:
  UptimeWindow
  UptimeWeightBps
  SlashHistoryWeightBps
  GovernanceWeightBps
  SelfBondWeightBps
  ConcentrationWeightBps
  MinScore
  MaxScore
  RewardModifierEnabled
  MaxRewardPenaltyBps

ValidatorScore:
  OperatorAddress
  Score
  UptimeScore
  SlashScore
  GovernanceScore
  SelfBondScore
  ConcentrationScore
  MissedBlocks
  SignedBlocks
  JailCount
  SlashCount
  LastUpdatedHeight
```

### State Implementation Contract

Required catalog properties:

- `AetraValidatorScoreStateSpecEvidence` must list all required `Params` and `ValidatorScore` fields;
- `DefaultAetraValidatorScoreStateSpecEvidence` must include all twenty-two state fields from section 24.2;
- `BuildAetraValidatorScoreStateSpecReport` must reject missing, duplicate, and unexpected state fields;
- `ValidateAetraValidatorScoreStateSpec` must reject wrong module identity and incomplete state catalogs;
- state field coverage must be tested before the module is considered production-ready.

### Current Implementation Mapping

Current code uses more explicit field names while preserving the 24.2 intent:

- `UptimeWindow` maps to `UptimeWindowBlocks`;
- `UptimeWeightBps`, `SlashHistoryWeightBps`, `GovernanceWeightBps`, `SelfBondWeightBps`, and `ConcentrationWeightBps` map to `ScoreWeights`;
- `MinScore` and `MaxScore` map to the enforced `0..10000` bps score domain;
- `RewardModifierEnabled` maps to `ObjectiveRewardModifierEnabled`;
- `MaxRewardPenaltyBps` maps to `MinRewardMultiplierBps`;
- `Score` maps to `OverallScoreBps`;
- `UptimeScore`, `SlashScore`, `GovernanceScore`, `SelfBondScore`, and `ConcentrationScore` map to the component score fields;
- `JailCount` maps to `JailEvents`;
- `SlashCount` maps to `SlashEventCount`;
- `LastUpdatedHeight` is represented by deterministic score update context and should become an explicit height field if score updates are moved from epoch-only accounting to height-indexed accounting.
