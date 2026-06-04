# Prototype Security And Determinism Audit Gate

This gate blocks an Orbitalis prototype release when consensus-critical or fund-safety risks are untriaged.

Scope: custom modules `x/tokenfactory`, `x/dex`, `x/fees`; app wiring and ABCI paths; genesis/bootstrap; fees ante policy; DEX math/accounting; tokenfactory admin rights; localnet/prototype acceptance scripts.

The consensus node does not require Redis, PostgreSQL, or any external database. Do not commit database URLs, credentials, mnemonics, validator keys, localnet keyrings, or environment dumps. Off-chain indexers can use databases later through environment variables or a secret manager, outside this gate.

## Release Rule

Prototype release is blocked when any `Critical` or `High` finding lacks one of:

- merged fix and regression test,
- documented mitigation accepted for the prototype scope,
- explicit downgrade with rationale and owner.

`Medium` findings require an owner and target milestone. `Low` findings can be batched but must remain visible.

## Severity

| Severity | Blocks Prototype | Examples |
| --- | --- | --- |
| Critical | Yes | nondeterministic state transition, ABCI panic from malformed tx/query/genesis, unauthorized mint/burn/admin, bank supply corruption |
| High | Yes | missing denom validation, missing bank balance/accounting check, DEX reserve/LP supply desync, wrong fee denom bypass, unbounded state scan in tx path |
| Medium | Owner required | unpaginated prototype query with bounded cap, dependency advisory without reachable symbol, localnet-only insecure default |
| Low | Track | docs inconsistency, noisy generated-code scanner result with narrow suppression |

## Local Command

Fast PR gate:

```powershell
.\scripts\security\prototype-audit.ps1 -Profile Fast
```

Full release gate:

```powershell
.\scripts\security\prototype-audit.ps1 -Profile Full
.\tests\e2e\prototype_acceptance.ps1 -Profile Smoke
```

Nightly/manual scale gate:

```powershell
.\scripts\security\prototype-audit.ps1 -Profile Nightly
```

Outputs are written under ignored `.work\security\prototype-audit-*`:

- `summary.md`: pass/fail/triage table.
- `results.json`: machine-readable result list.
- `determinism-gate/summary.md`: deterministic execution findings classified by AppHash/consensus impact.
- `gosec.json`: source findings with generated protobuf excluded.
- tool logs for `go vet`, `go test`, `buf lint`, `gitleaks`, `govulncheck`, and `go mod verify`.

Use `-Strict` to make triage-required tool findings return non-zero. Default mode permits known, documented triage so developers can reproduce the gate locally.

## Required Checks

| Check | Command | Gate |
| --- | --- | --- |
| Go tests | `go test -p=1 ./...` | Fail on error |
| Vet | `go vet -p=1 ./...` | Fail on error |
| Proto lint | `buf lint` | Fail on error |
| Staged secret scan | `gitleaks protect --staged --redact` | Fail on leak |
| History secret scan | `gitleaks detect --source . --redact --log-opts --all` | Full gate; fail on leak |
| Source static scan | `gosec -exclude-generated` over `go list` package dirs | Triage all source findings |
| Dependency advisories | `govulncheck -scan=package ./...` | Triage all package findings |
| Module integrity | `go mod verify` | Full gate; fail unless local cache mutation is documented |
| Deterministic execution | `scripts\security\determinism-gate.ps1` | Fail on untriaged High/Critical finding |
| Acceptance smoke | `tests\e2e\prototype_acceptance.ps1 -Profile Smoke` | Release gate |
| 5-validator full profile | `prototype-audit.ps1 -Profile Nightly` | Manual/nightly |

Generated `.pb.go`/`.pb.gw.go` scanner noise must be excluded through tool flags or narrow config, not by hiding source `.proto` files. If a generated-code finding maps to a source proto design problem, track the proto issue.

## Cosmos Manual Checklist

Each release candidate needs a reviewer to mark every item `PASS`, `FINDING`, or `N/A`.

| Area | Review |
| --- | --- |
| Nondeterminism | No consensus path uses map iteration order, wall clock, `math/rand`, goroutines, select races, floating point, pointer addresses, or platform-dependent serialization. |
| ABCI panic | BeginBlocker/EndBlocker, InitGenesis, ExportGenesis, ante decorators, and msg servers return errors for malformed input instead of panicking. Panics are limited to impossible app wiring/module registration failures. |
| Authorization | `Msg` signers match msg server sender/admin checks. Tokenfactory mint/burn/change-admin requires current admin. Fees params update requires governance authority. |
| Denom validation | Native `norb`/`ORB` cannot be spoofed by tokenfactory or LP denoms. All user-provided denoms use SDK validation before bank movement. |
| Balance checks | Bank sends/mints/burns propagate errors. DEX module balances equal recorded reserves. LP supply equals pool `total_shares`. |
| Rounding | DEX math uses integer arithmetic only. Rounding favors protocol safety and slippage checks reject zero/tiny output surprises. |
| State bloat DoS | Tx paths use direct key lookups. List queries are bounded or paginated. Prototype unpaginated list endpoints have explicit caps and MUST FIX before public high-cardinality use. |
| Genesis/bootstrap | Genesis validates, native metadata is present, custom module defaults round-trip, validator stake is positive, and tracked files contain no secrets. |
| Local scripts | Destructive operations validate resolved paths stay inside the workspace and never delete repository root or arbitrary paths. |
| Observability | Health/diagnostic bundle excludes keyring and validator private material. |

## Regression Test Map

| Risk | Regression Coverage |
| --- | --- |
| Deterministic default genesis/export/empty block sequence | `app/determinism_test.go`, `scripts/security/determinism-gate.ps1` |
| Native metadata, staking/fees/mint denom consistency | `app/app_test.go`, `tests/e2e/native_token_smoke.ps1` |
| Invalid PoS delegation denom/funds/validator | `app/pos_test.go` |
| Wrong/malformed fee denom and non-FeeTx | `x/fees/keeper/ante_test.go`, `tests/e2e/fees_ante_smoke.ps1` |
| Tokenfactory unauthorized burn/native spoofing | `x/tokenfactory/keeper/msg_server_test.go` |
| Tokenfactory query malformed/not found/bounded list | `x/tokenfactory/keeper/query_server_test.go` |
| DEX duplicate pair, wrong denom, corrupted pool, LP accounting, slippage | `x/dex/keeper/msg_server_test.go`, `x/dex/keeper/math_test.go`, `tests/e2e/dex_smoke.ps1` |
| DEX query malformed/not found/bounded list | `x/dex/keeper/query_server_test.go` |
| Full prototype tx/query composition | `tests/e2e/prototype_acceptance.ps1` |

## Current Known Triage

| Finding | Source | Severity | Triage |
| --- | --- | --- | --- |
| `go mod verify` reports modified global CometBFT module cache | Local `C:\Users\Ryzen\go\pkg\mod\github.com\cometbft\cometbft@v0.39.3` | Medium | Environment cache issue. Clean module cache or run in CI before release. Do not commit vendored or modified module cache. |
| `GO-2026-5026` in `golang.org/x/net@v0.53.0` | `govulncheck -scan=package` | Medium/High by advisory | Fixed in `v0.55.0`. Upgrade when dependency graph permits; run symbol scan before release to confirm reachability. |
| `GO-2026-5024` in `golang.org/x/sys@v0.43.0` | `govulncheck -scan=package`, Windows | Medium/High by advisory | Fixed in `v0.44.0`. Upgrade when dependency graph permits; Windows prototype builds should keep this visible. |
| `GO-2026-4479` in `github.com/pion/dtls/v2@v2.2.12` | `govulncheck -scan=package` | Medium/High by advisory | No fixed version reported. Track upstream; confirm whether the vulnerable path is reachable from Orbitalis node/runtime before release. |
| `GO-2024-2584` in `github.com/cosmos/cosmos-sdk@v0.54.3` | `govulncheck -scan=package` | High | Cosmos SDK slashing advisory with no fixed version reported by current tool. Keep PoS/slashing smoke tests mandatory and track SDK patch guidance. |
| `gosec` `G115` in generated protobuf | Generated `.pb.go` encode/decode casts | Low | Excluded with `-exclude-generated`; source `.proto` remains linted by `buf lint`. |
| Localnet secrets in ignored directories | `.localnet*`, keyring, validator private files | Low if ignored | Full filesystem secret scans will find generated localnet material. Gate uses staged/history scans; diagnostic bundles exclude private material. |
| `crypto/rand` in vote extension handler | `app/abci.go` determinism scan | Low/Medium | Current handler creates dummy vote-extension bytes and does not write consensus state. It is not production-ready and must be replaced or disabled before any public validator network. |
| `cmttime.Now()` in local genesis generation | `cmd/l1d/cmd/testnet_genesis.go` determinism scan | Low | Genesis time differs per initialization by design; local profile requires identical `genesis.json` across nodes in one init run, not byte-identical fresh runs. |
| `time.Now()`/`math/rand` in speedtest | `cmd/l1d/cmd/speedtest.go` determinism scan | Low | CLI benchmarking only; not in consensus tx/ABCI state path. |

## CI

CodeQL and GitHub Dependency Review are required for PRs through `.github/workflows/security.yml`.

CI does not replace manual Cosmos review. CodeQL catches general Go issues; dependency review catches new vulnerable dependencies; this gate covers Cosmos-specific determinism, authorization, denom, accounting, rounding, and state-bloat risks.
