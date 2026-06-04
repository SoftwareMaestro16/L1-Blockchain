# Cosmos Security Audit Checklist

Use this checklist for every prototype change that touches app wiring, custom modules, proto/query surface, localnet tooling, or release artifacts. It is Cosmos-specific review, not a replacement for `go test`, `go vet`, `buf lint`, or the runnable gate in [Prototype Security And Determinism Audit Gate](prototype-audit-gate.md).

The reviewer must load and apply `cosmos-vulnerability-scanner` before marking the checklist complete. Required scanner themes: missing denom validation, insufficient authorization, missing balance check, ABCI panic, nondeterminism, rounding, and unbounded loops.

## Release Rule

Prototype release is blocked when any `Critical` or `High` finding is untriaged. Each `Critical` or `High` item must have one of:

- merged fix plus regression test,
- documented blocker with owner, issue or task link, and decision,
- explicit downgrade with rationale, owner, and accepted prototype scope.

Every change must record architecture, security, scalability, and test strategy review. Post-code review must also record any refactor opportunity that was intentionally deferred.

## Review Record

| Field | Required Entry |
| --- | --- |
| Change scope | App, module, proto/query, script, release artifact, or future module name |
| Architecture decision | Boundary kept, dependency direction, no keeper business logic leaked into CLI/proto |
| Security decision | Signer/authority, denom, balance/accounting, params, panic, and nondeterminism result |
| Scalability decision | Direct lookup, pagination/cap, bounded loop, or documented blocker |
| Test decision | Regression test, e2e smoke, static gate, or documented blocker |
| Refactor note | Follow-up opportunity, owner, or `none` |

## App

- InitGenesis and ExportGenesis validate module state, metadata, params, module accounts, and validator stake before state is trusted.
- BeginBlocker, EndBlocker, ante decorators, vote extensions, and app wiring do not depend on wall time, randomness, map iteration order, floats, goroutines, select races, external APIs, pointer addresses, or platform-dependent serialization.
- Panic is limited to impossible startup/wiring failures. Malformed tx, query, genesis input, and params input return errors.
- Module accounts are registered with the minimum permissions required for tokenfactory, dex, fees, staking, bank, mint, and distribution interactions.
- Native token assumptions use `norb` as the base denom and keep `ORB` as display metadata only.

## Tokenfactory

- Every user-provided denom and subdenom is SDK-validated and bounded by params before bank metadata, mint, burn, or supply mutation.
- Mint, burn, and admin transfer require the current denom admin signer. `MsgUpdateParams` requires the configured authority.
- Native `norb`, staking, fees, and DEX LP denoms cannot be spoofed or overwritten.
- Bank keeper mint/burn/send errors propagate without local bookkeeping changes after failure.
- Denom list/query endpoints are paginated or explicitly capped; not found and malformed requests return status errors.

## DEX

- Pool creation rejects duplicate unordered pairs, invalid denoms, same-denom pairs, zero liquidity, unsupported LP denoms, and params outside bounds.
- Add/remove/swap paths check user balances through bank errors and keep recorded reserves, module balances, and LP supply synchronized.
- Constant-product math uses integer arithmetic only. Rounding and fee handling favor protocol safety, and slippage/min-out checks reject zero or tiny-output surprises.
- Pool lookup and tx paths are direct-key or otherwise bounded. List queries are paginated or capped.
- Corrupted pool, reserve desync, invalid pool id, malformed request, and not found cases return errors without panic.

## Fees

- Ante fee policy accepts prototype tx fees only in allowed denom `norb`.
- Empty fee, zero fee, malformed fee, wrong denom, multi-denom fee, duplicate allowed denoms, and non-`FeeTx` inputs fail safely.
- Params validation bounds allowed denoms and fee policy values. `MsgUpdateParams` rejects invalid authority.
- Fee policy stays separate from future protocol fee accounting, and any accounting change must include bank balance/supply invariants.

## Localnet Scripts

- Scripts never print mnemonics, private keys, validator private material, keyring files, database URLs, or environment dumps by default.
- `--keyring-backend test` is documented as local-only.
- Destructive operations resolve paths and refuse repository root, parent directories, and paths outside the intended workspace.
- Generated genesis, config, and release inputs are reproducible for a single localnet run and contain no tracked secrets.

## Proto And Query

- Query methods are read-only and return gRPC status errors for nil, malformed, not found, and invalid id requests.
- List endpoints use pagination/default limits or an explicit prototype cap with a MUST FIX note before public high-cardinality use.
- Proto changes pass `buf lint` and follow generation policy. Public breaking changes require versioning.
- CLI, gRPC, and REST examples match [Query Surface](../query-surface.md); no local filesystem paths or secrets appear in API output.

## Release Artifacts

- Release packages include security docs, audit outputs, operator docs, query docs, and smoke-test references needed to reproduce the prototype.
- Staged/history secret scans pass or have documented local-only generated artifact triage.
- `go mod verify`, dependency advisories, CodeQL, and Dependency Review findings are either clean or triaged with owner and milestone.
- Known prototype exceptions are listed in [Prototype Security And Determinism Audit Gate](prototype-audit-gate.md) and do not hide untriaged `Critical` or `High` risk.

## Risk To Test Map

| Risk | Required Review | Regression Or Gate |
| --- | --- | --- |
| Nondeterministic AppHash | App, keeper, genesis/export, ABCI | `app/determinism_test.go`, `scripts/security/determinism-gate.ps1`, `tests/scripts/determinism_gate_test.ps1` |
| ABCI or tx panic | App, fees ante, msg servers, query servers | `go test ./...`, module keeper/query tests, `scripts/security/prototype-audit.ps1 -Profile Fast` |
| Missing denom validation | tokenfactory, dex, fees | `app/app_test.go`, `x/tokenfactory/keeper/msg_server_test.go`, `x/dex/keeper/msg_server_test.go`, `x/fees/keeper/ante_test.go` |
| Insufficient authorization | tokenfactory, dex params, fees params | `x/tokenfactory/keeper/msg_server_test.go`, `x/dex/keeper/msg_server_test.go`, `x/fees/keeper/msg_server_test.go` |
| Missing bank balance or accounting check | dex, tokenfactory, fees accounting | `x/dex/keeper/msg_server_test.go`, `x/dex/keeper/math_test.go`, `tests/e2e/dex_smoke.ps1` |
| Wrong fee policy | fees ante and params | `x/fees/keeper/ante_test.go`, `tests/e2e/fees_ante_smoke.ps1` |
| Rounding or invariant leakage | dex math/accounting | `x/dex/keeper/math_test.go`, `tests/e2e/dex_smoke.ps1` |
| Unbounded loop or state bloat | list queries, tx paths, localnet scripts | query server tests, `docs/query-surface.md`, `scripts/security/determinism-gate.ps1` |
| Malformed query request | fees, tokenfactory, dex, bank/staking examples | `x/*/keeper/query_server_test.go`, `tests/e2e/query_surface_smoke.ps1` |
| Local secret exposure | localnet scripts, diagnostics, release package | `scripts/security/prototype-audit.ps1`, `tests/scripts/prototype_release_package_test.ps1` |

Future modules must add rows before merging their first keeper or message implementation.
