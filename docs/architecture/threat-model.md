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

## 29.4 Governance Attack

Threat:

- malicious proposal changes economics, slashing, cap, or VM params dangerously.

Controls:

- param bounds;
- delayed activation;
- emergency review window for critical params;
- explicit authority checks;
- event monitoring.

Tests:

- malicious param proposal rejected;
- out-of-range values rejected;
- authority spoofing rejected;
- delayed activation works.

Security requirements:

- Economics, slashing, validator power cap, validator set growth, block gas/bytes, and VM/CosmWasm/AVM params must define explicit min/max bounds.
- Critical params must activate after a deterministic delay or epoch boundary so validators, dashboards, and delegators can observe the change before it becomes active.
- Emergency review window for critical params must be documented and observable; it must not depend on subjective off-chain consensus in the state transition.
- Governance and params messages must verify explicit authority and reject spoofed signers.
- Param update execution must emit stable events with old value, new value, activation height or epoch, authority, and criticality.
- Event monitoring must alert on critical economics, slashing, cap, or VM parameter changes before activation.

Acceptance gate:

- `BuildAetraGovernanceAttackThreatReport` must pass.
- Missing governance attack threat definition must fail validation.
- Missing param bounds, delayed activation, emergency review window, explicit authority checks, or event monitoring control must fail validation.
- Missing malicious proposal rejection, out-of-range rejection, authority spoofing rejection, or delayed activation test must fail validation.

## 29.5 Contract Attack

Threat:

- malicious CosmWasm contract consumes gas/storage, exploits permissions, or causes state bloat.

Controls:

- gas limits;
- storage pricing;
- upload policy;
- migration controls;
- contract size limit;
- malicious contract test suite.

Tests:

- gas exhaustion;
- storage abuse;
- unauthorized migration;
- invalid instantiate;
- export/import with malicious-but-contained contract state.

Security requirements:

- Gas limits must bound instantiate, execute, query, migrate, reply, and submessage paths.
- Storage pricing must make large writes economically bounded and must prevent free state bloat.
- Upload policy must be governance-gated or permissioned before security review and must require explicit fees/deposits before permissionless upload.
- Migration controls must reject unauthorized admin changes, disabled migration paths, and malformed migration targets.
- Contract size limit must reject oversized wasm code before it can enter contract state.
- Malicious contract test suite must include infinite loop, large storage write, failed execution rollback, unauthorized migration, invalid instantiate, query no-mutation, and export/import with contained malicious state.
- Export/import with malicious-but-contained contract state must preserve app hash and must not leak contract access to reserved module funds.

Acceptance gate:

- `BuildAetraContractAttackThreatReport` must pass.
- Missing contract attack threat definition must fail validation.
- Missing gas, storage pricing, upload, migration, code size, or malicious contract test-suite control must fail validation.
- Missing gas exhaustion, storage abuse, unauthorized migration, invalid instantiate, or malicious-contained export/import test must fail validation.
