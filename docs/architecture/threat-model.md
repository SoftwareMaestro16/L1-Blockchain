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
