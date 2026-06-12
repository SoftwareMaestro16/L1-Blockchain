> Note: historical native asset-factory and native exchange modules have been removed from the active app graph.
# Module Boundaries

## Launch Module Inventory

Machine-readable launch inventory: `app/launch_module_inventory.json`.
Generated launch scope doc: `docs/TESTNET.md`.

Classification counts:
- `disabled`: 2
- `future_avm_standard`: 15
- `launch_core`: 14
- `launch_support`: 28
- `prototype_only`: 7

Public testnet profile rejects `prototype_only` and `disabled` modules in app wiring, rejects memory-only consensus keepers, and rejects native application-asset modules.

Core SDK module responsibilities and Aetra-specific wrappers are documented
in [Core Module Architecture](architecture/core-module-architecture.md). This
file focuses on custom modules and readiness packages. Economy and interop
boundaries are documented in
[Economy And Interop Module Architecture](architecture/economy-interop-architecture.md).
Application module architecture is documented in
[Application Module Architecture](architecture/application-module-architecture.md).
Additional support modules are documented in
[Additional Modules](architecture/additional-modules.md).

Core SDK module responsibilities and Aetra-specific wrappers are documented
in [Core Module Architecture](architecture/core-module-architecture.md). This
file focuses on custom modules and readiness packages. Economy and interop
boundaries are documented in
[Economy And Interop Module Architecture](architecture/economy-interop-architecture.md).
Application module architecture is documented in
[Application Module Architecture](architecture/application-module-architecture.md).
Additional support modules are documented in
[Additional Modules](architecture/additional-modules.md).

## Historical `x/tokenfactory` Prototype Boundary

Historical purpose: create and manage custom denoms without EVM dependency
during prototype testing. Historical `x/tokenfactory` behavior is retained only
as migration compatibility evidence.

State:
- Denom registry keyed by full denom.
- Admin record per denom.
- Optional metadata record per denom.
- Module params.

Minimal Msg surface:
- `MsgCreateDenom`
- `MsgMint`
- `MsgBurn`
- `MsgChangeAdmin`
- `MsgUpdateParams`

Keeper dependencies:
- Bank keeper interface for mint, burn, send, and metadata operations.
- Account/address codec where required by the scaffolded SDK version.

Security invariants:
- Only authorized admins can mint, burn, or transfer admin rights.
- Total supply changes must match bank keeper mint/burn results.
- Subdenom length bounds and mint/burn/create emergency flags are governance-controlled.
- Governance cannot seize an existing factory denom admin role.

## Historical `x/dex` Prototype Boundary

Historical purpose: deterministic constant-product AMM during prototype
testing. Historical `x/dex` prototype language is retained only as a migration
compatibility reference.

Current direction:
- Application-level asset logic must not be reintroduced as active native modules.
- Future contract-based pools/routers must treat the historical prototype as the
  reference implementation or migration bridge until audited.

State:
- Pool registry keyed by pool ID.
- Asset pair index.
- LP share accounting.
- Fee accumulator references.
- Module params.

Minimal Msg surface:
- `MsgCreatePool`
- `MsgAddLiquidity`
- `MsgRemoveLiquidity`
- `MsgSwapExactIn`
- `MsgUpdateParams`

Keeper dependencies:
- Bank keeper interface for escrow, pool balances, and LP share movement.
- Fees keeper interface for protocol fee accounting.

Security invariants:
- Integer math only.
- No pool operation can create value.
- LP shares must remain backed by pool reserves.
- User-provided min-out values protect against slippage.
- Governance-controlled pool params are bounded and cannot mutate reserves or LP supply.
- Duplicate pair lookup uses a deterministic pair index instead of scanning all pools.
- Native AET metadata cannot be spoofed through display denoms, factory denoms,
  or arbitrary LP denoms.
- Recorded reserves must match module account balances, and LP supply must
  match pool shares after every pool state transition.

## `x/identity`

Purpose: future `.aet` domain registry and resolver ownership surface.

Current status:
- Pure validation helpers only.
- No SDK stores, keepers, module accounts, genesis, or CLI tx surface are
  registered yet.
- Registry records are the source of truth; UI/wallet proof layers are not a
  replacement for registry ownership checks.

State:
- Domain record keyed by normalized name.
- Owner address.
- Optional resolver address.
- Expiry and renewal state.
- Auction status.
- Contract item reference for wallet/UI representation.

Security invariants:
- Domain names are normalized and restricted to lowercase ASCII `a-z`, digits
  `0-9`, `-`, and `_`.
- Whitespace, invisible Unicode, mixed-script spoofing, unsupported symbols,
  and non-`.aet` TLDs are rejected.
- Owner and resolver addresses use central address validation and reject zero
  addresses by default.
- Resolver updates require current owner authorization.
- Expiry, renewal, and auction transitions must emit deterministic events.
- Domain lifecycle is `available -> auction -> active -> expired ->
  available/auction`.
- Auctions run for `24h`, use a `5%` minimum bid increment, apply bounded
  `10m` anti-snipe extensions, and assign ownership only during finalization.
- Auction proceeds split `40%` burn, `40%` treasury, and `20%` rewards.
- Renewal extends expiry and uses the deterministic start-price discount rule.
- Resolver records support domain-to-address resolution, reverse resolution,
  multi-address records, delegated manager grants, bounded metadata, and
  deterministic resolver update events.
- Resolver payment routing fails before funds move when `primary` is unset,
  the resolver target is zero, the registry owner does not match, or the domain
  is expired.
- Subdomains such as `app.alice.aet`, `bot.alice.aet`, and `pool.alice.aet`
  resolve through the base `.aet` registry owner.

## `x/workflow`

Purpose: future bounded multi-step orchestration for application flows that need
explicit synchronous composition.

Current status:
- Pure validation helpers only.
- No SDK stores, keepers, module accounts, genesis, or CLI tx surface are
  registered yet.

Examples:
- resolver-based payment.
- domain auction finalization.
- contract mint and wallet deploy.
- contract mint and metadata attach.
- contract deployment plus first message.

Security invariants:
- Workflow authority is a valid non-zero address.
- Workflows are bounded by maximum step count and payload size.
- Step IDs are unique inside one workflow.
- Orchestration must not bypass signer checks, replay protection, `naet` fee
  policy, zero-address rejection, or module-specific invariants.

## `x/memo`

Purpose: optional human-readable transaction metadata for notes, receipts, and
UI context.

Current status:
- Pure validation and fee helpers only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- No execution state.
- Future indexed transaction metadata may include `memo`, `memo_hash`, and
  `memo_visible`, but it must be immutable after block inclusion.

Security invariants:
- Memo text is optional and UTF-8 only.
- Memo length is governed within a hard protocol bound.
- Prohibited control characters are rejected.
- Memo metadata does not affect execution state transitions.
- Memo data is stored as transaction metadata, not keeper execution input.
- Memo fees are paid only in `naet`.
- Empty memo can have zero memo fee.
- Memo byte cost can scale by reputation and congestion so memo text cannot
  become cheap spam storage.
- Memo projection may index by tx hash, sender, receiver, domain, contract,
  asset, and event type.
- Full memo on-chain and hash-only on-chain storage policies are explicit.
- Consensus does not depend on search index results.
- `EventMemoAttached` is deterministic and includes tx hash, from, to, domain,
  memo hash, and memo according to storage policy.

## `x/execution`

Purpose: future execution OS for transaction orchestration, execution pipeline,
module dispatch, async entrypoint coordination, deterministic event collection,
and error handling.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- `ExecutionEnvelope` carries tx hash, sender, receiver, route, VM route, gas,
  fees, optional memo metadata, resolver/domain records, reputation record,
  load counters, async messages, module events, block height, and timestamp.
- Deterministic execution trace records pipeline stage order for tests.

Security invariants:
- CheckTx order is decode, signatures, fees, memo, stateless checks.
- DeliverTx/FinalizeBlock order is ante, execution context, module dispatch,
  async enqueue, event emit, state write.
- Resolver lookup, reputation limits, fee estimation, memo validation, and VM
  routing must not bypass signatures, `naet` fee validation, zero-address
  rejection, or module authorization.
- Event collection is deterministic and sorted where needed.

## `x/vm`

Purpose: AVM and gated CosmWasm runtime facade for execution routing.

Current status:
- Pure validation facade only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- Runtime policy for AVM params and CosmWasm feature gate.
- VM call descriptor for deploy, external call, internal call, bounced call,
  and query/getter.

Security invariants:
- AVM is the primary VM and AVM action-to-entrypoint mapping is deterministic.
- CosmWasm remains disabled by default as an optional compatibility layer.
- Runtime routing validates code size, gas, query response/depth, and feature
  gates before keeper wiring.
- VM routing cannot bypass base-chain signer, fee, denom, zero-address,
  transaction, or genesis validation.

## `x/messaging`

Purpose: async contract messaging facade for internal message envelopes,
outgoing message validation, bounce/refund behavior, and delivery through the
async execution spec.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- Message id, source, destination, value in `naet`, opcode, query id, body,
  bounce flag, deadline, gas limit, and created logical time.

Security invariants:
- Message ids are deterministic.
- Source and destination reject zero address.
- Value and forwarding fees are native `naet`.
- Bounce, expiry, and refund/no-double-spend behavior must match
  `x/aetravm/async`.
- Outgoing messages are bounded by async message body, gas, and queue limits.

## `x/queue`

Purpose: deterministic delayed execution and scheduled task queue.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- Scheduled height.
- Reputation class.
- Tx height, tx index, and message index.
- Source logical time.
- Sequence tie-breaker.
- Account and contract queue ownership.
- Attempts and last error for retry/failure handling.

Security invariants:
- Priority key is deterministic:
  `scheduled_height, reputation_class, tx_height, tx_index, message_index,
  source_logical_time, sequence`.
- Low reputation cannot starve forever because aged ready items receive an
  effective top class after the starvation window.
- Per-block processing, per-account queued messages, and per-contract queued
  messages are bounded.
- Queue observability exposes queued, processed, failed, lag, account count,
  and contract count.

## `x/events`

Purpose: deterministic event schema for protocol, indexer, contract, memo,
domain, reputation, and fee events.

Current status:
- Pure validation and canonicalization helpers only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- Event type, category, tx hash, height, sequence, actor, and sorted
  attributes.

Security invariants:
- Supported events include `EventTransfer`, `EventMemoAttached`,
  `EventDomainAuctionStarted`, `EventDomainResolved`,
  `EventContractMessageQueued`, `EventContractMessageProcessed`,
  `EventReputationUpdated`, and `EventFeeDistributed`.
- Event attributes are canonicalized and sorted by key/value.
- Event sorting is deterministic by height, sequence, type, and tx hash.
- Event actors reject zero address when present.

## `x/actors`

Purpose: contract actor model for isolated actor state, mailbox processing, and
actor lifecycle.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- Actor address.
- Code hash.
- State root.
- Logical time.
- Mailbox stats.
- Lifecycle status.
- Exported actor state with mailbox.

Security invariants:
- Each contract behaves as an actor.
- One actor state transition occurs per delivered message.
- Actor cannot mutate another actor state directly.
- All cross-actor effects go through messages.
- Exported state includes actor state and mailbox and is deterministically
  ordered.

## `x/scheduler`

Purpose: deterministic execution planning for sequential execution today and
future optimistic parallel execution.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, ABCI hooks, or
  concurrent state access are registered yet.

State:
- Scheduler task id.
- Tx height, tx index, and message index ordering keys.
- Declared read/write set.
- Deterministic batch plan.

Security invariants:
- Initial execution remains sequential deterministic execution.
- Optimistic parallel execution requires deterministic conflict detection.
- Read/write set tracking must identify write/write and read/write conflicts.
- DAG scheduler and safe concurrent state access are future production work.
- Scheduler output must fall back to sequential on conflict and must not
  introduce nondeterministic state writes.

## `x/storage`

Purpose: future KV state engine for versioned contract storage, snapshots, state
sync, and bounded iteration.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, ABCI hooks, snapshots,
  or state-sync integration are registered yet.

State:
- Contract namespace.
- Storage key format.
- Versioned key/value entry.
- Max state size.
- Storage rent/deposit accounting.
- Deterministic snapshot and state root.

Security invariants:
- Contract storage is namespaced.
- Storage keys and namespace lengths are bounded.
- Max state size is enforced before accepting a write.
- Bounded iteration requires an explicit positive limit.
- Export/import exact state must preserve version, entries, and state root.
- Snapshot/state-sync tests must prove imported state roots match exported
  roots before production wiring.

## `x/compute`

Purpose: measure CPU/compute usage separately from simple tx gas, price
expensive computation, and protect validators from CPU abuse.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, ABCI hooks, or fee
  deduction wiring are registered yet.

State:
- Compute unit schedule.
- Per-op cost.
- Per-contract compute stats.
- Per-block compute budget.

Security invariants:
- Expensive contract operations are charged more compute units than cheap
  operations.
- Per-contract and per-block compute caps are enforced deterministically.
- Compute accounting is sorted deterministically and rejects zero contract
  addresses.
- Compute pricing must complement gas limits and cannot bypass signer, fee,
  denom, zero-address, or VM validation.

## `x/permissions`

Purpose: ACL system for contracts, modules, resolver delegates, domain
managers, contract extensions, and governance-controlled permissions.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, ABCI hooks, or
  authority routing are registered yet.

State:
- Permission id.
- Owner address.
- Grantee address.
- Scope.
- Resource.
- Grant height.
- Expiry height.
- Revocation height.

Security invariants:
- All permissions have owner, scope, expiry, and revocation path.
- Permission checks are deterministic.
- Resolver delegate, domain manager, contract extension, module ACL,
  governance, and emergency permissions are explicit scopes.
- There is no hidden superuser outside explicit governance/emergency policy.
- Permission owner and grantee reject zero addresses.

## `x/indexer`

Purpose: fast query layer for state search, event search, memo search, domain
lookup, and asset discovery.

Current status:
- Pure executable projection specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, ABCI hooks, indexer
  service, or consensus dependency are registered yet.

State:
- Projection kind.
- Projection key.
- Owner.
- Height.
- Tx hash.
- Value.
- Canonical search fields.

Security invariants:
- Indexer must never be required for consensus.
- Query limits are required for bounded result sets.
- State, event, memo, domain, and asset indexes are projections only.
- Search output is deterministic for tests but cannot affect state
  transitions.

## `x/market`

Purpose: bounded, deterministic, and non-extractive market for compute,
storage, and execution priority.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, ABCI hooks, settlement,
  fee deduction, or scheduler integration are registered yet.

State:
- Market resource.
- Account.
- Quantity.
- Optional premium in `naet`.
- Base fee paid flag.
- Normal-user reservation flag.
- Sequence tie-breaker.

Security invariants:
- Market premiums cannot replace the base `naet` fee.
- Premiums are capped and priority score is capped by scheduler fairness.
- Wealthy users cannot fully starve normal users because normal-user reserved
  slots and per-account share caps are enforced deterministically.
- Market ordering must remain deterministic and cannot bypass fee, signer,
  denom, zero-address, scheduler, or VM validation.

## `x/scheduler-v2`

Purpose: DAG execution engine for parallel tx scheduling and async actor
mailbox planning.

Current status:
- Pure executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, ABCI hooks,
  parallel state access, or execution-pipeline integration are registered yet.

State:
- DAG task id.
- Actor id.
- Tx height, tx index, and message index ordering keys.
- Dependency ids.
- Deterministic read/write set.
- Replay hash.
- Actor mailbox plan.

Security invariants:
- Read/write sets are required and deterministic.
- Conflict resolution is deterministic and serializes conflicting tasks.
- Schedule replay hash is stable across input order.
- Actor mailbox planning is deterministic.
- Validators must get identical results before this can move from spec to
  production execution.

## `x/fees`

Purpose: centralize protocol fee policy and distribution.

State:
- Fee collector module account reference.
- Distribution weights.
- Accrued fee records where needed.
- Module params.

Minimal Msg surface:
- `MsgUpdateParams`
- Future fee claim/distribution messages only if they cannot be handled by hooks.

Keeper dependencies:
- Bank keeper interface for balances and transfers.
- Distribution or auth module interfaces only when explicitly required.

Security invariants:
- Distribution weights must sum to the configured denominator.
- Governance authority controls params.
- Fee collection must be idempotent for repeated block execution inputs.

## `x/reputation`

Purpose: future deterministic reputation state for anti-spam, scheduler
weighting, and progressive account or contract limits.

Current status:
- Pure validation and scoring helpers only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.

State:
- Reputation record keyed by account or contract address.
- Deterministic component scores for age, staking time, transaction success,
  bounded volume, domain ownership, and contract behavior.
- Penalties for spam, failed transactions, and slash events.
- Last updated epoch for inactivity decay.

Security invariants:
- Reputation scores are based only on deterministic on-chain events.
- There is no direct reputation purchase.
- Reputation staking may exist only as a bonded signal with slashing/risk.
- Domain ownership can add bounded reputation, but cannot dominate score.
- Contracts also have reputation.
- New accounts have progressive limits.
- Score is clamped to `0..100`; levels are restricted, new, normal, trusted,
  and elite.
- Low score lowers tx rate limit and async queue quota, raises memo/storage
  byte cost, and tightens contract deploy limits.
- High score may improve deterministic queue priority within bounded weights,
  but cannot bypass fees, signatures, or validation.
- Contract deployment may require a
  score threshold or bonded deposit.
- Domain auction spam can be rate-limited by score.
- Contract reputation updates on deterministic failed or successful executions.
- Reputation must not bypass signer checks, sequence replay protection,
  zero-address rejection, `naet` fee validation, or module authorization.


## `x/aetravm/async`

Purpose: define deterministic asynchronous contract message semantics before
keeper and VM runtime wiring.

State:
- Contract account model.
- Bounded contract state bytes.
- Message envelope model.
- Global message queue.
- Per-contract inbox/outbox views.
- Execution receipts and observability counters.

Current status:
- Pure Go executable specification only.
- No SDK stores, keepers, module accounts, or ABCI hooks are registered yet.
- Cosmos SDK delivered transactions remain synchronous; async semantics are
  modeled as deterministic queue processing inside blocks.
- Production partitioning or sharding is a later R&D track, not part of this
  module boundary.

Security invariants:
- Contract address derivation must be deterministic.
- Queue ordering must use tx index, message index, source logical time,
  destination key, and assigned sequence tie-breaker.
- Bounce and refund behavior must be deterministic and preserve failed-state
  rollback rules.
- Bounce/refund service messages must not create double-refund loops.
- All protocol fee and message value accounting is native `naet` only.
- Per-message gas limit and forward fee validation must be explicit.
- Per-tx, per-block, recursion, body, state, emitted-message, storage-write,
  and deploy limits must be bounded before VM wiring.
- Export/import must preserve queued messages, inbox, outbox, receipts, and
  metrics exactly, including `next_sequence`, `next_tx_index`, logical time,
  and ordering metadata.

## `x/aetravm/avm`

Purpose: define the native Aetra Virtual Machine before keeper and runtime
wiring.

State:
- Deterministic bytecode format.
- Module verifier.
- Local runner.
- Storage snapshot ABI.
- Gas schedule.
- Host function allowlist.
- Async handler adapter.

Current status:
- Pure Go executable specification only.
- No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks are
  registered yet.
- AVM cannot mutate production chain state until the base-chain safety gate,
  async queue semantics, security scans, determinism gate, and adversarial
  audit are green.

Security invariants:
- Bytecode serialization must be deterministic.
- Code hash must be computed from encoded bytecode.
- Message ABI must use the async `MessageEnvelope`.
- Storage ABI must use deterministic key/value snapshots with bounded memory.
- Host functions must be allowlisted.
- Wall-clock time, host randomness, filesystem/network access, floating point,
  unbounded iteration, and nondeterministic map iteration are forbidden.
- Gas, code size, memory, stack/register, import, and instruction limits must be
  bounded before keeper wiring.
- AVM must not bypass address validation, zero-address rejection, `naet` fee
  policy, signer checks, malformed transaction handling, or genesis validation.

## `x/sharding/sim`

Purpose: provide the sharding R&D simulator before any production sharding or
partitioning implementation.

State:
- In-memory masterchain state model.
- In-memory workchain registry.
- In-memory shardchain registry.
- Cross-shard message and receipt model.
- Equivocation evidence model.

Current status:
- Pure Go simulator only.
- No SDK stores, keepers, module accounts, genesis, ABCI hooks, consensus
  changes, or network partitioning are registered.
- No production sharding claim is allowed.

Security invariants:
- The simulator must not register SDK stores or mutate production chain state.
- Public wording must say sharding R&D or experimental sharding until the
  production gate passes.
- Masterchain state must commit validator set, staking snapshot, workchain
  registry, shard headers, cross-shard receipt roots, config updates, and
  equivocation evidence.
- Workchains must keep explicit VM set, address format, genesis hash, upgrade
  policy, and native `naet` fee policy.
- Shardchains must commit state root, message queue root, receipt root,
  validator subset, data availability status, and split/merge references.
- Cross-shard messages must reject duplicate receipts, missing receipts,
  invalid shard proofs, stale shard headers, wrong destination shards, replayed
  messages, validator equivocation, and data-unavailable shard blocks.
- Prototype keepers may begin only after simulator tests, fuzz tests,
  adversarial tests, long-run testnet, independent audit, and
  consensus-safety proof are complete.

## `app/wasmconfig`

Purpose: keep CosmWasm readiness gated until explicit app wiring is requested.

State:
- No chain state.
- Policy constants and validation helpers only.

Current status:
- AVM is the primary VM.
- CosmWasm is an optional gated compatibility layer.
- CosmWasm remains disabled by default.
- Enabling CosmWasm requires explicit config or feature gate.
- Upload, instantiate, admin/migration, gas, contract size, memory/cache, and
  query limits are defined before keeper wiring.
- Pinned code is disabled by default and governance-only if enabled later.
- Governance authority for enabling/disabling CosmWasm is explicit and must be
  a non-zero authority address.

Security invariants:
- CosmWasm readiness must not add a `wasm` store key, module account, genesis
  state, CLI tx surface, or keeper wiring by default.
- CosmWasm cannot bypass `naet` fee policy, address policy, zero-address
  policy, or genesis validation.
- Upload, instantiate, execute, query, migrate, pinning, code size, gas, and
  query response/depth limits must be enforced before state mutation when the
  real `x/wasm` keeper is wired.
- Any future non-AVM runtime requires a written binary serialization spec, message
  ABI, storage ABI, gas schedule, deterministic execution proof, fuzz tests,
  upgrade/migration policy, and adversarial audit before implementation.
- Contract standards must remain testable independent of optional CosmWasm or
  future VM compatibility layers.

## `x/bridge`

Purpose: future interoperability. It remains out of scope for the first scaffold.

Activation requires a separate design covering light-client verification, replay domains, validator or relayer trust assumptions, finality, rate limits, and emergency controls.
