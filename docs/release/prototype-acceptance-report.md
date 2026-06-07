> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Prototype Acceptance Report

Date: 2026-06-04

Scope: final local prototype acceptance pass for current source commit `216b50ca6a43` on branch `build/one-command-aetrad`.

## Release Decision

Status: **NOT READY for a publishable prototype release**.

The 3-validator prototype lifecycle is functional: build, health, CLI/gRPC/REST queries, bank tx, fee-denom rejection, tokenfactory, DEX, staking, restart, tests, vet, buf lint, determinism gate, and dry-run package creation all completed successfully.

The release must not be published until the MUST FIX blockers below are resolved and the checks are rerun from a clean tree.

## MUST FIX Blockers

1. Clean the working tree before release packaging.
   - Clean package preflight without `-AllowDirty` currently fails.
   - Current `git status --short --untracked-files=all` includes untracked `STEP.md`, `STEP_V2.md`, `STEP_V3.md`, and this report before it is committed.
   - Project policy says `STEP*.md` must not be committed as release evidence.

2. Resolve full audit gate triage items.
   - `prototype-audit.ps1 -Profile Full` passed the runnable gate but reported `triage_required` for `govulncheck` and `go-mod-verify`.
   - Release criteria require no untriaged security or artifact-integrity findings.

3. Fix or explicitly de-scope 5-validator acceptance.
   - Default-port 5-validator Full run failed because `26657`/`6060` were already in use in the local environment.
   - Shifted-port 5-validator Full run also failed because `tests/e2e/prototype_acceptance.ps1` calls `scripts/localnet/health.ps1` without forwarding the configured base RPC/REST/gRPC/pprof ports, so the health check waits on default RPC `26657`.
   - The acceptance request requires 3/5 validator observations, so this remains a release blocker unless the owner explicitly scopes the prototype release to 3 validators only.

4. Rebuild the artifact and checksums after blockers are fixed.
   - The checksum below proves the dry-run package can be created, not that the clean release preflight was satisfied.

## Architecture

The prototype keeps release, audit, determinism, localnet, and observability concerns outside consensus code. The acceptance pass did not introduce release-time changes to keepers, message servers, ante decorators, protocol parameters, or query servers.

Observed module boundaries are acceptable for the current prototype:

- `fees`: ante fee policy is exercised by e2e wrong-denom rejection.
- `tokenfactory`: lifecycle is exercised in the 3-validator Full acceptance run.
- `dex`: create/swap/slippage flows are exercised in the 3-validator Full acceptance run.
- `staking`/`bank`: local validator and bank lifecycle are exercised in CLI/e2e flows.
- `scripts`: release, audit, localnet, health, and acceptance scripts remain non-consensus tooling.

## Module Breakdown

| Area | Acceptance evidence | Status |
| --- | --- | --- |
| Build/version | `prototype_acceptance.ps1 -Profile Full` and release dry-run package | pass, dirty |
| CLI/gRPC/REST queries | Base query matrix in Full acceptance | pass |
| Bank tx | `bank send updated node1 balance to 400001000naet` | pass |
| Fees ante policy | `wrong fee denom rejected` | pass |
| Tokenfactory | factory denom created and minted to node0 | pass |
| DEX | slippage guard rejected, swap increased factory balance `90000000` to `90098715` | pass |
| Staking | delegation increased voting power `300` to `305` | pass |
| Restart/health | restart preserved chain progress `15->16` | pass |
| 5 validators | default and shifted-port attempts failed | MUST FIX |
| Security gate | Full audit runnable checks complete | triage_required |
| Determinism | blocking findings `0` | pass |

## Implementation

No production code was changed during this acceptance pass. This report records the final gate result and the exact blockers required before publishing.

Evidence artifacts were written only under ignored `.work` and `.localnet*` paths.

## Command Transcript

Commands were run on Windows PowerShell with local Go toolchain `.work/tools/go1.25.11/go/bin/go.exe`.

| Command | Result | Notes |
| --- | --- | --- |
| `git status --short --untracked-files=all; git rev-parse --short=12 HEAD` | dirty, `216b50ca6a43` | release tree is not clean |
| `go test ./...` | pass | included current dirty working tree tests |
| `go vet ./...` | pass | no vet failures |
| `buf lint` | pass | proto lint clean |
| `scripts/security/determinism-gate.ps1 -Json` | pass | `findings=76`, `blocking=0`, `strict_blocking=0` on current tree |
| `go test -run '^$' -bench 'Benchmark(EmptyBlock\|Dex)' -benchtime=1x ./app ./x/dex/keeper` | pass | benchmark results below |
| `tests/scripts/prototype_release_package_test.ps1` | pass | package script smoke passed |
| `scripts/security/prototype-audit.ps1 -Profile Full` | triage_required | rerun on current tree; see security section |
| `tests/e2e/prototype_acceptance.ps1 -Profile Full` | pass | 3 validators, height `17`; run before subsequent docs/lifecycle commits landed, while their code changes were present in the dirty tree |
| `tests/e2e/prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5 -SkipBuild` | fail | default port collision on `26657`/`6060` |
| `tests/e2e/prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5 -BaseP2PPort 27656 -BaseRPCPort 27657 -BaseRESTPort 1417 -BaseGRPCPort 9190 -BasePprofPort 6160 -SkipBuild` | fail | health helper still waited on default `26657` |
| `scripts/release/prototype-package.ps1 -Version prototype-final-pass-216b50c -TargetOS windows -TargetArch amd64 -OutputRoot .work/release-final-pass-latest -AllowDirty` | pass | dry-run artifact created for `216b50ca6a43` |
| `scripts/release/prototype-package.ps1 -Version prototype-clean-check -TargetOS windows -TargetArch amd64 -OutputRoot .work/release-clean-check -SkipBuild` | fail as expected | clean-tree preflight rejects untracked `STEP*.md` and this report |

## Tests

Baseline test results:

- `go test ./...`: pass on current tree
- `go vet ./...`: pass on current tree
- `buf lint`: pass on current tree
- Release package script test: pass
- 3-validator Full e2e: pass
- 5-validator Full e2e: fail, blocked as above

3-validator Full e2e output highlights:

This e2e transcript shows binary commit `27d88b54f069` because it ran before later docs/lifecycle commits landed. Current quick gates, full audit, determinism gate, and dry-run package were rerun at `216b50ca6a43`.

```text
Binary: C:\Users\Ryzen\Desktop\L1\build\aetrad.exe
Version: dev-27d88b54f069
Commit: 27d88b54f069
Dirty: true
localnet healthy at height 4
base CLI/gRPC/REST queries passed
bank send updated node1 balance to 400001000naet
wrong fee denom rejected
factory denom ... minted to node0
DEX slippage guard rejected excessive min_amount_out
DEX swap increased factory balance from 90000000 to 90098715
delegation increased voting power from 300 to 305
restart preserved chain progress: 15->16
prototype acceptance Full passed at height 17
```

## Security Review

Final Cosmos security pass used the `cosmos-vulnerability-scanner` workflow focus areas: denom validation, signer/authority checks, missing balance checks, ABCI panic paths, nondeterminism, rounding, and unbounded loops.

Full audit summary:

| Check | Status |
| --- | --- |
| go-vet | pass |
| go-test | pass |
| buf-lint | pass |
| gitleaks-staged | pass |
| gitleaks-history | pass |
| gosec | pass |
| govulncheck | triage_required |
| go-mod-verify | triage_required |
| deterministic-execution-gate | pass |

Determinism gate:

- Output: `.work/security/determinism-gate-20260604-235307`
- Findings: `76`
- Blocking findings: `0`
- Strict blocking findings: `0`

Secret scan:

- Staged secret scan: pass
- History secret scan: pass

Artifact hygiene:

- Dry-run package excludes runtime directories by release script policy.
- Current artifact is not publishable because it was built with `-AllowDirty`.

## Performance Notes

Benchmark pass:

```text
BenchmarkEmptyBlockFinalizeCommit-12    1   1937500 ns/op
BenchmarkDexCreatePoolsAndSwap-12       1    411700 ns/op
```

3-validator acceptance:

- Full localnet reached height `17`.
- Restart preserved chain progress from height `15` to `16`.
- Bank, fees, tokenfactory, DEX, staking, query, and health flows completed within the test timeout.

5-validator acceptance:

- Not measured successfully.
- Default-port run was blocked by local port collisions.
- Shifted-port run exposed a test harness port-forwarding bug.

## Checksums

Dry-run package:

```text
82b8ad12e73b4da4729d14aec102184df8803a7b21c438af5678d5c98110a9ea  orbitalis-prototype-final-pass-216b50c-windows-amd64.zip
```

Dry-run paths:

- Package: `.work/release-final-pass-latest/prototype-final-pass-216b50c/orbitalis-prototype-final-pass-216b50c-windows-amd64.zip`
- Checksum: `.work/release-final-pass-latest/prototype-final-pass-216b50c/orbitalis-prototype-final-pass-216b50c-windows-amd64.zip.sha256`
- Manifest commit: `216b50ca6a43`
- Clean preflight status: failed before report commit because untracked files remain

## Limitations

- This is a prototype acceptance pass, not a mainnet readiness assessment.
- The clean release artifact was not produced because the working tree is dirty.
- 5-validator Full acceptance is not yet usable with shifted ports.
- `govulncheck` and `go-mod-verify` findings need owner triage before publication.
- DEX coverage validates prototype create/swap/slippage behavior, not market-depth, high-load, or adversarial liquidity scenarios.
- Query/list scalability remains prototype-bounded; future modules must preserve pagination/default-limit discipline.

## Refactor Suggestions

- Forward base port parameters from `prototype_acceptance.ps1` into `scripts/localnet/health.ps1`, including restart health checks.
- Add a preflight localnet port availability check for all validators before starting processes.
- Split audit gate output into `pass`, `triage_required`, and `blocker` summaries in release notes so owner decisions are visible.
- Promote the untracked tokenfactory smoke script only after deciding whether it belongs in fast PR smoke or manual/nightly e2e.

## Next Iteration Plan

1. Decide ownership of the dirty working tree changes and remove or commit them according to release policy.
2. Fix 5-validator shifted-port acceptance or explicitly de-scope 5-validator release readiness.
3. Triage `govulncheck` and `go-mod-verify`; document advisory decisions or update dependencies/cache handling.
4. Rerun full checks from a clean tree.
5. Rebuild the release artifact without `-AllowDirty`.
6. Record the clean checksum and publish only if all MUST FIX blockers are closed.
