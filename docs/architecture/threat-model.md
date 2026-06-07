# Threat Model

This document defines section 29 of the Aetra architecture backlog. Threat model entries must be concrete enough to produce controls, simulations, tests, observability, and governance requirements.

The implementation gate is `app/params/threat_model_spec.go`.

## 29.1 Validator Cartel

Threat:

- several validators coordinate censorship or governance capture.

Controls:

- 100-300 validator target;
- validator power cap;
- top-N monitoring;
- commission floor;
- identity transparency;
- governance participation metrics;
- delegation warnings.

Tests/simulations:

- top-10 concentration simulation;
- split-identity validator simulation;
- delegation overflow simulation;
- governance capture threshold analysis.

Security requirements:

- Cartel detection must use objective chain data and public validator metadata.
- Controls must prefer economic signals, warnings, caps, and metrics over centralized validator admission.
- Identity transparency must not become mandatory KYC at consensus level.
- Concentration warnings must not halt staking or break delegation; they should guide delegators, explorers, governance, and risk dashboards.
- Governance capture threshold analysis must model proposal quorum, voting period, bonded ratio, validator voting power, delegation concentration, and top-N concentration.
- Split-identity simulation must assume one operator can run multiple validators and must test whether power caps alone are insufficient.
- Delegation overflow simulation must prove over-cap stake cannot increase effective voting power or normal reward weight after the configured cap is active.

Acceptance gate:

- `BuildAetraValidatorCartelThreatReport` must pass.
- Missing validator cartel threat definition must fail validation.
- Missing any required control must fail validation.
- Missing top-10 concentration, split-identity, delegation overflow, or governance capture simulation must fail validation.
- Mandatory KYC, staking halt on warnings, or subjective off-chain enforcement must fail architecture review.

## 29.2 Stake Centralization Through Rewards

Threat:

- large validators grow faster because delegators chase apparent safety/APR.

Controls:

- overflow rewards reduced;
- over-cap warnings;
- commission floor;
- concentration metrics;
- reward multiplier based on cap.

Tests:

- rewards for over-cap validator lower than normal;
- delegator APR estimate reflects overflow penalty;
- cap changes do not create accounting corruption.

Security requirements:

- Reward accounting must not let overflow stake receive the same effective reward weight as in-cap stake.
- APR estimates shown to delegators must include overflow penalty and commission impact.
- Cap changes must be epoch-boundary or otherwise deterministic, and must preserve delegation shares, historical rewards, slashing exposure, and export/import state.
- Over-cap warnings must be queryable and indexer-friendly so wallets and explorers can steer delegators away from concentration.
- Commission floor must avoid race-to-zero commission campaigns that pull stake into a few already trusted validators.

Acceptance gate:

- `BuildAetraStakeCentralizationRewardsThreatReport` must pass.
- Missing stake centralization through rewards threat definition must fail validation.
- Missing overflow reward reduction, over-cap warning, commission floor, concentration metric, or cap-based reward multiplier control must fail validation.
- Missing reward, APR estimate, or cap-change accounting test must fail validation.

## 29.3 Downtime And Weak Operators

Threat:

- too many low-quality validators reduce liveness.

Controls:

- minimum self-bond;
- validator score;
- downtime slashing;
- jail;
- public metrics;
- gradual validator set growth.

Tests:

- liveness with < 1/3 voting power offline;
- halt behavior with > 1/3 offline documented;
- recovery after validators return;
- downtime penalties applied.

Security requirements:

- Minimum self-bond must make operators carry direct risk without making validator entry only accessible to large funds.
- Validator score must use objective chain data such as missed blocks, jail history, slash history, and governance participation.
- Downtime slashing and jail must be deterministic, bounded, and covered by delegator impact tests.
- Public metrics must expose uptime, missed blocks, jail/unjail events, slash history, and active validator count.
- Validator set growth must be gradual; Aetra should not expand active validators faster than operator quality, monitoring, and liveness tests support.
- Halt behavior with more than one third voting power offline must be documented clearly as a BFT liveness boundary, not treated as a software bug.

Acceptance gate:

- `BuildAetraDowntimeWeakOperatorsThreatReport` must pass.
- Missing downtime and weak operators threat definition must fail validation.
- Missing self-bond, validator score, downtime slashing, jail, public metrics, or gradual validator set growth control must fail validation.
- Missing liveness, halt-boundary, recovery, or downtime penalty test must fail validation.
