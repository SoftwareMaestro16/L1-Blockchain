# Aetheris Roadmap V2

`Roadmap_v2.md` is a local operator planning file and must not be committed or pushed. It is intentionally listed in `.gitignore`.

## Direction

Aetheris should become a production-grade asynchronous L1 with:

- hardened PoS and staking;
- native-only base fees in `naet`;
- deterministic async contract execution;
- AVM, the Aetheris Virtual Machine;
- executable wallet, fungible token, NFT, and SBT standards;
- gated CosmWasm support;
- consensus-safe sharding research before any production sharding claim.

Production claims require tests, benchmarks, state export/import checks, adversarial tests, and audit gates. Design notes are not enough.

## Track 1: PoS And Staking To Production

Goal: make validator operation, staking state, rewards, slashing, upgrades, and network restart behavior safe enough for public testnet and later production.

Tasks:

- Define canonical staking policy:
  - bond denom: `naet`;
  - min self-delegation bounds;
  - commission min/max/rate-change bounds;
  - unbonding period;
  - max validators;
  - redelegation limits;
  - jailed/unjailed behavior.
- Add parameter validation tests for every staking, slashing, distribution, mint, and fee parameter.
- Add validator lifecycle tests:
  - create validator;
  - edit validator;
  - delegate;
  - redelegate;
  - unbond;
  - cancel unbond if supported;
  - jail;
  - unjail;
  - remove inactive validator from active set.
- Add validator set propagation tests:
  - power changes after delegation;
  - power changes after unbonding;
  - power changes after redelegation;
  - CometBFT validator updates match staking keeper state.
- Add rewards tests:
  - inflation mints `naet`;
  - rewards accrue deterministically;
  - validator commission is bounded;
  - delegator withdraw works;
  - community pool accounting is deterministic.
- Add slashing tests:
  - downtime signing info;
  - missed block window;
  - tombstone/equivocation path if enabled;
  - jailed validator cannot keep producing;
  - slashed stake and distribution accounting are consistent.
- Add localnet profiles:
  - 1 validator dev;
  - 3 validator smoke;
  - 5 validator public-testnet rehearsal;
  - 10 validator stress;
  - 20 validator long-run profile.
- Add restart and persistence tests:
  - stop/start after delegation;
  - stop/start after unbonding;
  - stop/start after slashing state changes;
  - state-sync join after staking changes;
  - snapshot restore after staking changes.
- Add chaos tests:
  - kill one validator in 3-validator profile;
  - kill two validators in 5-validator profile;
  - restart validator with stale local state and verify safe failure;
  - corrupt local node data outside consensus state and verify diagnostics.
- Add invariants:
  - bonded + unbonded + not-bonded pools match staking state;
  - validator tokens and delegator shares are internally consistent;
  - distribution outstanding rewards match module balances;
  - total `naet` supply changes only through mint/reward/slash/burn paths.
- Add benchmarks:
  - create validator;
  - delegate;
  - redelegate;
  - unbond;
  - validator set update;
  - reward withdrawal;
  - slashing window scan.

Acceptance:

- 3-validator and 5-validator networks produce blocks for a long-run smoke.
- Validator power changes after staking transactions.
- Restart preserves staking state and consensus progress.
- State-sync and snapshot restore produce the same staking state root.
- No staking path accepts non-`naet` bond or fee denoms.

## Track 2: Base Chain Safety Before Contracts

Goal: finish the hardening needed before AVM or CosmWasm can mutate production state.

Tasks:

- Keep address validation centralized in `app/addressing`.
- Keep native token and fee constants centralized in `app/params`.
- Reject zero address everywhere by default.
- Reject old public `0:`, `orb1`, and `ORB` formats outside migration tooling.
- Add signed transaction replay helpers that submit identical signed bytes twice.
- Add wrong chain-id signing tests.
- Add malformed protobuf transaction tests.
- Add invalid signer tests for bank, staking, gov, fees, tokenfactory, DEX, AVM, and CosmWasm messages.
- Add consensus panic tests for every custom module message and genesis type.
- Add deterministic event contract tests for important state transitions.
- Require `go test ./...`, `go vet ./...`, `buf lint`, security scans, and determinism gate before public testnet.

Acceptance:

- Bad transactions fail before state mutation.
- Same signed transaction cannot execute twice.
- Invalid signer cannot mutate state.
- Malformed inputs do not panic consensus code.

## Track 3: Async Blockchain Execution Model

Goal: make Aetheris contract execution asynchronous while preserving Cosmos SDK determinism.

Core concepts:

- contract account;
- contract state;
- incoming message;
- outgoing message;
- deterministic message queue;
- bounce message;
- refund/excess message;
- logical time per contract;
- bounded per-block processing;
- gas and storage accounting;
- failure result code.

Tasks:

- Define message envelope:
  - source;
  - destination;
  - value in `naet`;
  - opcode;
  - query id;
  - body;
  - bounce flag;
  - created logical time;
  - deadline;
  - gas limit;
  - forward fee.
- Define deterministic queue ordering:
  - tx index;
  - message index;
  - source logical time;
  - destination address;
  - sequence tie-breaker.
- Define contract logical time:
  - monotonically increasing per contract;
  - included in message IDs;
  - exported/imported exactly.
- Define bounce behavior:
  - when bounce is created;
  - which sender state is final;
  - which recipient state is rolled back;
  - how `naet` refunds are computed;
  - how double-refund is prevented.
- Define per-block limits:
  - max inbound messages per tx;
  - max processed messages per block;
  - max recursive depth;
  - max body size;
  - max state size;
  - max emitted messages;
  - max deployments;
  - max storage writes.
- Define gas and fee model:
  - execution gas;
  - message forwarding fee;
  - deployment fee;
  - storage rent or storage deposit if used;
  - all protocol fees in `naet`.
- Add state export/import:
  - contract code;
  - contract state;
  - queue state;
  - logical time;
  - pending bounces;
  - pending refunds.
- Add observability:
  - queued messages;
  - processed messages;
  - bounced messages;
  - failed messages;
  - queue lag;
  - gas used;
  - contract state size.
- Add tests:
  - contract emits internal message;
  - recipient executes later in deterministic order;
  - failed send bounces;
  - refund cannot double-spend;
  - export/import preserves queue state;
  - queue limits prevent DoS.

Acceptance:

- Async execution is deterministic across nodes.
- Failed internal sends produce deterministic bounce/refund behavior.
- Queue state survives export/import exactly.
- Queue limits prevent block and memory DoS.

## Track 4: AVM, Aetheris Virtual Machine

Goal: build AVM as the native Aetheris VM for async contracts.

Non-negotiable requirements:

- binary serialization spec;
- message ABI;
- storage ABI;
- deterministic execution proof;
- gas schedule;
- memory limits;
- code size limits;
- stack/register limits;
- host function allowlist;
- fuzz tests;
- differential tests;
- upgrade and migration policy;
- adversarial audit.

Tasks:

- Define AVM bytecode format:
  - module header;
  - version;
  - code hash;
  - imports;
  - exported entrypoints;
  - metadata hash.
- Define AVM execution entrypoints:
  - deploy;
  - receive external;
  - receive internal;
  - receive bounced;
  - query/getter;
  - migrate.
- Define AVM storage:
  - key/value namespace per contract;
  - deterministic encoding;
  - bounded iteration;
  - query pagination;
  - state size limits.
- Define AVM host functions:
  - read/write storage;
  - emit internal message;
  - inspect message envelope;
  - get block context;
  - charge gas;
  - return result code.
- Define forbidden behavior:
  - wall-clock time;
  - random host entropy outside consensus-approved randomness;
  - filesystem/network access;
  - floating point;
  - unbounded iteration;
  - nondeterministic map iteration.
- Add AVM toolchain:
  - bytecode verifier;
  - disassembler;
  - local runner;
  - gas profiler;
  - contract test harness;
  - state snapshot inspector.
- Add AVM keeper:
  - store code;
  - instantiate contract;
  - route external message;
  - process internal queue;
  - execute getters;
  - export/import state.
- Add AVM tests:
  - deploy valid contract;
  - reject malformed bytecode;
  - reject oversized code;
  - reject nondeterministic opcode;
  - run simple counter;
  - send internal message;
  - bounce failed message;
  - preserve state across restart;
  - preserve queue across export/import.

Acceptance:

- AVM can execute a minimal contract deterministically.
- AVM gas is deterministic and bounded.
- AVM contracts can participate in async message queue.
- AVM cannot weaken base-chain signer, fee, denom, zero-address, or genesis validation.

## Track 5: Executable Contract Standards

Goal: standards are not only documents. They must become executable AVM contract systems with conformance tests.

### AW-5: Aetheris Wallet Standard

Storage:

- signature allowed flag;
- seqno;
- wallet id;
- public key;
- owner address;
- extension map;
- recovery policy;
- spending limits if enabled.

Messages:

- signed external command;
- internal extension command;
- update signature allowed;
- install extension;
- remove extension;
- multi-send;
- query wallet state.

Rules:

- `seqno` prevents replay;
- `wallet_id` prevents cross-wallet replay;
- `valid_until` prevents stale commands;
- signature validation happens before accepting external command state changes;
- extensions are explicit and revocable;
- multi-send is bounded;
- protocol fees are paid in `naet`.

Tests:

- deploy wallet;
- signed send;
- replay rejected;
- wrong wallet id rejected;
- expired command rejected;
- invalid signature rejected;
- extension install/remove;
- unauthorized extension rejected;
- multi-send bounded;
- relayer pays `naet`.

### AFT-44: Aetheris Fungible Token Standard

Model:

- token master contract;
- token wallet contract per holder;
- native AET is not AFT-44.

Token master storage:

- token id;
- admin;
- total supply;
- mintable flag;
- burnable flag;
- metadata reference;
- name;
- symbol;
- decimals;
- wallet code hash;
- admin transfer state.

Token wallet storage:

- master address;
- owner address;
- balance;
- wallet code version;
- pending query ids.

Messages:

- create token;
- mint;
- transfer;
- internal transfer;
- receive transfer;
- burn;
- burn notification;
- change admin;
- renounce admin;
- change metadata;
- close minting;
- bounce handling.

Rules:

- wallet address is derived from `(master, owner, wallet_code_hash)`;
- mint deploys or credits recipient wallet;
- transfer is async and deterministic;
- burn decreases wallet balance before master supply decrement;
- master supply and wallet balances must not diverge;
- native `AET`/`naet` spoofing is rejected;
- token balances never pay protocol fees.

Tests:

- create token master;
- derive wallet address;
- mint;
- transfer;
- deploy missing recipient wallet;
- burn;
- admin transfer;
- renounce admin;
- non-admin mint rejected;
- native spoof rejected;
- bounce restores or finalizes state deterministically;
- replayed wallet message rejected.

### ANFT-66: Aetheris NFT Standard

Storage:

- collection owner/admin;
- collection metadata;
- next item index;
- item code hash;
- mutable metadata flag;
- royalty policy if enabled.

Item storage:

- collection address;
- item index;
- owner address;
- item metadata;
- initialized flag;
- transfer policy.

Messages:

- create collection;
- mint item;
- batch mint with strict limit;
- transfer item;
- change collection metadata;
- change owner/admin;
- get collection data;
- get item data;
- prove collection membership;
- bounce handling.

Rules:

- collection derives item address;
- item proves collection membership;
- transfer requires current owner authorization;
- metadata changes are bounded;
- batch mint cannot DoS a block or queue.

Tests:

- create collection;
- mint item;
- verify derived address;
- transfer;
- unauthorized transfer rejected;
- malformed collection rejected;
- metadata spoofing rejected;
- batch mint bounded.

### ASBT-67: Aetheris Soulbound Token Standard

Storage:

- collection address;
- item index;
- immutable owner;
- content;
- authority address;
- revoked timestamp;
- revoke reason/content.

Messages:

- mint SBT;
- prove ownership;
- ownership proof response;
- request current owner;
- revoke by authority;
- destroy if policy allows;
- transfer rejected.

Rules:

- owner is immutable after mint;
- transfer always fails;
- revoke does not transfer ownership;
- authority is explicit;
- proof responses are deterministic.

Tests:

- mint;
- transfer rejected;
- prove ownership;
- revoke by authority;
- unauthorized revoke rejected;
- owner cannot be changed;
- malformed proof rejected.

Acceptance:

- AW-5, AFT-44, ANFT-66, and ASBT-67 run as AVM contracts.
- Each standard has a conformance suite.
- Native `naet` remains the only protocol fee denom.
- Async bounce/refund behavior is tested for every standard.

## Track 6: CosmWasm Production Gate

Goal: support CosmWasm without weakening the base chain or AVM direction.

Policy:

- CosmWasm remains disabled by default.
- Enabling requires explicit config or feature gate.
- CosmWasm does not bypass `naet` fee policy, address policy, zero-address policy, or genesis validation.

Tasks:

- Define upload permissions.
- Define instantiate permissions.
- Define admin and migration policy.
- Define gas multiplier and gas limits.
- Define contract code size limits.
- Define memory/cache limits.
- Define query limits.
- Define pinned code policy if used.
- Define governance authority for enabling/disabling CosmWasm.
- Add tests:
  - disabled by default;
  - unauthorized upload rejected;
  - unauthorized instantiate rejected;
  - unauthorized migrate rejected;
  - gas limit enforced;
  - query limit enforced;
  - contract cannot pay fees in non-`naet`;
  - contract cannot use zero address as admin/recipient.
- Add smoke:
  - upload simple contract;
  - instantiate;
  - execute;
  - query;
  - migrate only if policy permits.

Acceptance:

- CosmWasm can be enabled only intentionally.
- CosmWasm contracts cannot bypass base-chain safety.
- AVM and CosmWasm boundaries are documented and tested.

## Track 7: Production Sharding And Partitioning R&D

Goal: design masterchain/workchain/shardchain architecture without claiming production sharding before it is consensus-safe.

Terminology:

- masterchain: global coordination chain for validator set, workchain configs, shardchain headers, finality references, and cross-chain routing commitments;
- workchain: execution domain with its own rules, VM set, address space, and state transition function;
- shardchain: partition of a workchain state and message space;
- cross-shard message: async message routed through deterministic proof and receipt flow.

Research tasks:

- Write consensus-safe sharding spec:
  - finality model;
  - validator assignment;
  - randomness source;
  - shard header commitments;
  - data availability;
  - fraud/equivocation evidence;
  - slashing rules;
  - cross-shard message ordering;
  - cross-shard replay protection.
- Define masterchain state:
  - validator set;
  - staking snapshot;
  - workchain registry;
  - shardchain registry;
  - shard headers;
  - cross-shard receipt roots;
  - config updates.
- Define workchain model:
  - workchain id;
  - allowed VMs;
  - fee policy;
  - address format;
  - genesis state;
  - upgrade policy.
- Define shardchain model:
  - shard id;
  - state root;
  - message queue root;
  - receipt root;
  - validator subset;
  - split/merge rules.
- Define cross-shard async messages:
  - source workchain/shard;
  - destination workchain/shard;
  - message id;
  - proof;
  - receipt;
  - timeout;
  - bounce/refund semantics.
- Build simulator before implementation:
  - deterministic message routing;
  - shard split;
  - shard merge;
  - validator reassignment;
  - delayed receipts;
  - failed shard block;
  - equivocation evidence.
- Add prototype implementation only after simulator:
  - masterchain header keeper;
  - workchain registry keeper;
  - shard routing keeper;
  - cross-shard queue keeper;
  - export/import of sharded state.
- Add adversarial tests:
  - duplicate cross-shard receipt;
  - missing receipt;
  - invalid shard proof;
  - stale shard header;
  - wrong destination shard;
  - replayed message;
  - validator equivocation;
  - data unavailable shard block.
- Add benchmarks:
  - routing table lookup;
  - cross-shard proof verification;
  - queue processing;
  - shard split;
  - shard merge;
  - state export/import.

Acceptance:

- Before production: written spec, simulator, prototype, fuzz tests, adversarial tests, long-run testnet, independent audit.
- Public language must say "sharding R&D" or "experimental sharding" until the production gate is passed.
- No production sharding claim without consensus-safety proof and audit.

## Track 8: Public Testnet And Production Gates

Public testnet gate:

- `go test ./...` passes.
- `go vet ./...` passes.
- `buf lint` passes.
- security scans pass or findings are triaged.
- deterministic execution gate passes.
- 3-validator and 5-validator localnet profiles pass.
- snapshot/state-sync works.
- validator onboarding docs are clean.
- faucet plan is implemented or explicitly deferred.
- explorer/indexer plan is implemented or explicitly deferred.
- incident response and rollback docs are tested.
- CosmWasm smoke passes if CosmWasm is enabled.
- AVM smoke passes if AVM is enabled.

Production gate:

- long-running public testnet has no untriaged consensus or fund-safety issues;
- validator set can upgrade safely;
- staking, fees, DEX, AVM, and contract standards have adversarial tests;
- state export/import is deterministic;
- independent audit completed;
- emergency governance and halt/restart process tested.

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
