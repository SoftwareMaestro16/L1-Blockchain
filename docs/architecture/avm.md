# Aetra Virtual Machine

AVM is the native Aetra Virtual Machine research track for asynchronous
contracts. The current implementation is a pure Go executable specification in
`x/aetravm/avm`; it is not wired into SDK keepers, module accounts, genesis,
CLI, or ABCI hooks.

AVM cannot mutate production chain state until the base-chain safety gate,
async queue semantics, security scans, determinism gate, and adversarial audit
are green.

## Non-Negotiable Requirements

- binary serialization spec
- message ABI
- storage ABI
- deterministic execution proof
- gas schedule
- memory limits
- code size limits
- stack/register limits
- host function allowlist
- fuzz tests
- differential tests
- upgrade and migration policy
- adversarial audit

## Bytecode Format

The AVM bytecode format is deterministic and big-endian:

```text
magic               4 bytes, "AVM1"
version             uint16
metadata_hash       32 bytes
import_count        uint16
imports             repeated uint16 host function ids
export_count        uint16
exports             repeated (uint8 entrypoint, uint32 instruction offset)
instruction_count   uint32
instructions        repeated (uint8 opcode, uint64 arg, uint16 data_len, data)
```

The module code hash is `sha256(encoded_module)`. The verifier rejects malformed
headers, unsupported versions, oversized code, unknown imports, missing exports,
invalid export offsets, unknown opcodes, nondeterministic opcodes, and oversized
instruction data.

## Execution Entrypoints

AVM entrypoints are explicit:

- `deploy`
- `receive external`
- `receive internal`
- `receive bounced`
- `query/getter`
- `migrate`

Async messages use `receive internal` by default and `receive bounced` when the
message envelope has `bounced = true`. Missing bounced handlers fail
deterministically and must not create bounce/refund loops.

## Message ABI

AVM receives the Aetra async message envelope:

- source
- destination
- value in `naet`
- opcode
- query id
- body
- bounce flag
- bounced flag
- created logical time
- deadline
- gas limit
- forward fee
- depth

AVM output messages are normal async `MessageEnvelope` values. They inherit
deterministic queue semantics from `x/aetravm/async`.

## Storage ABI

AVM storage is a per-contract deterministic key/value namespace:

- keys are byte strings with bounded length;
- values are byte strings;
- integer helpers use big-endian encoding;
- snapshots are sorted by key;
- iteration must be bounded and paginated;
- exported snapshots must be deterministic;
- state size and memory limits are enforced before committing state.

## Host Function Allowlist

Allowed host functions:

- read storage
- write storage
- emit internal message
- inspect message envelope
- get block context
- charge gas
- return result code

Any host function outside the allowlist is rejected by the verifier.

## Forbidden Behavior

AVM bytecode and host functions must not allow:

- wall-clock time
- random host entropy outside consensus-approved randomness
- filesystem or network access
- floating point
- unbounded iteration
- nondeterministic map iteration

The executable verifier rejects the forbidden opcode set used to model these
classes.

## Gas And Limits

AVM execution is bounded by:

- gas schedule per opcode;
- per-message gas limit;
- max code bytes;
- max instructions;
- max imports;
- max stack depth;
- max memory bytes;
- max key/data sizes;
- async emitted-message limits;
- async storage-write limits.

Gas accounting is deterministic and local to the runner. When AVM is wired into
keepers, keeper gas and AVM gas must be reconciled without allowing non-`naet`
protocol fees.

## Toolchain

Required AVM toolchain components:

- bytecode verifier
- disassembler
- local runner
- gas profiler
- contract test harness
- state snapshot inspector

The current pure Go package includes the verifier, deterministic encoder and
decoder, local runner, storage snapshot encoder, and async handler adapter.
Disassembler, gas profiler CLI, and snapshot inspector CLI remain future work.

## Keeper Gate

AVM keeper wiring is a separate production gate. It must provide:

- store code
- instantiate contract
- route external message
- process internal queue
- execute getters
- export/import state

Keeper wiring must reuse the async queue export/import semantics and must not
bypass address validation, zero-address rejection, `naet` fee policy, signer
checks, malformed transaction handling, or genesis validation.

## Required Tests

```powershell
go test ./x/aetravm/avm
go test ./x/aetravm/async
go test ./...
```

The current package tests cover:

- deploy valid contract module
- reject malformed bytecode
- reject oversized code
- reject nondeterministic opcode
- reject non-allowlisted host function
- run simple counter deterministically
- deterministic gas bound
- deterministic storage snapshot
- storage memory limit
- send internal message through async queue
- bounce failed message
- preserve state across runner calls
- preserve queue across export/import

Fuzz tests, differential tests, keeper tests, and adversarial audit are required
before AVM can be enabled beyond the executable specification.

## Acceptance

- AVM can execute a minimal contract deterministically.
- AVM gas is deterministic and bounded.
- AVM contracts can participate in the async message queue.
- AVM cannot weaken base-chain signer, fee, denom, zero-address, transaction,
  or genesis validation.
