# Security Audit Pack

Phase 14 defines the minimum Aetra security audit pack before public
testnet. Public testnet cannot proceed with untriaged high/critical
fund-safety, consensus-safety, or secret-leak findings.

## Required Documents

- [Manual Security Audit Checklist](manual-audit-checklist.md)
- [Cosmos Security Audit Checklist](cosmos-security-checklist.md)
- [Prototype Security And Determinism Audit Gate](prototype-audit-gate.md)
- [Security Triage Policy](security-triage-policy.md)
- [Determinism Gate](determinism-gate.md)
- [Security Risks And Controls](security-risks-controls.md)
- [Aetra Slashing System](slashing-system.md)

## Required Workflows

The `.github/workflows/security.yml` Security Gate must keep:

- `govulncheck`
- `gosec high severity`
- `gitleaks secrets`
- `dependency review`
- `CodeQL`

The release workflow must keep staged/history secret scanning and the local
prototype audit gate. New high/critical findings from any workflow are
untriaged by default until a finding record names severity, reachability,
decision, owner, target, and evidence.

## Cosmos-Specific Checks

Every security review must cover:

- non-determinism
- incorrect signers
- ABCI panics
- unsafe rounding
- unbounded iteration
- malformed genesis
- replay paths
- invalid authority paths

Regression evidence must map to app tests, keeper tests, adversarial tests,
determinism-gate output, or a documented blocker.

## Contract-Specific Checks

Every contract-standard or async execution review must cover:

- wallet replay
- wrong wallet_id
- extension takeover
- token supply divergence
- NFT unauthorized transfer
- SBT transfer bypass
- async queue DoS
- bounce/refund double-spend
- metadata spoofing
- admin takeover

Regression evidence currently lives in:

- `x/aetravm/standards/aw/*_test.go`
- `x/aetravm/standards/aft/*_test.go`
- `x/aetravm/standards/anft/*_test.go`
- `x/aetravm/async/*_test.go`
- `tests/scripts/*standard_doc_test.ps1`
- `tests/scripts/async_contract_execution_doc_test.ps1`

## Triage Rule

A high/critical finding is triaged only when it has one of:

- merged fix with regression test or scanner evidence;
- explicit downgrade with technical rationale, owner, and review date;
- accepted risk record with owner, issue link, mitigation, and target date.

Secret leaks are release blockers unless the leaked material is proven local,
ignored, rotated when needed, and absent from tracked history and release
artifacts.

## Public Testnet Acceptance

Public testnet readiness requires:

- no untriaged high/critical fund-safety findings;
- no untriaged high/critical consensus-safety findings;
- no untriaged high/critical secret-leak findings;
- clean or owner-triaged `govulncheck`, `gosec`, CodeQL, gitleaks, and
  Dependency Review results;
- completed manual Cosmos and contract-standard checklist records.
