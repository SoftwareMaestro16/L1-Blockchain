# Aetra Prototype Release Package

This document defines the release package for the Aetra working L1 prototype.

The package is a prerelease/testnet artifact. It is not a mainnet-ready release and must not be marketed as production validator software.

## Contents

Each package produced by `scripts\release\prototype-package.ps1` contains:

- `bin/aetrad` or `bin/aetrad.exe`
- `release-manifest.json`
- `QUICKSTART.md`
- `RELEASE-NOTES.md`
- `SHA256SUMS.txt`
- copied operator docs:
  - `README.md`
  - `docs/operator-commands.md`
  - `docs/operator-troubleshooting.md`
  - `docs/prototype-acceptance-suite.md`
  - `docs/security/prototype-audit-gate.md`
  - `docs/query-surface.md`
  - `docs/observability.md`
  - `docs/release/prototype-package.md`
  - `docs/release/prototype-limitations.md`
- optional evidence under `evidence/`

The script also creates:

- `orbitalis-<version>-<os>-<arch>.zip`
- `orbitalis-<version>-<os>-<arch>.zip.sha256`

Artifacts must not contain `.work`, `.localnet`, keyrings, mnemonics, validator private keys, node keys, environment files, Redis/PostgreSQL URLs, or local diagnostic bundles unless explicitly copied as sanitized evidence.

## Local Build

Build a Windows package from a clean current checkout:

```powershell
git status --short
go test -p=1 ./...
go vet -p=1 ./...
buf lint
.\scripts\security\prototype-audit.ps1 -Profile Fast
.\tests\e2e\prototype_acceptance.ps1 -Profile Smoke
.\scripts\release\prototype-package.ps1 -Version prototype-local -TargetOS windows -TargetArch amd64
```

`prototype-package.ps1` refuses a dirty git tree by default. `-AllowDirty` is only for local dry-runs and package-script tests; publishable artifacts must come from a clean checkout.

Reuse an existing binary:

```powershell
.\scripts\release\prototype-package.ps1 -Version prototype-local -SkipBuild -Binary build\aetrad.exe -AllowDirty
```

Attach sanitized evidence:

```powershell
.\scripts\release\prototype-package.ps1 -Version prototype-local -EvidencePath .work\security\prototype-audit-YYYYMMDD-HHMMSS\summary.md
```

Run package-owned checks from the script when preparing a manual prerelease:

```powershell
.\scripts\release\prototype-package.ps1 -Version prototype-v0.1.0-rc.1 -RunChecks -RunAcceptanceSmoke
```

## GitHub Workflow

Manual workflow: `.github/workflows/prototype-release.yml`.

The workflow:

1. Checks out a clean tree.
2. Sets up Go from `go.mod`.
3. Installs release tooling.
4. Runs unit tests, vet, `buf lint`, gitleaks, and the prototype audit gate.
5. Optionally runs `tests\e2e\prototype_acceptance.ps1 -Profile Smoke`.
6. Builds matrix artifacts for Windows amd64, Linux amd64, and Linux arm64.
7. Generates checksums.
8. Uploads artifacts.
9. Optionally creates a GitHub prerelease.

## Required Evidence

A prototype prerelease should attach or link:

- workflow run URL,
- `go test -p=1 ./...` output,
- `go vet -p=1 ./...` output,
- `buf lint` output,
- `scripts\security\prototype-audit.ps1 -Profile Full` summary,
- `tests\e2e\prototype_acceptance.ps1 -Profile Smoke` transcript,
- checksums for each binary/archive.

Known `triage_required` findings from the audit gate are acceptable only when they match [prototype-audit-gate.md](../security/prototype-audit-gate.md) and have an owner. Untriaged Critical/High findings are blockers, not accepted limitations; see [prototype-limitations.md](prototype-limitations.md).

`release-manifest.json` records `version`, `commit`, `dirty`, `target_os`, `target_arch`, required checks, checks executed by the script, copied evidence names, and excluded runtime/private material. `RELEASE-NOTES.md` repeats the commit, dirty flag, test evidence expectations, known limitations, non-goals, and blocker rule so an operator can verify the artifact without reading CI logs first.

## Known Limitations

The authoritative scope boundary is [prototype-limitations.md](prototype-limitations.md). Each prototype release must keep non-goals, accepted limitations, and blockers separate.

- Non-goals include mainnet launch, IBC/external bridge, production governance economics, exchange-grade DEX behavior, public faucet, full external audit, and explorer/API SLA.
- Accepted limitations include local-only key material, local prototype min-gas behavior, bounded but not public-load-proven list queries, local-only load profiling, and prototype-only vote extension behavior.
- Blockers include untriaged Critical/High Cosmos security findings, accepted wrong fee denoms, unauthorized mint/burn/admin actions, DEX invariant failures, nondeterministic AppHash divergence, public malformed-input panics, unbounded tx/list paths, secrets in tracked/release artifacts, and failed build/genesis/localnet/acceptance gates.

## Tagging

Use prerelease-style versions until the project has a mainnet readiness gate:

```text
prototype-v0.1.0
prototype-v0.1.0-rc.1
testnet-v0.1.0
```

Before tagging, require:

```powershell
git status --short
.\scripts\security\prototype-audit.ps1 -Profile Full
.\tests\e2e\prototype_acceptance.ps1 -Profile Smoke
```
