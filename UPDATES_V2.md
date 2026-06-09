Ты работаешь в репозитории Aetra Blockchain: C:\Users\Ryzen\Desktop\L1.

Главный backlog для текущей работы: UPDATES_V2.md.
Этот файл локальный и не должен попадать в git. Сначала изучи нужные разделы UPDATES_V2.md и существующую архитектуру repo, потом реализуй только те задачи, которые я явно скину в этом чате.

Работаем быстро, но не в ущерб корректности:
- не зависай на долгом планировании;
- сначала проверь git status;
- не трогай unrelated dirty files;
- не делай широкие рефакторы ради красоты;
- следуй существующим package/module patterns;
- делай изменения маленькими, понятными и проверяемыми;
- все важные behavior/reject/security/determinism cases покрывай тестами;
- после Go-изменений запускай gofmt;
- сначала targeted tests для затронутых пакетов;
- потом по возможности go test ./...;


Критичные правила проекта:
- user-facing address всегда AE...;
- raw/internal/proof address всегда 4:...;
- не использовать aevaloper/aevalcons в user-facing API;
- private key и seed phrase никогда не хранить on-chain;
- balances остаются в bank/native balance layer;
- tokens/NFT/DEX остаются AVM smart contracts/standards, не native x/ asset modules и вообще их реализовывать не надо; 
- ordinary user staking только через official pool/index flow;
- пользователь не выбирает валидатора;
- direct user delegation to validators disabled;
- storage rent применяется к active persistent state;
- protocol/system state не должен freeze/delete из-за user rent path;
- frozen wallet/contract recoverable через top-up + pay debt + unfreeze;
- все параметры fees/rent/commissions/validator limits/emissions/governance-controlled, не hardcoded навсегда.

Особые backlog-направления:
- dynamic deterministic fee market: средний transfer target 0.01 AET = 10_000_000 naet, но fee не фиксированная навсегда;
- fee зависит от gas, tx bytes, message count, deterministic congestion/load, reputation premium/discount, storage/rent side effects;
- live native economy: fee collector, burn, validator rewards, treasury/community, emissions, storage/system reserve accounting;
- reputation реально влияет на bounded fee/priority/allocation, но не блокирует базовые права;
- ordinary contracts например token/NFT/DEX/domain contracts не имеют persistent reputation; 
- Core Zones/layers должны быть runtime-gated, keeper-backed, deterministic, export/import stable;
- sharding only behind explicit governance/genesis feature gate;
- raw 4: address bug: ordinary user/validator/contract raw address must be 256-bit high-entropy-looking, not 20-byte address padded with 12 zero bytes. Do not quick-fix address derivation without explicit version/migration gate.

Git rules:
- UPDATES_V2.md не добавлять в git;
- не пушить просто так;
- commit/push делать только если я явно попрошу;
- если push requested: пушить только в main через `git push origin main`;
- перед commit staging должен быть точечным, только файлы текущей задачи;
- не revert чужие изменения без явного запроса.

Definition of done для каждой задачи:
- implementation complete;
- tests cover main behavior, reject/security cases, determinism where relevant;
- export/import or invariants covered if stateful;
- no full scans where bounded/lazy behavior required;
- gofmt run for changed Go files;
- targeted tests run and reported;
- repo state summarized;
- no unrelated dirty files touched.

жди задания

<!-- # Aetra Testnet Completion Backlog V2

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
- A user identity has one unified reputation score. Staking, transactions,
  contract interactions, lifecycle behavior, and configured system signals feed
  that same identity score; they do not create separate wallet reputation
  states.
- Low identity reputation must not block basic transactions, token creation, NFT
  creation, smart contract deployment, contract execution, or normal wallet
  rights. Reputation is a soft weighting/QoS signal, not a permission gate.
- `AE...` remains the only user-facing address format.
- `4:...` remains raw/internal/proof format.
- Cosmos wallet compatibility must preserve the standard signer/key model:
  BIP-39 seed, BIP-32/BIP-44 derivation, secp256k1, SignDoc-based signing, and
  normal transaction broadcasting. Aetra adds account runtime metadata on top;
  it does not store secrets or change the signing model.
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
  single canonical operator path. -->
<!-- 
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
- Export/import after 10 AET deposit returns the same pool share and pool totals. -->
<!-- 
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
  changes it with a deterministic migration record. -->
<!-- 
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
- Protocol-critical module action still executes under system rent stress. -->
<!-- 
### Blocker 3A - Deterministic Dynamic Fee Market

Current risk:

- A forever-fixed transfer fee makes spam pricing, congestion response, storage
  side effects, and reputation weighting unrealistic for public testnet.

Requirement:

- Normal transfer fees must be deterministic but not permanently fixed:

```text
transfer_fee_naet =
  max(min_tx_fee_naet, base_transfer_fee_naet)
  + gas_used * current_base_fee_per_gas_naet
  + tx_size_bytes * byte_fee_naet
  + message_count * message_fee_naet
  + bounded_congestion_surcharge_naet
  + low_reputation_premium_naet
  + storage_rent_side_effects_naet
  - bounded_reputation_discount_naet
```

- The neutral long-run target for a normal wallet transfer is `0.01 AET`,
  expressed as `10_000_000 naet`, but this is a target/anchor, not a forever
  fixed fee. Individual deterministic transfer fees may be slightly lower or
  higher because of gas, tx bytes, message count, congestion, reputation, and
  storage/rent side effects.
- `current_base_fee_per_gas_naet` is derived from deterministic block or epoch
  utilization, never wall-clock time, local mempool contents, randomness, or map
  iteration order.
- Low identity reputation may increase fees only through a small bounded
  premium or reduce priority; it must not block basic transfers, staking, token
  creation, NFT creation, contract deployment, or contract execution.
- High identity reputation may receive only a small bounded discount or priority
  improvement; it must never make protocol fees zero.
- Contract calls and account actions that create, grow, or touch persistent
  state must add their deterministic storage rent/debt side effects to the
  required fee budget.
- All fee inputs and caps are governance/genesis params:
  - `min_tx_fee_naet`;
  - `base_transfer_fee_naet`;
  - `target_transfer_fee_naet` defaulting to `10_000_000 naet` (`0.01 AET`);
  - `base_fee_per_gas_naet`;
  - `byte_fee_naet`;
  - `message_fee_naet`;
  - `congestion_surcharge_cap_naet`;
  - `low_reputation_premium_cap_naet`;
  - `reputation_discount_cap_naet`;
  - storage rent fee caps and recovery thresholds.
- Fees are paid only in native `naet`.
- Fee calculation is export/import stable.

Acceptance:

- Golden transfer fee tests cover low, medium, and high load fixtures such as
  `1024`, `1045`, and `1130 naet` when params intentionally set that scale.
- Public testnet genesis uses the configured `target_transfer_fee_naet` of
  `10_000_000 naet` (`0.01 AET`) as the neutral transfer fee anchor, with
  `min_tx_fee_naet` as the lower safety floor.
- `1000 naet` is only a tiny math fixture unless governance explicitly lowers
  the floor and target for a test profile.
- Low-reputation transfer fixture pays more than the neutral account for the
  same tx, within the configured premium cap.
- High-reputation transfer fixture may pay less than the neutral account but
  never below `min_tx_fee_naet` and never zero.
- Contract execution fixture includes gas, bytes, congestion, reputation, and
  storage rent/debt side effects in the required fee.
- Repeated runs with the same state produce the exact same required fee.
- Export/import before and after fee-market state changes preserves fee
  calculation results exactly. -->
<!-- 
### Blocker 3B - Live Native Economy Wiring

Current risk:

- Fees, emissions, validator rewards, burn, treasury, and storage rent can exist
  as separate helpers while the runnable chain still behaves like a flat-fee
  prototype. Public testnet needs one connected native economy path.

Requirement:

- Wire a real native `naet` economy loop:

```text
tx execution
  -> deterministic required fee
  -> fee collector module account
  -> burn share
  -> validator reward pool
  -> treasury/community pool
  -> storage rent/system reserve accounting

epoch/block emissions
  -> bounded native AET mint policy
  -> validator reward pool
  -> pool-user rewards after validator commission and pool protocol fee
  -> treasury/community allocations
```

- All native economic accounting is in `naet`.
- Native emission is governance/genesis controlled and capped:
  - no hardcoded forever inflation;
  - no unauthorized mint path;
  - no minting of AVM token/NFT/DEX assets through native modules.
- Fee collector must be the single canonical entry point for collected tx fees:
  - fees collected in ante/tx execution;
  - split by deterministic params;
  - burned amount reduces total supply;
  - validator reward amount is credited to validator reward accounting;
  - treasury amount is credited to treasury/community module account;
  - protection/storage/system reserve amount is credited separately if enabled.
- Validator rewards combine:
  - fee share;
  - configured emissions;
  - performance/allocation weights;
  - commission accounting;
  - slashing/jailing penalties.
- Pool-user rewards receive delegator/index yield only:
  - after validator commission;
  - after pool protocol fee;
  - after slashing losses;
  - never full validator/operator economics.
- Storage rent and contract storage side effects must be charged or reserved as
  part of the same economic accounting flow, without freezing protocol-critical
  state through a user rent path.
- All splits use deterministic integer math with explicit remainder policy.
- All economic params are governance/genesis params:
  - fee burn bps;
  - validator reward bps;
  - treasury/community bps;
  - protection/system reserve bps;
  - emission rate/cap/schedule;
  - validator commission bounds;
  - pool protocol fee;
  - storage rent rates and caps.
- Economic state is export/import stable and invariant checked.

Acceptance:

- Transfer fee is collected into fee collector and split deterministically.
- Burned fee share reduces native `naet` total supply exactly.
- Validator reward share is credited to validator reward accounting.
- Treasury/community share is credited to the configured module account.
- Emission epoch mints only the allowed native `naet` amount and credits the
  reward pool according to params.
- Unauthorized mint, wrong denom fee, mixed denom fee, and fee bypass attempts
  are rejected before state mutation.
- Validator commission is credited separately from pool-user rewards.
- Pool-user reward distribution uses the lazy pool reward index and cannot
  exceed allocated emissions plus fee rewards.
- Storage rent side effects are charged/accounted in the same tx/epoch economic
  report where applicable.
- Remainders are handled by one documented deterministic policy.
- Bank supply invariant, module account invariant, fee accounting invariant,
  emission cap invariant, burn invariant, treasury invariant, and export/import
  invariant pass after fee collection and reward distribution.
- Golden economy test covers:
  - fee collection;
  - burn;
  - validator rewards;
  - treasury;
  - emissions;
  - pool-user rewards;
  - storage rent side effects;
  - exact supply before/after. -->
<!-- 
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
  - AVM scheduler. -->
<!-- 
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

### CHAT F - Native Account And Wallet Compatibility

Owned paths:

- `x/native-account`;
- `app/addressing`;
- account activation/auth policy docs;
- wallet compatibility docs;
- account genesis/export-import tests;
- account/auth/storage-rent integration tests through stable interfaces.

Do not touch:

- PoS allocation/reward math;
- AVM opcode semantics;
- release workflow except adding agreed wallet compatibility test commands;
- address derivation rules.

Temporary integration boundary:

- If AVM or PoS needs account status/auth checks before APIs are stable, define
  a minimal interface in the consuming package and mark it as an AWCE-1
  temporary integration boundary. -->
<!-- 
## Phase 0 - Launch Scope Freeze

### Task 0.1 - Define Testnet Kernel

Implementation:

- Create a short `docs/TESTNET.md` that states the testnet kernel:
  - Cosmos SDK + CometBFT node;
  - AWCE-1 wallet compatibility layer;
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

- Testnet scope is small and not confused by prototype-era docs. -->
<!-- 
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

- The launch scope is enforceable by CI instead of tribal knowledge. -->
<!-- 
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

- Validator can download/build one binary and verify what it is. -->
<!-- 
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

- Local testnet can be booted repeatedly with clear commands. -->

<!-- ### Task 1.4 - Export/Import Roundtrip

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

- Testnet can rehearse a coordinated upgrade before public launch. -->
<!-- 
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
- AWCE-1 wallet compatibility summary;
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

- A validator can join without reading source code. -->
<!-- 
### Task 2.2 - Docker Image

Aetra logo in /assets/aetra.png

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

- Release can publish a validator-ready image. -->
<!-- 
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

- A release candidate cannot be published without the runnable gates. -->
<!-- 
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
  10 AET without choosing a validator. -->
<!-- 
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

- Pool state survives restart and scales past prototype size. -->
<!-- 
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

- Rewards are understandable and not open-ended. -->
<!-- 
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

- Validators and pool participants have deterministic slashing risk. -->
<!-- 
### Task 3.5 - Validator Score V1 And Identity Reputation Signals

Implementation:

- Minimal deterministic validator score:
  - uptime;
  - missed blocks;
  - commission;
  - slashing risk;
  - stake efficiency;
  - pool allocation limit;
  - operational history.
- Score output drives allocation engine weights.
- No nondeterministic inputs.
- Validator score is a consensus-role score, not the user's wallet reputation.
- Pool staking contributes to the user's unified `IdentityReputation` only
  through stake-time exposure signals:
  - amount staked;
  - duration;
  - consistency;
  - claim/unbond settlement height;
  - slashing exposure.
- Do not store a separate stake reputation record for the wallet.
- Pool share records may keep accounting fields needed to compute stake-time,
  but the final reputation update goes into the single identity reputation
  record.
- Low identity reputation may apply only a small bounded liquid-staking yield or
  allocation priority adjustment; it must never block pool deposits or exits.

Tests:

- same input -> same weights;
- score changes with uptime/commission/slashing;
- inactive/ineligible validators rejected;
- export/import preserves scores and snapshots.
- pool claim computes identity reputation delta from stake amount and duration;
- longer stake duration produces larger bounded reputation delta;
- larger stake amount produces larger bounded reputation delta;
- unstake/claim cannot create duplicate reputation credit;
- low reputation user can still deposit 10 AET and create token/NFT contracts;
- no separate wallet stake reputation state is exported.

Done:

- Allocation engine has a transparent v1 validator score, and pool stake-time
  feeds the unified wallet identity reputation. -->
<!-- 
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

2.1 ❗ Missing VM execution control errors (КРИТИЧНО)

У тебя нет ошибок уровня control flow / call / continuation.

Это важно для любой stack/continuation VM.

❌ отсутствует:
continuation missing
invalid jump / branch
call stack depth exceeded
recursive call limit exceeded

👉 почему важно:
без этого VM будет “ломаться молча” при сложных вызовах

➕ нужно добавить:
14 = invalid jump / branch
15 = call stack overflow
16 = continuation not found
17 = recursion limit exceeded
2.2 ❗ Missing type / memory safety errors (очень важно для safety VM)

У тебя есть type check error (7), но этого мало.

❌ нет:
invalid memory access
null reference / empty slice read
invalid chunk reference (если у тебя content-addressing)
➕ добавить:
18 = invalid memory access
19 = null reference
20 = invalid chunk reference
21 = corrupted state object
2.3 ❗ Missing arithmetic edge cases

У тебя есть integer overflow, но не хватает:

division by zero
negative shift / invalid shift
underflow arithmetic (если unsigned)
➕ добавить:
22 = division by zero
23 = invalid shift operation
24 = arithmetic underflow
2.4 ❗ Missing execution safety / gas edge cases

У тебя есть “out of gas”, но нет градации.

В более зрелых VM обычно есть:

gas limit exceeded
gas reservation failure
execution timeout
➕ добавить:
25 = gas limit exceeded
26 = gas reservation failed
27 = execution timeout
2.5 ❗ Missing action pipeline errors (сейчас у тебя только базовые)

У тебя есть basic message errors, но нет:

message routing failure
queue overflow
shard / routing failure (если message-driven VM)
➕ добавить:
38 = message routing failed
39 = queue overflow (у тебя partially есть message too large, но это другое)
40 = shard unavailable / routing failure
2.6 ❗ Missing storage / state consistency errors

Ты указал:

account state too big

Но нет:

state corruption
state version mismatch
snapshot failure
➕ добавить:
41 = state corruption
42 = state version mismatch
43 = snapshot failure
2.7 ❗ Missing developer / contract-level controlled aborts

У тебя есть THROWARG (11), но нет:

explicit user abort
assert failure
➕ добавить:
44 = explicit contract abort
45 = assertion failed
3. ВАЖНЫЙ ПРОБЕЛ В АРХИТЕКТУРЕ
Сейчас у тебя есть проблема:

Ты смешал:

VM errors
contract errors
action errors

Но у тебя нет разделения по слоям

4. Что реально нужно исправить в ТЗ

Добавь концепт:

4.1 Error domains
0–31   = VM execution errors
32–63  = action/message errors
64–95  = state/storage errors
96–127 = system/host errors
5. КРАТКОЕ СРАВНЕНИЕ
У тебя есть:

✔ stack safety
✔ basic arithmetic
✔ basic actions
✔ gas stop

НЕ ХВАТАЕТ:
1. Control flow safety (call/jump/continuation errors)
2. Memory/chunk safety errors
3. Full gas lifecycle errors
4. Message routing / queue system errors
5. State consistency / versioning errors
6. Explicit abort/assert semantics
6. Самое важное (если упростить)

Твоя VM сейчас:

“compute + basic actions VM”

А не хватает:

“full deterministic message-driven execution engine”

проверь что есть exit code о нехватке AET газ выполнение транзакции. 
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
Core model
Chunk {
  data_bits <= 2048
  refs <= 8
  type_tag optional
  level uint8 (0..N)
  hash_layers [1..K] optional
  hash = BLAKE3(canonical_chunk_encoding)
}
1. 🔁 Добавь MULTI-LAYER HASHING (ключевая фишка)

Это самая важная адаптация.

Идея:

Не один hash, а несколько уровней абстракции.

Rules:
Each Chunk MUST support hierarchical hashing levels:

H0 = BLAKE3(raw encoding)
H1 = BLAKE3(encoding where refs replaced by H0)
H2 = BLAKE3(encoding where refs replaced by H1)
...
Зачем это тебе:
быстрые partial proofs
можно проверять subgraphs без полного rehash
ускоряет state sync
даёт возможность light verification
2. 🌳 LEVEL FIELD (важно для DAG scaling)

Добавь:

level = max(ref.level) + 1

Rules:

leaf chunks → level = 0
parent inherits max level + 1
optional cap: level <= 255
Почему это важно:

Это даёт тебе:

fast pruning
selective sync
proof compression
incremental verification
3. 🧬 TYPE_TAG → “EXOTIC CHUNK MODEL”

Добавь расширяемую систему типов:

type_tag defines chunk semantics:
- normal
- pruned
- snapshot
- diff
- proof
- system
Rules:
- type_tag MAY change serialization rules
- but MUST NOT break hash determinism
- unknown type_tag = reject or treat as opaque
Зачем это тебе:

Это заменяет “exotic cells” концептуально:

pruned graph storage
state diffs
proof chunks
snapshot compression
4. 🔗 REFS RULE UPGRADE (твои 8 refs — это мощно)

У тебя 8 refs → значит можно добавить структуру:

Rule upgrade:
refs[0..7] are ordered and typed:

refs {
  data_edges (0..3)
  control_edges (4..5)
  metadata_edges (6)
  system_edges (7)
}
Почему это важно:

Ты превращаешь Chunk из “node” в:

multi-channel graph node

5. 🔒 CANONICAL ENCODING (очень важно)

Добавь строгость как в серьёзных VM:

Canonical rule:
Canonical encoding MUST ensure:

- deterministic bit order
- fixed ref ordering
- no optional field ambiguity
- no encoding variants allowed
Правило:
Two identical chunks MUST always produce identical bytes regardless of:
- compiler
- runtime
- serialization library
6. 🧩 MERKLE-DAG INVARIANT (ключевая архитектурная идея)

Добавь как обязательный invariant:

Chunk system MUST form:

Directed Acyclic Graph (DAG)

Rules:
- no cycles allowed
- every chunk hash depends only on:
  - its data
  - its refs' hashes
7. ⚡ IMMUTABILITY + VERSION SAFETY

Добавь:

Chunks are immutable.

Any modification produces a new chunk:

new_hash = f(old_chunk + diff)
optional upgrade:
If type_tag = diff:
  chunk represents transformation over parent chunks
8. 🧠 STATE MODEL UPGRADE (очень важный insight)

Теперь можно прямо усилить VM:

VM State = Root Chunk

Execution = deterministic transformation:

(StateChunk, InputMessage) → (NewStateChunk, OutputChunks, ExitCode)
9. 🔥 ЧТО ТЫ ФАКТИЧЕСКИ ЗАИМСТВУЕШЬ (без копирования)

Ты берёшь 4 идеи, но в своей форме:

1. Graph-based storage

✔ chunks = DAG nodes

2. Multi-level hashing

✔ H0/H1/H2 abstraction layers

3. Typed exotic semantics

✔ type_tag replaces exotic cells

4. Deterministic canonical serialization

✔ strict encoding rules

10. 🚀 ИТОГ — ЧТО У ТЕБЯ СТАЛО ЛУЧШЕ

После этой доработки:

Было:
просто DAG
1 hash
8 refs
static chunk
Стало:
multi-layer Merkle DAG
typed execution/storage model
deterministic canonical encoding
structured edge semantics
proof-ready state system
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
1. 🧠 ДОБАВИТЬ TYPE LATTICE (очень важно)

Сейчас у тебя типы плоские.

Нужно добавить идею:

типы образуют систему совместимости (lattice)

➕ Добавить:
Types form a directed compatibility graph (type lattice):

- each runtime value has a canonical type_id
- subtyping is explicit
- implicit casting is forbidden unless defined in Codec
Почему это важно:

Без этого:

Map<T> ломается на edge cases
tuple/struct превращаются в хаос
codec становится непредсказуемым
2. 🧬 TYPE AS CHUNK (очень сильное улучшение)

Сейчас type_hash есть, но слабый.

Нужно усилить:

➕ Добавить:
All complex types MUST compile into deterministic Chunk schemas:

TypeSchema = Chunk {
  schema AST
  hash = BLAKE3(schema)
}
Что это даёт:
тип = объект DAG
schema versioning
возможность ончейн эволюции типов
3. ⚙️ TYPED VALUE CARRIES TYPE_ID (важное упущение)

Сейчас VM хранит “universal values”.

Но нет привязки к типу во время исполнения.

➕ Добавить:
Each runtime value MAY carry:

Value {
  payload
  type_id (optional but recommended)
}
И правило:
If type_id exists:
  Codec MUST validate strict type match
4. 🔁 ADD CANONICAL VALUE NORMALIZATION

Сейчас у тебя нет нормализации значений.

➕ Добавить:
All values MUST have canonical form:

- no alternate encodings allowed
- no redundant padding
- no ambiguous integer representation
Пример:
uint32(1) = always same bytes
string encoding = strictly length-prefixed UTF-8 only
5. 📦 CODEC SHOULD BE STATEFUL (очень важное улучшение)

Сейчас Codec = stateless encode/decode.

Но в DAG VM это недостаточно.

➕ Добавить:
Codec<T> MAY depend on:

- Chunk context
- schema version
- execution_frame
Почему это важно:
Map<K,V> требует context
chunk references требуют resolution
execution_frame влияет на decoding
6. 🔥 ADD “LAZY DECODING” MODEL

Сейчас Reader<T> линейный.

Добавь:

➕ Добавить:
Reader<T> MUST support lazy decoding:

- values are decoded on access
- nested chunks remain references until dereferenced
Зачем:
экономия gas
faster execution
partial evaluation of DAG
7. 🧩 MAP → CHUNK MAP SHOULD BE INDEXED DAG

Сейчас:

Map<K,V> compiled to ChunkMap

но нет структуры.

➕ Добавить:
ChunkMap MUST be implemented as:

- radix-tree or hash-trie
- stored as Chunk DAG
- immutable updates produce new root chunk
8. 🧠 ADD EXECUTION FRAME TYPE RULES

У тебя есть execution_frame как тип, но нет правил.

➕ Добавить:
execution_frame contains:

- stack snapshot
- gas state
- continuation pointer
- active chunk references
Почему это важно:

Это связывает:

VM execution
DAG state
control flow
9. ⚡ ADD CODEC COST MODEL (ОЧЕНЬ ВАЖНО)

Сейчас у тебя есть gas в VM, но нет gas в codec.

➕ Добавить:
Each Codec<T> MUST define:

- encode_gas_cost
- decode_gas_cost
- max_size
Почему это критично:

иначе:

сериализация станет DOS-уязвимостью
Map decode может быть экспоненциальным
10. 🔒 STRICT FAIL MODEL FOR DECODING

Сейчас:

invalid decode reverts

но нужно уточнить:

➕ Добавить:
All decode failures MUST:

- revert execution frame
- not partially mutate state
- not leak invalid chunk state
11. 🧬 FINAL STRUCTURAL GAP SUMMARY

Твоя текущая версия не хватает:

❌ Missing core layers:
1. Type lattice (нет системы типов как графа)
2. Type = chunk (нет ончейн схем)
3. Typed runtime values (нет строгой привязки)
4. Lazy decoding (всё eagerly)
5. Codec gas model (нет защиты от DOS)
6. Execution-frame coupling (слабая связь VM ↔ types)
7. Canonical normalization rules (не полная строгость)
🚀 КАК ТЕПЕРЬ ВЫГЛЯДИТ УРОВЕНЬ ТВОЕЙ VM

После апгрейда:

Было:
typed serialization system
static types
simple codec
Стало:
DAG-based type system
schema-as-chunk architecture
lazy evaluation codec layer
gas-aware serialization engine
execution-frame aware decoding
fully deterministic canonical VM

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

🔥 1. НЕ ХВАТАЕТ: EXPLICIT TREE INVARIANTS

Сейчас у тебя “persistent trie”, но нет формального определения корректности дерева.

➕ Добавить:
ChunkMap MUST satisfy strict structural invariants:

1. Each key maps to exactly one leaf Chunk
2. No duplicate paths allowed
3. Internal nodes are deterministic function of children
4. Empty branches are canonicalized (no null ambiguity)
Почему это важно:

Без этого:

возможны “разные валидные деревья”
state becomes non-canonical
proofs become ambiguous
🔥 2. НЕТ MERKLE COMMITMENT MODEL (КРИТИЧНО)

Сейчас root есть, но не формализован как криптографическое commitment.

➕ Добавить:
ChunkMap root MUST be a cryptographic commitment:

root_hash = BLAKE3(serialize(root_chunk))

AND:

- every node hash includes children hashes
- any modification propagates to root deterministically
Зачем:

Это превращает ChunkMap в:

verifiable state structure (not just data structure)

🔥 3. НЕТ PATH ENCODING SPECIFICATION

Ты пишешь:

“deterministic nibbles/buckets”

но не фиксируешь формат.

➕ Добавить:
Path encoding MUST be defined as:

key_hash = BLAKE3(key)

path = fixed-length sequence of base-8 digits (fanout=8)

Each level consumes 3 bits:

0–7 → branch index
Почему это важно:

Сейчас у тебя ambiguity:

nibble?
byte split?
bit chunk?

👉 без фиксации невозможно interop

🔥 4. НЕТ BALANCED TREE / DEGENERACY CONTROL

Trie может стать:

глубоко линейным
или неравномерным
➕ Добавить:
ChunkMap MUST enforce maximum depth:

max_depth = ceil(hash_length / log2(fanout))

AND optionally:

- path compression allowed for single-child chains
Зачем:
защищает от worst-case O(n) depth
делает gas predictable
🔥 5. НЕТ PATH COMPRESSION (ОЧЕНЬ ВАЖНО)

Сейчас у тебя классический trie.

Но тебе нужно ускорение:

➕ Добавить:
If a subtree contains only one branch:

MUST be compressed into a single Chunk node:

compressed_path + leaf
Почему:
уменьшает depth
ускоряет lookup
уменьшает state size
🔥 6. НЕТ VERSIONED STATE MODEL

Сейчас нет истории/версий.

➕ Добавить:
Each ChunkMap update produces new root version:

ChunkMap {
  root_chunk
  version_id (monotonic or hash-chained)
}
Можно усилить:
version_id = BLAKE3(prev_root_hash + operation_hash)
🔥 7. НЕТ PROOF MODEL (КРИТИЧНО ДЛЯ DAG VM)

Сейчас есть “can be proven independent”, но нет формализации.

➕ Добавить:
ChunkMap MUST support:

- inclusion proof (key exists)
- exclusion proof (key does not exist)
- partial subtree proof
Почему:

Это превращает структуру в:

verifiable state machine (light-client ready)

🔥 8. НЕТ DETERMINISTIC ITERATION ORDER

Ты пишешь:

“deterministic iteration only through bounded proof/index APIs”

но не фиксируешь порядок.

➕ Добавить:
Iteration order MUST be:

- lexicographically by key_hash
OR
- level-order traversal of trie

AND MUST be identical across all nodes
🔥 9. НЕТ CONCURRENCY / PARALLEL UPDATE MODEL (ВАЖНО ДЛЯ ТВОЕЙ VM)

Ты пишешь “parallel-friendly buckets”, но не формализовано.

➕ Добавить:
ChunkMap updates MAY be parallelized if:

- different top-level fanout buckets are modified
- no overlapping path prefixes exist
Это даёт:
shard execution model
parallel contract execution
conflict-free updates
🔥 10. GAS MODEL СЛИШКОМ ОБЩИЙ

Сейчас:

gas depends on depth and encoded bytes

но нужно точнее.

➕ Добавить:
Gas MUST be computed as:

gas = O(depth) + O(chunks visited) + O(encoded_bytes)

Where:

- lookup = O(depth)
- insert = O(depth)
- delete = O(depth)
- proof generation = O(subtree size)
🔥 11. НЕТ COLLISION RESOLUTION MODEL

Ты пишешь:

key collision handling

но не описано как.

➕ Добавить:
If key_hash collision occurs:

- store multiple values in collision bucket
- bucket MUST be encoded as secondary ChunkMap
🔥 12. САМЫЙ ВАЖНЫЙ ПРОБЕЛ (АРХИТЕКТУРНО)

Сейчас ChunkMap = data structure

но тебе нужно:

ChunkMap = state transition engine primitive

➕ Добавить финальный invariant:
ChunkMap MUST behave as a pure function:

update(map, op) → new_map

No in-place mutation is ever allowed
🚀 ИТОГ — ЧТО ТЫ СЕЙЧАС УЛУЧШАЕШЬ
Было:
persistent trie
8-fanout
hash-based keys
immutable updates
Стало:
cryptographically committed state trie
proof-capable structure (inclusion/exclusion)
deterministic iteration model
concurrency-safe bucket execution
canonical encoding spec
gas-predictable operations
versioned state machine

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
  contract maps use ChunkMap, not global storage slots. -->

<!-- ### Task 4.5 - AVM Execution Model

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

🔥 1. НЕ ХВАТАЕТ FORMAL STATE TRANSITION FUNCTION (КРИТИЧНО)

Сейчас у тебя phases описаны как список, но нет математической модели.

➕ ДОБАВИТЬ:
Execution MUST be defined as a pure state transition:

(StateChunk, Message) → (NewStateChunk, OutputActions, Receipt, ExitCode)
Почему это важно:

Без этого:

нет формальной детерминированности
сложно делать replay
нельзя доказывать корректность execution
🔥 2. НЕТ EXPLICIT READ/WRITE SEPARATION

Сейчас “Storage Phase” есть, но не формализовано чтение/запись.

➕ ДОБАВИТЬ:
Execution MUST separate:

READ STATE:
- immutable snapshot of input Chunks

WRITE STATE:
- only produced in Finalization Phase

No phase MAY mutate input state directly.
🔥 3. НЕТ ACTION DETERMINISM MODEL (очень важно)

Action Phase у тебя есть, но не гарантируется детерминизм.

➕ ДОБАВИТЬ:
Action Phase MUST be deterministic:

- same inputs → identical ordered action list
- action ordering MUST be canonical
- no hidden randomness allowed
Почему это критично:

иначе:

разные nodes → разные messages
consensus break
🔥 4. НЕТ RECEIPT STRUCTURE FORMALIZATION

Ты пишешь “records receipt”, но не описано что это.

➕ ДОБАВИТЬ:
Receipt MUST contain:

- exit_code
- gas_used
- gas_limit
- state_root_before
- state_root_after
- emitted_actions_hash
- execution_trace_hash (optional)
🔥 5. НЕТ EXECUTION TRACE MODEL (DEBUG + VERIFIABILITY)

Сейчас нет trace layer.

➕ ДОБАВИТЬ:
Execution MAY produce deterministic trace:

- instruction steps
- stack deltas
- chunk reads/writes
- gas consumption steps
Зачем:
debugging
light client verification
zk-proof compatibility (если захочешь дальше)
🔥 6. НЕТ MESSAGE CONTEXT SEPARATION

У тебя:

pending calls/messages

но нет разделения типов сообщений.

➕ ДОБАВИТЬ:
Messages MUST be categorized:

- external message (user → contract)
- internal message (contract → contract)
- system message (VM/host generated)
Почему:

иначе action phase становится неуправляемой

🔥 7. НЕТ GAS ACCOUNTING PER PHASE

Ты пишешь “gas limits enforced”, но нет breakdown.

➕ ДОБАВИТЬ:
Gas MUST be tracked per phase:

- storage_load_gas
- compute_gas
- action_gas
- finalization_gas
Почему:
предотвращает hidden DoS
делает execution predictable
🔥 8. НЕТ REVERT MODEL FORMALIZATION

Сейчас:

out-of-gas reverts state changes

но не определено “как именно”.

➕ ДОБАВИТЬ:
On failure:

- all WRITE operations are discarded
- READ snapshot remains intact
- receipt is still persisted
- emitted actions are dropped unless explicitly marked as "system-bounce"
🔥 9. НЕТ DETERMINISTIC SCHEDULING MODEL

Execution phases есть, но порядок message execution не формализован.

➕ ДОБАВИТЬ:
Message execution order MUST be:

1. sorted by (block_height, message_hash)
2. tie-breaker: lexicographic sender address
🔥 10. НЕТ STATE ROOT CONSISTENCY RULE

Ты говоришь “commit new Chunk roots”, но нет правил.

➕ ДОБАВИТЬ:
Finalization MUST ensure:

- exactly one new root Chunk is produced
- root MUST include all state changes
- root MUST be hash-stable across nodes
🔥 11. НЕТ PARTIAL FAILURE MODEL

Сейчас failure = abort.

Но нет granular failures.

➕ ДОБАВИТЬ:
Failures MUST be categorized:

- recoverable (retryable message)
- non-recoverable (abort contract)
- system-fatal (node error)
🔥 12. НЕТ ISOLATION MODEL (ВАЖНО ДЛЯ SECURITY)

Сейчас ExecutionFrame есть, но isolation не определён.

➕ ДОБАВИТЬ:
Each ExecutionFrame MUST be isolated:

- no shared mutable memory between frames
- all dependencies passed via Chunks
- no global VM state mutation allowed
🚀 ИТОГ — ЧТО У ТЕБЯ СЕЙЧАС И ЧТО СТАЛО
❌ Сейчас:
pipeline phases (описательные)
weak determinism guarantees
implicit state model
partial gas model
loose action semantics
✅ После улучшений:
formal state transition function
deterministic execution ordering
strict read/write separation
verifiable receipts
traceable execution model
phased gas accounting
message classification system
replay-safe execution engine

Tests:

- deploy;
- execute external;
- execute internal;
- out-of-gas rollback;
- abort exit code;
- same input/state/code -> same root/gas/receipt;
- forbidden nondeterministic opcode rejected.

Done:

- AVM can execute deterministic contracts safely. -->
<!-- 
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

🔥 1. НЕТ CAPABILITY MODEL (КРИТИЧНО)

Сейчас:

"Allowed list"

Но это не защита, а просто whitelist.

➕ ДОБАВИТЬ:
Each contract MUST explicitly declare host capabilities it is allowed to use.

Host calls are NOT globally available.

ExecutionFrame carries capability mask.
Пример:
capabilities = {
  crypto: [sha256, blake3, ed25519_verify],
  chain: [get_height, get_chain_id],
  messaging: [send_internal],
  storage: [read_chunk]
}
Почему это важно:

Без этого:

любой контракт может вызвать любой host function
невозможно sandbox enforcement
future security model ломается
🔥 2. НЕТ PURE VS EFFECTFUL HOST CLASSIFICATION

Сейчас все host functions равны.

Но это ошибка.

➕ ДОБАВИТЬ:
Host functions MUST be classified:

PURE:
- hash functions
- address parsing
- verification

EFFECTFUL:
- storage read/write/delete
- emit event
- send message
Почему:

Это основа:

determinism proofs
replay systems
parallel execution
🔥 3. НЕТ STRICT GAS MODEL PER FUNCTION

Сейчас:

each allowed host has gas cost

но не формализовано.

➕ ДОБАВИТЬ:
Each host function MUST define:

- base gas cost
- per-byte cost (if applicable)
- per-ref cost (for Chunk operations)
Пример:
sha256:
  base = 10 gas
  per 32 bytes = 1 gas

send_internal:
  base = 50 gas
  per message byte = 2 gas
🔥 4. НЕТ SIDE-EFFECT ATOMICITY MODEL

Сейчас storage + message + event не связаны формально.

➕ ДОБАВИТЬ:
All EFFECTFUL host calls MUST be:

- staged during execution
- committed only in Finalization Phase
- rollback-safe if execution fails
Это критично:

иначе у тебя:

half-written state
ghost messages
inconsistent DAG
🔥 5. НЕТ MESSAGE SEMANTICS MODEL

Сейчас:

send internal message

но нет гарантий.

➕ ДОБАВИТЬ:
Internal messages MUST be:

- deterministic
- ordered
- content-addressed
- stored as Chunk objects before emission
И правило:
Message emission is NOT execution.
It is a queued effect.
🔥 6. НЕТ BLOCK CONTEXT OBJECT (ОЧЕНЬ ВАЖНО)

Сейчас:

get block height
get chain id

но нет unified context.

➕ ДОБАВИТЬ:
Host functions access immutable BlockContext:

BlockContext {
  height
  chain_id
  block_hash
  timestamp (consensus-based)
  entropy_seed
}
Почему:
prevents fake time
ensures reproducibility
simplifies verification
🔥 7. RANDOMNESS СЕЙЧАС НЕДОФОРМАЛИЗОВАН

У тебя уже есть идея, но нужно усилить.

➕ ДОБАВИТЬ:
secure_random MUST be derived only from:

random = BLAKE3(
  previous_state_root ||
  block_entropy ||
  message_hash ||
  contract_address
)
И важное правило:
Randomness MUST be deterministic per block.
No external entropy sources allowed.
🔥 8. НЕТ EXECUTION ISOLATION BOUNDARY

Сейчас host functions выглядят как глобальные API.

➕ ДОБАВИТЬ:
Host functions execute inside sandboxed ExecutionFrame:

- no shared memory
- no global state access
- all inputs must come from Chunk or BlockContext
🔥 9. НЕТ ERROR BOUNDARY MODEL (ВАЖНО)

Сейчас:

abort with exit code

но не определено поведение ошибок host layer.

➕ ДОБАВИТЬ:
Host function failure MUST:

- produce VM exit code (mapped)
- not corrupt state
- not partially apply effects
🔥 10. НЕТ AUDIT TRAIL (СИЛЬНОЕ УЛУЧШЕНИЕ)

Это часто забывают.

➕ ДОБАВИТЬ:
All host calls MUST be recorded in execution trace:

- function_id
- input hash
- output hash
- gas used
🚀 ИТОГ — ЧТО ТЫ СЕЙЧАС УЛУЧШАЕШЬ
❌ Сейчас:
whitelist host API
implicit effects
weak randomness model
no capability isolation
no formal gas breakdown
partial determinism
✅ После улучшения:
capability-based security model
pure/effectful separation
deterministic BlockContext system
fully auditable host execution trace
sandboxed execution environment
cryptographically safe randomness
strict gas accounting per function

Tests:

- each allowed host has gas cost;
- unknown host rejected;
- forbidden host rejected;
- host storage respects Chunk/ChunkMap limits;
- send_internal respects queue limits;
- abort returns contract-defined small exit code.

Done:

- Host surface is deterministic and auditable. -->

<!-- ### Task 4.7 - Get Methods

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

  🔥 1. НЕТ FORMAL QUERY EXECUTION DOMAIN (КРИТИЧНО)

Сейчас get-method = “execute with restrictions”.

Это опасная модель.

➕ ДОБАВИТЬ:
Get methods MUST run in a separate Query Execution Domain:

QueryState ≠ ExecutionState

Rules:
- no mutation allowed
- no action queue exists
- no side-effect buffer exists
Почему это важно:

иначе:

accidental state writes
hidden side-effects через bugs
non-replayable queries
🔥 2. НЕТ QUERY SNAPSHOT MODEL (СИЛЬНЫЙ ПРОБЕЛ)

Сейчас:

loads code and state root

но не определено как именно.

➕ ДОБАВИТЬ:
Get method executes on immutable snapshot:

QuerySnapshot = {
  state_root_chunk,
  block_context,
  contract_code_chunk
}
И правило:
Snapshot MUST NOT change during execution.
🔥 3. НЕТ QUERY DETERMINISM GUARANTEE (ВАЖНО)

Сейчас есть “deterministic”, но не формализовано.

➕ ДОБАВИТЬ:
Get method MUST be deterministic:

same (snapshot + args) → identical output bytes + gas usage
И запрет:
Get method MUST NOT depend on:
- wall clock
- external calls
- nondeterministic iteration
🔥 4. НЕТ QUERY GAS MODEL (СИЛЬНО НЕДООПИСАНО)

Сейчас:

executes with query gas limit

но нет структуры.

➕ ДОБАВИТЬ:
Query gas is separate from execution gas:

query_gas = {
  compute_gas
  decode_gas
  serialization_gas
}
Важно:
query gas does NOT affect state gas
query gas is non-refundable in some modes (optional design)
🔥 5. НЕТ QUERY ISOLATION BOUNDARY (КРИТИЧНО)

Сейчас:

cannot write storage / send messages

но это “runtime check”, не архитектура.

➕ ДОБАВИТЬ:
Query VM MUST NOT instantiate:

- action queue
- storage writer
- message emitter
- event emitter

Only read-only execution frame is created.
🔥 6. НЕТ PROOF MODE (ОЧЕНЬ СИЛЬНОЕ УЛУЧШЕНИЕ)

Ты упомянул “proof metadata”, но не описал модель.

➕ ДОБАВИТЬ:
Get methods MAY run in Proof Mode:

- execution produces inclusion proof
- returns partial Chunk path
- allows light client verification
Это даёт:
light client queries
verifiable API responses
trustless indexing
🔥 7. НЕТ RESPONSE CANONICALIZATION

Сейчас:

return typed value bytes/Chunk

но нет строгой формы.

➕ ДОБАВИТЬ:
All query responses MUST be canonical encoded:

- deterministic serialization
- no field ordering variance
- no optional ambiguity
🔥 8. НЕТ ARGUMENT VALIDATION LAYER

Сейчас:

malformed args rejected

но не определено как.

➕ ДОБАВИТЬ:
Query arguments MUST be validated via Codec<T> before execution:

- decode before VM starts
- invalid decode = immediate rejection (no gas charged beyond decode phase)
🔥 9. НЕТ QUERY CACHE MODEL (ОЧЕНЬ ВАЖНО ДЛЯ ПРОДАКШЕНА)

Сейчас нет кэширования.

➕ ДОБАВИТЬ:
Get method results MAY be cached if:

- same state_root_chunk
- same method_id
- same arguments hash

Cache MUST be invalidated on state root change
🔥 10. НЕТ QUERY EXECUTION STACK LIMITS

Сейчас есть gas, но нет safety caps.

➕ ДОБАВИТЬ:
Query execution MUST enforce:

- max stack depth
- max recursion depth
- max chunk traversal depth
🔥 11. НЕТ METHOD DISCOVERY MODEL (очень полезное улучшение)

Сейчас CLI вызывает метод напрямую.

➕ ДОБАВИТЬ:
Contract MUST expose method registry:

- method_id
- input schema
- output schema
- gas estimate
🔥 12. НЕТ QUERY TRACE MODE (DEBUG/DEV TOOLING)
➕ ДОБАВИТЬ:
Query MAY return execution trace:

- chunk reads
- gas breakdown
- opcode-level steps (optional debug mode)
🚀 ИТОГ — ЧТО У ТЕБЯ СЕЙЧАС И ЧТО СТАЛО
❌ Сейчас:
read-only execution (ad hoc)
weak isolation model
implicit snapshot
no formal proof layer
no caching model
minimal determinism guarantees
✅ После улучшения:
separate Query Execution Domain
immutable snapshot execution model
proof-capable query system
canonical response encoding
cacheable deterministic queries
codec-based argument validation
structured gas accounting
method registry system
💡 ГЛАВНАЯ СУТЬ ПРОБЛЕМЫ ТВОЕГО ТЕКСТА

Сейчас:

“get methods = restricted execution”

После улучшения:

“get methods = deterministic verifiable query runtime”

Tests:

- get method reads state;
- attempted write rejected;
- attempted send message rejected;
- gas limit enforced;
- response deterministic;
- malformed args rejected;
- proof query stable.

Done:

- Contracts have practical read APIs without state mutation. -->

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

🔥 1. НЕТ FORMAL MODULE TRUST MODEL (КРИТИЧНО)

Сейчас:

verifier accepts/rejects bytecode

Но не определено на каком уровне доверия.

➕ ДОБАВИТЬ:
AVM Module MUST have explicit trust classification:

- untrusted (user-uploaded)
- verified (passed static verifier)
- canonical (standard library / system)
Почему это важно:

иначе:

невозможно различать system vs user code
невозможно безопасно делать upgrades
verifier становится единственной точкой доверия
🔥 2. НЕТ MODULE HASH SEMANTIC RULE

Ты пишешь metadata hash, dependency hash — но не определено главное:

➕ ДОБАВИТЬ:
ModuleHash MUST be defined as:

hash = BLAKE3(canonical_encoding(all module sections))
И правило:
Any change in ANY section MUST change ModuleHash
🔥 3. НЕТ LINKING MODEL (IMPORT/EXPORT НЕ ПОЛНЫЙ)

Сейчас import/export есть, но нет semantics.

➕ ДОБАВИТЬ:
Imports MUST be resolved via deterministic linking:

- import name → module_id → function_index
- no runtime symbol resolution allowed
- all linking MUST happen at verification time
🔥 4. НЕТ BYTECODE EXECUTION SAFETY MODEL (важно)

Сейчас:

stack underflow, overflow, invalid jump

но это только часть.

➕ ДОБАВИТЬ:
Verifier MUST statically guarantee:

- all jumps target valid instruction boundaries
- no infinite control flow loops without gas check
- no unreachable dead instruction execution paths (optional analysis mode)
🔥 5. НЕТ CONTROL FLOW GRAPH VALIDATION (ОЧЕНЬ ВАЖНО)

Это missing core feature.

➕ ДОБАВИТЬ:
Bytecode MUST be converted to Control Flow Graph (CFG) before acceptance:

- all basic blocks identified
- all branches validated
- no dangling entrypoints allowed
🔥 6. НЕТ STACK SAFETY MODEL (сейчас слишком поверхностно)
➕ ДОБАВИТЬ:
Verifier MUST compute stack effect per instruction:

- max stack depth
- min stack depth
- net stack delta per block
Почему:

иначе:

runtime crashes
undefined execution states
🔥 7. НЕТ DETERMINSITIC VERIFICATION RESULT MODEL (ВАЖНО ДЛЯ CONSENSUS)

Ты пишешь:

stored with code metadata

но не определяешь формат.

➕ ДОБАВИТЬ:
VerificationResult MUST be canonical:

VerificationResult {
  module_hash
  verifier_version
  pass/fail
  error_code
  analyzed_stack_bound
  cfg_hash
}
🔥 8. НЕТ REPLAY SAFETY MODEL (КРИТИЧНО ДЛЯ NODE CONSENSUS)
➕ ДОБАВИТЬ:
Verification MUST be replay-safe:

- identical bytecode → identical verification result
- no machine-dependent analysis
- no parallel nondeterministic ordering in verification
🔥 9. НЕТ DEPENDENCY GRAPH VALIDATION

Ты упоминаешь dependency hashes, но не модель.

➕ ДОБАВИТЬ:
Module dependencies MUST form DAG:

- no cycles allowed
- each dependency MUST be verified before linking
- dependency hash MUST match exact module hash
🔥 10. НЕТ EXECUTION GUARANTEE BOUNDARY

Сейчас verifier проверяет, но не гарантирует runtime safety formally.

➕ ДОБАВИТЬ:
If module passes verification:

- execution MUST NOT trigger undefined opcode behavior
- execution MUST stay within verified stack bounds
- execution MUST respect CFG constraints
🔥 11. НЕТ FUZZ / MALFORMED INPUT MODEL (у тебя частично есть, но слабый)
➕ ДОБАВИТЬ:
Verifier MUST be:

- memory safe on arbitrary input
- panic-proof
- deterministic under fuzz conditions
🔥 12. НЕТ VERSION COMPATIBILITY RULE (ОЧЕНЬ ВАЖНО)
➕ ДОБАВИТЬ:
Verifier MUST enforce:

- ABI version compatibility
- instruction set version match
- backward incompatible modules rejected
🚀 ИТОГ — ЧТО У ТЕБЯ СЕЙЧАС И ЧТО СТАНЕТ
❌ Сейчас:
module format defined
basic validation rules
import/export structure
simple safety checks
✅ После улучшения:
full deterministic module verification pipeline
CFG-based static analysis
stack effect validation
dependency DAG enforcement
replay-safe verification result model
formal trust classification system
ABI/version compatibility enforcement
consensus-safe bytecode acceptance layer

Done:

- A node can safely accept/reject AVM code before deployment.

## 🧩 Task 4.X - Execution Kernel Architecture Upgrade (TVM-inspired, AVM-native)
Objective

Extend AVM execution core with a deterministic, message-driven, continuation-based execution model built on immutable Chunk state graphs and ExecutionFrames.

This must NOT introduce cell/slice/builder primitives.

All memory, state, and execution context MUST be expressed via Chunks, ChunkMaps, and typed VM values.

1. 🧠 Execution Model Upgrade (Stack + Continuation Hybrid)
Implementation:

AVM MUST operate as:

Stack Machine + Continuation Engine + Chunk DAG State Model
Execution properties:
Stack-based evaluation remains primary compute model
Control flow MUST be continuation-driven (not direct call stack mutation)
All execution state MUST be representable in ExecutionFrame
No mutable heap or linear memory allowed
Continuation model:

Introduce:

ContinuationSlot {
  return_ptr
  alt_return_ptr
  error_handler_ptr
  dispatcher_ptr
}
Mapping:
return_ptr → normal completion continuation
alt_return_ptr → alternate execution path
error_handler_ptr → exception recovery path
dispatcher_ptr → entry point resolver for contract logic
2. 🧩 ExecutionFrame Expansion (VM Context Layer)

Extend ExecutionFrame:

ExecutionFrame {
  ip (instruction pointer)
  stack_snapshot
  local_context_chunk
  continuation_slot
  env_chunk
  pending_actions
  error_state
}
Mapping rules:
env_chunk replaces runtime environment tuple
local_context_chunk replaces temporary VM state storage
pending_actions replaces outbound message buffer
3. 📦 Persistent State Model (Chunk Storage Layer)

Replace traditional storage model with:

StateRootChunk

Rules:

all persistent contract state MUST be stored in Chunk graphs
no mutable memory allowed
state transitions MUST produce new root Chunk
State mapping:
storage → StateRootChunk
persistent variables → sub-chunks inside state graph
4. 📡 Message + Action System (ActionQueue Layer)

Introduce:

ActionQueueChunk
Supported actions:
emit message
internal call
state update event
external notification event
Rules:
actions are accumulated during compute phase
actions are NOT executed immediately
actions are flushed only during Finalization Phase
action ordering MUST be deterministic
5. ⚙️ VM Register Subsystem (Rebranded Control Registers)

Replace TVM c0–c7 concept with AVM Execution Slots:

SLOT_RETURN        → normal continuation
SLOT_ALT_RETURN    → alternate continuation
SLOT_ERROR         → exception handler
SLOT_DISPATCH      → contract entry resolver
SLOT_STATE         → StateRootChunk
SLOT_ACTIONS       → ActionQueueChunk
SLOT_ENV           → ExecutionContextChunk
Key rule:

All slots MUST be:

immutable during execution
replaced only via continuation transitions
6. 🧠 Environment Model (BlockContext Integration)

Replace runtime tuple with:

ExecutionContextChunk {
  caller
  origin
  attached_value
  block_height
  chain_id
  contract_address
  message_hash
  timestamp (consensus-based)
}
Rules:
no wall-clock time allowed
no process-level entropy
all values MUST come from block-provided context
7. 📊 Stack Model (Typed Evaluation Stack)

AVM stack supports:

int256
bool
ChunkRef
ExecutionFrameRef
tuple
address
hash
Rules:
stack is strictly typed at opcode level
no implicit conversions
invalid type usage → runtime trap
8. ⚙️ Instruction Execution Phases

Execution MUST follow deterministic pipeline:

Phase 1 — State Load Phase
load StateRootChunk
load ExecutionContextChunk
Phase 2 — Credit Phase
apply attached value
update ExecutionContextChunk
Phase 3 — Compute Phase
execute bytecode
modify stack + continuation slots
accumulate actions
Phase 4 — Action Build Phase
finalize ActionQueueChunk
validate message constraints
Phase 5 — Commit Phase
produce new StateRootChunk
persist execution result
9. 🚨 Error + Exit Model Upgrade

Replace flat exit codes with structured model:

ExitCode = {
  category
  subcode
  message_hash
}
Categories:
SUCCESS
VM_ERROR
TYPE_ERROR
EXECUTION_ERROR
ACTION_ERROR
STATE_ERROR
GAS_ERROR
Rule:
VM errors MUST NOT mutate state
only SUCCESS reaches commit phase
all failures produce deterministic receipt
10. 🧮 Instruction Set Extension (Continuation-Aware Ops)

Add new opcode categories:

Control flow (continuation-aware):
CALL_FRAME
RETURN_FRAME
JUMP_COND
JUMP_UNCOND
RAISE_ERROR
TRY_BEGIN
TRY_END
State operations:
LOAD_STATE_CHUNK
STORE_STATE_CHUNK
CLONE_STATE
MERGE_STATE
Action system:
EMIT_ACTION
QUEUE_MESSAGE
FLUSH_ACTIONS
Context ops:
GET_CALLER
GET_ORIGIN
GET_VALUE
GET_BLOCK_HEIGHT
GET_CHAIN_ID
Cryptographic ops:
HASH_CHUNK
HASH_DATA
VERIFY_SIGNATURE
11. 🔁 Execution Semantics Rule (VERY IMPORTANT)

VM execution MUST satisfy:

(StateRootChunk, Message) 
→ (NewStateRootChunk, ActionQueueChunk, ExitCode, Receipt)
12. ⚡ Determinism Guarantees
no nondeterministic iteration
no process-dependent randomness
no external IO
identical input → identical output
13. 🧩 Mental Model (AVM version of TVM concept)

AVM can be modeled as:

Stack-based deterministic execution engine
+ continuation slots instead of call stack
+ Chunk DAG state instead of memory heap
+ ActionQueue instead of immediate side effects
+ BlockContext-driven environment
🚀 Result

После этого апгрейда ты получаешь:

continuation VM без call stack
DAG-based persistent state
message-driven execution model
deterministic action pipeline
block-context sealed environment
fully content-addressed runtime


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

🔥 1. НЕТ ОБЯЗАТЕЛЬНОЙ ARCHITECTURE DEMO PURPOSE

Сейчас:

counter, token, NFT…

Но не сказано что именно они должны демонстрировать.

➕ ДОБАВИТЬ:
Each example MUST demonstrate one core AVM primitive:

- counter → ExecutionFrame + state mutation via Chunk
- KV map → ChunkMap + persistent DAG updates
- registry → string encoding + ownership model
- token → ActionQueue + transfer semantics
- NFT → unique Chunk identity + metadata immutability
- staking adapter → cross-contract messaging + deterministic accounting
🔥 2. НЕТ MINIMAL SURFACE RULE (очень важно)

Сейчас нет ограничения сложности.

➕ ДОБАВИТЬ:
Examples MUST be minimal:

- no unnecessary abstraction layers
- no external dependencies
- no hidden runtime magic
- no framework-specific helpers
Почему:

иначе примеры становятся:

“framework demo”, а не VM reference

🔥 3. НЕТ STATE MODEL EXPLICIT LINK

Ты используешь Chunk / ChunkMap, но не заставляешь это показать.

➕ ДОБАВИТЬ:
Each contract MUST explicitly show:

- StateRootChunk structure
- how state evolves per call
- how ChunkMap is used internally
🔥 4. НЕТ ACTION QUEUE DEMONSTRATION REQUIREMENT

Очень важный пробел.

➕ ДОБАВИТЬ:
At least 2 examples MUST demonstrate ActionQueue usage:

- token transfer → emits internal message
- staking adapter → emits batch actions
🔥 5. НЕТ GET METHODS BINDING RULE

Сейчас get methods просто перечислены.

➕ ДОБАВИТЬ:
Each contract MUST define at least 1 get method:

- must be pure
- must not mutate Chunk state
- must demonstrate deterministic query execution
🔥 6. НЕТ SERIALIZATION / EXPORT LINK

Ты пишешь:

compile/encode/export/import

но не определяешь что именно проверяется.

➕ ДОБАВИТЬ:
Each example MUST verify:

- Chunk canonical encoding stability
- identical source → identical StateRootChunk hash
- import/export roundtrip equality
🔥 7. НЕТ DETEMINISTIC BEHAVIOR REQUIREMENT (КРИТИЧНО)
➕ ДОБАВИТЬ:
All examples MUST be fully deterministic:

same input state + message → same:
- state root
- actions
- receipts
- gas usage
🔥 8. НЕТ ERROR SCENARIO COVERAGE

Сейчас только happy path.

➕ ДОБАВИТЬ:
At least one example MUST demonstrate:

- invalid input rejection
- gas failure
- state revert behavior
🔥 9. НЕТ CROSS-CONTRACT INTERACTION RULE (ВАЖНО)
➕ ДОБАВИТЬ:
At least 1 example MUST demonstrate:

- internal message sending
- deterministic message ordering
- state isolation between contracts
🔥 10. НЕТ STANDARD INTERFACE DISCIPLINE

Ты упоминаешь AFT / ANFT, но не фиксируешь интерфейс.

➕ ДОБАВИТЬ:
AFT and ANFT MUST define:

- required state schema
- required actions
- required get methods
- required ChunkMap structure
🔥 11. НЕТ “REFERENCE CONTRACT SUITE STRUCTURE”

Сейчас это список.

Нужно превратить в систему.

➕ ДОБАВИТЬ:
All examples MUST be organized as:

/examples
  /counter
  /chunk_map_kv
  /registry
  /aft_token
  /anft_token
  /staking_adapter
🔥 12. НЕТ TEST BINDING TO EXECUTION MODEL

Сейчас tests общие.

➕ ДОБАВИТЬ:
Tests MUST validate:

- ExecutionFrame correctness
- Chunk state transitions
- ActionQueue emission
- continuation correctness

Done:

- A developer can write and run a minimal contract.


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


Define a minimal, deterministic, verifiable AVM Instruction Set Architecture (ISA) with:

strict stack discipline
fully specified execution semantics
deterministic gas model
formal error propagation rules
verifier-compatible bytecode constraints
1. 🧠 Core ISA Model

AVM is a:

Stack-based, strictly typed, deterministic execution machine
+ Chunk-based memory model
+ continuation-controlled control flow
Execution rule:
Every instruction MUST define:

(InputStackState, VMState) → (OutputStackState, VMState, ExitCode)
2. 📦 Instruction Structure (MANDATORY FORM)

Every opcode MUST be defined as:

Instruction {
  mnemonic
  opcode
  stack_input_signature
  stack_output_signature
  gas_cost_model
  overflow_behavior
  failure_exit_code
  determinism_rule
}
3. 🧩 STACK DISCIPLINE RULE (CRITICAL)
Rules:
- stack is LIFO
- no implicit casting
- all operations are type-checked at execution time
- invalid stack state → immediate trap
Stack safety invariant:
For every instruction:

stack_depth_out = stack_depth_in + push - pop
must always be ≥ 0
4. 🔢 ARITHMETIC MODEL (STRICT CHECKED SEMANTICS)

Replace “checked add/sub/mul/div/mod” with formal semantics:

Arithmetic instructions MUST:

- operate on int256 only
- detect overflow BEFORE commit
- return deterministic failure exit code on overflow
Division rule:
DIV(x, 0) → EXIT_DIV_ZERO
5. 🔀 CONTROL FLOW MODEL (IMPORTANT UPGRADE)

Replace simple jump/call with verified control model:

Control flow instructions MUST operate on validated targets:

- jump target MUST be CFG block entry
- call MUST push continuation frame
- return MUST restore continuation slot
- abort MUST terminate execution frame
Rule:
All control transfers MUST be CFG-prevalidated
6. 🧠 CHUNK INTEGRATION LAYER (IMPORTANT)
Chunk operations:
new_chunk
read_chunk
ref_chunk
hash_chunk
type_hash_chunk
Rules:
- all chunks are immutable
- all chunk refs are content-addressed
- invalid chunk reference → EXIT_CHUNK_ERROR
7. 🗺 CHUNKMAP OPS (STATE LAYER)
get
put
delete
proof
Formal rule:
ChunkMap operations MUST:

- operate only on StateRootChunk snapshot
- return new root after mutation ops
- never mutate in-place
8. 📡 MESSAGE MODEL (ACTION LAYER)

Replace “emit via host calls” with structured action ops:

read_caller
read_source
read_value
read_body
emit_message
emit_event
Rule:
All emit operations MUST go to ActionQueue buffer, not immediate execution
9. 🔒 TYPE SYSTEM ENFORCEMENT LAYER

Add:

encode<T>
decode<T>
load<T>
Rule:
All encode/decode MUST be Codec<T>-verified
invalid decode → EXIT_TYPE_ERROR
10. ⚡ GAS MODEL (CRITICAL FOR CONSENSUS)

Each instruction MUST define:

gas_cost = base + stack_cost + memory_cost + chunk_cost
Rule:
Gas consumption MUST be deterministic and independent of hardware
11. 🚨 ERROR PROPAGATION MODEL (VERY IMPORTANT FIX)

Replace simple exit codes with structured propagation:

On failure:

- instruction MUST NOT mutate state
- VM MUST unwind to ExecutionFrame checkpoint
- ExitCode MUST be recorded in Receipt
12. 🔁 DETERMINISM INVARIANT
Same program + same StateRootChunk + same input → ALWAYS:

- same final StateRootChunk
- same ActionQueue
- same Gas usage
- same TraceHash
13. 🧾 VERIFIER INTEGRATION RULE
Every instruction MUST be verifier-compatible:

- stack effect must be statically analyzable
- control flow must be CFG-resolved
- chunk ops must be bounded
14. 🧠 MENTAL MODEL (FINAL)

AVM ISA =

Stack machine core
+ CFG-controlled execution
+ Chunk DAG memory model
+ ChunkMap state engine
+ deterministic action queue
+ codec-enforced type system
🚀 ЧТО ТЫ ПОЛУЧИЛ ПОСЛЕ УЛУЧШЕНИЯ
❌ Было:
список категорий opcodes
частичные stack rules
описательная ISA
✅ Стало:
formal ISA specification
deterministic execution semantics
verifier-ready instruction model
CFG-based control flow system
full Chunk + ChunkMap integration
gas model tied to execution semantics
consensus-safe instruction semantics

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

Objective

Define a minimal, deterministic, size-bounded value system for AVM that:

guarantees canonical encoding
enforces strict runtime type safety
prevents unbounded runtime polymorphism
supports compile-time abstraction without runtime reflection
1. 🧠 Core Value Model

AVM runtime values MUST be a tagged union:

Value =
  Null
  Bool
  Int(Signed/Unsigned, bit_width)
  Coins(uint128)
  Timestamp(uint64)
  Address
  Hash
  Bytes
  String
  Tuple
  ChunkRef
  ReaderCursor
  WriterHandle
  ExecutionFrameRef
Rule:
Each Value MUST carry a deterministic type tag.
No untyped values are allowed in runtime.
2. 📦 Canonical Encoding Rule (CRITICAL)

All values MUST serialize into:

(ValueTag || CanonicalPayloadBytes)
Canonicalization rules:
integers → fixed-width big-endian
strings → UTF-8 + length prefix
tuples → ordered deterministic encoding
chunks → content-addressed hash reference
Rule:
Same semantic value MUST always produce identical byte representation.
3. 🧠 TYPE SAFETY MODEL (STRICT)
Invalid type operations MUST result in deterministic trap.
No implicit casting is allowed.
Example:
Int + String → EXIT_TYPE_ERROR
ChunkRef used as Int → EXIT_TYPE_ERROR
4. 🔢 INTEGER MODEL (NO AMBIGUITY)
Supported integers:

- int8 ... int256
- uint8 ... uint256
Rule:
All arithmetic MUST be width-preserving unless explicitly widening opcode is used.
5. 🧩 COMPOSITE TYPES (TUPLES)
Tuple = ordered fixed-length heterogeneous value array
Rule:
order is semantically significant
empty tuple is valid canonical value
nested tuples allowed but bounded
6. 🧠 STRING MODEL (IMPORTANT FOR CONSENSUS)
String MUST be:

- UTF-8 only
- length-bounded
- immutable
- canonical encoded
Rule:
No hidden normalization except canonical UTF-8 validation.
7. 📦 CHUNK VALUE INTEGRATION (CRITICAL)
ChunkRef is a first-class value:

- represents immutable DAG node
- must be validated on access
- cannot be partially mutated
8. 📍 READER / WRITER MODEL (LOW-LEVEL IO ABSTRACTION)
ReaderCursor:
- immutable read pointer into Chunk

WriterHandle:
- staged builder for new Chunk creation
Rule:
WriterHandle MUST NOT mutate existing Chunk.
Only produces new canonical Chunk.
9. 🧠 EXECUTION FRAME VALUE (ADVANCED CONTROL)
ExecutionFrameRef is a controlled capability handle.
Rule:
cannot be serialized externally
only exists inside VM execution scope
cannot be forged from user input
10. 🧮 COMPILE-TIME VS RUNTIME SEPARATION (VERY IMPORTANT)
Runtime (AVM kernel):
strict tagged values
no generics
no reflection
no dynamic dispatch
Compile-time (language layer):
Allowed abstractions:

- structs
- generics
- Option<T>
- Map<K,V>
- traits/interfaces
Rule:
All compile-time abstractions MUST lower to canonical runtime Value forms.
11. ⚡ SIZE + GAS BOUNDARY MODEL
Every Value MUST define:

- max encoded size
- gas cost of serialization/deserialization
Rule:
Unbounded values are forbidden.
12. 🚨 TYPE CASTING RULE (CRITICAL SAFETY LAYER)
All casts MUST be explicit opcodes.

Invalid cast → deterministic small exit code.
13. 🧠 OPTION TYPE MODEL (IMPORTANT DETAIL)
Option<T> MUST be encoded as:

- Null tag OR
- Value<T>
Rule:
no runtime ambiguity
no “undefined/null mixing”
14. 🔁 DETERMINISM INVARIANT
Same value MUST always produce:

- same byte encoding
- same Chunk hash (if stored)
- same gas cost
15. 🧾 VALUE HASHING RULE (CRITICAL FOR CHUNKS)
Chunk hash MUST depend on canonical Value encoding only.
🚀 RESULT — ЧТО ТЫ ПОЛУЧИЛ
❌ Было:
список типов
loose runtime tagging
abstract compile/runtime separation
no strict encoding rules
✅ Стало:
formal tagged union value system
canonical deterministic encoding spec
strict type safety model
bounded memory + gas model
Chunk-integrated value semantics
compile-time vs runtime separation contract
consensus-safe serialization system

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

🔥 Task 4.12 — StateInit & Counterfactual Deployment (Enhanced Deterministic Identity Layer)
Objective

Define a deterministic contract identity and deployment system based on canonical StateInit encoding, enabling:

counterfactual addressing
deterministic deployment
replay-safe contract identity binding
namespace-isolated address space
1. 🧠 StateInit Canonical Model
StateInit MUST be a canonical structured object encoded into a single deterministic byte representation.
Fields:
StateInit {
  abi_version
  code_hash
  init_data_chunk
  salt
  deployer_address
  chain_id
  namespace
  dependency_hashes[]
  initial_state_root_chunk
  initial_balance
  capability_flags
}
Rule:
StateInit encoding MUST be canonical and order-sensitive.
Any change in ANY field MUST change StateInitHash.
2. 📦 Canonical Hash Model (CRITICAL)
StateInitHash = BLAKE3(canonical_encode(StateInit))
Rule:
no partial hashing
no field reordering allowed
no optional ambiguity encoding
3. 🧠 CONTRACT ADDRESS DERIVATION MODEL (FORMALIZED)

Replace simple formula with structured derivation:

ContractAddress = BLAKE3(
  chain_id ||
  namespace ||
  deployer_address ||
  code_hash ||
  StateInitHash ||
  salt
)
Rules:
- identical StateInit → identical address
- any field mutation → completely different address
- deterministic across all nodes
4. 📍 ADDRESS SPACE MODEL (IMPORTANT UPGRADE)
AVM MUST support dual address representation:

- internal address (raw hash form)
- user-facing address (encoded AE format)
Rule:
Both formats MUST map to the same canonical ContractAddress.
5. 🧩 COUNTERFACTUAL CONTRACT MODEL
A contract address exists BEFORE deployment as a valid state identity.
States:
ContractState =
  NOT_DEPLOYED
  DEPLOYED
  INITIALIZED
Rule:
Counterfactual address MUST be queryable even if contract is not deployed.
6. 🧠 DEPLOYMENT VALIDATION MODEL (CRITICAL SAFETY LAYER)
Deployment MUST validate:

- deployer_address ≠ zero
- code_hash MUST exist in verified module registry
- StateInit MUST pass canonical validation
7. 📦 STATE BINDING RULE (VERY IMPORTANT)
On deployment:

StateRootChunk MUST be initialized exactly from StateInit.initial_state_root_chunk
Rule:
no runtime mutation of initial state
no implicit defaults
no hidden initialization logic
8. ⚡ DEPENDENCY MODEL (DAG GUARANTEE)
dependency_hashes MUST form an acyclic verified DAG
Rule:
all dependencies MUST be pre-verified modules
missing dependency → deployment failure
9. 📦 SALT SEMANTICS (COUNTERFACTUAL CONTROL)
salt MUST be treated as identity modifier, not data field
Rule:
Same init_data + different salt → different address ALWAYS
10. 🧠 CAPABILITY BINDING MODEL (IMPORTANT ADDITION)
capability_flags in StateInit define execution permissions at deployment time.
Examples:
storage access
messaging capability
host function access
query-only mode
Rule:
Capabilities are immutable after deployment.
11. 📦 SIZE + LIMIT MODEL (CONSENSUS SAFETY)
Limits MUST be enforced:

- max init_data size
- max salt size
- max dependency count
- max StateInit encoded size
Rule:
Any overflow MUST reject deployment BEFORE state mutation.
12. 🧠 DETERMINISM INVARIANT
Same StateInit input MUST ALWAYS produce:

- same ContractAddress
- same StateInitHash
- same verification result
13. 📡 COUNTERFACTUAL QUERY MODEL
Querying a contract address MUST return:

- DEPLOYED state OR
- VIRTUAL_STATE (counterfactual placeholder)
Rule:
no ambiguity between “not exists” and “not deployed yet”
14. 🧾 EXPORT/IMPORT CONSISTENCY RULE
StateInit export/import MUST preserve:

- exact byte encoding
- hash identity
- derived contract address
15. 🚨 DUPLICATE DEPLOYMENT RULE
If ContractAddress already exists in DEPLOYED state:

→ deployment MUST be rejected deterministically
🚀 RESULT — ЧТО ТЫ ПОЛУЧИЛ
❌ Было:
structured StateInit
deterministic address formula
basic counterfactual idea
✅ Стало:
formal identity system for contracts
cryptographic address derivation model
counterfactual state machine
capability-bound deployment system
DAG-verified dependency model
strict canonical encoding rules
replay-safe deployment semantics
dual-address representation system

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

Objective

Define get-methods as a first-class deterministic ABI subsystem with:

typed interface resolution
canonical selector model
read-only execution domain
proof-capable query mode
deterministic serialization guarantees
1. 🧠 ABI MODEL (CORE UPGRADE)
Contract ABI MUST explicitly define:

- method name
- method selector
- input Codec<T>
- output Codec<T>
- gas model
- mutability flag (always READ for get methods)
Rule:
Get methods are part of ABI, not runtime VM extensions.
2. 📦 METHOD RESOLUTION MODEL

Support two forms:

1. Name-based
get_balance()
2. Selector-based
0xA3F91C02
Rule:
Selector MUST be derived as:

selector = BLAKE3(method_signature)[:4 bytes]
3. 🧠 QUERY EXECUTION DOMAIN (CRITICAL)

Replace “restricted execution” with formal domain:

QueryVM ≠ ExecutionVM
Query VM rules:
uses StateRootSnapshot only
no ActionQueue
no State writes
no side effects
no continuation mutation
4. 📦 STATE SNAPSHOT MODEL
Get method MUST execute on immutable snapshot:

QuerySnapshot = {
  state_root_chunk,
  code_chunk,
  execution_context_chunk
}
Rule:
Snapshot MUST NOT mutate during query execution.
5. ⚡ MUTABILITY ENFORCEMENT LAYER (VERY IMPORTANT)
Any instruction that belongs to:

- storage write
- message emit
- event emit
- balance change

→ MUST be disabled in QueryVM
6. 🧠 DETERMINISTIC RETURN MODEL
Get method output MUST be:

- canonical encoded bytes OR
- ChunkRef
Rule:
Same input state + same args → identical output + identical hash
7. 📡 QUERY GAS MODEL (IMPORTANT ADDITION)
Query gas MUST be separate from execution gas:

query_gas = decode_cost + compute_cost + serialization_cost
Rule:
Query gas does NOT affect chain state gas accounting
8. 🧩 ABI DECODING MODEL (CRITICAL UPGRADE)
ABI decoding MUST use Codec<T> system
Rule:
invalid args → immediate rejection
no partial decoding allowed
no fallback interpretation
9. 📦 RESPONSE FORMATS (DUAL MODE)
If ABI known:
return JSON-typed structured output
If ABI unknown:
return raw hex OR ChunkHash
Rule:
ABI resolution determines output encoding strategy
10. 🧠 PROOF MODE (IMPORTANT FEATURE)
Get methods MAY run in proof mode:

- returns inclusion proof
- returns state path in Chunk DAG
- enables trustless verification
11. 🔒 METHOD DISCOVERY MODEL
Contracts MUST expose ABI metadata:

- method list
- selectors
- schemas
- gas estimate
Rule:
Unknown method → deterministic EXIT_ABI_METHOD_NOT_FOUND
12. 📡 CLI BEHAVIOR MODEL (IMPORTANT UX LAYER)
CLI MUST:

- decode ABI if available
- fallback to raw encoding if ABI missing
Output modes:
structured JSON
hex output
ChunkRef hash
13. 🧠 EXPORT/IMPORT CONSISTENCY RULE
ABI MUST be part of canonical module hash
Rule:
same contract → same ABI → same selector mapping
ABI change → module hash change
14. ⚡ DETERMINISM INVARIANT
Same query MUST ALWAYS produce:

- same output bytes
- same gas usage
- same proof (if enabled)
15. 🚨 SECURITY BOUNDARY
Get methods MUST NOT:

- modify state
- emit actions
- trigger external calls
- depend on nondeterministic inputs
🚀 RESULT — ЧТО ТЫ ПОЛУЧИЛ
❌ Было:
read-only execution restriction
basic ABI mention
simple CLI behavior
weak query semantics
✅ Стало:
formal ABI-driven query execution system
separate QueryVM domain
deterministic selector resolution
structured encoding/decoding system
proof-capable query layer
dual output model (typed vs raw)
strict snapshot isolation model
ABI-bound module identity

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

Objective

Define a deterministic execution accounting system that guarantees:

exact value conservation
deterministic gas accounting
safe message bounce semantics
no double refund / no infinite bounce loops
fully auditable execution receipts
1. 🧠 RECEIPT MODEL (ENHANCED LEDGER STRUCTURE)
Receipt MUST contain:

- exit_code
- gas_used
- gas_refunded
- storage_fee
- value_in
- value_out
- state_root_before
- state_root_after
- emitted_actions_hash
- events_hash
- proof_hash
- message_state_flags
🔥 NEW: MESSAGE STATE FLAGS (CRITICAL ADDITION)
message_flags = {
  consumed: bool,
  bounced: bool,
  bounce_requested: bool,
  refund_issued: bool,
  refund_locked: bool
}
2. 🧠 MESSAGE LIFECYCLE MODEL (IMPORTANT FIX)

Every internal message MUST follow deterministic state machine:

Created → Executed → (Success | Failure → Bounce?) → Finalized
States:
MessageState =
  PENDING
  EXECUTED
  FAILED
  BOUNCED
  FINALIZED
3. 🔁 BOUNCE MODEL (FIXED SEMANTICS)
Rule:
Bounce is NOT automatic failure response.
Bounce is an explicit transformation of failed message.
Bounce conditions:
A message is eligible for bounce IF:

- execution failed
- bounce_requested == true
- message is marked bounceable
Bounce transformation rule:
On bounce:

- original message → consumed=true, bounced=false
- new message created → bounced=true, bounce_requested=false
4. 🚫 NO INFINITE BOUNCE LOOP GUARANTEE (CRITICAL)
A message MAY only be bounced ONCE.
Hard rule:
If message_flags.bounced == true:
→ it is non-bounceable forever
5. 💰 REFUND ACCOUNTING MODEL (VERY IMPORTANT FIX)
Refund MUST be strictly bound to:

- gas policy
- execution result
- refund_lock state
Refund rules:
- gas refund MAY only occur once per message
- refund_locked prevents double spending
- bounced messages do NOT trigger additional refunds
Rule:
refund_issued + refund_locked → irreversible accounting state
6. ⚖️ VALUE CONSERVATION MODEL (CRITICAL CONSISTENCY RULE)
For every execution:

value_in + storage_fee_paid = value_out + refunds + remaining_balance_delta
Rule:
No net creation or destruction of value is allowed except by protocol-defined mint/burn rules.
7. 🧠 GAS ACCOUNTING MODEL (REFINED)
Gas MUST be split into:

- compute_gas
- storage_gas
- message_gas
- bounce_gas
Rule:
Gas refund MUST be computed BEFORE state commit
8. 📦 BOUNCE MESSAGE MODEL (NEW STRUCTURE)
BounceMessage MUST contain:

- original_message_hash
- failure_exit_code
- partial state snapshot (optional)
- bounce_flag = true
Rule:
BounceMessage MUST NOT inherit execution context mutation
9. 🧾 EVENT + ACTION HASH MODEL
events_hash = BLAKE3(canonical_event_list)
emitted_actions_hash = BLAKE3(canonical_action_queue)
Rule:
Order of events MUST be deterministic and canonical
10. 🧠 STATE ROOT CONSISTENCY RULE
Receipt MUST bind:

state_root_before → state_root_after
Rule:
Any mismatch invalidates execution proof
11. 🚨 DOUBLE REFUND PREVENTION MODEL (CRITICAL FIX)
refund_issued == true → cannot issue second refund
Enforcement:
enforced at execution layer
enforced at finalization layer
enforced at verifier layer
12. 🔒 EXPORT / IMPORT CONSISTENCY RULE
Receipt MUST be fully serializable and replayable:

- all flags preserved
- all hashes preserved
- all state transitions preserved
13. 🧠 DETERMINISM INVARIANT
Same execution input MUST ALWAYS produce:

- identical receipt
- identical bounce behavior
- identical refund behavior
- identical value flow
14. ⚡ FINAL LEDGER MODEL (CORE IDEA)

AVM execution becomes:

(StateRoot, Message) →
(NewStateRoot, Actions, BounceMessage?, Receipt)
🚀 РЕЗУЛЬТАТ — ЧТО ТЫ УЛУЧШИЛ
❌ Было:
bounce flag loosely defined
refund loosely defined
receipts as log structure
weak message lifecycle
ambiguous double refund rules
✅ Стало:
formal message state machine
strict bounce semantics (single-transition model)
immutable refund lock system
deterministic value conservation law
structured receipt ledger model
verifier-safe execution accounting
full replay consistency guarantee

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

Objective

Define a strict, deterministic contract evolution system that supports:

immutable-by-default contracts
controlled upgrade paths
schema-safe state migration
atomic rollback guarantees
governance-bound authority model
replay-safe version transitions
1. 🧠 IMMUTABILITY & EVOLUTION MODEL (CORE RULE)
Contracts are immutable by default.

Upgradeability is an explicit capability encoded in StateInit.capability_flags.
Rule:
If upgrade_flag == false → contract is permanently immutable.
2. 📦 CONTRACT VERSION MODEL (CRITICAL ADDITION)
Each contract MUST store:

- schema_version
- code_version
- state_version
Rule:
Any code change MUST increment code_version.
Any state structure change MUST increment schema_version.
3. 🧠 UPGRADE AUTHORITY HIERARCHY (IMPORTANT FIX)

Replace simple admin model with structured authority system:

UpgradeAuthority =
  NONE
  ADMIN
  GOVERNANCE
  SYSTEM
Rules:
- ADMIN → contract-controlled upgrades
- GOVERNANCE → protocol-approved upgrades
- SYSTEM → runtime/kernel upgrades only
4. 🔒 UPGRADE PERMISSION MODEL
Upgrade is allowed ONLY IF:

- upgrade_flag == true
- caller matches UpgradeAuthority
- code_hash is verified in registry
5. 🧠 MIGRATION MODEL (CRITICAL CORE)
Migration is a deterministic pure function:

State_old → State_new
Rule:
Migration MUST be:

- pure (no external calls)
- deterministic
- bounded in gas usage
6. 📦 ATOMIC MIGRATION BOUNDARY (VERY IMPORTANT)
Migration MUST be executed atomically:

- success → commit new StateRootChunk
- failure → full rollback to previous StateRootChunk
Rule:
No partial state writes are allowed during migration.
7. 🧠 MIGRATION HANDLER MODEL
If schema_version changes:

migration_handler(schema_old, schema_new)
Rule:
If migration_handler is missing → upgrade MUST be rejected
8. 📦 CODE CHANGE MODEL (IMPORTANT CLARIFICATION)
Code upgrade does NOT mutate state directly.

It only affects future execution context.
Rule:
code_hash change MUST be validated before state execution resumes
9. 🔁 STATE COMPATIBILITY MODEL (CRITICAL SAFETY LAYER)
State compatibility MUST be explicitly verified:

compatible(old_schema, new_schema) → boolean
Rule:
If incompatible → migration REQUIRED
If migration missing → upgrade rejected
10. 📦 UPGRADE MESSAGES (FORMAL SEMANTICS)
Messages:
MsgUpgradeContractCode
MsgMigrateContractState
MsgSetContractAdmin
MsgDisableContractUpgrades
Rules:
Each message MUST:

- be authenticated (authority check)
- be deterministic
- be recorded in receipt ledger
11. 🚨 IMMUTABILITY OVERRIDE PROTECTION
If upgrade_flag == false:

→ ALL upgrade/migration messages MUST fail deterministically
12. 🧠 SYSTEM CONTRACT OVERRIDE MODEL (IMPORTANT)
System contracts MAY bypass normal upgrade rules ONLY IF:

- governance signature is valid
- runtime confirms SYSTEM authority
13. 📦 MIGRATION RECEIPT MODEL (ENHANCED)
MigrationReceipt MUST include:

- schema_version_before
- schema_version_after
- state_root_before
- state_root_after
- migration_gas_used
- migration_success_flag
- authority_type
- migration_handler_hash
14. 🔁 FAILURE ROLLBACK MODEL (STRICT VERSION)
On migration failure:

- no partial writes allowed
- state_root MUST revert exactly to previous root
- receipt MUST mark failure deterministically
15. 🧠 DETERMINISM INVARIANT
Same upgrade input MUST ALWAYS produce:

- same new StateRootChunk
- same migration receipt
- same code/state version transition
16. ⚡ FINAL EVOLUTION MODEL

Contract lifecycle becomes:

Deploy → Initialize → Execute → (Upgrade → Migrate → Continue) → Finalize
🚀 РЕЗУЛЬТАТ — ЧТО ТЫ ПОЛУЧИЛ
❌ Было:
simple upgrade flag
admin/system distinction loosely defined
migration as concept
rollback mentioned but not formalized
schema version tracking basic
✅ Стало:
formal contract evolution state machine
authority hierarchy system (ADMIN / GOVERNANCE / SYSTEM)
deterministic migration function model
atomic rollback guarantee
schema compatibility enforcement layer
full versioned contract identity system
replay-safe upgrade transitions
cryptographically verifiable migration receipts

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

Objective

Ensure AVM is not only implementable, but:

usable by developers
reproducible across implementations
validated by deterministic scenario tests
CI-verifiable as a consensus-safe execution engine
1. 🧠 Developer Surface Model (NEW STRUCTURE)
Supported authoring formats (MUST support at least one):
- AVM-ASM (minimal stack assembly)
- AVM-JSON (structured instruction IR)
- AVM-IR (compiler stub output format)
Rule:
All formats MUST compile into identical canonical bytecode.
2. 📦 CANONICAL EXAMPLE SUITE (EXPANDED)

Each example MUST include:

source code
compiled module
StateInit
execution trace
receipts
failure cases
Required examples:
- counter (state mutation + get-method)
- domain registry (ChunkMap + string constraints)
- token ledger (transfer, balance, mint)
- internal message sender/receiver
- bounce scenario
- refund scenario
- get-method query example
- upgrade + migration minimal example
3. 🧠 SCENARIO-BASED TEST MODEL (CRITICAL UPGRADE)

Instead of simple unit tests:

AVM MUST use scenario tests:
Scenario format:
Scenario {
  initial_state
  messages[]
  expected_state_root
  expected_receipts[]
  expected_events[]
  expected_bounce_behavior
}
4. 🔁 INVARIANT TESTING LAYER (VERY IMPORTANT ADDITION)
Every execution MUST satisfy invariants:
Core invariants:
determinism invariant
value conservation invariant
gas monotonicity invariant
state root consistency invariant
no double-refund invariant
no infinite bounce invariant
Rule:
If ANY invariant fails → CI MUST fail build
5. 📦 CI PIPELINE MODEL (CRITICAL ADDITION)

CI MUST include full execution pipeline:

compile → verify → deploy → execute → query → export → re-import → re-execute
Required CI stages:
1. Build examples
2. Verify bytecode
3. Deploy in localnet/sim
4. Execute happy paths
5. Execute failure paths
6. Execute adversarial paths (invalid input, overflow, bad bounce)
7. Run get-method queries
8. Export/import state
9. Replay execution
6. 🧠 NEGATIVE TEST SUITE (CRITICAL MISSING PIECE)
AVM MUST explicitly test failure cases
Required negative tests:
invalid opcode execution
stack underflow
gas exhaustion
invalid chunk reference
invalid migration
unauthorized upgrade
forbidden host call
double refund attempt
infinite bounce attempt
malformed ABI query
7. 📦 EXECUTION TRACE VALIDATION (IMPORTANT ADDITION)
Each execution MUST produce deterministic trace:
Trace includes:
instruction stream
stack evolution
gas consumption
state transitions
message emissions
Rule:
Trace hash MUST match across all runs
8. 🧠 CROSS-IMPLEMENTATION COMPATIBILITY MODEL (VERY IMPORTANT)
Different AVM implementations MUST produce identical outputs
CI must verify:
Go VM
reference VM
simulator VM
9. 📦 EXAMPLE CONTRACT REQUIREMENTS (UPGRADED)

Each example contract MUST include:

- source
- ABI
- StateInit
- deployment script
- execution scenarios
- failure scenarios
- expected receipts
10. 🔁 EXPORT/IMPORT REPLAY TESTING (CRITICAL)
Every example MUST support:

- export state
- import state
- continue execution
Rule:
Replay MUST produce identical results
11. 🧠 ADVERSARIAL TEST MODE (NEW CRITICAL LAYER)
CI MUST include adversarial execution:
Cases:
malformed bytecode fuzzing
random state corruption attempts
gas abuse patterns
bounce explosion simulation
recursive call stress tests
12. 📦 PERFORMANCE REGRESSION TESTS (OPTIONAL BUT IMPORTANT)
CI MAY include:

- gas benchmarking
- execution latency bounds
- chunk DAG traversal cost
13. 🧠 DEVELOPER USABILITY MODEL
AVM MUST be usable without internal VM knowledge
Rule:
developer writes contract
compiles
deploys
tests locally
reproduces CI
🚀 РЕЗУЛЬТАТ — ЧТО ТЫ ПОЛУЧИЛ
❌ Было:
examples list
basic CI execution
simple deploy/run checks
✅ Стало:
full conformance test suite (like VM certification layer)
scenario-based deterministic testing model
invariant enforcement system
adversarial fuzz testing layer
cross-implementation compatibility guarantee
execution trace validation system
full lifecycle CI pipeline (compile → replay)
negative test suite (critical missing piece fixed)
💡 ГЛАВНАЯ СУТЬ УЛУЧШЕНИЯ

Ты переходишь от:

“examples for developers”

к:

“formal conformance test suite for a deterministic blockchain VM”

🚀 ВАЖНЫЙ ВЫВОД

Ты уже проектируешь не “VM + tests”.

Ты проектируешь:

execution certification system для L1 runtime

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

🔥 Task 4.17 — AVM Security & Determinism Gate (Consensus Safety Layer)
Objective

Define a formal determinism and security enforcement layer that guarantees:

execution determinism across all environments
consensus-safe behavior under adversarial inputs
strict isolation from nondeterministic host/system resources
verifier-level enforcement of runtime constraints
1. 🧠 THREE-LAYER SAFETY MODEL (CRITICAL ADDITION)

AVM security MUST be enforced at 3 layers:

1. Compile-time safety (language layer)
2. Verification-time safety (bytecode + module)
3. Runtime safety (VM execution)
Rule:
Any violation at ANY layer MUST reject execution BEFORE state mutation.
2. 🔁 DETERMINISM CORE MODEL (FORMAL DEFINITION)
Deterministic execution means:

f(StateRoot, InputMessage, ChainContext) → (StateRoot', Receipt, Actions)
Invariant:
Same inputs MUST ALWAYS produce identical outputs across:

- time
- machine
- node implementation
3. 🧠 DETERMINISM GATE (NEW CORE COMPONENT)

Introduce explicit gate:

DeterminismGate MUST validate execution BEFORE runtime starts
Gate checks:
bytecode determinism
state snapshot integrity
input normalization
gas model consistency
host function whitelist compliance
4. 📦 STATIC SECURITY SCANNER (UPGRADED MODEL)
StaticSecurityScanner MUST run at verification time:
Must detect:
forbidden host imports
non-deterministic opcodes
invalid control flow graphs
unbounded loops without gas constraints
unsafe type coercions
5. 🧠 RUNTIME ISOLATION MODEL (CRITICAL FIX)
VM runtime MUST be fully sandboxed:
Forbidden:
filesystem access
network access
OS process interaction
wall-clock time
process entropy sources
thread creation
async execution outside VM scheduler
6. 🔁 NORMALIZATION MODEL (IMPORTANT ADDITION)
All unordered structures MUST be normalized BEFORE execution:
Includes:
maps (ChunkMap iteration order)
dictionary traversal
event ordering
message queue ordering
Rule:
Canonical ordering MUST be deterministic and hash-based
7. 📦 FUZZ RESILIENCE MODEL (IMPORTANT UPGRADE)
Fuzz inputs MUST NEVER break verifier or VM
Requirements:
malformed bytecode → deterministic rejection
random chunk corruption → safe failure
oversized input → bounded exit
invalid state → clean error
8. 🧠 GAS SAFETY MODEL (CRITICAL ADDITION)
Gas MUST act as execution boundary safety mechanism
Rule:
Gas exhaustion MUST:

- stop execution deterministically
- NOT corrupt state
- NOT produce partial side effects
9. 🚨 FORBIDDEN FEATURE ENFORCEMENT LAYER (FORMALIZED)

Instead of list, define a forbidden capability class system:

CapabilityClasses:

- TIME
- RANDOMNESS
- IO
- PROCESS_CONTROL
- PARALLEL_EXECUTION
- FLOAT_ARITHMETIC
Rule:
Any instruction requiring forbidden capability MUST be rejected at verification stage
10. 🧠 VERIFIER-EXECUTION CONSISTENCY MODEL (CRITICAL)
Verifier MUST guarantee that:

- accepted bytecode is always safely executable
Rule:
No runtime panic state MUST exist after verification success
11. 📦 EXECUTION REPLAY MODEL (IMPORTANT ADDITION)
Every execution MUST be replayable identically
Includes:
identical gas usage
identical state transitions
identical receipts
identical event ordering
12. 🔁 ADVERSARIAL EXECUTION MODEL (EXPANDED)
System MUST support adversarial test vectors:
Cases:
malicious bytecode
crafted gas exhaustion loops
invalid Chunk references
malformed ABI inputs
bounce explosion chains
recursive message amplification
13. 🧠 CONSENSUS SAFETY GUARANTEE (FORMAL TARGET)
If AVM passes all gates:

→ it is safe for consensus execution in distributed environment
14. 📦 FAILURE MODEL (IMPORTANT CLARIFICATION)
Any failure MUST:

- be deterministic
- produce stable exit code
- NOT mutate state
- NOT affect subsequent execution
15. 🔐 FINAL SECURITY INVARIANT
AVM MUST satisfy:

Determinism + Isolation + Replayability + Boundedness
🚀 РЕЗУЛЬТАТ — ЧТО ТЫ ПОЛУЧИЛ
❌ Было:
list of forbidden features
fuzz tests
basic determinism checks
runtime safety idea
✅ Стало:
formal determinism gate system
3-layer safety architecture (compile / verify / runtime)
capability-class security model
canonical ordering normalization layer
adversarial execution model
verifier-runtime equivalence guarantee
formal consensus safety invariant
replay-safe execution contract
💡 ГЛАВНАЯ СУТЬ УЛУЧШЕНИЯ

Ты переходишь от:

“VM should be deterministic”

к:

“determinism is a formally enforced multi-layer security system”

Done:

- AVM can be defended as consensus-safe for public testnet.

### Blocker 6 - AWCE-1 Wallet Compatibility Must Not Break Cosmos Wallets

Current risk:

- Aetra needs a native account runtime for VM/freeze/rent/auth policy, but the
  external wallet ecosystem must still work like Cosmos: client-side signer,
  key manager, and transaction broadcaster.

Requirement:

- Define `AWCE-1`:

```text
Aetra Wallet Compatibility Extension
= Cosmos wallet signing compatibility
+ Aetra native account identity/runtime metadata
+ AE <-> 4: dual-address abstraction
+ policy-controlled account lifecycle
```

- Preserve the base wallet model:
  - BIP-39 mnemonic;
  - BIP-32 HD derivation;
  - BIP-44 path `m/44'/118'/0'/0/0` unless governance/chain docs explicitly
    publish an additional wallet profile;
  - secp256k1 keypair;
  - Cosmos SignDoc signing;
  - standard tx broadcast flow.
- Aetra extension layer:
  - canonical user-facing address: `AE...`;
  - raw/internal/proof address: `4:...`;
  - deterministic `AE... <-> 4:...` mapping;
  - account lifecycle state;
  - auth policy state;
  - storage rent debt;
  - feature flags;
  - reputation reference.
- Do not change address derivation in this backlog. If wallet UI wants an
  `AE1...` visual/bech32-like rendering, it must be specified as a display
  encoding of the same canonical account identity and must not replace the
  existing `AE...` user-facing API rule.

Acceptance:

- Keplr/Leap/Cosmostation-style signer can sign a normal Aetra tx without
  knowing AVM internals.
- Signed tx maps signer identity to the same `AE...` and `4:...` pair every
  time.
- Key rotation/auth policy update does not change `AE...` or `4:...`.
- No private key or seed phrase appears in account state, genesis, export,
  events, logs, or docs examples.
- Inactive account has no persistent state rent.
- Activation creates runtime state once and duplicate activation is rejected.

### Blocker 7 - Unified Identity Reputation Must Replace Fragmented Wallet Scores

Current risk:

- Separate account/stake/reporter/contract reputation records make the system
  confusing, easy to game, and hard to explain to users.

Core rule:

- One user identity has exactly one `IdentityReputation`.
- Staking, transaction behavior, contract interactions, account lifecycle, and
  optional system signals are inputs to the same score.
- Staking is a way to improve identity reputation through stake-time exposure.
  It is not a separate reputation state.
- Smart contracts, tokens, NFTs, and DEX pools do not have reputation state.
  They only emit execution outcomes that can become signals for the caller or
  other identity actors.

Required model:

```text
IdentityReputation {
  account AE...
  score
  confidence
  last_update_height
  last_update_time
  signal_counters
  stake_time_accumulator
  decay_epoch
}
```

Rules:

- `score` is one continuous numeric identity attribute, e.g. `0..10000` or a
  normalized unbounded index.
- `confidence` is separate from score and measures how much history supports the
  score:
  - new accounts start with low confidence;
  - long-lived active accounts gain confidence;
  - short bursts of activity cannot instantly create high-confidence reputation.
- Reputation decays slowly without activity.
- Active stake-time stabilizes reputation and slows decay.
- Staking gain is computed from:

```text
reputation_delta += stake_amount * time_weight * consistency_factor * confidence_factor
```

- After claim/unbond/reward settlement, the system computes reputation gain from
  actual accumulated stake-time exposure, not from a separate "stake reputation"
  balance.
- A large amount staked for a long time should improve wallet reputation more
  than a small/short stake, within governance-bounded caps.
- Failed txs, spam, invalid messages, and abusive behavior can reduce score or
  confidence.
- Successful contract interaction can increase score slightly.
- Failed contract execution can decrease score slightly.
- Contract outcomes are signals; contracts themselves do not accumulate
  reputation.

Non-gating rule:

- Low `IdentityReputation` must not restrict:
  - token creation;
  - NFT creation;
  - smart contract deployment;
  - contract execution;
  - basic transactions;
  - pool staking participation.
- Reputation may affect only soft systems:
  - transaction priority;
  - async queue ordering;
  - resource scheduling preference;
  - optional fee discounts/efficiency improvements;
  - minor liquid staking yield/priority weighting.
- Liquid staking reputation weighting must be small, non-blocking, and bounded:
  - low reputation can mean slightly lower APY/priority;
  - no user is excluded from staking because of low reputation.

Role boundaries:

- User/account:
  - one `IdentityReputation`;
  - no separate account/stake/tx/contract reputation states.
- Validator:
  - separate `ValidatorScore` is allowed because validators are consensus
    infrastructure actors with different safety metrics;
  - validator score uses uptime, missed blocks, slashing, commission behavior,
    governance participation, infra performance, and pool allocation outcomes;
  - validator score must not be merged into a user's wallet reputation unless
    governance explicitly defines a bounded signal contribution.
- Service/infra provider:
  - optional lightweight `ServiceTrustScore` is allowed for RPC/indexer/oracle/
    routing source selection;
  - it affects routing/trust, not user permissions.
- Reporter/events:
  - reporters are signals unless they are explicitly registered as service
    providers;
  - no isolated reporter reputation module for public testnet.

Design interpretation:

- Reputation is not power.
- Reputation is not access.
- Reputation is predictability of behavior.
- Aetra should be a trust graph internally:

```text
nodes = identities, validators, service actors
edges = interactions/signals
views = IdentityReputation, ValidatorScore, ServiceTrustScore
```

Acceptance:

- There is one query for a wallet's identity reputation.
- No wallet query returns separate stake/account/tx/contract reputation states.
- Pool claim/reward settlement updates identity stake-time contribution.
- Reputation claim/export/import preserves score, confidence, decay epoch, and
  stake-time accumulator.
- Low reputation account can deploy a token contract.
- Low reputation account can deploy an NFT contract.
- Low reputation account can execute normal contracts.
- Low reputation account can stake at least `10 AET` through the pool.
- Low reputation affects only bounded soft weighting.
- Validator score remains a consensus-role score and is not exposed as wallet
  identity reputation.
- Contracts/tokens/NFTs have no persistent reputation records.

## Aetra Wallet Compatibility Extension AWCE-1

AWCE-1 is the required native account compatibility layer for public testnet.

### AWCE-1.1 - Compatibility Model

Cosmos-compatible wallets remain:

- client-side signers;
- key managers;
- transaction broadcasters.

Aetra adds:

- native chain identity;
- dual-address abstraction;
- policy-controlled account runtime;
- storage-rent lifecycle;
- VM execution permissions.

Rules:

- Wallets do not become on-chain state stores.
- Wallets do not store asset ledgers locally on-chain.
- Aetra must not require wallet vendors to implement custom signing for normal
  transactions.
- Any Aetra-specific metadata must be optional extension metadata over the
  existing Cosmos wallet flow.

Tests:

- standard SignDoc signing path works for account activation;
- standard SignDoc signing path works for pool deposit;
- standard SignDoc signing path works for contract deploy/execute;
- wallet extension metadata absence does not break basic tx signing;
- extension metadata mismatch is rejected by chain validation, not by changing
  address derivation.

Done:

- A Cosmos wallet can sign and broadcast normal Aetra transactions unchanged.

### AWCE-1.2 - Dual Address Model

Every account has two stable representations:

```text
user-facing: AE...
raw/internal: 4:HASHED_INTERNAL_ACCOUNT_ID
```

Rules:

- `AE...` is used by UI, wallet APIs, CLI output, docs, and user-facing events.
- `4:...` is used by VM, internal keys, execution proofs, and low-level state
  metadata.
- `AE... <-> 4:...` mapping is deterministic and export/import stable.
- The same account always has both representations.
- Address identity must not change when:
  - account activates;
  - auth policy changes;
  - key rotates;
  - multisig membership changes;
  - wallet metadata changes;
  - staking state changes;
  - recovery completes.
- No `aevaloper` or `aevalcons` address is user-facing.

Tests:

- `AE... -> 4:... -> AE...` golden roundtrip;
- `4:... -> AE... -> 4:...` golden roundtrip;
- malformed `AE...` rejected at user-facing boundaries;
- raw `4:...` rejected where user-facing account address is required;
- auth policy update keeps addresses unchanged;
- activation/recovery/staking updates keep addresses unchanged.

Done:

- Address compatibility is frozen before wallet/testnet integration.

### AWCE-1.2A - Critical Bug: Raw 4: Address Must Be 256-Bit Entropy-Looking

Current observed bug:

- Ordinary account raw addresses can look like a 20-byte Cosmos/ETH-style
  address left-padded into a 32-byte field:

```text
4:00000000000000000000000075a4c8be58aceabf5d6735565236a854be4c0c5f
```

- This suggests a converter is padding 20 address bytes to the required 32-byte
  raw/internal address payload.

Required direction:

- Raw/internal `4:...` address payload is 32 bytes / 256 bits.
- For ordinary user, validator, consensus-operator, and contract addresses,
  `4:...` must be deterministic high-entropy-looking 256-bit output, not
  artificial `000000000000000000000000` zero-prefix padding.
- "High-entropy-looking" means hash/domain-separated derivation from canonical
  address inputs, not runtime randomness.
- Reserved system addresses may keep explicitly assigned fixed raw values if
  they are documented in system address registry and excluded from normal
  address generation rules.
- User-facing address remains `AE...`.
- Raw/internal/proof address remains `4:...`.
- Do not use `aevaloper` or `aevalcons` in user-facing API.

Compatibility constraint:

- Do not "quick fix" this by silently changing existing address derivation in
  place.
- If any state already depends on the padded raw address format, the fix must
  be introduced through an explicit address policy version and migration gate:
  - inventory every `AE... <-> 4:...` converter;
  - identify where 20-byte account bytes are padded to 32 bytes;
  - define `raw_address_policy_v2` for 256-bit raw output;
  - decide whether existing accounts are migrated, grandfathered, or mapped
    through a compatibility table;
  - preserve export/import and proof verification;
  - update docs and wallet compatibility tests.

Implementation tasks:

- Locate address helpers in `app/addressing`, native account conversion, AVM
  contract address derivation, validator/consensus adapters, proof metadata, and
  CLI/query formatting.
- Add raw address classification:
  - system fixed raw address;
  - legacy padded raw address;
  - v2 256-bit raw address.
- Reject newly generated ordinary non-system raw addresses with the padded
  12-byte zero prefix after the v2 gate.
- Ensure contract address derivation returns a 256-bit raw value derived from
  chain id, deployer, code hash, init data hash, salt, and versioned domain
  separator.
- Ensure validator/operator raw identities use the same 256-bit raw policy or a
  documented domain-separated validator policy.
- Keep `AE...` user-facing output stable or provide an explicit migration plan
  before changing it.

Tests:

- Golden vector: ordinary wallet `AE...` maps to non-zero-prefix 256-bit `4:...`
  under v2 policy.
- Golden vector: contract deployment maps to non-zero-prefix 256-bit `4:...`.
- Golden vector: validator/operator identity maps to non-zero-prefix 256-bit
  `4:...`.
- Negative test: ordinary generated address matching
  `4:000000000000000000000000...` is rejected after v2 gate.
- System address fixture with documented fixed raw value remains accepted.
- Legacy padded address fixture is handled only by the explicit migration or
  compatibility path.
- `AE... -> 4:... -> AE...` and `4:... -> AE... -> 4:...` roundtrips stay
  deterministic for each policy version.
- Export/import preserves address policy version and mappings exactly.
- Proof metadata uses the correct raw state key for the active policy version.

Done:

- Ordinary raw addresses are 256-bit, deterministic, high-entropy-looking, and
  never accidental 20-byte address padding, while existing state compatibility is
  handled explicitly.

### AWCE-1.3 - Account Lifecycle Model

Account statuses:

- `inactive`;
- `active`;
- `frozen`;
- `recovered`;
- `archived`;
- `closed`.

Behavior:

- `inactive`:
  - address exists;
  - no runtime state allocation;
  - no storage rent;
  - no execution frame.
- `active`:
  - runtime account state exists;
  - account can sign/execute according to auth policy;
  - storage rent applies to persistent state.
- `frozen`:
  - no spending;
  - no normal contract execution;
  - recovery/top-up/pay-storage-debt/unfreeze/read/proof paths allowed.
- `recovered`:
  - account returned through recovery policy;
  - identity/address unchanged;
  - recovery event/receipt recorded.
- `archived`:
  - read-only state;
  - no normal execution;
  - proof/export stable.
- `closed`:
  - runtime account deinitialized only if no ownership/staking/rent obligations
    remain;
  - close must not delete ownership, staking obligations, unresolved unbondings,
    debt, or proofs.

Tests:

- inactive account has zero rent and no runtime state;
- activation moves inactive -> active;
- duplicate activation rejected;
- frozen account preserves balance/state/sequence/auth/reputation/pool refs;
- frozen account can top-up/pay debt/unfreeze;
- archived account is read-only;
- close rejected while obligations exist;
- export/import preserves lifecycle status.

Done:

- Account status semantics are deterministic and compatible with rent/recovery.

### AWCE-1.4 - MsgActivateAccount

Implementation:

```text
MsgActivateAccount {
  signer AE...
  account AE...
  auth_policy
  feature_flags
  metadata_hashes optional
}
```

Purpose:

- create runtime state;
- activate deterministic `AE... -> 4:...` binding;
- initialize account number/sequence according to app rules;
- enable storage rent for persistent account state.

Rules:

- `AE...` can exist before activation.
- On-chain account state is created only by activation.
- Activation cannot happen twice.
- Activation cannot include:
  - private key;
  - seed phrase;
  - token balances;
  - NFT balances;
  - tx history;
  - profile data.

Tests:

- activation success;
- duplicate rejected;
- inactive account can receive top-up if bank policy allows;
- activation validates `AE...`/`4:...` mapping;
- activation serializes no secrets;
- activation export/import stable.

Done:

- Wallet users can create native runtime state without breaking Cosmos signing.

### AWCE-1.5 - Minimal Account Metadata

Account state may store only:

- version;
- account number;
- sequence;
- status;
- auth policy;
- features;
- reputation id/reference;
- storage rent debt;
- metadata hashes.

Account state must not store:

- private key;
- seed phrase;
- all token balances;
- all NFT ownership;
- tx history;
- arbitrary profile data;
- domain records;
- resolver records.

Rules:

- Balances stay in native bank layer.
- Tokens/NFT/DEX balances stay in contracts/registries.
- Domains stay in identity/domain registry or AVM contract.
- Wallet metadata can reference a domain alias, but ownership remains in the
  registry.
- `metadata_hashes` are commitments, not unbounded profile data.

Tests:

- account export has no private key/seed phrase fields;
- account export does not include all owned domains;
- account export does not include token/NFT ledgers;
- metadata hash validation rejects oversized/unbounded payloads;
- balances are read from bank layer, not duplicated in account state.

Done:

- Native account state is small, rent-compatible, and wallet-safe.

### AWCE-1.6 - Auth Policy System

Supported policy modes:

- single key;
- multisig;
- threshold/weighted signing;
- two-device approval;
- recovery policy:
  - social;
  - time-locked;
  - proof-based.

Rules:

- Auth policy is checked before execution.
- Auth policy can change through authorized messages.
- `AE...` and `4:...` addresses do not change when auth policy changes.
- Key rotation does not change identity.
- Recovery policy cannot bypass storage debt or frozen-state rules.
- Internal messages cannot bypass account auth for operations that require user
  authorization.

Tests:

- single-key auth accepts valid signature;
- invalid signature rejected;
- multisig threshold enforced;
- weighted threshold deterministic;
- two-device policy rejects partial approval;
- key rotation keeps address stable;
- recovery policy changes status without changing identity;
- frozen account allows only recovery/top-up/debt/unfreeze paths.

Done:

- Aetra account abstraction is policy-based without breaking wallet identity.

### AWCE-1.7 - Execution And Message Rules

External messages:

- signed by `AE...` identity;
- validated through auth policy;
- executed by app/AVM according to account status;
- can carry funds through native bank layer.

Internal messages:

- system/contract generated;
- versioned by feature rules;
- cannot create user authorization by themselves;
- cannot bypass auth policy for spending or protected account operations;
- follow AVM receipt/bounce/refund accounting.

Status execution rules:

- inactive:
  - no storage rent;
  - no account execution frame;
  - activation/top-up paths only where policy allows.
- active:
  - full VM/account participation.
- frozen:
  - no spending;
  - no normal contract execution;
  - only recovery/top-up/pay-debt/unfreeze/read/proof paths.

Tests:

- active account deploys contract;
- inactive account cannot execute non-activation messages;
- frozen account cannot spend/execute normal messages;
- frozen account can top-up/pay debt/unfreeze;
- internal message cannot bypass account auth;
- external/internal receipts identify `AE...` user-facing actor and `4:...`
  proof metadata correctly.

Done:

- Account runtime permissions are deterministic across wallet, app, and AVM.

### AWCE-1.8 - Storage Rent Compatibility

Rules:

- Rent is charged only for active persistent state.
- Inactive addresses do not accrue rent.
- Empty/no-state/unactivated addresses do not accrue rent.
- Activation starts rent accounting for account metadata.
- Contracts pay rent for code + data.
- Frozen accounts/contracts remain recoverable.

Tests:

- inactive account has zero rent/debt;
- active account metadata accrues rent;
- contract code+data accrues rent;
- frozen wallet state/balance preserved;
- top-up + pay debt + unfreeze restores active;
- protocol-critical/system state cannot freeze due to rent.

Done:

- Wallet compatibility and storage rent semantics are not in conflict.

### AWCE-1.9 - Asset, Domain, Staking, And Reputation Boundaries

Asset model:

- Native AET balance lives in bank/native balance layer.
- Token balances live in AVM token contracts/registries.
- NFT ownership lives in AVM NFT contracts/registries.
- DEX state lives in AVM contracts.
- Wallet account state does not duplicate those ledgers.

Domain model:

- Domains live in domain/identity registry or AVM domain contract.
- Ownership is `AE...`.
- Wallet account may store only optional alias/reference.

Staking model:

- User does not choose validators.
- User stakes through official liquid staking/nominator pool.
- Minimum normal wallet deposit is `10 AET`.
- Reputation is pool-weighted and stake-time based, not validator-choice based.

Reputation model:

- Account stores one identity reputation reference/score, not multiple
  reputation states.
- `IdentityReputation` accrues from:
  - pool exposure;
  - stake-time;
  - transaction behavior;
  - contract interaction outcomes;
  - account behavior signals;
  - protocol-defined events.
- Staking is a reputation amplifier:
  - larger stake amount increases possible reputation gain;
  - longer stake duration increases possible reputation gain;
  - claim/unbond/reward settlement calculates actual stake-time contribution;
  - duplicate credit is forbidden.
- Reputation decays slowly without activity.
- Active stake-time slows decay.
- Reputation confidence tracks how reliable the score is.
- Low identity reputation must not block token creation, NFT creation, smart
  contract deployment, contract execution, basic transactions, or pool staking.
- Low identity reputation may only affect bounded soft weighting such as queue
  priority, scheduling, optional fee efficiency, or minor liquid-staking yield
  weighting.
- Reputation cannot increase without stake-time or configured behavior evidence.
- Contracts, tokens, NFTs, DEX pools, and ordinary contract instances do not
  keep persistent reputation state.

Tests:

- token/NFT query reads contract/registry state, not account state;
- domain owner query reads registry/contract state, not account metadata;
- pool deposit updates pool/share/reputation refs only;
- direct validator delegation rejected;
- pool claim updates unified identity reputation from stake-time;
- low reputation account can deploy token/NFT/contracts;
- contract execution emits behavior signals but no contract reputation state;
- reputation export/import preserves score, confidence, decay, and stake-time
  accumulator.

Done:

- Wallet account remains an identity/auth/runtime layer, not an asset database.

### AWCE-1.10 - Wallet Extension Metadata And Docs

Implementation:

- Add docs section/spec:
  - `docs/AWCE-1.md` or `docs/wallet-compatibility.md`;
  - wallet flow diagrams;
  - activation flow;
  - dual-address examples;
  - frozen/recovery flow;
  - pool staking from 10 AET;
  - contract deploy/execute/get examples.
- Add machine-readable extension descriptor if needed:

```json
{
  "standard": "AWCE-1",
  "canonical_user_address": "AE...",
  "raw_address": "4:...",
  "signing": "cosmos-signdoc-secp256k1",
  "default_hd_path": "m/44'/118'/0'/0/0",
  "features": ["activate_account", "auth_policy", "storage_rent", "avm"]
}
```

Rules:

- Docs must not instruct users to export seed phrases into Aetra tools.
- Docs must not teach direct validator selection for normal staking.
- Docs must not call `4:...` the normal wallet address.

Tests:

- static docs test for AWCE required sections;
- static docs test rejects seed/private-key examples;
- static docs test rejects `aevaloper`/`aevalcons` user-facing flows;
- wallet compatibility examples validate against CLI/API schema.

Done:

- Wallet vendors and app developers have one concise compatibility spec.

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

### Task 6.3 - Core Zones, Layers, And Sharding Runtime Gate

Current risk:

- Load, routing, zones, sharding, and Aether Mesh can remain executable specs or
  simulators while the runnable app is still a single flat execution path.

Requirement:

- Implement a real gated runtime path for Aetra Core Zones/layers:
  - Aether Core control plane;
  - Financial/Bank Zone for native `naet` settlement, fees, treasury, rewards;
  - Staking/Validator Zone for registry, election, pool allocation, slashing;
  - Account/Auth Zone for native account lifecycle, reputation refs, storage
    rent status;
  - Contract/AVM Zone for deploy/execute/query/storage/receipts;
  - Identity/Domain Zone only if it is wired as bounded native system state or
    AVM contract standard;
  - optional Service/Infra Zone for RPC/indexer/oracle/service actor trust.
- Each zone must have:
  - deterministic zone id;
  - store key/prefix layout;
  - state root or commitment;
  - bounded message ingress/egress queue;
  - receipt root;
  - export/import format;
  - invariant coverage.
- Implement deterministic routing from tx/message to zone:
  - stable message type classification;
  - `AE...` actor and `4:...` raw/proof key extraction;
  - no wall-clock, random, local mempool order, goroutine race, or map order;
  - same tx/state -> same zone/shard/priority decision on every node.
- Implement sharding only behind explicit governance/genesis feature gate until
  long-run evidence exists:
  - `sharding_enabled = false` by default for conservative public testnet;
  - if enabled, active shard count is derived from deterministic load score;
  - shard assignment uses domain-separated hash of zone id, primary actor, and
    routing epoch;
  - split/merge is epoch-gated and deterministic;
  - queue partitioning and merge preserve canonical message order;
  - cross-shard/cross-zone messages require receipts/proofs and replay markers.
- Core zones must be real app/runtime state when marked enabled:
  - no "spec-only" module may satisfy readiness;
  - no in-memory-only keeper may be used for enabled zone state;
  - no unbounded full-state scan in block lifecycle;
  - export/import must restart with identical zone roots, queues, receipts, and
    shard assignments.
- Aether Core remains the control plane and must not execute application smart
  contract logic directly.

Messages/queries:

- Add or expose bounded queries for:
  - `ZoneInfo`;
  - `ZoneStateRoot`;
  - `ZoneRoutingDecision`;
  - `ShardInfo`;
  - `ShardAssignment`;
  - `CrossZoneReceipt`;
  - `AetherCoreStatus`.
- Add internal message envelopes for cross-zone/cross-shard delivery with:
  - source zone/shard;
  - destination zone/shard;
  - sender/recipient;
  - value in `naet` where applicable;
  - payload hash;
  - timeout height;
  - proof/receipt references;
  - replay marker key.

Tests:

- App starts with Core Zone registry wired and disabled experimental sharding.
- Enabled zones have mounted store keys and keeper-backed state.
- Same tx routes to same zone/shard across repeated runs.
- Financial tx routes to Financial/Bank Zone and preserves `naet` accounting.
- AVM contract execute routes to Contract/AVM Zone and emits receipt/root.
- Staking pool action routes to Staking/Validator Zone and never asks user for
  validator choice.
- Cross-zone message success writes one receipt and one replay marker.
- Duplicate cross-zone message is rejected by replay marker.
- Shard assignment golden vectors are deterministic.
- Split/merge fixture preserves every queued message exactly once.
- Export/import preserves zone registry, zone roots, shard assignments, queues,
  receipts, and replay markers.
- Readiness gate fails if a required runtime zone is only a prototype/spec.

Done:

- Aetra can demonstrate real layered Core Zone execution with deterministic
  routing, optional gated sharding, receipts, roots, and restart safety.

### Task 6.4 - Reputation Module Consolidation

Implementation:

- Inventory every reputation-like module/type/query:
  - account reputation;
  - stake reputation;
  - validator reputation;
  - reporter reputation;
  - contract reputation;
  - service/infra trust.
- Replace wallet-facing fragmented reputation with one identity record:

```text
reputation/identity/{account}
```

- Convert activity-specific modules into signal providers:
  - staking emits stake-time signals;
  - tx execution emits success/failure/spam signals;
  - AVM emits contract interaction outcome signals;
  - lifecycle emits recovery/freeze/unfreeze stability signals;
  - validator subsystem emits validator-role metrics into `ValidatorScore`;
  - service subsystem emits infra-role metrics into `ServiceTrustScore`.
- Remove or disable:
  - contract reputation state;
  - token/NFT reputation state;
  - reporter-only reputation state unless reporter is a service actor.
- Add migration from old fragmented records:
  - deterministic merge order;
  - bounded score conversion;
  - confidence recalculation;
  - no double-counting stake-time;
  - migration receipt/report.

Queries:

- `QueryIdentityReputation(account AE...)`;
- `QueryValidatorScore(validator AE...)`;
- `QueryServiceTrustScore(service_id)`;
- no wallet-facing `QueryStakeReputation`, `QueryContractReputation`, or
  `QueryReporterReputation` in public testnet API.

Tests:

- fragmented fixture migrates to one identity reputation;
- migration is deterministic;
- migration rejects duplicate stake-time credit;
- low identity reputation does not block token/NFT/contract deployment;
- validator score query remains separate from wallet identity reputation;
- service trust query remains separate and affects only routing/source choice;
- contract execution signal updates caller identity signal counters, not
  contract reputation state;
- export/import preserves identity reputation, validator score, and service
  trust records.

Done:

- Reputation is one user identity property plus two role-specific views:
  validator consensus score and service routing trust.

### Task 6.5 - Reputation Must Affect Runtime Decisions

Current risk:

- Reputation can exist as queryable state but have no real effect on fees,
  routing, validator allocation, or scheduling.

Requirement:

- Implement reputation for selected actor types only:
  - wallet/user identity: `IdentityReputation`;
  - validator/operator: `ValidatorScore`;
  - registered service/infra actor: `ServiceTrustScore` if enabled.
- Do not implement persistent reputation for ordinary smart contracts, token
  contracts, NFT collections/items, DEX pools, or domain contracts. Contracts
  emit behavior signals; the signal belongs to the caller, owner/admin, service
  actor, or validator only when policy explicitly maps it.
- Reputation must affect real runtime systems through bounded soft weighting:
  - tx priority/scheduling;
  - fee premium/discount within caps;
  - async queue ordering;
  - pool allocation/reward efficiency where applicable;
  - validator election/allocation score;
  - service endpoint/source selection.
- Reputation must not become an access gate for ordinary wallet rights:
  - low reputation cannot block basic transfers;
  - cannot block token/NFT/contract deployment;
  - cannot block normal contract execution;
  - cannot block pool staking from `10 AET`;
  - cannot replace signer/auth checks.
- Reputation updates must be explained by deterministic signals:
  - stake-time exposure through pool shares;
  - tx success/failure/spam behavior;
  - account lifecycle stability/recovery/freeze events;
  - contract interaction outcomes for the initiating identity;
  - validator uptime/missed blocks/slashing/commission/performance;
  - service availability/proof quality if service actors are enabled.
- Reputation uses deterministic integer/fixed-point math:
  - score bounds;
  - confidence bounds;
  - decay schedule;
  - per-epoch update caps;
  - no floating point;
  - no local time;
  - no random sampling.
- All effect caps are governance/genesis params:
  - max fee premium bps;
  - max fee discount bps;
  - max queue priority boost;
  - max validator allocation boost/penalty;
  - max service trust routing boost;
  - decay rate;
  - confidence gain/loss rates.

Acceptance:

- Low-reputation wallet pays a higher deterministic fee than neutral wallet for
  the same tx, within cap, and tx still succeeds when fees are paid.
- High-reputation wallet may receive bounded discount/priority but never zero
  fee and never bypasses auth/ante checks.
- Spam/failed tx fixture reduces wallet score or confidence deterministically.
- Pool stake-time claim increases wallet `IdentityReputation` within cap and
  cannot be double-claimed.
- Jailed/slashed validator receives no positive validator score bonus while
  jailed/slashed.
- Validator score changes pool allocation/reward efficiency deterministically.
- Service trust, if enabled, changes endpoint/source selection only and cannot
  move funds or bypass fees.
- Ordinary contract execution updates caller/actor signal counters only; no
  `contract reputation` state key exists.
- Export/import preserves reputation scores, confidence, decay state, pending
  signal accumulators, and last-applied heights.
- Golden tests prove same state/signals produce same fee, priority, allocation,
  and reputation deltas across repeated runs.

Done:

- Reputation is not decorative: it safely influences real runtime economics and
  scheduling while preserving user rights and excluding ordinary contract
  reputation.

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
- Core Zones runtime wiring smoke;
- deterministic routing/shard-assignment smoke;
- cross-zone receipt/replay-marker smoke;
- AVM deploy/execute/get method smoke;
- pool staking smoke;
- slashing/reputation v1 smoke;
- unified identity reputation migration and non-gating smoke;
- runtime reputation effects smoke for fee premium/discount, queue priority, and
  validator allocation;
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
3. CHAT F: freeze AWCE-1 account lifecycle, activation, dual-address mapping,
   auth policy, and wallet compatibility tests without changing address
   derivation.
4. CHAT D: create canonical `docs/VALIDATOR.md`, `docs/TESTNET.md`,
   `docs/COSMOVISOR.md`, Dockerfile, health docs.
5. CHAT B: finish 10 AET pool-based staking, direct delegation rejection, real
   consensus validator-set alignment, reward policy, slashing, validator score
   v1.
6. CHAT C: finish AVM module format/verifier, exit-code mapping, instruction
   set, Chunk core, typed Codec, ChunkMap, StateInit, get methods,
   receipts/bounce/refund, upgrades, examples, determinism gate.
7. CHAT E: wire Core Zones runtime gate, deterministic routing, optional
   sharding feature gate, cross-zone receipts, and replay markers.
8. CHAT E: remove stale native DEX/token/NFT launch docs and enforce future AVM
   standards boundary.
9. CHAT A: final app wiring, live invariants, 4/5-node localnet, upgrade rehearsal, release
   workflow.
10. All chats: run full testnet release candidate checklist.

## Definition Of Done For V2

Aetra is V2 testnet-ready only when:

- a clean checkout can build one `aetrad` binary;
- genesis can be generated, validated, and published;
- 4-5 node localnet reaches height and has stable validators;
- export/import restart works;
- upgrade rehearsal works;
- AWCE-1 preserves Cosmos wallet signing compatibility while adding native
  account lifecycle/auth policy/runtime metadata;
- `AE... <-> 4:...` mapping is deterministic and unchanged by activation, auth
  policy updates, key rotation, recovery, metadata, staking, or VM usage;
- account activation creates runtime state once; inactive accounts do not accrue
  storage rent;
- account state stores no private keys, seed phrases, asset ledgers, tx history,
  domain records, resolver records, or unbounded profile data;
- normal staking uses the pool path with a `10 AET` minimum deposit;
- normal staking messages, CLI, API, and docs do not accept validator selection;
- direct user validator delegation is disabled in public testnet profile and
  rejected before staking mutation;
- each wallet identity has one unified reputation score with confidence/decay;
- staking improves wallet reputation through bounded stake-time contribution
  after claim/settlement, without creating a separate stake reputation state;
- low wallet reputation does not block token/NFT creation, smart contract
  deployment, contract execution, basic txs, or pool staking;
- contracts/tokens/NFTs/DEX pools have no persistent reputation state;
- validator score and service trust remain role-specific views, not wallet
  reputation fragments;
- reputation affects real runtime behavior only through bounded soft weighting:
  fee premium/discount, queue priority, validator allocation, and service source
  selection where enabled;
- ordinary contracts, token contracts, NFT contracts, DEX pools, and domain
  contracts cannot own persistent reputation state;
- the actual CometBFT/SDK validator set is controlled by or proven identical to
  the Aetra registry/election output;
- reward/slashing/reputation v1 are understandable and tested;
- Core Zones/layers are keeper-backed runtime state when enabled, with
  deterministic zone routing, state roots, receipts, replay markers, bounded
  queues, and export/import restart safety;
- sharding is disabled by default unless governance/testnet profile enables it,
  and any enabled shard assignment/split/merge path is deterministic and covered
  by golden tests;
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
подключи нормалтьно чтобы комисии работали все распрделение эмитация живой экономики все как в реальном блокчейне и транзакции скорость сделай чтобы как нативно было чтобы можно было проверить реальный TPS и финальный блок время.