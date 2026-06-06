> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Prototype Non-Goals And Limitations

Version scope: current Aetra working prototype prerelease line. Update this file for every prototype release decision that changes scope, risk acceptance, or blocker status.

This document prevents scope creep and avoids implying mainnet readiness. It does not override the security gate: any untriaged Critical or High security finding remains a release blocker, not an accepted limitation.

## Classification Rule

| Class | Meaning | Release Decision |
| --- | --- | --- |
| Non-goal | Explicitly outside the prototype release scope. | Do not build or imply support in prototype docs. |
| Accepted limitation | Known prototype weakness with bounded local/test impact and no untriaged Critical/High risk. | May ship in a prototype package when documented and linked from release notes. |
| Blocker | Correctness, security, determinism, build, artifact, or release-evidence issue that invalidates the release decision. | Must be fixed, regression-tested, or explicitly owner-triaged before publishing. |

If a limitation is later found to create fund loss, consensus halt, unauthorized mint/burn/admin, wrong-fee bypass, secret exposure, or other Critical/High risk, reclassify it as a blocker immediately.

## Non-Goals

| Non-goal | Boundary |
| --- | --- |
| Mainnet launch | No production validator onboarding, no genesis ceremony, no public network operations, no validator SLOs. |
| IBC or external bridge | No bridge trust model, relayer operations, packet verification, or cross-chain asset safety claims. |
| Production governance economics | No production token economics, validator incentives, parameter-change process, treasury policy, or emergency governance runbook. |
| Exchange-grade DEX | DEX is a prototype constant-product AMM. No routing, oracle integration, MEV policy, market-maker guarantees, price-quality guarantees, or production liquidity risk model. |
| Public faucet | Local funding uses genesis-funded local accounts and normal `bank send`; there is no production faucet, privileged mint endpoint, or public funding SLA. |
| Full external audit | Internal Cosmos-specific gates and scanner-guided checks exist, but they are not a complete independent external audit. |
| Explorer/API SLA | CLI/gRPC/REST are observable for prototype flows, but no public explorer uptime, indexing completeness, latency, retention, or support SLA is promised. |

## Accepted Prototype Limitations

| Limitation | Current mitigation | Exit condition |
| --- | --- | --- |
| Local-only key material | Localnet uses ignored node homes and `--keyring-backend test`; scripts and diagnostics avoid printing mnemonics/private keys. | Public testnet/mainnet key-management plan and runbook. |
| Localnet minimum gas price may be `0naet` | Docs and e2e examples still use `--fees 1000000naet`; `x/fees` ante enforces allowed fee denom. | Public network fee policy and validator min-gas configuration. |
| Tokenfactory/DEX are prototype modules | Unit/e2e/adversarial tests cover signer, denom, balance, DEX accounting, and negative flows. | External audit, expanded invariants, and production economics/risk model. |
| Query/list endpoints are bounded but not public-load proven | List endpoints use bounded pagination and smoke/load scripts record minimal local evidence. | High-cardinality API and explorer load tests with documented limits. |
| Minimal load profile is local baseline only | `load-profile.ps1` records block progress, tx latency, and failures under small local mixed load. | Dedicated performance plan for public testnet/mainnet capacity. |
| Vote extension behavior is prototype-only | Security docs flag dummy/test-oriented vote extension behavior as not production-ready. | Replace, disable, or formally specify vote extensions before any public validator network. |
| Release package may include triage-required dependency/tool findings only when documented | Audit gate records triage-required findings; release notes must link evidence and owner decision. | Clean full audit gate or documented owner acceptance for each reachable/non-reachable finding. |

## Blockers

These are never accepted limitations:

- untriaged Critical or High Cosmos security findings,
- wrong fee denom accepted or fee policy bypass,
- unauthorized tokenfactory mint, burn, or admin transfer,
- DEX invariant failure, including reserve, module-balance, or LP-supply mismatch,
- nondeterminism, nondeterministic consensus state write, or AppHash divergence,
- ABCI/query/genesis panic reachable from malformed public input,
- unbounded tx/list path, including any unbounded tx path or public list/query path without a cap,
- secrets or secret material in tracked files, release packages, diagnostics, or logs,
- build, genesis validation, localnet startup, or acceptance smoke failure,
- artifact produced from an unapproved dirty tree.

Current release-decision blockers, if any, are recorded in [Prototype Acceptance Report](prototype-acceptance-report.md). A publishable prototype package must either show no blockers or link owner-approved blocker de-scoping with explicit release status.
