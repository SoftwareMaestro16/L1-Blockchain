# VM Direction

Phase 11 defines the VM decision boundary for Aetra. Contract standards must
remain testable independent of the final VM choice, and no new VM may be
introduced without a written specification and adversarial evidence.

## Near-Term CosmWasm Candidate

CosmWasm remains the gated near-term VM candidate already reflected in
`app/wasmconfig`.

Required near-term rules:

- CosmWasm is disabled by default.
- CosmWasm is enabled only by explicit config or feature gate.
- Upload permissions are explicit: governance-only by default, allowlist only
  for dev/test networks with a non-empty valid address allowlist.
- Instantiate permissions are explicit: code-owner-only by default, everybody
  only for explicitly gated dev/test networks.
- Admin and migration policy is explicit: migration requires the current
  non-zero contract admin.
- Gas limits are explicit and bounded.
- Contract size limits are explicit and bounded.
- Memory/cache limits are explicit and bounded.
- Query limits are explicit and bounded.
- Pinned code is disabled by default and governance-only if enabled later.
- Governance authority for enabling/disabling CosmWasm is explicit.
- CosmWasm contracts cannot bypass `naet` fee policy, address policy,
  zero-address policy, or genesis validation.

The current readiness policy defines:

- recommended wasmd version: `v0.70.2`
- recommended wasmvm version: `v3.0.6`
- recommended Cosmos SDK minor: `v0.54`
- max stored contract size: `800 KiB`
- max proposal contract size: `3 MiB`
- smart query gas limit: `3,000,000`
- simulation gas limit: `20,000,000`
- gas multiplier: `140,000`
- memory cache: `100 MiB`, hard cap `256 MiB`
- smart query response limit: `256 KiB`, hard cap `1 MiB`
- smart query depth limit: `8`, hard cap `16`
- pinned code count: `0` by default, hard cap `128`

CosmWasm readiness must not add a `wasm` store key, genesis state, module
account, CLI upload/instantiate/execute/migrate surface, or keeper wiring unless
the explicit gate is enabled and the full security checklist is satisfied.

## Aetra VM R&D Decision

Aetra contracts may eventually be one of these models:

- CosmWasm contracts with Aetra async/message standards.
- AVM, the Aetra Virtual Machine, with its own execution, storage, and
  message ABI.
- Both, with strict compatibility boundaries and explicit migration rules.

No future Aetra VM may be implemented before the R&D spec defines:

- binary serialization spec
- message ABI
- storage ABI
- gas schedule
- deterministic execution proof
- fuzz tests
- upgrade/migration policy
- adversarial audit

The async execution package `x/aetravm/async` is not a VM. It is an
executable semantics specification for deterministic queue behavior, bounce
behavior, fee/value denomination, limits, observability, and export/import.
Production sharding is separate sharding R&D and remains experimental until
`docs/architecture/sharding-rd.md`, `x/sharding/sim`, prototype keepers,
adversarial tests, long-run testnet, independent audit, and consensus-safety
proof are complete.

The AVM package `x/aetravm/avm` is the current pure Go executable
specification for the native Aetra Virtual Machine. It defines deterministic
bytecode encoding, verifier, local runner, storage snapshot ABI, gas schedule,
host function allowlist, forbidden opcode checks, and an async handler adapter.
It is not wired into SDK stores or keepers.

## Compatibility Boundaries

Contract standards are VM-independent until a runtime is selected. The Go
packages under `x/aetravm/standards` define executable specifications,
deterministic message codecs, and async/AVM-compatible conformance handlers,
not SDK modules or store wiring.

Every contract standard must define:

- explicit storage schema
- explicit inbound messages
- explicit outbound messages
- explicit getters
- explicit unknown-message policy
- explicit bounce behavior
- explicit fee behavior
- explicit deployment behavior

If CosmWasm is used first, CosmWasm message handlers must conform to these
standard packages. If a native Aetra VM is later introduced, its ABI must
either match the same standards or declare a versioned migration.

## Acceptance Gates

- CosmWasm readiness does not weaken base chain security.
- Aetra async VM research has a written spec before implementation.
- Contract standards can be tested independent of the VM choice.
- `go test ./app/wasmconfig`
- `go test ./x/aetravm/standards/...`
- `go test ./x/aetravm/async`
- `go test ./x/aetravm/avm`
