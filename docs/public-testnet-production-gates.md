# Public Testnet And Production Gates

This file is the release gate ledger for Aetra public testnet and later
production readiness. It is stricter than the prototype acceptance suite and
does not replace module-specific security checklists. The full Track 10 test
matrix, production gate, build order, and economic summary are recorded in
[Test And Production Gates](test-production-gates.md).

## Public Testnet Gate

Public testnet cannot open until all required items are green or explicitly
triaged with owner, severity, mitigation, and target milestone.

Required checks:

- `go test ./...` passes.
- `go vet ./...` passes.
- `buf lint` passes.
- Security scans pass or findings are triaged:
  - `govulncheck`
  - `gosec`
  - CodeQL
  - gitleaks
  - dependency review
- Deterministic execution gate passes:
  - `scripts\security\determinism-gate.ps1`
- 3-validator, 5-validator, and 10-validator localnet profiles pass:
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 3`
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 5`
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 10`
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile All`
- Snapshot and state-sync work from published trust height, trust hash, and at
  least two RPC servers.
- Validator onboarding docs are clean and tested from a fresh machine or clean
  home directory.
- Faucet plan is implemented or explicitly deferred.
- Explorer/indexer plan is implemented or explicitly deferred.
- Incident response and rollback docs are tested.
- CosmWasm smoke passes if CosmWasm is enabled:
  - `tests\e2e\cosmwasm_smoke.ps1 -EnableWasm`
- AVM smoke passes if AVM is enabled:
  - `go test ./x/aetravm/avm ./x/aetravm/async`
- Contract standard smoke passes if async contracts are enabled:
  - `go test ./x/aetravm/standards/...`

Blocking rule:

- Any untriaged `Critical` or `High` fund-safety, consensus-safety, or
  secret-leak finding blocks public testnet.

## Phase 12 Modular Execution Public Testnet Gate

Modular execution features cannot be advertised on public testnet until their
specific gates have passed. This section is the minimum Phase 12 gate for
Execution Zones, Compute Shards, Aether Mesh, Identity, routing, load scoring,
and VM readiness. Passing the base public testnet gate does not automatically
promote any modular execution feature from experimental to advertised.

Required gates:

- Base chain hardening complete.
- Determinism gate passes:
  - `tests\scripts\determinism_gate_test.ps1`
  - `scripts\security\determinism-gate.ps1`
- Export/import gate passes.
- Genesis migration gate passes.
- 3-validator localnet long-run passes.
- 5-validator localnet long-run passes.
- 10-validator stress profile passes:
  - `scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 10`
- State-sync and snapshot restore pass.
- Load score and routing simulator pass:
  - `go test ./x/load/... ./x/routing/...`
  - `aetrad execution-os smoke --profile execution-os-sim`
- Sharding simulator pass:
  - `go test ./x/sharding/sim`
- Mesh simulator pass:
  - `go test ./x/mesh/...`
- Identity executable spec pass:
  - `go test ./x/identity/types`
- VM readiness tests pass:
  - `go test ./x/aetravm/avm ./x/aetravm/async ./x/vm/types`
- Security scans pass or findings are owner-triaged:
  - `govulncheck`
  - `gosec`
  - CodeQL
  - gitleaks
  - dependency review
- Independent audit findings are triaged.
- Public docs do not overclaim production sharding or production
  smart-contract execution.

Acceptance:

- Public testnet can advertise only the features that passed their gate.
- Any feature still behind R&D remains documented as experimental.

Suggested branch order:

1. `feature/load-score-spec`
2. `feature/routing-engine-spec`
3. `feature/zone-registry-sim`
4. `feature/load-driven-sharding-sim`
5. `feature/aether-mesh-sim`
6. `feature/identity-zone-spec`
7. `feature/contract-zone-readiness`
8. `prototype/execution-os-keepers`
9. `prototype/aether-core-routing-wiring`
10. `tooling/execution-os-localnet`
11. `security/execution-os-invariants`
12. `testnet/modular-execution-gate`

Each branch must end with:

```powershell
go test ./...
go vet ./...
buf lint
powershell -NoProfile -ExecutionPolicy Bypass -File tests\scripts\determinism_gate_test.ps1
```

Run `buf generate` only when protobuf files are changed.

Traceability matrix:

| Design requirement | Implementation phase |
| --- | --- |
| Aether Core as control plane | Phase 9 |
| Execution Zones | Phase 3, Phase 8 |
| Compute Shards | Phase 4 |
| Deterministic `LOAD_SCORE` | Phase 1 |
| Load spike resistance | Phase 1, Phase 4 |
| Deterministic routing | Phase 2 |
| Aether Mesh | Phase 5 |
| `.aet` Identity Layer | Phase 6 |
| Economic security | Phase 0, Phase 9, Phase 12 |
| Low-fee congestion model | Phase 1, Phase 2, Phase 11 |
| Trilemma claim support | Phase 12 only after accepted gates |

Final rule:

Do not wire production Execution Zones, production Compute Shards, production
Aether Mesh, or production contract execution into Aether Core until the
corresponding executable spec and simulator have passed deterministic tests,
adversarial tests, export/import tests, benchmarks, long-run localnet tests,
and independent audit review.

## Production Gate

Production cannot be claimed until all public testnet gates remain green over
a long-running public testnet and the additional production requirements are
met.

Required production evidence:

- Long-running public testnet has no untriaged consensus-safety or fund-safety
  issues.
- Validator set can upgrade safely.
- Staking, fees, DEX, AVM, and contract standards have adversarial tests.
- State export/import is deterministic.
- Independent audit is completed and high/critical findings are fixed or
  explicitly accepted by governance with public rationale.
- Emergency governance and halt/restart process is tested.
- Snapshot/state-sync restore produces the same expected app hash.
- Public RPC, explorer/indexer, faucet, validator docs, incident response, and
  rollback process have owners and operational runbooks.

Production exclusions:

- Sharding remains `sharding R&D` or `experimental sharding` until
  `docs/architecture/sharding-rd.md` production gate is complete.
- CosmWasm remains disabled unless explicitly enabled by config and gate tests.
- AVM remains non-production until keeper wiring, adversarial tests, fuzz
  tests, export/import, and audit gates are complete.

## Immediate Build Order

1. Finish base-chain safety and Phase 2 helper cleanup.
2. Finish PoS/staking production hardening.
3. Build deterministic async queue without AVM first.
4. Build minimal AVM with a counter contract.
5. Implement AW-5 wallet.
6. Implement AFT-44 token master/wallet.
7. Implement ANFT-66 NFT collection/item.
8. Implement ASBT-67 soulbound item.
9. Gate CosmWasm behind explicit config and tests.
10. Start sharding simulator and spec.
11. Only after simulator and audit, prototype masterchain/workchain/shardchain.

## Evidence Map

| Gate Area | Evidence |
| --- | --- |
| Base chain tests | `go test ./...`, `go vet ./...`, `buf lint` |
| Security scans | `docs/security/security-audit-pack.md`, `.github/workflows/security.yml` |
| Determinism | `scripts\security\determinism-gate.ps1`, `docs/security/prototype-audit-gate.md` |
| Localnet 3/5/10 profiles | `scripts\testnet\public-testnet-preflight.ps1` |
| Snapshot/state-sync | `docs/public-testnet-preparation.md`, `scripts\localnet\snapshot.ps1`, `scripts\localnet\statesync.ps1` |
| Load/routing simulator | `x/load`, `x/routing`, `aetrad execution-os smoke --profile execution-os-sim` |
| Sharding simulator | `x/sharding/sim`, `docs/architecture/sharding-rd.md` |
| Mesh simulator | `x/mesh`, `tests/adversarial/modular_execution_invariants_test.go` |
| Identity executable spec | `x/identity/types` |
| VM readiness | `x/aetravm/avm`, `x/aetravm/async`, `x/vm/types`, `app/wasmconfig` |
| Validator onboarding | `docs/validator-onboarding.md` |
| Faucet | `docs/public-testnet-preparation.md#faucet-plan` |
| Explorer/indexer | `docs/public-testnet-preparation.md#explorer-and-indexer-plan` |
| Incident/rollback | `docs/testnet-incident-response.md`, `docs/public-testnet-preparation.md#rollback-and-restart-procedure` |
| CosmWasm | `app/wasmconfig`, `tests\e2e\cosmwasm_smoke.ps1`, `docs/security/cosmwasm-readiness.md` |
| AVM | `x/aetravm/avm`, `docs/architecture/avm.md` |
| Contract standards | `x/aetravm/standards/...`, `docs/standards` |
| Sharding R&D | `x/sharding/sim`, `docs/architecture/sharding-rd.md` |

## Decision Record

Before public testnet launch, operators must publish:

- release commit,
- binary checksum,
- genesis hash,
- seed/persistent peers,
- chain id,
- expected native denom `naet`,
- public RPC endpoints,
- snapshot/state-sync trust data,
- faucet status: implemented or explicitly deferred,
- explorer/indexer status: implemented or explicitly deferred,
- enabled experimental features, if any.
