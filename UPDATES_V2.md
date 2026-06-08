# Aetra Testnet Completion Backlog V2

This file is the practical backlog for turning Aetra into a runnable, testable,
validator-understandable public testnet candidate.

The goal of V2 is not to add more native modules. The goal is:

- one stable node binary;
- deterministic genesis;
- 4-5 node localnet;
- export/import restart;
- upgrade rehearsal;
- validator join guide;
- pool-based PoS that matches the product model;
- AVM v1 that can run real contracts;
- minimal, boring, validator-grade infrastructure.

## Non-Negotiable Direction

- Keep the runnable chain small.
- Do not introduce native DEX/token/NFT modules for application assets.
- Token, NFT, DEX, wallet standards, and domains must be AVM contracts or
  registries unless explicitly listed as system state.
- Normal user staking goes through the official liquid staking/nominator pool
  path only. A wallet user deposits at least `10 AET` into the official pool and
  never chooses a validator address.
- Direct wallet-to-validator delegation is disabled for the public testnet
  product model. If an operator-only escape hatch exists, it must be disabled
  by default, governance-gated, excluded from normal CLI/API/docs, and covered
  by rejection tests.
- `AE...` remains the only user-facing address format.
- `4:...` remains raw/internal/proof format.
- Private keys, seed phrases, validator private keys, keyrings, and localnet
  secrets never appear in genesis, events, exports, logs, docs examples, or
  release artifacts.
- BeginBlock/EndBlock and block lifecycle must not scan all users/contracts.
- Export/import must preserve every consensus state field needed for restart.

## Current Repo Observations

This backlog was written after inspecting the repo state around:

- `cmd/l1d/cmd`;
- `scripts/localnet`;
- `scripts/testnet`;
- `scripts/release`;
- `.github/workflows/testnet-readiness.yml`;
- `x/contracts`;
- `x/aetravm/avm`;
- `x/aetravm/async`;
- `x/aetravm/standards`;
- `x/nominator-pool`;
- validator and testnet docs.

Observed status:

- Testnet readiness CI exists, but it must become the canonical release gate,
  not only an auxiliary workflow.
- `x/nominator-pool/types/state.go` already defines `DefaultMinPoolDeposit` as
  `10 AET` in base units and has direct-delegation params. V2 must make this the
  user-facing rule everywhere: wallet -> pool deposit, no validator selection.
- Exit codes already exist under `x/contracts/types/exit_codes.go` and are small
  stable values. Keep them small; do not introduce huge opaque error numbers.
- AVM has a stack VM, gas table, host registry, storage ABI, async messages, and
  receipts/events work in progress.
- AVM has `EntryQuery`, but contract get methods are not yet a full user-facing,
  read-only contract query workflow.
- AVM storage currently looks more like bounded key/value storage. The target
  testnet AVM model should move persistent contract data toward immutable
  content-addressed Chunks and ChunkMap indexes.
- `app/invariants.go` exists and registers many important app invariants, but
  some checks still use exported-genesis/model-level shortcuts or placeholder
  inputs. V2 must bind launch invariants to live keeper state and CI.
- `app/keeperwiring/native.go` still wires some policy/economics/score keepers
  without store services, while other modules are KV-backed. Any runtime keeper
  used for testnet consensus must persist mutations through KV and survive
  restart/export/import.
- The repo contains many prototype modules under `x/`. Testnet launch should not
  attempt to productionize every module. Freeze, remove from app wiring, or mark
  future-only anything outside the minimal kernel.
- Docs still contain historical tokenfactory/DEX language in places. For V2,
  keep those as future AVM standards or remove from testnet launch docs.
- Docker image, seed/peer publication, Cosmovisor guide, and top-level
  `docs/VALIDATOR.md` / `docs/TESTNET.md` / `docs/COSMOVISOR.md` are not yet the
  single canonical operator path.

## V2.1 Launch Blockers From Code Audit

These are explicit blockers for public testnet readiness.

### Blocker 1 - Pool-Only User Staking Must Be Enforced Everywhere

Requirement:

- Wallet, CLI, API, docs, examples, and smoke tests must expose one default
  staking action:

```text
AE wallet -> official liquid staking/nominator pool -> pool allocations -> validators
```

- Minimum user deposit: `10 AET`.
- User deposit message must not contain:
  - validator operator address;
  - consensus address;
  - `aevaloper`;
  - `aevalcons`;
  - raw `4:...` validator target.
- Validator choice belongs to deterministic pool allocation/accounting, not to
  the user.
- Direct wallet-to-validator delegation must fail before any staking mutation.
- Native staking hooks may be callable only by the official pool/contract
  capability path.

Acceptance:

- CLI help for normal staking has no validator argument.
- Normal tx/API protobuf for pool deposit has no validator field.
- E2E smoke includes `delegate-direct-disabled`.
- A 9.999999 AET deposit is rejected.
- A 10 AET deposit from an active wallet succeeds.
- Pool share minting for 10 AET is deterministic across two runs.
- Export/import after 10 AET deposit returns the same pool share and pool totals.

### Blocker 2 - Consensus Validator Set Must Match The Aetra PoS Model

Current risk:

- Validator registry/election/score modules can exist while CometBFT still uses
  the standard SDK staking validator set as the actual consensus source.

Requirement:

- Either:
  - Aetra validator election writes the effective validator set used by
    CometBFT; or
  - app invariants enforce that SDK staking state exactly mirrors the Aetra
    registry/election output at every boundary.
- No validator can be active in consensus unless it satisfies:
  - registry active status;
  - min stake/funding mode;
  - pool-backed self/nominator split;
  - commission bounds;
  - slashing/insurance requirements;
  - score/reputation eligibility.

Acceptance:

- Active CometBFT validator set query equals Aetra elected set.
- Attempted validator activation below entry rules is rejected before consensus
  power changes.
- Export/import preserves validator power and election snapshots.
- Upgrade test preserves the exact validator set unless migration explicitly
  changes it with a deterministic migration record.

### Blocker 3 - Storage Rent Must Be Runtime-Enforced

Current risk:

- Storage rent models and frozen states exist, but launch must prove rent is
  charged in ante/account lifecycle/contract execution, not only in helpers.

Requirement:

- Every active stateful account and contract accrues rent from:
  - code bytes;
  - data bytes;
  - unpaid debt;
  - elapsed consensus time/height.
- Rent is collected during transactions/actions from configured payer:
  - owner account;
  - contract balance;
  - pool reserve;
  - module balance;
  - protocol payer.
- Frozen wallet/contract keeps code, data, sequence, ownership, balance, and
  proof access.
- Frozen state recovers through top-up -> pay storage debt -> unfreeze.
- Protocol-critical state cannot freeze/archive/delete because of rent.

Acceptance:

- Contract execute accrues rent before compute.
- Account tx path accrues rent before normal action.
- Frozen contract rejects normal execute but accepts top-up/pay-debt/unfreeze and
  read/proof queries.
- Protocol-critical module action still executes under system rent stress.

### Blocker 4 - Prototype Keepers Must Not Rely On Runtime Memory

Requirement:

- Any keeper used in testnet consensus must:
  - write prefix records in `InitGenesis`;
  - read/write KV in every mutating method;
  - export deterministic sorted genesis from KV;
  - paginate over bounded prefixes;
  - migrate old single-genesis blobs to prefix layout in migration `1 -> 2`.
- In-memory state is allowed only for tests or explicit fixtures with no
  store service.

Acceptance:

- mutate -> export -> new app import -> query returns same state for:
  - nominator pool;
  - validator registry;
  - validator election;
  - validator insurance;
  - config/config-voting;
  - system registry;
  - storage rent;
  - actor registry;
  - AVM scheduler.

### Blocker 5 - AVM Must Be A Real VM, Not A Demo Interpreter

Requirement:

- AVM v1 must have:
  - canonical bytecode/module format;
  - verifier;
  - instruction set spec;
  - gas schedule;
  - typed values;
  - deterministic stack execution;
  - Chunk/ChunkMap persistent state;
  - deploy/execute/get methods;
  - receipts/events/proofs;
  - host function allowlist;
  - no forbidden nondeterminism;
  - examples that compile/run in CI.

Acceptance:

- Same code/state/message always gives same:
  - exit code;
  - gas used;
  - receipt hash;
  - state root;
  - outgoing messages.
- Out-of-gas and failed migration roll back state.
- Get methods cannot mutate state or emit consensus messages.
- Minimal examples cover counter, domain registry, token-like ledger, internal
  message, bounce/refund, and get methods.

## Parallel Workstream Ownership

Use these ownership boundaries so multiple Codex chats can work in parallel.

### CHAT A - Testnet Core And Release Gate

Owned paths:

- `.github/workflows/*`;
- `cmd/l1d/cmd`;
- `scripts/localnet/*`;
- `scripts/testnet/*`;
- `scripts/release/*`;
- `app/genesis*`;
- `app/upgrades*`;
- `app/params/testnet_readiness*`;
- top-level release/testnet docs.

Do not touch:

- AVM instruction implementation except build/CLI integration;
- PoS accounting internals except smoke-test integration.

### CHAT B - PoS, Nominator Pool, Slashing, Validator Reputation

Owned paths:

- `x/nominator-pool`;
- `x/validator-registry`;
- `x/validator-election`;
- `x/validator-insurance`;
- `x/aetra-validator-score`;
- `x/aetra-staking-policy`;
- staking/slashing docs and e2e smoke tests.

Do not touch:

- AVM runtime implementation;
- release workflow except adding job commands agreed with CHAT A.

### CHAT C - AVM Runtime V1

Owned paths:

- `x/aetravm/avm`;
- `x/aetravm/async`;
- `x/aetravm/messageabi`;
- `x/aetravm/standards`;
- `x/contracts`;
- AVM examples and contract docs.

Do not touch:

- PoS allocation math;
- validator registry/election internals;
- app-level release workflow except adding AVM tests commands agreed with CHAT A.

### CHAT D - Infrastructure And Operator Docs

Owned paths:

- `docs/VALIDATOR.md`;
- `docs/TESTNET.md`;
- `docs/COSMOVISOR.md`;
- `docs/AVM.md`;
- `docs/HEALTH.md`;
- Docker files;
- release artifact documentation;
- peer/seed list publication templates.

Do not touch:

- consensus state internals;
- AVM opcode semantics;
- PoS accounting internals.

### CHAT E - Noise Reduction And Launch Scope

Owned paths:

- stale docs mentioning native DEX/token/NFT;
- module-boundary docs;
- app module wiring tests for "no native app asset modules";
- future standards docs under `x/aetravm/standards`.

Do not touch:

- active PoS/AVM implementation unless removing stale references requires tests.

## Phase 0 - Launch Scope Freeze

### Task 0.1 - Define Testnet Kernel

Implementation:

- Create a short `docs/TESTNET.md` that states the testnet kernel:
  - Cosmos SDK + CometBFT node;
  - native bank balance layer;
  - native account/auth/freeze/rent only where already wired;
  - pool-based staking;
  - AVM contracts;
  - no native token/NFT/DEX app modules for application assets.
- Add a machine-checkable launch scope test.
- Make the release workflow fail if docs teach normal users to stake by choosing
  validators directly.

Tests:

- Static doc test: no user-facing `aevaloper` / `aevalcons`.
- Static doc test: no native DEX/token/NFT launch instructions.
- Static doc test: user staking examples mention official pool deposit.

Done:

- Everyone can read one page and know what the testnet actually launches.

### Task 0.2 - Remove Or Quarantine Prototype Noise

Implementation:

- Identify modules/docs that are prototype-only or future-only.
- Move future concepts into `docs/future/` or mark them as AVM standards, not
  launch-critical modules.
- Ensure app wiring does not include native DEX/token/NFT modules.
- Keep `x/aetravm/standards/aft`, `anft`, `adex` as future AVM standards only
  unless they are runnable as contracts.

Tests:

- App module wiring test rejects `tokenfactory`, `dex`, `nft`, `market` native
  asset modules in launch profile.
- Release docs test rejects launch instructions for native application-asset
  modules.

Done:

- Testnet scope is small and not confused by prototype-era docs.

### Task 0.3 - Launch Module Inventory

Implementation:

- Create a machine-readable launch inventory that classifies every `x/*` module:
  - `launch_core`;
  - `launch_support`;
  - `future_avm_standard`;
  - `prototype_only`;
  - `disabled`.
- For every module in app wiring, record:
  - why it is needed for public testnet;
  - whether it owns consensus state;
  - whether it has KV-backed runtime mutations;
  - export/import status;
  - invariant status;
  - block lifecycle scanning risk.
- Public testnet profile must fail if a `prototype_only` or `disabled` module is
  wired into launch state.
- Future token/NFT/DEX/market modules must be either AVM standards/contracts or
  disabled from launch profile.

Tests:

- inventory covers every `x/*` directory;
- app wiring modules are all listed in inventory;
- launch profile rejects native DEX/token/NFT app asset modules;
- launch profile rejects memory-only consensus keepers;
- docs generated from inventory match module-boundary docs.

Done:

- The launch scope is enforceable by CI instead of tribal knowledge.

## Phase 1 - Runnable Testnet Core

### Task 1.1 - One Stable Binary

Implementation:

- Build one binary: `aetrad` / `aetrad.exe`.
- Ensure `aetrad version --long --output json` includes:
  - app name;
  - version;
  - git commit;
  - build date;
  - dirty flag;
  - Cosmos SDK version;
  - CometBFT version;
  - AVM version.
- Add release ldflags for version metadata.
- Release artifact must include the binary and checksums.

Tests:

- CLI test parses version JSON.
- Release script test verifies binary checksum file.
- CI job runs built binary version command.

Done:

- Validator can download/build one binary and verify what it is.

### Task 1.2 - Deterministic Genesis

Implementation:

- Make localnet genesis generation deterministic within one run across all nodes.
- Chain-id must pass `ValidateAetraTestnetChainID`.
- Genesis must reject:
  - malformed chain-id;
  - secrets;
  - wrong denom;
  - missing staking params;
  - invalid module params;
  - validator count outside local/testnet profile rules.
- Genesis validation command:
  - `aetrad genesis validate-genesis <path>`;
  - `scripts/localnet/validate-genesis.ps1`.

Tests:

- Genesis validate CI job.
- Golden test for default app genesis shape.
- Localnet script test: all node genesis files have identical hash.
- Negative tests: malformed chain-id, secret-like field, wrong denom.

Done:

- A launch operator can validate published genesis before starting.

### Task 1.3 - 4-5 Node Localnet

Implementation:

- Add canonical 4-node and 5-node localnet profiles.
- Keep 3-node smoke for fast CI; 4/5-node is launch rehearsal.
- Scripts:
  - init;
  - validate genesis;
  - start;
  - wait for height;
  - query validator set;
  - stop;
  - collect diagnostics.
- No secret material in diagnostics artifacts.

Tests:

- CI fast: 3-node smoke.
- Manual/release gate: 4-node and 5-node localnet smoke.
- Negative: occupied ports, missing binary, invalid validator count.

Done:

- Local testnet can be booted repeatedly with clear commands.

### Task 1.4 - Export/Import Roundtrip

Implementation:

- Export state after blocks and committed txs.
- Validate exported genesis.
- Import into fresh home.
- Restart from imported state.
- Verify app hash/critical state consistency where feasible.

Tests:

- `tests/e2e/export_import_smoke.ps1`.
- App-level export/import restart test.
- Negative: corrupted exported state is rejected.

Done:

- Runtime mutations survive restart/export/import.

### Task 1.5 - Upgrade Rehearsal

Implementation:

- Add a canonical no-op upgrade handler for rehearsal.
- Add upgrade dry-run:
  - pre-upgrade state;
  - version map;
  - handler execution;
  - post-upgrade export validation.
- Document Cosmovisor upgrade path.

Tests:

- `go test ./app/upgrades ./app -run Upgrade`.
- Localnet upgrade smoke:
  - schedule upgrade;
  - stop at height;
  - swap binary or no-op handler;
  - restart;
  - export validate.

Done:

- Testnet can rehearse a coordinated upgrade before public launch.

## Phase 2 - Validator-Grade Infrastructure

### Task 2.1 - Canonical Operator Docs

Create:

- `docs/VALIDATOR.md`;
- `docs/TESTNET.md`;
- `docs/COSMOVISOR.md`;
- `docs/HEALTH.md`;
- `docs/AVM.md`.

Current audit:

- Top-level `docs/VALIDATOR.md`, `docs/TESTNET.md`, and
  `docs/COSMOVISOR.md` are missing and must be created as canonical docs, even
  if older partial docs such as `validator-onboarding.md` or
  `public-testnet-preparation.md` exist.

`docs/VALIDATOR.md` must cover:

- hardware;
- OS;
- build/download binary;
- version verification;
- chain-id;
- genesis validation;
- keyring;
- validator key safety;
- state sync;
- snapshots;
- create validator;
- monitor;
- restart;
- upgrade;
- incident response.

`docs/TESTNET.md` must cover:

- chain-id;
- genesis URL/checksum placeholder;
- seed nodes;
- persistent peers;
- RPC endpoints;
- faucet path if enabled;
- minimum fees;
- expected block time;
- launch profile;
- known non-goals.

`docs/COSMOVISOR.md` must cover:

- install;
- directory layout;
- current binary;
- upgrades directory;
- environment;
- upgrade handler naming;
- rollback policy.

Tests:

- Static doc coverage test for all required sections.
- Static docs test rejects normal user direct-delegation examples.
- Static docs test requires the `10 AET` pool minimum in user staking docs.
- Release package includes all docs.

Done:

- A validator can join without reading source code.

### Task 2.2 - Docker Image

Implementation:

- Add Dockerfile for `aetrad`.
- Add minimal runtime image.
- Add non-root user.
- Add healthcheck command.
- Add build args for version/commit.
- Add docker-compose localnet sample only if it does not replace existing
  PowerShell localnet scripts.

Current audit:

- Top-level `Dockerfile` is missing. Public testnet release is incomplete until
  a reproducible image can be built and version-checked.

Tests:

- Docker build CI job.
- Container `aetrad version --long --output json` passes.
- Healthcheck command returns healthy against a local node.

Done:

- Release can publish a validator-ready image.

### Task 2.3 - Health Checks And Peer Lists

Implementation:

- Define health endpoints/commands:
  - process alive;
  - RPC status;
  - latest height increasing;
  - catching_up false;
  - peer count;
  - validator signing info;
  - app invariant command/test.
- Add `docs/HEALTH.md`.
- Add seed/persistent peer list templates:
  - `docs/testnet/peers.example.json`;
  - `docs/testnet/seeds.example.txt`.

Tests:

- Localnet health script validates 3-node/5-node profile.
- Peer list parser rejects malformed node IDs/endpoints.

Done:

- Operators can monitor and join peers with published data.

### Task 2.4 - Release Workflow

Implementation:

- Make testnet release workflow run:
  - `go test ./...`;
  - `go vet ./...`;
  - `buf lint`;
  - genesis validate;
  - localnet smoke;
  - export/import smoke;
  - invariants;
  - release artifact build;
  - binary version command;
  - Docker build.
- Upload:
  - binary;
  - checksums;
  - docs;
  - readiness report;
  - localnet diagnostics on failure.

Tests:

- Workflow static test contains all required jobs.
- Release package script test verifies required docs and checksums.

Done:

- A release candidate cannot be published without the runnable gates.

## Phase 3 - PoS V1 Completion

### Task 3.1 - Pool-Based User Staking

Implementation:

- Normal user staking path:
  - User `AE...`;
  - deposit amount `>= 10 AET`;
  - official liquid staking contract/pool;
  - pool shares/receipt;
  - allocation to validators.
- Normal CLI/API must not ask user for a validator address and must not accept a
  hidden validator override in the default staking command.
- User-facing deposit message shape:

```text
MsgDepositToStakingPool {
  depositor AE...
  pool_id or official_contract AE...
  amount >= 10 AET
}
```

- The normal deposit message must not contain `validator`, `operator_address`,
  `consensus_address`, `aevaloper`, `aevalcons`, or raw target address fields.
- Direct SDK `MsgDelegate` must be:
  - disabled for normal user path;
  - rejected by app-level invariant if enabled in testnet profile; and
  - explicitly guarded as operator-only/advanced governance-enabled path if it
    exists outside public testnet.
- Pool deposits must support small users at `10 AET` and above.
- The pool, not the user, selects validators through deterministic allocation
  weights.

Tests:

- User deposits into pool without validator address.
- 10 AET deposit succeeds from active `AE...` wallet.
- Deposit below 10 AET rejected.
- Deposit message containing a validator address is rejected by validation.
- Direct user delegation rejected before staking mutation.
- Pool shares deterministic.
- Export/import preserves shares and pool totals.
- CLI help snapshot does not mention validator selection for normal staking.
- Docs static test rejects user-facing validator-choice examples.

Done:

- User staking UX is pool/index-based and a wallet user can start staking from
  10 AET without choosing a validator.

### Task 3.2 - Nominator Pool Accounting

Implementation:

- Pool state must live in KV/prefix records, not runtime-only memory.
- Mutating methods read/write KV.
- Export reads KV and emits deterministic sorted state.
- Import writes prefix records.
- No full scans in block lifecycle.
- Pool allocations update only touched keys.

Tests:

- mutate -> export -> import -> query same state.
- pagination bounded.
- deterministic order.
- storage rent debt preserved.

Done:

- Pool state survives restart and scales past prototype size.

### Task 3.3 - Reward Policy V1

Implementation:

- Define simple reward policy:
  - fees/inflation source;
  - pool share accounting;
  - validator commission;
  - lazy reward index;
  - rounding rule;
  - cap by collected/emitted rewards.
- Rewards distributed by pool shares, not manual validator choice.
- Reputation cannot increase without stake-time exposure.

Tests:

- reward cap invariant;
- deterministic reward index;
- claim idempotency;
- export/import after rewards;
- jailed/slashed validator does not produce positive bonus.

Done:

- Rewards are understandable and not open-ended.

### Task 3.4 - Slashing V1

Implementation:

- Ensure slashing params are genesis/governance params.
- Wire downtime/double-sign policy to validator status and pool exposure.
- Pool users inherit validator slashing exposure through allocation accounting.
- Slashed state export/import stable.

Tests:

- downtime slash fixture;
- double-sign/tombstone fixture if available;
- pool allocation principal decreases or records exposure;
- cannot recover slashed stake through export/import/migration.

Done:

- Validators and pool participants have deterministic slashing risk.

### Task 3.5 - Validator Score/Reputation V1

Implementation:

- Minimal deterministic score:
  - uptime;
  - missed blocks;
  - commission;
  - slashing risk;
  - stake efficiency;
  - pool allocation limit;
  - reputation accumulator.
- Score output drives allocation engine weights.
- No nondeterministic inputs.

Tests:

- same input -> same weights;
- score changes with uptime/commission/slashing;
- inactive/ineligible validators rejected;
- export/import preserves scores and snapshots.

Done:

- Allocation engine has a transparent v1 score.

## Phase 4 - AVM V1 Completion

### AVM Direction

Aetra VM should become a deterministic, stack-based, message-driven VM using
immutable content-addressed Chunks as the persistent data model.

Important constraint:

- Do not copy TON slice/builder exactly.
- Do not expose raw byte fiddling as the main developer experience.
- Prefer typed Reader/Writer/Codec over manual bit parsing.

### Task 4.1 - Small Stable Exit Codes

Current status:

- Small exit codes exist in `x/contracts/types/exit_codes.go`.

Implementation:

- Keep exit codes under `100` for core VM/contract errors.
- Define one canonical list:
  - ok;
  - validation failed;
  - unauthorized;
  - inactive/frozen;
  - code rejected;
  - out of gas;
  - storage limit;
  - storage rent debt;
  - message expired;
  - queue limit;
  - execution failed;
  - internal bounce;
  - forbidden host call;
  - contract abort.
- Map AVM runtime exit codes to contract receipt exit codes.
- Add `ExitCodeName` coverage for every runtime code.

Tests:

- golden exit code list;
- all codes `< 100`;
- unknown returns `unknown`;
- receipt stores code and name/proof metadata.

Done:

- Operators and contract developers can understand failures.

### Task 4.2 - Chunk Core

Implementation:

Define AVM Chunk:

```text
Chunk {
  data_bits <= 2048
  refs <= 8
  type_hash optional
  hash = BLAKE3(canonical_chunk_encoding)
}
```

Rules:

- immutable;
- content-addressed;
- DAG only;
- cycles rejected;
- identical chunk -> identical hash;
- hash is stable across export/import;
- refs are ordered and bounded.

Tests:

- hash golden vectors;
- data over 2048 bits rejected;
- refs over 8 rejected;
- cycle rejected;
- canonical encoding stable;
- export/import preserves hash/root.

Done:

- AVM state can be represented as content-addressed immutable chunks.

### Task 4.3 - Type System And Codec

Runtime primitive types:

- `bool`;
- `uint8`, `uint16`, `uint32`, `uint64`, `uint128`, `uint256`;
- `int8`, `int16`, `int32`, `int64`, `int128`, `int256`;
- `address`;
- `hash`;
- `coins = uint128`;
- `timestamp = uint64`;
- `null`;
- `tuple`;
- `chunk`;
- `execution_frame`.

Compile-time/developer types:

- `struct`;
- `Option<T>`;
- `Tuple<T...>`;
- `Map<K,V>` compiled to ChunkMap;
- bounded UTF-8 `string`;
- bytes.

Implementation:

- VM stores universal values, not runtime generics.
- Add optional `type_hash = BLAKE3(schema_descriptor)`.
- Add `Reader<T>` as read cursor over Chunk data.
- Add `Writer<T>` as immutable Chunk constructor.
- Add `Codec<T>` descriptors:
  - canonical schema string;
  - encode;
  - decode;
  - gas cost;
  - max encoded size.
- Strings:
  - UTF-8;
  - byte-length bounded;
  - encoded as length-prefixed bytes;
  - no unbounded string concatenation.

Tests:

- primitive encoding golden vectors;
- string UTF-8 valid/invalid;
- Option null/value encoding;
- type_hash stable;
- invalid decode reverts;
- same typed value -> same chunk hash.

Done:

- Contracts can be written with short typed code, not manual byte parsing.

### Task 4.4 - ChunkMap

Do not implement EVM-style mutable mapping.

Implementation:

Define persistent Chunk Trie Map:

```text
ChunkMap {
  root: Chunk
  fanout: 8
  key_hash = BLAKE3(canonical_key)
  path = deterministic nibbles/buckets from key_hash
  leaf = value Chunk
}
```

Rules:

- immutable tree;
- lazy node creation;
- update copies only changed path;
- no global mutable hashmap;
- no full scan for lookup/update;
- deterministic iteration only through bounded proof/index APIs;
- parallel-friendly buckets.

Tests:

- put/get/delete;
- update changes only path root;
- same operations -> same root;
- different buckets can be proven independent;
- key collision handling;
- export/import preserves root;
- gas depends on depth and encoded bytes.

Done:

- Domains, token ownership, NFT ownership, balances inside contracts, and other
  contract maps use ChunkMap, not global storage slots.

### Task 4.5 - AVM Execution Model

Execution phases:

1. Storage Phase - load state Chunks.
2. Credit Phase - apply attached value.
3. Compute Phase - execute VM.
4. Action Phase - emit outgoing messages/events.
5. Finalization Phase - commit new Chunk roots.

Implementation:

- ExecutionFrame:
  - instruction pointer;
  - stack snapshot;
  - local context;
  - pending calls/messages;
  - error handler/abort state.
- Stack values are typed VM values.
- No classic RAM or linear mutable memory.
- All state updates produce new Chunks.
- Out-of-gas reverts state changes but records receipt.
- Per-message and per-block gas limits enforced.

Tests:

- deploy;
- execute external;
- execute internal;
- out-of-gas rollback;
- abort exit code;
- same input/state/code -> same root/gas/receipt;
- forbidden nondeterministic opcode rejected.

Done:

- AVM can execute deterministic contracts safely.

### Task 4.6 - Host Functions V1

Allowed:

- hash SHA256;
- hash BLAKE3;
- verify ed25519;
- parse/format Aetra address;
- read storage chunk;
- write storage chunk;
- delete storage chunk;
- emit event;
- send internal message;
- get block height;
- get chain id;
- get contract address;
- get caller/source;
- get attached value;
- abort with exit code.

Careful:

- Wall-clock time is forbidden.
- If `time.now()` is added, it must mean consensus block time from header, not
  local process clock.
- Randomness must not be process randomness. If `secure_random()` is added, it
  must be a deterministic/verifiable block entropy value, e.g.:

```text
random = BLAKE3(previous_state_root || block_entropy || message_hash || domain)
```

and only after the chain defines block entropy/proof rules.

Forbidden:

- filesystem;
- network;
- floating point;
- goroutines/threads;
- process/env;
- nondeterministic map iteration;
- local wall-clock;
- unverified randomness.

Tests:

- each allowed host has gas cost;
- unknown host rejected;
- forbidden host rejected;
- host storage respects Chunk/ChunkMap limits;
- send_internal respects queue limits;
- abort returns contract-defined small exit code.

Done:

- Host surface is deterministic and auditable.

### Task 4.7 - Get Methods

Implementation:

- Add contract get methods as read-only AVM entrypoints.
- Get method call:
  - loads code and state root;
  - executes with query gas limit;
  - cannot write storage;
  - cannot send internal messages;
  - cannot emit consensus events;
  - can return typed value bytes/Chunk;
  - can include proof metadata.
- CLI/API:
  - `aetrad query contracts get <contract> <method> <args-json>`;
  - gRPC query endpoint;
  - bounded response size.

Tests:

- get method reads state;
- attempted write rejected;
- attempted send message rejected;
- gas limit enforced;
- response deterministic;
- malformed args rejected;
- proof query stable.

Done:

- Contracts have practical read APIs without state mutation.

### Task 4.8 - Minimal Contract Examples

Examples must be real AVM examples, not native modules:

- counter contract;
- key/value ChunkMap contract;
- domain registry sample:
  - name string;
  - owner AE address;
  - resolver records map;
- token sample using AFT standard;
- NFT sample using ANFT standard;
- pool deposit adapter sample for official staking if applicable.

Tests:

- examples compile/encode;
- deploy;
- execute;
- get methods;
- export/import;
- receipts/events.

Done:

- A developer can write and run a minimal contract.

### Task 4.9 - Canonical AVM Module And Bytecode Verifier

Implementation:

- Define one canonical AVM module format:
  - magic/version;
  - ABI version;
  - code version selector;
  - import table;
  - export table;
  - metadata hash;
  - instruction stream;
  - optional schema/type descriptors;
  - dependency hashes.
- Verifier must reject:
  - unknown version;
  - missing required entrypoints;
  - forbidden imports;
  - unreachable invalid jump targets;
  - stack underflow/overflow;
  - instruction count over params;
  - bytecode over params;
  - nondeterministic opcodes;
  - malformed dependency hashes.
- Verification result must be deterministic and stored with code metadata.

Tests:

- bytecode golden vectors;
- malformed module rejected;
- forbidden host import rejected;
- same module hash across export/import;
- verifier cannot panic on random bytes/fuzz fixtures.

Done:

- A node can safely accept/reject AVM code before deployment.

### Task 4.10 - Production Instruction Set V1

Implementation:

- Freeze a small v1 instruction set:
  - stack: push/pop/dup/swap/drop;
  - arithmetic: checked add/sub/mul/div/mod for supported ints;
  - comparison: eq/neq/lt/lte/gt/gte;
  - boolean: and/or/not;
  - control: jump/jump_if/call/return/abort;
  - typed load/decode/encode through Reader/Writer/Codec;
  - chunk ops: new/read/ref/hash/type_hash;
  - ChunkMap ops: get/put/delete/proof;
  - message ops: read caller/source/value/op/query_id/body;
  - event/internal-message emit through host calls only.
- Every instruction has:
  - mnemonic;
  - opcode;
  - stack input/output contract;
  - gas rule;
  - overflow behavior;
  - exit code on failure.

Tests:

- opcode table golden test;
- stack contract tests for each opcode;
- integer overflow exits deterministically;
- invalid jump rejected by verifier;
- same program always same trace hash.

Done:

- AVM is specified as a real machine with stable execution semantics.

### Task 4.11 - AVM Value Model For Short Contracts

Implementation:

- Runtime value tags:
  - null;
  - bool;
  - signed/unsigned int widths;
  - coins;
  - timestamp;
  - address;
  - hash;
  - bytes;
  - string;
  - tuple;
  - chunk ref;
  - reader cursor;
  - writer handle;
  - execution frame.
- Developer/compiler layer may expose:
  - structs;
  - generics;
  - `T?`/Option;
  - `Map<K,V>`;
  - methods and traits.
- VM must not implement runtime reflection or unbounded generics.
- Value encoding must be canonical, size-bounded, and gas-metered.

Tests:

- value tag golden vectors;
- tuple nested encoding;
- string/domain name encoding;
- Option null/value;
- invalid type cast exits with small code;
- same value bytes produce same chunk hash.

Done:

- Contracts can be concise without making the VM nondeterministic or huge.

### Task 4.12 - StateInit And Counterfactual Deploy

Implementation:

- `StateInit` must include:
  - ABI version;
  - code hash/code id;
  - init data bytes;
  - salt bytes/string;
  - owner `AE...`;
  - optional library/dependency hashes;
  - initial storage root;
  - initial balance in `naet`;
  - flags/capabilities.
- Add:
  - canonical encoding;
  - `HashStateInit`;
  - max init data size;
  - max salt size;
  - max dependency count;
  - deterministic normalization.
- Contract address derivation:

```text
DeriveContractAddress(chain_id, namespace, deployer, code_hash, init_data_hash, salt)
```

- Output both:
  - raw/internal `4:...`;
  - user-facing `AE...`.
- Reject zero deployer and malformed code hash.

Tests:

- same StateInit -> same address;
- init data change -> address changes;
- salt change -> address changes;
- oversized init data rejected;
- duplicate address deploy rejected;
- counterfactual address query returns virtual/not_deployed before deploy;
- export/import preserves derived address.

Done:

- Contract deployment is deterministic and supports counterfactual workflows.

### Task 4.13 - Get Methods As First-Class Contract ABI

Implementation:

- Each contract can expose named get methods in ABI metadata.
- Get method call includes:
  - method name or selector;
  - typed args;
  - query gas limit;
  - optional proof request.
- Get method execution context:
  - read-only state root;
  - no writes;
  - no outgoing messages;
  - no consensus events;
  - no balance movement;
  - deterministic return bytes/chunk.
- CLI must print typed JSON when ABI is known and raw hex/chunk hash when ABI
  is unknown.

Tests:

- query by method name;
- query by selector;
- ABI decode success;
- unknown method rejected;
- write/send/event forbidden in get context;
- response hash stable across repeated query and export/import.

Done:

- Users can inspect contract state without custom off-chain parsers.

### Task 4.14 - Receipts, Events, Bounce, And Refund Accounting

Implementation:

- Every deploy/execute/get/migrate/internal message produces a receipt or query
  response record with:
  - small exit code;
  - gas used;
  - gas refunded if policy ever allows it;
  - storage fee/rent charged;
  - value in/out;
  - state root before/after;
  - outgoing messages;
  - events;
  - bounced/refunded flags;
  - proof hash.
- Bounce rules:
  - failed internal message with `bounce=true` creates one bounce;
  - bounced message sets `bounced=true` and `bounce=false`;
  - failed bounce cannot create another bounce loop;
  - double refund is impossible by state flag.

Tests:

- failed internal message creates bounce;
- non-bounceable message does not create bounce;
- failed bounce ends without loop;
- value accounting conserved;
- double refund rejected;
- export/import preserves bounced/refunded status.

Done:

- Operators can explain every value movement and contract failure.

### Task 4.15 - Contract Upgrade And Migration Model

Implementation:

- Contracts immutable by default.
- Upgradeable only if deployed with upgrade flag.
- Admin-controlled upgrade only by current admin.
- System/governance upgrade only for explicitly system-owned contracts.
- Code hash change requires migration handler unless state root is empty.
- Contract state stores schema version.
- Migration transforms state deterministically and emits migration receipt.
- Failed migration rolls back.

Messages:

- `MsgUpgradeContractCode`;
- `MsgMigrateContractState`;
- `MsgSetContractAdmin`;
- `MsgDisableContractUpgrades`.

Tests:

- immutable contract cannot upgrade;
- admin upgrade works only if allowed;
- non-admin rejected;
- migration changes schema version;
- failed migration rolls back;
- system upgrade requires governance/system authority.

Done:

- Upgradeability is explicit and safe enough for testnet system contracts.

### Task 4.16 - AVM Developer Surface And CI Examples

Implementation:

- Add minimal contract authoring format for v1:
  - either tiny assembly format;
  - or JSON/module builder;
  - or first compiler stub if available.
- Add examples under one canonical path:
  - counter;
  - domain registry with bounded string names;
  - token-like ledger using ChunkMap;
  - internal message sender/receiver;
  - bounce/refund example;
  - get-method example.
- Add CI command that builds/encodes examples, deploys them in localnet or app
  simulator, executes happy/reject paths, and checks receipts.

Tests:

- examples build in CI;
- examples deploy;
- examples execute external/internal;
- examples query get methods;
- examples export/import and continue executing.

Done:

- AVM is usable by a developer, not only by Go unit tests.

### Task 4.17 - AVM Security And Determinism Gate

Implementation:

- Add a dedicated AVM determinism test suite:
  - repeated execution;
  - shuffled map input normalization;
  - export/import/re-execute;
  - randomized bytecode verifier fuzz;
  - forbidden host call attempts;
  - gas overflow attempts;
  - storage limit attempts.
- Add static checks that forbid:
  - wall-clock/process time in runtime;
  - process randomness in runtime;
  - filesystem/network in host functions;
  - goroutines/threads inside execution;
  - floating point arithmetic in consensus execution.

Tests:

- `go test ./x/aetravm/...`;
- static determinism scanner;
- fuzz corpus does not panic verifier;
- gas overflow returns deterministic error.

Done:

- AVM can be defended as consensus-safe for public testnet.

## Phase 5 - Contract Standards Instead Of Native App Assets

### Task 5.1 - Move Token/NFT/DEX To AVM Standards

Implementation:

- Keep standards:
  - AFT for fungible tokens;
  - ANFT for NFTs/SBTs;
  - ADEX as future DEX standard.
- Do not wire native app modules for these.
- Launch docs should say:
  - token/NFT/DEX are contract standards;
  - DEX is not required for initial public testnet unless AVM contract example is
    ready.

Tests:

- App wiring rejects native asset modules.
- Docs do not teach native tokenfactory/DEX launch flows.
- Standards registry has deterministic descriptors.

Done:

- Asset story is clear and not split across native and AVM paths.

### Task 5.2 - Domain Registry As AVM Or Minimal Identity Registry

Implementation options:

Option A for testnet:

- minimal native identity registry if already wired and bounded;
- domain ownership separate from wallet account state;
- resolver records not stored in account state.

Option B preferred for AVM maturity:

- domain registry contract backed by ChunkMap.

Rules:

- owner is `AE...`;
- name is bounded UTF-8 string;
- records are bounded ChunkMap entries;
- storage rent applies.

Tests:

- register;
- transfer;
- set resolver;
- get owner;
- export/import;
- storage rent accounting.

Done:

- Domain feature does not pollute wallet state.

## Phase 6 - App Wiring And Invariants

### Task 6.1 - Runtime Invariants

Implementation:

- Register global app invariants for:
  - bank/module accounting;
  - rewards cap;
  - rent reserve/runway;
  - pool shares;
  - validator entry;
  - direct delegation rejection;
  - AVM receipt/queue/state root;
  - no native app asset modules.
- Invariants must read live keeper state, not only model fixtures or placeholder
  input values.
- Existing placeholder-style inputs such as max integer budgets, empty activation
  attempts, or hardcoded "export/import stable = true" must be replaced by
  state-derived evidence or removed from launch gates.
- Each invariant must be runnable:
  - in Go tests;
  - from CI;
  - against a localnet/exported app state where feasible.
- Add CLI/test command or CI target that runs them against app state.

Tests:

- registry includes every required invariant;
- each invariant has a failing fixture;
- default app state passes;
- post-core-flow app state passes.
- invariant failures include deterministic IDs and messages.
- direct delegation invariant fails if either direct delegation param is true.
- validator-set invariant fails if CometBFT/SDK staking set diverges from Aetra
  registry/election output.

Done:

- Testnet release can prove app-level invariant coverage.

### Task 6.2 - No Full Scans In Block Lifecycle

Implementation:

- Audit BeginBlock/EndBlock/PreBlock.
- Replace all-account/all-contract scans with:
  - bounded queues;
  - touched-key updates;
  - epoch snapshots;
  - explicit indexes.

Tests:

- static scan test for risky iteration paths;
- block lifecycle bounded-iteration test with large fixture;
- no query pagination bypass.

Done:

- Chain can scale beyond toy localnet.

## Phase 7 - Final Public Testnet Gate

### Task 7.1 - Testnet Release Candidate Checklist

A release candidate is not ready until all pass:

- `go test ./...`;
- `go vet ./...`;
- `buf lint`;
- binary build;
- version command;
- genesis validate;
- 4-node localnet;
- 5-node localnet;
- export/import restart;
- upgrade rehearsal;
- app invariants;
- AVM deploy/execute/get method smoke;
- pool staking smoke;
- slashing/reputation v1 smoke;
- release package build;
- Docker image build;
- validator docs coverage.

Done:

- The release artifact has enough evidence for a public testnet launch decision.

### Task 7.2 - Public Testnet Launch Artifacts

Create release artifacts:

- binary archives for Linux amd64/arm64 and Windows dev tooling;
- Docker image;
- checksums;
- genesis JSON;
- genesis checksum;
- seed nodes list;
- persistent peers list;
- RPC endpoint list;
- faucet instructions if enabled;
- validator guide;
- Cosmovisor guide;
- health guide;
- incident response guide.

Done:

- External validators can join from published artifacts only.

## Recommended Execution Order

1. CHAT E: create launch module inventory and quarantine prototype-only modules.
2. CHAT A: stabilize binary/version/genesis/localnet/export-import CI.
3. CHAT D: create canonical `docs/VALIDATOR.md`, `docs/TESTNET.md`,
   `docs/COSMOVISOR.md`, Dockerfile, health docs.
4. CHAT B: finish 10 AET pool-based staking, direct delegation rejection, real
   consensus validator-set alignment, reward policy, slashing, validator score
   v1.
5. CHAT C: finish AVM module format/verifier, exit-code mapping, instruction
   set, Chunk core, typed Codec, ChunkMap, StateInit, get methods,
   receipts/bounce/refund, upgrades, examples, determinism gate.
6. CHAT E: remove stale native DEX/token/NFT launch docs and enforce future AVM
   standards boundary.
7. CHAT A: final app wiring, live invariants, 4/5-node localnet, upgrade rehearsal, release
   workflow.
8. All chats: run full testnet release candidate checklist.

## Definition Of Done For V2

Aetra is V2 testnet-ready only when:

- a clean checkout can build one `aetrad` binary;
- genesis can be generated, validated, and published;
- 4-5 node localnet reaches height and has stable validators;
- export/import restart works;
- upgrade rehearsal works;
- normal staking uses the pool path with a `10 AET` minimum deposit;
- normal staking messages, CLI, API, and docs do not accept validator selection;
- direct user validator delegation is disabled in public testnet profile and
  rejected before staking mutation;
- the actual CometBFT/SDK validator set is controlled by or proven identical to
  the Aetra registry/election output;
- reward/slashing/reputation v1 are understandable and tested;
- AVM can deploy, execute external/internal messages, charge gas, store state,
  emit receipts/events, return small exit codes, and run get methods;
- AVM has canonical module encoding, verifier, instruction set, StateInit,
  typed value/Codec model, Chunk/ChunkMap storage, get methods, receipts,
  bounce/refund, upgrade/migration policy, and CI contract examples;
- AVM persistent data uses Chunk/ChunkMap design before public contract
  developer use;
- token/NFT/DEX are AVM standards/contracts, not native app asset modules;
- validator docs, testnet docs, Cosmovisor docs, Docker image, health checks,
  seed/peer templates, and release artifacts exist;
- app invariants are live keeper checks, not placeholder/model-only checks;
- CI gates enforce the above.
