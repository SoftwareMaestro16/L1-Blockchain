> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# TO_AUDIT: Aetra Automated Fuzzing And Exploit Campaign

This file defines the audit task list to run before a fuzzing/invariant
campaign. It is based on `docs/security/aetheris-fuzzing-invariant-pipeline.md`
and the Aetra attack catalog.

Runtime reports, generated scenarios, minimized exploits, and state diffs
should be written under `.work/aexs/` unless a later implementation explicitly
chooses another output path.

## Audit Goal

Treat every Aetra module as hostile input surface and prove, through
automated fuzzing, invariant checks, deterministic replay, and exploit
minimization, that the tested scope has:

- no exploitable red-team path;
- no deterministic execution divergence;
- no economic loop that mints, burns, drains, or bypasses fees incorrectly;
- no routing or load manipulation path;
- no domain/resolver hijack path;
- no AVM, queue, bounce, or refund double-spend path.

The audit is not passed until mandatory invariants have a `100%` pass rate and
every required module attack surface has at least `95%` planned fuzz/invariant
coverage.

## Source Documents And Scope

- [ ] Use `docs/security/aetheris-fuzzing-invariant-pipeline.md` as the primary
  pipeline task source.
- [ ] Use existing module-boundary and security docs to map module invariants.
- [ ] Include current implemented modules: `x/fees`, `x/tokenfactory`, `x/dex`,
  `x/aetherisvm`, and the Cosmos SDK base modules wired through `app`.
- [ ] Include future architecture surfaces where specs already exist:
  `x/execution`, `x/vm`, `x/messaging`, `x/queue`, `x/events`, `x/actors`,
  `x/scheduler`, `x/storage`, `x/identity`, `x/reputation`, `x/sharding/sim`.
- [ ] Keep all generated runtime artifacts under `.work/aexs/`.
- [ ] Do not mark a module safe if it has no invariant coverage.

## Audit Task Generation Rule

Every module must be decomposed into atomic audit tasks. No module can have
fewer than five audit tasks.

Each module must include at least:

- [ ] normal behavior verification;
- [ ] edge case testing;
- [ ] adversarial attack simulation;
- [ ] state integrity validation;
- [ ] economic manipulation test, if the module can affect value, fees,
  rewards, balances, routing cost, priority, or resource allocation.

Every atomic task must record:

- [ ] module;
- [ ] task id;
- [ ] function or flow covered;
- [ ] state transition covered;
- [ ] attack surface covered;
- [ ] invariant tested;
- [ ] defensive analysis result;
- [ ] adversarial simulation result;
- [ ] pass/fail result;
- [ ] reproduction seed or exact manual reproduction steps.

If any module has fewer than five tasks, the audit is incomplete. If any task
does not include both defensive analysis and adversarial simulation, the task is
invalid.

## Per-Task Red-Team Mode

Every task must be executed from two perspectives:

- Defensive analysis: define expected behavior, expected state transition,
  expected events, expected error path, and expected invariant.
- Adversarial simulation: attempt to break the expected behavior with malformed
  input, replay, unauthorized signer, bad fee, boundary values, state
  corruption attempt, or module-specific exploit.

No task is valid unless both perspectives are executed and recorded.

## Campaign Setup Tasks

- [ ] Create a deterministic campaign id.
- [ ] Record git commit, branch, dirty status, Go version, OS, and test command
  set.
- [ ] Record the fuzz seed list.
- [ ] Record the target modules.
- [ ] Record enabled runtime modes:
  - [ ] stateless fuzzing;
  - [ ] stateful multi-block fuzzing;
  - [ ] adversarial red-team fuzzing;
  - [ ] deterministic replay;
  - [ ] stress mode;
  - [ ] chaos mode.
- [ ] Record simulator mode:
  - [ ] in-memory app runner;
  - [ ] single-validator localnet;
  - [ ] multi-validator localnet;
  - [ ] sharding simulator.
- [ ] Define stop conditions:
  - [ ] first critical exploit;
  - [ ] max run count;
  - [ ] max wall-clock duration;
  - [ ] coverage threshold reached;
  - [ ] deterministic divergence.

## Scenario Generator Tasks

- [ ] Generate random bank transfer sequences.
- [ ] Generate random staking, delegation, unbonding, redelegation, and reward
  sequences.
- [ ] Generate random validator lifecycle and slashing evidence sequences.
- [ ] Generate random fee and spam bursts.
- [ ] Generate random tokenfactory create, mint, burn, and admin sequences.
- [ ] Generate random DEX create-pool, add-liquidity, remove-liquidity, and swap
  sequences.
- [ ] Generate random governance proposal, vote, and parameter update sequences.
- [ ] Generate random identity domain registration, auction, renewal, resolver,
  reverse lookup, and subdomain sequences.
- [ ] Generate random AVM deploy, external call, internal call, bounced call,
  query, and migrate sequences.
- [ ] Generate random async message, queue, delayed execution, bounce, and refund
  sequences.
- [ ] Generate random `LOAD_SCORE`, routing, and shard activation scenarios.
- [ ] Preserve seed and step list for every generated scenario.

## Transaction Mutator Tasks

- [ ] Inject invalid signatures.
- [ ] Replay already accepted transaction bytes.
- [ ] Manipulate nonce and sequence values.
- [ ] Corrupt fee denom and fee amount fields.
- [ ] Inject missing fee and non-`naet` fee paths.
- [ ] Inject extreme gas values.
- [ ] Inject malformed addresses.
- [ ] Inject zero address in signer, recipient, admin, authority, resolver, and
  DEX actor fields.
- [ ] Corrupt memo fields, including invalid UTF-8 and oversized memo payloads.
- [ ] Inject malformed routing hints.
- [ ] Inject invalid domain resolution and expired domain actions.
- [ ] Inject fake cross-zone messages.
- [ ] Inject queue depth abuse.
- [ ] Inject oversized AVM payloads.
- [ ] Inject invalid AVM entrypoint inputs.
- [ ] Inject malformed genesis fragments for simulator startup tests.
- [ ] Record mutation metadata for every scenario.

## Atomic Module Audit Tasks

Each task below must include both defensive analysis and adversarial simulation.

### `x/auth`

- [ ] AUTH-01 Normal behavior: verify valid signatures, signer extraction, and
  sequence increment for accepted transactions.
- [ ] AUTH-02 Edge cases: test empty signer set, duplicate signers, malformed
  public keys, wrong chain id, and max-size auth info.
- [ ] AUTH-03 Adversarial: attempt invalid signature injection, replayed signed
  bytes, tx malleability, and nonce manipulation.
- [ ] AUTH-04 State integrity: prove rejected auth paths do not increment
  sequence or mutate account state.
- [ ] AUTH-05 Economic abuse: prove auth failure cannot bypass fee deduction,
  priority rules, or rate limits.

### `x/bank`

- [ ] BANK-01 Normal behavior: verify `naet` sends, module account transfers,
  and multi-send success paths.
- [ ] BANK-02 Edge cases: test zero amount, max amount, insufficient funds,
  malformed denom, zero address, and self-transfer.
- [ ] BANK-03 Adversarial: attempt double spend, partial multi-send failure,
  negative balance creation, and overflow.
- [ ] BANK-04 State integrity: prove balances, module balances, and total supply
  remain consistent after accepted and rejected sends.
- [ ] BANK-05 Economic abuse: prove bank paths cannot mint, burn, spoof native
  denom metadata, or pay protocol fees with non-`naet` assets.

### `x/staking`

- [ ] STAKE-01 Normal behavior: verify validator creation, delegation,
  redelegation, unbonding, and reward eligibility.
- [ ] STAKE-02 Edge cases: test invalid validator address, non-`naet` bond denom,
  zero self-delegation, max commission, and unbonding boundary windows.
- [ ] STAKE-03 Adversarial: attempt stake grinding, delegation manipulation,
  reward farming loop, and validator power spoofing.
- [ ] STAKE-04 State integrity: prove validator tokens, delegator shares, staking
  pools, and validator-set updates remain consistent.
- [ ] STAKE-05 Economic abuse: prove staking cannot inflate rewards, bypass
  unbonding risk, or create slash-immune stake.

### `x/slashing`

- [ ] SLASH-01 Normal behavior: verify downtime and equivocation evidence
  updates validator status and stake penalties.
- [ ] SLASH-02 Edge cases: test stale evidence, duplicate evidence, unknown
  validator, jailed validator, and evidence at boundary heights.
- [ ] SLASH-03 Adversarial: attempt slashing bypass through redelegation,
  unbonding, delayed evidence, and malformed proof.
- [ ] SLASH-04 State integrity: prove slash accounting, jailed/tombstoned state,
  and validator-set removal are deterministic.
- [ ] SLASH-05 Economic abuse: prove slashed stake cannot be recovered through
  timing, migration, export/import, or governance parameter race.

### `x/gov`

- [ ] GOV-01 Normal behavior: verify proposal creation, voting, tallying,
  parameter update, and delayed execution.
- [ ] GOV-02 Edge cases: test proposal spam, malformed params, expired voting
  period, zero deposit, and invalid authority.
- [ ] GOV-03 Adversarial: attempt governance replay, proposal front-running,
  emergency parameter abuse, and upgrade hijack.
- [ ] GOV-04 State integrity: prove accepted proposals update only authorized
  params and rejected proposals leave state unchanged.
- [ ] GOV-05 Economic abuse: prove governance cannot set fee, inflation, staking,
  or burn params outside hard protocol bounds.

### `x/distribution`

- [ ] DIST-01 Normal behavior: verify validator commission, delegator rewards,
  community pool accounting, and reward withdrawal.
- [ ] DIST-02 Edge cases: test tiny rewards, rounding remainders, jailed
  validators, zero delegations, and repeated withdrawals.
- [ ] DIST-03 Adversarial: attempt reward double claim, commission bypass,
  reward inflation, and module balance desync.
- [ ] DIST-04 State integrity: prove outstanding rewards, module balances, and
  supply accounting stay deterministic.
- [ ] DIST-05 Economic abuse: prove distribution cannot mint outside the
  authorized reward path or leak treasury/community-pool funds.

### `app` / BaseApp

- [ ] APP-01 Normal behavior: verify BaseApp module wiring, ante ordering,
  genesis validation, export, and deterministic empty-block execution.
- [ ] APP-02 Edge cases: test malformed genesis fragments, missing module
  state, duplicate accounts, invalid params, and export from partial state.
- [ ] APP-03 Adversarial: attempt app hash divergence through replayed tx
  sequences, nondeterministic module ordering, malformed tx bytes, and invalid
  proposal data.
- [ ] APP-04 State integrity: prove same genesis plus same block and tx
  sequence yields the same app hash, exported state, validator updates, and
  module roots.
- [ ] APP-05 Economic abuse: prove app wiring cannot bypass ante fee checks,
  signer checks, staking/slashing authority, governance authority, or module
  account permissions.

### `x/fees`

- [ ] FEES-01 Normal behavior: verify valid `naet` fee collection, min fee,
  split accounting, and query params.
- [ ] FEES-02 Edge cases: test missing fee, zero fee, multi-denom fee, malformed
  fee, max fee, and simulation mode.
- [ ] FEES-03 Adversarial: attempt fee underpayment, fee denom spoofing,
  fee-griefing spam, and non-FeeTx bypass.
- [ ] FEES-04 State integrity: prove failed ante checks do not execute messages
  or corrupt fee accounting.
- [ ] FEES-05 Economic abuse: prove fee split, burn, treasury, and validator
  reward accounting cannot be manipulated by tx shape or load state.

### `x/tokenfactory`

- [ ] TF-01 Normal behavior: verify create denom, mint, burn, change admin, and
  metadata query paths.
- [ ] TF-02 Edge cases: test invalid subdenom, duplicate denom, zero admin,
  native denom spoof, and max metadata size.
- [ ] TF-03 Adversarial: attempt unauthorized mint, unauthorized burn, admin
  takeover, metadata spoofing, and burn-from mismatch.
- [ ] TF-04 State integrity: prove supply changes exactly by minted or burned
  amount and authority metadata remains consistent.
- [ ] TF-05 Economic abuse: prove tokenfactory assets cannot pay protocol fees,
  spoof AET, or inflate native supply.

### `x/dex`

- [ ] DEX-01 Normal behavior: verify pool creation, add liquidity, remove
  liquidity, swap, LP mint, and LP burn.
- [ ] DEX-02 Edge cases: test duplicate pair, tiny reserves, zero liquidity,
  same denom pair, invalid pool id, and max amount.
- [ ] DEX-03 Adversarial: attempt pool drain, LP inflation, reserve desync,
  failed bank movement partial update, and slippage bypass.
- [ ] DEX-04 State integrity: prove reserves match module balances, LP supply
  matches shares, and failed operations leave pool state unchanged.
- [ ] DEX-05 Economic abuse: prove constant-product and fee-adjusted swap math
  cannot be exploited through rounding, ordering, or malformed denoms.

### `x/identity`

- [ ] ID-01 Normal behavior: verify domain auction, assignment, renewal,
  expiry, resolver update, reverse lookup, and subdomain flow.
- [ ] ID-02 Edge cases: test invalid name, duplicate name, expired domain,
  missing resolver, zero resolver, and max metadata.
- [ ] ID-03 Adversarial: attempt domain hijack, auction manipulation, resolver
  overwrite, reverse lookup poisoning, and subdomain collision.
- [ ] ID-04 State integrity: prove registry owner, resolver record, expiry, and
  NFT representation do not diverge.
- [ ] ID-05 Economic abuse: prove auction bids, renewal fees, refunds, and
  domain payments cannot be stolen or routed to invalid targets.

### `x/reputation`

- [ ] REP-01 Normal behavior: verify score updates, decay, level assignment,
  rate limit, and priority signal.
- [ ] REP-02 Edge cases: test score floor/ceiling, inactive accounts, new
  accounts, zero activity, and max activity.
- [ ] REP-03 Adversarial: attempt reputation farming, sybil bypass, spam with
  low score, and priority manipulation.
- [ ] REP-04 State integrity: prove score updates are deterministic and do not
  diverge across replay/export/import.
- [ ] REP-05 Economic abuse: prove reputation cannot be bought directly and
  cannot bypass required fees, deposits, or signer checks.

### `x/execution`

- [ ] EXEC-01 Normal behavior: verify transaction pipeline order, dispatch,
  route output, events, and deterministic trace.
- [ ] EXEC-02 Edge cases: test malformed payload, missing route, invalid module,
  failed dispatch, and max tx size.
- [ ] EXEC-03 Adversarial: attempt partial rollback, wrong module dispatch,
  invalid state transition after ante failure, and execution desync.
- [ ] EXEC-04 State integrity: prove failed execution does not commit partial
  writes and accepted execution emits stable receipts.
- [ ] EXEC-05 Economic abuse: prove execution cannot bypass fee, gas, memo,
  reputation, or routing constraints.

### `x/vm` And AVM

- [ ] VM-01 Normal behavior: verify AVM deploy, external call, internal call,
  bounced call, query, and migrate entrypoint validation.
- [ ] VM-02 Edge cases: test max code size, missing entrypoint, bad code hash,
  zero gas, max gas, and malformed bytecode.
- [ ] VM-03 Adversarial: attempt VM crash input, infinite loop, stack overflow,
  sandbox escape, and nondeterministic host behavior.
- [ ] VM-04 State integrity: prove contract state changes are deterministic and
  rejected execution cannot commit state.
- [ ] VM-05 Economic abuse: prove AVM cannot underpay gas, pay protocol fees in
  non-`naet`, double-refund, or bypass storage limits.

### `x/messaging`

- [ ] MSG-01 Normal behavior: verify async send, internal message delivery,
  proof/receipt fields, and cross-zone message classification.
- [ ] MSG-02 Edge cases: test missing destination, expired message, max body,
  zero value, malformed opcode, and invalid query id.
- [ ] MSG-03 Adversarial: attempt message replay, message ordering attack,
  forged proof, stale receipt replay, and message starvation.
- [ ] MSG-04 State integrity: prove message state, receipts, and queue entries
  are deterministic across replay/export/import.
- [ ] MSG-05 Economic abuse: prove message forwarding fees, value transfer,
  bounce, and refund cannot double-spend.

### `x/queue`

- [ ] QUEUE-01 Normal behavior: verify enqueue, delayed execution, dequeue,
  bounce, refund, and per-block processing limit.
- [ ] QUEUE-02 Edge cases: test empty queue, max queue, max depth, expired item,
  duplicate sequence, and missing actor.
- [ ] QUEUE-03 Adversarial: attempt queue flooding, message loop, starvation,
  priority manipulation, and duplicate sequence injection.
- [ ] QUEUE-04 State integrity: prove queue ordering and sequence counters are
  deterministic and export/import stable.
- [ ] QUEUE-05 Economic abuse: prove queued value cannot be refunded twice,
  forwarded without fee, or trapped by malformed bounce path.

### `x/events`

- [ ] EVENTS-01 Normal behavior: verify deterministic event emission for bank,
  fees, DEX, identity, execution, queue, and memo paths.
- [ ] EVENTS-02 Edge cases: test empty attributes, max attribute size,
  duplicate event keys, and failed tx event behavior.
- [ ] EVENTS-03 Adversarial: attempt event spoofing, inconsistent event order,
  and misleading success event after failure.
- [ ] EVENTS-04 State integrity: prove events match committed state and receipts.
- [ ] EVENTS-05 Economic abuse: prove events cannot be used as authority for
  balances, fees, resolver targets, or execution success.

### `x/actors`

- [ ] ACTOR-01 Normal behavior: verify actor lifecycle, mailbox processing,
  logical time, and isolated state transition.
- [ ] ACTOR-02 Edge cases: test missing actor, inactive actor, max mailbox,
  max state size, and actor deletion/migration boundaries.
- [ ] ACTOR-03 Adversarial: attempt cross-actor direct state mutation, mailbox
  flood, logical-time spoof, and actor takeover.
- [ ] ACTOR-04 State integrity: prove one actor cannot mutate another actor
  except through committed messages.
- [ ] ACTOR-05 Economic abuse: prove actor storage, execution, and message costs
  cannot be avoided through actor splitting or mailbox abuse.

### `x/scheduler`

- [ ] SCHED-01 Normal behavior: verify deterministic ordering, task selection,
  read/write set handling, and priority class handling.
- [ ] SCHED-02 Edge cases: test empty plan, duplicate task id, max tasks,
  conflicting read/write sets, and dependency boundaries.
- [ ] SCHED-03 Adversarial: attempt scheduling manipulation, starvation,
  priority gaming, and nondeterministic tie-break.
- [ ] SCHED-04 State integrity: prove the same tasks and state produce the same
  execution plan across nodes.
- [ ] SCHED-05 Economic abuse: prove priority or market signals cannot starve
  normal users or bypass fee/reputation caps.

### `x/storage`

- [ ] STORE-01 Normal behavior: verify KV writes, reads, versioning, snapshots,
  export/import, and state sync.
- [ ] STORE-02 Edge cases: test max key, max value, empty value, duplicate key,
  deleted key, and pagination boundaries.
- [ ] STORE-03 Adversarial: attempt state root collision, snapshot poisoning,
  malformed import, and unbounded iteration.
- [ ] STORE-04 State integrity: prove committed state root, snapshot root, and
  exported state are deterministic.
- [ ] STORE-05 Economic abuse: prove storage growth, storage rent/deposit, and
  contract state size limits cannot be bypassed.

### `x/memo`

- [ ] MEMO-01 Normal behavior: verify optional UTF-8 memo on bank, identity,
  token, DEX, and contract calls.
- [ ] MEMO-02 Edge cases: test empty memo, max memo, invalid UTF-8, control
  chars, and oversized byte length.
- [ ] MEMO-03 Adversarial: attempt memo spam, binary payload injection,
  indexing abuse, and misleading memo on failed tx.
- [ ] MEMO-04 State integrity: prove memo metadata is immutable after block
  inclusion and cannot alter execution result.
- [ ] MEMO-05 Economic abuse: prove memo cost, byte fee, and reputation
  multiplier cannot be bypassed.

### `x/indexer`

- [ ] INDEX-01 Normal behavior: verify query indexing for tx hash, sender,
  receiver, domain, contract, memo, event, token, and NFT surfaces.
- [ ] INDEX-02 Edge cases: test empty result, pagination, duplicate records,
  deleted state, and max query size.
- [ ] INDEX-03 Adversarial: attempt index poisoning, stale resolver lookup,
  fake event indexing, and inconsistent domain cache.
- [ ] INDEX-04 State integrity: prove index output never overrides consensus
  state and can be rebuilt from committed events/state.
- [ ] INDEX-05 Economic abuse: prove index priority/search cannot route funds,
  change balances, or bypass protocol fees.

### `x/sharding/sim` And Load/Routing

- [ ] SHARD-01 Normal behavior: verify `LOAD_SCORE`, zone selection, shard
  activation, shard assignment, and commitment output.
- [ ] SHARD-02 Edge cases: test zero load, max load, oscillating load, empty
  shard, max shard count, and routing epoch changes.
- [ ] SHARD-03 Adversarial: attempt load poisoning, shard overload targeting,
  routing loop, route desync, and shard starvation.
- [ ] SHARD-04 State integrity: prove same tx and state produce same route,
  same shard, same commitment, and same replay output.
- [ ] SHARD-05 Economic abuse: prove fee level, reputation, or priority cannot
  manipulate routing outside deterministic protocol rules.

## Invariant Checklist

### Economic Invariants

- [ ] `sum(account balances) + module balances + burned + staked = total_supply`.
- [ ] No negative balances.
- [ ] No unauthorized mint.
- [ ] No unauthorized burn.
- [ ] No fee denom other than `naet` is accepted.
- [ ] Fee distribution totals match collected fees.
- [ ] Treasury, burn, validator reward, and community pool accounting is
  deterministic.
- [ ] Staking rewards cannot be farmed through state loops.
- [ ] Supply cannot drift after export/import.

### Consensus And State Invariants

- [ ] Same block input produces same app hash.
- [ ] Same signed transaction cannot execute twice.
- [ ] Invalid signer cannot mutate state.
- [ ] Malformed transaction fails before state mutation.
- [ ] Validator set matches staking keeper state.
- [ ] Slashing evidence is objective and deterministic.
- [ ] Genesis validation rejects malformed accounts, balances, params, and
  module state.
- [ ] Upgrade and migration paths preserve state roots and module invariants.

### DEX Invariants

- [ ] Pool reserves match module account balances.
- [ ] LP supply matches pool total shares.
- [ ] No negative liquidity.
- [ ] No fake LP token.
- [ ] Swap output is non-negative.
- [ ] Slippage bounds are enforced.
- [ ] Constant-product constraints hold after fee-adjusted swaps.
- [ ] Failed bank movement cannot mutate pool state.

### Load, Routing, And Sharding Invariants

- [ ] `LOAD_SCORE` is always in `[0,1]`.
- [ ] EMA smoothing is deterministic.
- [ ] `LOAD_SCORE` cannot jump beyond `MAX_DELTA`.
- [ ] Same transaction and state produce the same zone decision.
- [ ] Same transaction and state produce the same shard decision.
- [ ] No routing loop.
- [ ] No shard starvation.
- [ ] No hot-zone monopolization.
- [ ] No priority ordering divergence across nodes.

### Identity And Resolver Invariants

- [ ] Domain names are unique.
- [ ] Active domains cannot be re-auctioned.
- [ ] Expired domains require auction or explicit renewal path.
- [ ] Resolver cannot point to malformed or zero address.
- [ ] Resolver-based payment fails before funds move if unresolved.
- [ ] Reverse lookup is consistent with owner-approved mapping.
- [ ] Domain registry owner and NFT representation owner do not diverge.
- [ ] Subdomain ownership and resolver delegation do not bypass parent rules.

### Execution, AVM, And Queue Invariants

- [ ] AVM malformed input does not panic.
- [ ] AVM gas is bounded and deterministic.
- [ ] Infinite loops are stopped by gas or instruction limits.
- [ ] Contract state updates are deterministic.
- [ ] Queue ordering is deterministic.
- [ ] Cross-zone message replay is rejected.
- [ ] Bounce/refund cannot double-spend.
- [ ] Message loops are bounded by depth and per-block limits.
- [ ] Export/import preserves queue state exactly.

## Exploit Task Catalog

Every exploit task must produce a seed, step list, expected state, actual state,
affected module list, severity, and fix recommendation if it succeeds.

### 1. Consensus And Aether Core Exploits

- [ ] Attempt double-sign fork creation.
- [ ] Attempt equivocation across heights and rounds.
- [ ] Attempt long-range history rewrite.
- [ ] Attempt stake grinding.
- [ ] Attempt validator cartel concentration scenario.
- [ ] Attempt stake delegation manipulation.
- [ ] Attempt self-delegation inflation.
- [ ] Attempt fake validator liveness.
- [ ] Attempt validator eclipse simulation.
- [ ] Attempt block withholding.
- [ ] Attempt fork choice manipulation.
- [ ] Attempt finality delay manipulation.
- [ ] Attempt Byzantine majority simulator scenario.

### 2. Slashing Bypass Exploits

- [ ] Attempt delayed evidence submission bypass.
- [ ] Attempt malformed equivocation proof acceptance.
- [ ] Attempt slashing race condition.
- [ ] Attempt redelegation-based partial slash evasion.
- [ ] Attempt unbonding window slash evasion.
- [ ] Attempt jail escape through upgrade timing.
- [ ] Attempt invalid evidence replay.

### 3. Transaction, Auth, And Bank Exploits

- [ ] Attempt signature replay.
- [ ] Attempt cross-context replay with wrong chain id.
- [ ] Attempt invalid nonce bypass.
- [ ] Attempt transaction malleability.
- [ ] Attempt fee underpayment bypass.
- [ ] Attempt fee inflation manipulation.
- [ ] Attempt low-fee spam griefing.
- [ ] Attempt multi-send partial failure exploit.
- [ ] Attempt race-condition double spend.
- [ ] Attempt rollback exploit during replayed state transition.
- [ ] Attempt zero-address transfer or signer path.

### 4. Token And Economy Exploits

- [ ] Attempt tokenfactory mint authority takeover.
- [ ] Attempt unauthorized burn bypass.
- [ ] Attempt inflation manipulation through governance timing.
- [ ] Attempt fee routing manipulation.
- [ ] Attempt treasury drain via governance proposal.
- [ ] Attempt staking reward inflation.
- [ ] Attempt staking reward farming loop.
- [ ] Attempt supply manipulation through edge-case mint path.
- [ ] Attempt native denom spoofing.
- [ ] Attempt display/base decimal mismatch exploit.

### 5. DEX Exploits

- [ ] Attempt constant-product invariant break.
- [ ] Attempt liquidity drain through swap sequence.
- [ ] Attempt pool initialization manipulation.
- [ ] Attempt LP token inflation.
- [ ] Attempt liquidity removal race.
- [ ] Attempt zero-liquidity swap edge case.
- [ ] Attempt reserve/module balance desync.
- [ ] Attempt failed bank movement partial update.
- [ ] Attempt slippage bypass.
- [ ] Attempt rounding exploit.

### 6. Load System Exploits

- [ ] Attempt `LOAD_SCORE` manipulation through spam bursts.
- [ ] Attempt artificial mempool inflation.
- [ ] Attempt block saturation.
- [ ] Attempt execution delay amplification.
- [ ] Attempt EMA slow-poison attack.
- [ ] Attempt load spike oscillation.
- [ ] Attempt shard overload targeting through load manipulation.
- [ ] Attempt priority fee gaming.
- [ ] Attempt adaptive fee destabilization.

### 7. Routing Engine Exploits

- [ ] Attempt routing bias exploitation.
- [ ] Attempt zone congestion targeting.
- [ ] Attempt compute shard starvation.
- [ ] Attempt hot-zone monopolization.
- [ ] Attempt deterministic route prediction abuse.
- [ ] Attempt cross-zone routing loop.
- [ ] Attempt routing desync between nodes.
- [ ] Attempt transaction misclassification.
- [ ] Attempt fee-based routing gaming.

### 8. Execution Zone And AVM Exploits

- [ ] Attempt state divergence between zones.
- [ ] Attempt cross-zone replay.
- [ ] Attempt AVM determinism violation.
- [ ] Attempt contract execution desync.
- [ ] Attempt parallel execution race condition.
- [ ] Attempt state corruption.
- [ ] Attempt partial execution rollback.
- [ ] Attempt nondeterministic opcode or host behavior.
- [ ] Attempt gas exhaustion denial-of-service.
- [ ] Attempt infinite loop griefing.
- [ ] Attempt storage collision.
- [ ] Attempt contract upgrade hijack.
- [ ] Attempt uninitialized storage exploit.
- [ ] Attempt stack overflow.
- [ ] Attempt sandbox escape.

### 9. Compute Shard Exploits

- [ ] Attempt shard partition imbalance.
- [ ] Attempt shard starvation.
- [ ] Attempt shard overflow collapse.
- [ ] Attempt cross-shard inconsistency.
- [ ] Attempt load spoofing for shard activation.
- [ ] Attempt shard duplication.
- [ ] Attempt state split inconsistency.
- [ ] Attempt parallel execution collision.
- [ ] Attempt scheduling manipulation.
- [ ] Attempt queue flooding.

### 10. Aether Mesh And Cross-Zone Exploits

- [ ] Attempt cross-zone message replay.
- [ ] Attempt message delay manipulation.
- [ ] Attempt message ordering attack.
- [ ] Attempt asset duplication across zones.
- [ ] Attempt double spend across zones.
- [ ] Attempt proof forgery.
- [ ] Attempt relay censorship simulation.
- [ ] Attempt message starvation.
- [ ] Attempt finality mismatch.
- [ ] Attempt stale receipt replay.

### 11. Identity And `.aet` Domain Exploits

- [ ] Attempt domain hijack through resolver overwrite.
- [ ] Attempt expired domain takeover without auction.
- [ ] Attempt auction manipulation.
- [ ] Attempt resolver spoofing.
- [ ] Attempt subdomain collision.
- [ ] Attempt reverse lookup poisoning.
- [ ] Attempt domain binding race condition.
- [ ] Attempt index-layer cache poisoning.
- [ ] Attempt fake domain resolution injection.
- [ ] Attempt multi-resolver inconsistency.

### 12. Governance Exploits

- [ ] Attempt governance capture through voting power manipulation.
- [ ] Attempt proposal spam.
- [ ] Attempt emergency parameter abuse.
- [ ] Attempt upgrade hijack.
- [ ] Attempt delayed execution exploitation.
- [ ] Attempt governance replay.
- [ ] Attempt proposal front-running.
- [ ] Attempt staking-loop voting power manipulation.
- [ ] Attempt parameter griefing.

### 13. Genesis, Upgrade, And State Exploits

- [ ] Attempt malformed genesis injection.
- [ ] Attempt state export tampering.
- [ ] Attempt upgrade rollback.
- [ ] Attempt partial migration corruption.
- [ ] Attempt module initialization bypass.
- [ ] Attempt hidden privileged account injection.
- [ ] Attempt `InitGenesis` validation bypass.
- [ ] Attempt version mismatch exploit.
- [ ] Attempt state root collision.
- [ ] Attempt snapshot poisoning.

### 14. Mempool And Network Exploits

- [ ] Attempt mempool flooding.
- [ ] Attempt transaction prioritization gaming.
- [ ] Attempt gossip poisoning.
- [ ] Attempt node eclipse simulation.
- [ ] Attempt P2P partition simulation.
- [ ] Attempt block propagation delay.
- [ ] Attempt transaction reordering.
- [ ] Attempt network latency exploitation.
- [ ] Attempt bandwidth exhaustion.
- [ ] Attempt peer targeting.

### 15. Combined Full-Stack Exploits

- [ ] Attempt coordinated spam plus routing attack.
- [ ] Attempt load plus governance combined attack.
- [ ] Attempt DEX plus mempool plus routing exploit.
- [ ] Attempt validator collusion plus slashing delay exploit.
- [ ] Attempt cross-zone value extraction coordination.
- [ ] Attempt identity plus routing hijack.
- [ ] Attempt shard overload plus fee manipulation cascade.
- [ ] Attempt consensus plus mempool denial-of-service hybrid.
- [ ] Attempt economic plus staking starvation.
- [ ] Attempt full-stack destabilization.

## Coverage Matrix Tasks

- [ ] Build module-to-scenario coverage matrix.
- [ ] Build module-to-invariant coverage matrix.
- [ ] Build module-to-exploit coverage matrix.
- [ ] For every module, fill functions covered, state transitions covered,
  attack surfaces covered, and invariants tested.
- [ ] Fail audit if any coverage matrix cell is empty.
- [ ] Fail audit when any required module attack surface has less than `95%`
  planned fuzz/invariant coverage.
- [ ] Fail final audit for the selected tested scope when mandatory executed
  coverage is less than `100%`.
- [ ] Fail production-safe decision unless mandatory invariant pass rate is
  `100%`.
- [ ] Track coverage for:
  - [ ] `x/auth`;
  - [ ] `x/bank`;
  - [ ] `x/staking`;
  - [ ] `x/slashing`;
  - [ ] `x/fees`;
  - [ ] `x/tokenfactory`;
  - [ ] `x/dex`;
  - [ ] `x/identity`;
  - [ ] `x/reputation`;
  - [ ] `x/execution`;
  - [ ] `x/vm`;
  - [ ] `x/messaging`;
  - [ ] `x/queue`;
  - [ ] `x/sharding/sim`;
  - [ ] AVM;
  - [ ] load/routing.

## Mandatory Coverage Matrix

Before finalizing an audit campaign, fill and verify this matrix. If any cell
is missing, the audit is incomplete.

| Module/system | Functions covered | State transitions covered | Attack surfaces covered | Invariants tested |
| --- | --- | --- | --- | --- |
| `x/auth` | signature verification, signer extraction, sequence validation | account sequence increment, auth rejection | invalid signature, replay, nonce manipulation, tx malleability | no replay, invalid signer cannot mutate state |
| `x/bank` | send, multi-send, module transfer, balance query | account balance changes, module balance changes | double spend, overflow, zero address, partial multi-send | no negative balances, total supply consistency |
| `x/staking` | create validator, delegate, redelegate, unbond, rewards eligibility | validator power, delegation shares, staking pools | stake grinding, reward farming, non-`naet` bond, slash evasion | validator set consistency, staking pool consistency |
| `x/slashing` | downtime, equivocation evidence, jail, tombstone | slash amount, validator status, active set removal | delayed evidence, malformed proof, redelegation evasion | deterministic evidence, slash accounting consistency |
| `x/gov` | proposal, vote, tally, param update, upgrade execution | proposal status, params, delayed execution state | governance replay, parameter abuse, upgrade hijack | authorized params only, hard bounds preserved |
| `x/distribution` | reward accrual, commission, withdraw, community pool | reward balances, outstanding rewards, commission state | double claim, reward inflation, module balance desync | reward determinism, distribution balance consistency |
| `app` / BaseApp | module wiring, ante ordering, genesis validation, export, deterministic empty blocks | app hash, validator updates, module roots, exported state | replayed tx sequence, malformed tx bytes, nondeterministic module order, invalid proposal data | same inputs produce same app hash, failed ante cannot mutate state |
| `x/fees` | ante fee validation, fee split, params query | fee collection, burn/treasury/reward accounting | fee denom spoof, underpayment, non-FeeTx bypass | `naet`-only fees, exact fee distribution |
| `x/tokenfactory` | create denom, mint, burn, change admin, metadata | denom state, supply, admin authority | unauthorized mint/burn, admin takeover, native spoof | supply delta exact, authority consistency |
| `x/dex` | create pool, add liquidity, remove liquidity, swap | reserves, LP supply, pool indexes | pool drain, reserve desync, LP inflation, slippage bypass | reserves match balances, LP supply matches shares |
| `x/identity` | auction, assign, renew, expire, resolver update | domain record, resolver record, NFT representation | resolver hijack, expired reuse, auction manipulation | domain uniqueness, resolver validity, owner consistency |
| `x/reputation` | score update, decay, rate limit, priority signal | score, level, limits, priority class | sybil, farming, low-score spam, priority manipulation | score bounds, deterministic replay, no fee bypass |
| `x/execution` | validation pipeline, dispatch, route, receipt, event | execution result, state writes, rollback | partial commit, wrong dispatch, ante bypass | failed tx no mutation, deterministic trace |
| `x/vm` / AVM | deploy, external/internal/bounced call, query, migrate | contract state, gas use, output messages | crash input, infinite loop, sandbox escape, bad bytecode | bounded gas, deterministic state, no panic |
| `x/messaging` | send, receive, receipt, proof, replay check | message state, receipt state, cross-zone state | replay, forged proof, stale receipt, ordering attack | deterministic ordering, no double spend |
| `x/queue` | enqueue, dequeue, delay, bounce, refund | queue item, sequence, depth, refund state | queue flood, message loop, double refund, starvation | queue order, depth bounds, refund uniqueness |
| `x/events` | event emit, event ordering, event query | event stream, receipt linkage | spoofed event, misleading success, inconsistent order | events match committed state |
| `x/actors` | actor lifecycle, mailbox, logical time, state transition | actor state, mailbox state, logical time | direct state mutation, mailbox flood, actor takeover | actor isolation, monotonic logical time |
| `x/scheduler` | task planning, priority, read/write sets, dependencies | plan output, task status, conflict result | starvation, priority gaming, nondeterministic tie-break | same input yields same plan |
| `x/storage` | KV read/write, snapshot, export/import, pagination | key/value state, version, snapshot root | state root collision, snapshot poisoning, unbounded iteration | deterministic root, export/import equality |
| `x/memo` | memo validation, memo fee, memo indexing, memo event | tx metadata, memo record, index entry | memo spam, invalid UTF-8, binary injection | memo immutable, memo cannot affect execution |
| `x/indexer` | tx/event/domain/memo/token query, pagination | index record, rebuild state, query cursor | stale cache, fake event, index poisoning | index rebuildable, no consensus authority |
| `x/sharding/sim` and load/routing | `LOAD_SCORE`, zone route, shard route, commitment | load state, route decision, shard assignment | load poisoning, route loop, shard starvation | score bounds, deterministic route, no starvation |

## Exploit Report Requirements

Every successful exploit or invariant break must generate an `AUDIT_RESULT.md`
entry with:

- [ ] campaign id;
- [ ] seed;
- [ ] block height;
- [ ] attacker profile;
- [ ] exploit category;
- [ ] affected modules;
- [ ] transaction sequence;
- [ ] mutation sequence;
- [ ] expected state;
- [ ] actual state;
- [ ] state diff;
- [ ] invariant that failed;
- [ ] severity;
- [ ] minimal reproduction steps;
- [ ] suspected root cause;
- [ ] fix recommendation;
- [ ] regression test recommendation.

Severity rules:

- Critical: unauthorized mint/burn, double spend, app hash divergence, fee
  bypass at scale, slashing bypass, zone/root commitment corruption.
- High: DEX drain, resolver hijack, routing manipulation, queue double-refund,
  validator-set inconsistency, AVM state corruption.
- Medium: bounded DoS, non-critical accounting drift, delayed execution
  inconsistency caught before finality.
- Low: non-consensus reporting, indexing, or UX issue without fund-safety impact.

## Audit Execution Gates

- [ ] Run stateless fuzzing before stateful fuzzing.
- [ ] Run stateful fuzzing before chaos mode.
- [ ] Run invariant checks after every simulated block.
- [ ] Run deterministic replay for every failing seed.
- [ ] Minimize every exploit sequence before triage.
- [ ] Generate `AUDIT_RESULT.md` for every campaign.
- [ ] Do not mark safe if any Critical or High exploit remains.
- [ ] Do not mark safe if deterministic replay cannot reproduce the result.
- [ ] Do not mark safe if coverage thresholds are not met.

## Final Audit Decision

The tested scope is safe only if:

- [ ] mandatory invariant pass rate is `100%`;
- [ ] no exploitable red-team path remains;
- [ ] no deterministic execution violation occurs;
- [ ] no economic loop is exploitable;
- [ ] no unauthorized mint, burn, fee bypass, double spend, resolver hijack, or
  queue double-refund remains;
- [ ] every coverage target is at or above `95%`;
- [ ] all generated exploit reports are triaged;
- [ ] every fixed exploit has a regression test task.
