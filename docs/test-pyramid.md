> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Prototype Test Pyramid

This matrix makes the working L1 flows observable across unit, integration, adversarial, e2e, determinism, and benchmark layers. It is the coverage companion to [Security And Testing Strategy](security-testing.md), [Prototype Transaction Lifecycle Matrix](transaction-lifecycle-matrix.md), [Prototype Tx Event Contract](event-contract.md), [Cosmos Security Audit Checklist](security/cosmos-security-checklist.md), and [Prototype Security And Determinism Audit Gate](security/prototype-audit-gate.md).

Rule: define the target test layer before implementation where practical. No behavior change should merge without a targeted unit, integration, adversarial, e2e, determinism, or benchmark check, or a documented `MUST FIX` with owner and release decision.

## Fast PR Suite

Run locally before commit:

```powershell
go test ./...
.\tests\scripts\test_pyramid_doc_test.ps1
.\tests\scripts\cosmos_security_checklist_doc_test.ps1
.\tests\scripts\transaction_lifecycle_matrix_doc_test.ps1
.\tests\scripts\event_contract_doc_test.ps1
.\tests\scripts\determinism_gate_test.ps1
.\scripts\security\prototype-audit.ps1 -Profile Fast
```

Smoke after localnet or tx rejection changes:

```powershell
.\tests\e2e\prototype_acceptance.ps1 -Profile Smoke
.\tests\e2e\mempool_negative_smoke.ps1
.\tests\e2e\load_profile_smoke.ps1
```

## Nightly Or Manual

Benchmarks and high-cardinality generators are not part of every PR run:

```powershell
go test -run '^$' -bench 'Benchmark(EmptyBlock|Dex)' ./app ./x/dex/keeper
.\scripts\security\prototype-audit.ps1 -Profile Nightly
.\tests\e2e\prototype_acceptance.ps1 -Profile Full
```

## Matrix

| Module Or Flow | Unit | Integration | Adversarial | E2E Smoke | Determinism | Benchmark Or Perf |
| --- | --- | --- | --- | --- | --- | --- |
| App genesis/export/module accounts | `app/app_test.go`, `app/export_import_test.go`, `app/params/token_test.go` | `TestCustomModuleGenesisInitExportRoundTrip`, `TestStateExportImportPreservesPrototypeModuleData` | corrupted module genesis, corrupted exported DEX reserve, invalid staking/fees/tokenfactory/DEX state | `tests/e2e/export_import_smoke.ps1`, `tests/e2e/native_token_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | `app/determinism_test.go` | `BenchmarkEmptyBlockFinalizeCommit` |
| Fees ante and params | `x/fees/types/genesis_test.go`, `x/fees/keeper/msg_server_test.go` | app ante decorator via `x/fees/keeper/ante_test.go` | wrong denom, multi-denom, malformed fee, duplicate denom, zero/empty fee, non-`FeeTx`, invalid authority | `tests/e2e/fees_ante_smoke.ps1`, `tests/e2e/mempool_negative_smoke.ps1` | covered by app export and determinism gate | MUST FIX before public fee accounting: benchmark protocol fee distribution once implemented |
| Tokenfactory denom lifecycle | `x/tokenfactory/types/genesis_test.go`, `x/tokenfactory/keeper/msg_server_test.go` | bank supply/metadata checks through keeper tests | invalid signer, unauthorized admin, burn-from mismatch, native spoofing, malformed denom/admin | `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/native_token_smoke.ps1`, `tests/e2e/mempool_negative_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | default export plus deterministic static gate | bounded denom list pagination covered; add high-cardinality benchmark before public API load |
| DEX pool and swap lifecycle | `x/dex/keeper/math_test.go`, `x/dex/types/genesis_test.go` | `x/dex/keeper/msg_server_test.go` accounting/invariant checks | duplicate pair, wrong denom, corrupted pool, LP denom mismatch, slippage, tiny rounding output | `tests/e2e/dex_smoke.ps1`, `tests/e2e/mempool_negative_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | deterministic static gate and app export | `BenchmarkDexCreatePoolsAndSwap`, `tests/e2e/load_profile_smoke.ps1`, future invariant benchmark for many pools |
| Query surface CLI/gRPC/REST | `x/*/keeper/query_server_test.go` | app/localnet query composition through smoke helpers | nil request, malformed denom/id, not found, invalid pagination key, excessive limit | `tests/e2e/query_surface_smoke.ps1` | read-only; covered by no-state-write review in security checklist | add REST/gRPC high-cardinality pagination benchmark before public API load |
| PoS and bank native flow | `app/pos_test.go`, `app/app_test.go` | staking delegation updates validator power | invalid validator address, wrong denom, insufficient funds, invalid sequence/replay | `tests/e2e/pos_smoke.ps1`, `tests/e2e/native_token_smoke.ps1`, `tests/e2e/mempool_negative_smoke.ps1` | app default genesis/export determinism | empty block benchmark covers app path; minimal bank load in `tests/e2e/load_profile_smoke.ps1`; staking load benchmark is nightly/manual |
| Localnet scripts and release artifacts | `tests/scripts/*_test.ps1` | `tests/e2e/localnet_smoke.ps1` | unsafe path deletion, local-only keyring docs, no mnemonic/private key output | `tests/e2e/prototype_acceptance.ps1` | reproducible single-run genesis/config review | `prototype-audit.ps1 -Profile Nightly` |
| Security gates and docs | `tests/scripts/cosmos_security_checklist_doc_test.ps1`, `tests/scripts/test_pyramid_doc_test.ps1`, `tests/scripts/transaction_lifecycle_matrix_doc_test.ps1`, `tests/scripts/event_contract_doc_test.ps1`, `tests/scripts/prototype_docs_path_test.ps1` | `scripts/security/prototype-audit.ps1 -Profile Fast` | untriaged Critical/High, secrets, generic scanner issues, Cosmos checklist gaps, tx lifecycle gaps, event contract gaps | release package tests | `scripts/security/determinism-gate.ps1` | Nightly audit profile |

## Priority Gaps

| Gap | Priority | Decision |
| --- | --- | --- |
| Protocol fee accounting benchmark | High when accounting lands | MUST FIX before fee distribution code leaves prototype scope |
| Tokenfactory and query pagination load benchmark | High before public high-cardinality API | Required before public API/load testing milestone |
| Cross-architecture app hash replay | Medium | Manual/nightly until CI has multi-arch runners |
| Stateful DEX invariant generator for many pools/accounts | Medium | Add when pool count or routing expands beyond one-hop prototype swaps |

Shared helpers are acceptable when they reduce repeated Cosmos setup. They must keep signer, funding, and keeper dependencies visible at the call site.
