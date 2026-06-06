> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Aetra Automated Fuzzing And Invariant Testing Pipeline

This document defines the implementation task list for the Aetra automated
exploit simulator and invariant testing pipeline. It is a security engineering
roadmap, not runtime code.

The pipeline turns Aetra into a continuously attacked simulation
environment: every module, transaction path, routing decision, execution queue,
and economic rule is treated as a hostile target under adversarial fuzzing and
mandatory invariant enforcement.

## Purpose

The pipeline is intended to:

- automatically discover vulnerabilities;
- verify economic invariants;
- test consensus, base-chain, and execution-layer behavior;
- simulate red-team attack patterns;
- stress load, routing, queue, and sharding R&D systems;
- generate reproducible exploit reports when an invariant breaks.

The system is considered safe only when mandatory invariants pass at 100%,
red-team mode finds no exploitable path, deterministic execution does not
diverge, and economic loops cannot be exploited.

## Architecture

The pipeline has seven layers:

```text
Scenario Generator
  -> Transaction Mutator / Attack Generator
  -> Chain Simulator
  -> State Snapshot Comparator
  -> Invariant Checker
  -> Exploit Classifier
  -> Report Generator
```

Required targets:

- core modules: `x/auth`, `x/bank`, `x/staking`, `x/slashing`, `x/gov`,
  `x/distribution`;
- economy modules: `x/fees`, future `x/token`;
- application modules: `x/tokenfactory`, `x/dex`, `x/identity`;
- execution modules: `x/execution`, `x/vm`, `x/messaging`, `x/queue`,
  `x/events`, `x/actors`, `x/scheduler`, `x/storage`;
- anti-spam and query modules: `x/reputation`, `x/indexer`;
- scaling R&D: `x/sharding/sim`, load/routing model, compute shard simulator;
- AVM executable specification and async contract standards.

## Scenario Generator

The scenario generator creates random and adversarial scenario families.

Required scenario types:

- random transaction sequences;
- malformed single transactions;
- staking lifecycle sequences;
- validator lifecycle sequences;
- DEX pool, liquidity, and swap sequences;
- fee and spam bursts;
- governance proposal/vote sequences;
- tokenfactory create, mint, burn, and admin sequences;
- domain registration, auction, renewal, resolver, and reverse lookup flows;
- async contract messages;
- queue, bounce, refund, and delayed execution flows;
- load/routing/shard activation scenarios.

Required fuzz strategies:

- stateless fuzzing: single malformed transaction or message;
- stateful fuzzing: multi-block transaction chains;
- adversarial fuzzing: known attack patterns;
- economic fuzzing: mint, burn, fee, supply, staking, DEX, and reward paths;
- determinism fuzzing: repeated execution under the same inputs;
- boundary fuzzing: max gas, max memo, max tx size, max queue, max state size.

Implementation tasks:

- [ ] Define canonical scenario schema.
- [ ] Add deterministic seed handling.
- [ ] Add module target selection.
- [ ] Add scenario minimization hooks for exploit reproduction.
- [ ] Add scenario replay by seed and step list.
- [ ] Add `.work/aexs/scenarios/` as default runtime output path.

## Transaction Mutator And Attack Generator

The mutator transforms valid scenarios into hostile inputs.

Required mutations:

- invalid signatures;
- replayed signed transactions;
- nonce and sequence manipulation;
- corrupted fee fields;
- non-`naet` fees;
- missing fees;
- extreme gas values;
- malformed addresses;
- zero address inputs;
- malformed memo fields;
- malformed routing hints;
- invalid domain resolution;
- expired domain actions;
- fake cross-zone messages;
- queue depth abuse;
- oversized VM payloads;
- invalid AVM entrypoint inputs;
- partial rollback attempts;
- malformed genesis fragments for simulator startup tests.

Required adversarial profiles:

- spam attacker;
- malicious validator;
- MEV-style bot;
- governance attacker;
- routing exploiter;
- shard overload attacker;
- tokenfactory authority attacker;
- DEX liquidity drain attacker;
- identity/domain squatter;
- AVM crash-input attacker.

Implementation tasks:

- [ ] Add attacker profile catalog.
- [ ] Add mutation catalog.
- [ ] Add per-module mutation allowlist.
- [ ] Add seed-stable random mutation order.
- [ ] Add replay mutation for already accepted tx bytes.
- [ ] Add mutation metadata for reporting.

## Chain Simulator

The simulator runs Aetra execution locally with reproducible state.

Required modes:

- single-validator fast fuzz mode;
- multi-validator consensus mode;
- stateful multi-block mode;
- stress mode for DoS and load spikes;
- chaos mode with injected failures;
- deterministic replay mode for exploit reproduction.

Required simulated surfaces:

- Aether Core consensus and validator set;
- base-chain ante and transaction validation;
- staking, slashing, distribution, and governance;
- native fees in `naet`;
- tokenfactory and DEX;
- identity and resolver;
- AVM executable specification;
- async messaging and queue;
- load system and deterministic `LOAD_SCORE`;
- routing engine and compute shard simulator.

Implementation tasks:

- [ ] Add in-memory Aetra app runner.
- [ ] Add randomized but valid genesis generation.
- [ ] Add randomized validator set generation.
- [ ] Add block replay support.
- [ ] Add per-block state snapshot capture.
- [ ] Add multi-node deterministic replay comparison.
- [ ] Add localnet-backed stress mode after in-memory mode is stable.

## Invariant Engine

Every simulated block must run mandatory invariant checks. Invariant failure is
an audit failure even if no user-visible exploit is found yet.

### Economic Invariants

- [ ] Total supply consistency:
  `sum(account balances) + module balances + burned + staked = total_supply`.
- [ ] No negative balances.
- [ ] No unauthorized mint.
- [ ] No unauthorized burn.
- [ ] Fee distribution totals match collected fees.
- [ ] Non-`naet` fees are rejected.
- [ ] Mint, burn, slash, reward, and fee accounting are export/import stable.
- [ ] Validator/delegator rewards are deterministic.

### Consensus Invariants

- [ ] Same block inputs produce the same app hash.
- [ ] Same signed transaction cannot execute twice.
- [ ] Invalid signer cannot mutate state.
- [ ] Validator set matches staking state.
- [ ] Slashing evidence is deterministic and cryptographically verifiable.
- [ ] No state transition succeeds after malformed ante validation.

### Routing And Load Invariants

- [ ] Same transaction and state produce the same zone decision.
- [ ] Same transaction and state produce the same shard decision.
- [ ] No routing loops.
- [ ] No shard starvation.
- [ ] `LOAD_SCORE` is always in `[0,1]`.
- [ ] EMA smoothing is deterministic.
- [ ] `LOAD_SCORE` does not jump beyond `MAX_DELTA`.
- [ ] Load values are identical across simulated nodes.

### DEX Invariants

- [ ] Pool reserves match module account balances.
- [ ] LP supply matches pool shares.
- [ ] No negative liquidity.
- [ ] No fake LP tokens.
- [ ] Swap output is non-negative and slippage-bounded.
- [ ] Constant-product constraints hold after fee-adjusted swaps.
- [ ] Failed bank movement does not mutate pool state.

### Identity And Resolver Invariants

- [ ] Domain names are unique.
- [ ] Active domains cannot be re-auctioned.
- [ ] Expired domains require auction or explicit renewal path.
- [ ] Resolver cannot point to malformed or zero address.
- [ ] Resolver-based payment fails before funds move if unresolved.
- [ ] Reverse lookup is consistent with owner-approved mapping.
- [ ] Domain registry owner and NFT representation owner do not diverge.

### Execution, AVM, And Queue Invariants

- [ ] AVM malformed input does not panic.
- [ ] AVM gas is bounded and deterministic.
- [ ] Infinite loops are rejected by gas or instruction limits.
- [ ] Contract state updates are deterministic.
- [ ] Queue ordering is deterministic.
- [ ] Cross-zone message replay is rejected.
- [ ] Bounce/refund cannot double-spend.
- [ ] Message loops are bounded by depth and per-block limits.
- [ ] Export/import preserves queue state exactly.

Implementation tasks:

- [ ] Define invariant interface.
- [ ] Add invariant registry by module.
- [ ] Add pre-block and post-block invariant phases.
- [ ] Add state snapshot comparator.
- [ ] Add invariant failure minimizer.
- [ ] Add invariant coverage tracker.

## Attack Simulation Engine

Every fuzz run must attempt red-team attacks, not only random mutations.

Required attack families:

- consensus attacks:
  - double-sign simulation;
  - slashing bypass attempt;
  - validator set drift attempt;
  - deterministic replay divergence attempt.
- economic attacks:
  - inflation manipulation;
  - fee bypass;
  - unauthorized mint/burn;
  - staking reward farming loop;
  - supply accounting drift.
- DEX attacks:
  - pool drain;
  - fake LP;
  - reserve mismatch;
  - rounding exploit;
  - failed bank movement rollback exploit.
- load/routing attacks:
  - spam bursts to inflate `LOAD_SCORE`;
  - shard overload targeting;
  - routing poisoning;
  - priority manipulation;
  - low-reputation bypass.
- identity attacks:
  - duplicate domain;
  - expired domain reuse without auction;
  - resolver hijack;
  - reverse lookup spoof;
  - unresolved-domain payment bypass.
- execution attacks:
  - AVM crash input;
  - oversized payload;
  - infinite loop attempt;
  - state corruption attempt;
  - invalid bounce/refund replay.

Implementation tasks:

- [ ] Add red-team attack suite.
- [ ] Add attacker role selector per run.
- [ ] Add severity mapping by invariant family.
- [ ] Add reproduction output for every failing scenario.
- [ ] Add exploit path graph generation.

## Coverage Matrix

Audit readiness fails if any module attack surface has less than `95%` planned
fuzz/invariant coverage. Production-safe status requires `100%` mandatory
invariant pass rate and no exploitable red-team path.

Required coverage targets:

| Target | Required coverage |
| --- | --- |
| `x/auth` | signatures, replay, nonce, signer extraction, malformed tx |
| `x/bank` | transfers, balances, module accounts, overflow, zero address |
| `x/staking` | delegation, unbonding, redelegation, rewards, validator set |
| `x/slashing` | downtime, equivocation evidence, jail/tombstone, slash accounting |
| `x/fees` | fee denom, fee amount, fee split, bypass attempts |
| `x/tokenfactory` | create, mint, burn, admin transfer, unauthorized authority |
| `x/dex` | pools, swaps, liquidity, LP supply, reserve accounting |
| `x/identity` | domain lifecycle, auction, expiry, renewal, resolver |
| `x/reputation` | score bounds, decay, rate limits, priority influence |
| `x/execution` | pipeline order, dispatch, route output, deterministic trace |
| `x/vm` and AVM | entrypoints, gas, code limits, malformed bytecode |
| `x/messaging` | async calls, replay, cross-zone message proofs |
| `x/queue` | ordering, delayed execution, bounce, refund, depth limits |
| `x/sharding/sim` | shard assignment, overload, replay, state consistency |
| Load/routing | `LOAD_SCORE`, EMA, `MAX_DELTA`, deterministic zone/shard route |

Implementation tasks:

- [ ] Add coverage schema.
- [ ] Add module-to-scenario mapping.
- [ ] Add module-to-invariant mapping.
- [ ] Add report-time coverage percentage.
- [ ] Add audit failure when any required target is below threshold.

### Runnable Preflight

The pre-campaign structural audit is implemented by:

```powershell
.\scripts\security\aexs-audit.ps1 -OutputDir .work\aexs
```

The runner validates `TO_AUDIT.md`, this pipeline document, the mandatory
module list, atomic task counts, mandatory coverage matrix rows, per-task
defensive/adversarial records, reproduction seeds, campaign setup fields,
runtime/simulator modes, stop conditions, scenario generator coverage, and
transaction mutator coverage, and evidence links. It writes deterministic
campaign output under `.work\aexs\`:

- `summary.json`;
- `campaign-setup.json`;
- `coverage-matrix.json`;
- `atomic-tasks.json`;
- `atomic-tasks.md`;
- `invariant-checklist.json`;
- `invariant-checklist.md`;
- `exploit-catalog.json`;
- `exploit-catalog.md`;
- `scenario-generator.json`;
- `scenario-generator.md`;
- `transaction-mutator.json`;
- `transaction-mutator.md`;
- `AUDIT_RESULT.md`;
- `TO_AUDIT.md`.

The runner deliberately returns `NOT_SAFE_PRE_CAMPAIGN` until a real fuzzing
and invariant campaign records executed results. `-EnforceSafe` must fail until
mandatory invariants have a `100%` pass rate, planned fuzz/invariant coverage
is at least `95%`, and no untriaged Critical or High exploit remains.

Every generated atomic task record contains:

- module;
- task id;
- function or flow covered;
- state transition covered;
- attack surface covered;
- invariant tested;
- defensive analysis result;
- adversarial simulation result;
- pass/fail result;
- reproduction seed or exact manual reproduction steps.

Every generated mandatory invariant checklist record contains the invariant id,
invariant family, scope, state transition covered, attack surface covered,
defensive analysis result, adversarial simulation result, pass/fail result, and
reproduction seed or exact manual reproduction steps. The checklist is generated
from `TO_AUDIT.md` and explicitly includes the economic invariants for global
supply consistency, non-negative balances, authorized mint/burn, `naet`-only
protocol fees, fee distribution accounting, deterministic treasury/burn/reward
accounting, staking reward-loop resistance, and export/import supply stability.
It also includes the consensus and state invariants for same-input AppHash
determinism, signed transaction replay rejection, invalid signer no-mutation,
malformed transaction no-mutation, validator-set/staking consistency, objective
slashing evidence, malformed genesis rejection, and upgrade/migration root
preservation. DEX checklist records cover reserve/module-balance
reconciliation, LP supply/share reconciliation, non-negative liquidity, LP
denom authenticity, non-negative swap output, slippage enforcement,
fee-adjusted constant-product preservation, and atomic rollback when bank
movement fails. Load, routing, and sharding checklist records cover
`LOAD_SCORE` bounds, deterministic EMA smoothing, `MAX_DELTA` spike caps,
same-input zone and shard decisions, routing-loop prevention, shard starvation
prevention, hot-zone monopolization resistance, and deterministic priority
ordering across nodes. Identity and resolver checklist records cover canonical
domain uniqueness, active-domain re-auction rejection, expired-domain renewal or
auction lifecycle enforcement, malformed and zero-address resolver rejection,
resolver-payment rollback before funds move, owner-approved reverse lookup
consistency, registry/NFT owner reconciliation, and parent-policy enforcement
for subdomain ownership and resolver delegation.
Execution, AVM, and queue checklist records cover malformed AVM input panic
safety, bounded deterministic gas, infinite-loop termination, deterministic
contract writes, canonical queue ordering, cross-zone replay rejection,
bounce/refund double-spend prevention, message-loop depth and per-block bounds,
and exact queue export/import preservation.

Every generated consensus and Aether Core exploit catalog record contains the
exploit id, category, exploit path, deterministic seed, step list, expected
state, actual state, affected modules, severity, fix recommendation, and
execution status. The first catalog section covers double-sign fork creation,
equivocation across heights and rounds, long-range history rewrite, stake
grinding, validator cartel concentration, delegation manipulation,
self-delegation inflation, fake validator liveness, validator eclipse, block
withholding, fork-choice manipulation, finality-delay manipulation, and
Byzantine-majority simulator scenarios.
The slashing bypass catalog section covers delayed evidence submission,
malformed equivocation proof acceptance, slashing race conditions,
redelegation-based partial slash evasion, unbonding-window slash evasion, jail
escape through upgrade timing, and invalid evidence replay. The transaction,
auth, and bank catalog section covers signature replay, wrong-chain-id
cross-context replay, invalid nonce bypass, transaction malleability, fee
underpayment, fee inflation manipulation, low-fee spam griefing, multi-send
partial failure, race-condition double spend, rollback during replayed state
transitions, and zero-address transfer or signer paths.
The token and economy catalog section covers tokenfactory mint authority
takeover, unauthorized burn bypass, governance-timed inflation manipulation,
fee routing manipulation, treasury drain proposals, staking reward inflation,
staking reward farming loops, edge-case mint supply manipulation, native denom
spoofing, and display/base decimal mismatch. The DEX catalog section covers
constant-product invariant breaks, liquidity drain swap sequences, pool
initialization manipulation, LP token inflation, liquidity removal races,
zero-liquidity edge cases, reserve/module balance desynchronization, failed
bank movement partial updates, slippage bypass, and rounding exploits.
The load system catalog section covers `LOAD_SCORE` spam manipulation,
artificial mempool inflation, block saturation, execution-delay amplification,
EMA slow-poison attacks, load spike oscillation, shard overload targeting,
priority fee gaming, and adaptive fee destabilization. The routing engine
catalog section covers routing bias exploitation, zone congestion targeting,
compute shard starvation, hot-zone monopolization, deterministic route
prediction abuse, cross-zone routing loops, routing desync between nodes,
transaction misclassification, and fee-based routing gaming.
The execution zone and AVM catalog section covers state divergence between
zones, cross-zone replay, AVM determinism violations, contract execution
desync, parallel execution race conditions, state corruption, partial rollback,
nondeterministic opcode or host behavior, gas exhaustion denial-of-service,
infinite loop griefing, storage collisions, contract upgrade hijack,
uninitialized storage, stack overflow, and sandbox escape. The compute shard
catalog section covers shard partition imbalance, shard starvation, shard
overflow collapse, cross-shard inconsistency, load spoofing for shard
activation, shard duplication, state split inconsistency, parallel execution
collision, scheduling manipulation, and queue flooding.
The Aether Mesh and cross-zone catalog section covers cross-zone message
replay, message delay manipulation, message ordering attacks, asset duplication
across zones, double spends across zones, proof forgery, relay censorship
simulation, message starvation, finality mismatch, and stale receipt replay.
The identity and `.aet` catalog section covers resolver overwrite hijack,
expired domain takeover without auction, auction manipulation, resolver
spoofing, subdomain collision, reverse lookup poisoning, domain binding races,
index-layer cache poisoning, fake domain resolution injection, and
multi-resolver inconsistency.
The governance catalog section covers voting-power capture, proposal spam,
emergency parameter abuse, upgrade hijack, delayed execution exploitation,
governance replay, proposal front-running, staking-loop voting power
manipulation, and parameter griefing. The genesis, upgrade, and state catalog
section covers malformed genesis injection, state export tampering, upgrade
rollback, partial migration corruption, module initialization bypass, hidden
privileged account injection, `InitGenesis` validation bypass, version mismatch,
state root collision, and snapshot poisoning. The mempool and network catalog
section covers mempool flooding, transaction prioritization gaming, gossip
poisoning, node eclipse simulation, P2P partition simulation, block propagation
delay, transaction reordering, network latency exploitation, bandwidth
exhaustion, and peer targeting. The combined full-stack catalog section covers
coordinated spam plus routing, load plus governance, DEX plus mempool plus
routing, validator collusion plus slashing delay, cross-zone value extraction,
identity plus routing hijack, shard overload plus fee manipulation, consensus
plus mempool denial-of-service, economic plus staking starvation, and
full-stack destabilization.

The base-chain `x/auth`, `x/bank`, `x/staking`, `x/slashing`, `x/gov`,
`x/distribution`, `x/fees`, `x/tokenfactory`, `x/dex`, `x/identity`,
`x/reputation`, `x/execution`, `x/vm` / AVM, `x/messaging`, `x/queue`, and
`x/events`, `x/actors`, `x/scheduler`, `x/storage`, `x/memo`, `x/indexer`, and
`x/sharding/sim` atomic tasks are expanded into task-specific records for
signature validation, replay prevention, sequence integrity, fee/priority abuse,
bank sends, multi-send atomicity, supply consistency, zero-address handling,
native-denom spoofing, staking lifecycle, validator-set consistency, unbonding
risk, slashing evidence, tombstone/jail state, slash accounting, proposal
replay, upgrade hijack, hard parameter bounds, reward withdrawal, commission
accounting, rounding remainders, community-pool leakage, `naet` fee collection,
non-FeeTx bypass, fee split accounting, tokenfactory authority, burn-from
mismatch, exact supply deltas, factory asset fee spoofing, AMM reserve
accounting, LP supply consistency, constant-product swap safety, domain
ownership, resolver integrity, reverse lookup authorization, auction escrow,
refund safety, reputation farming, sybil bypass, priority manipulation,
deterministic reputation replay, execution dispatch, rollback safety, stable
receipts, deterministic traces, routing constraint enforcement, AVM gas bounds,
sandbox isolation, deterministic host behavior, rejected VM no-commit semantics,
async message proofs, receipt replay markers, canonical message ordering, bounce
handling, refund double-spend prevention, queue sequence counters, queue depth
bounds, per-block processing limits, deterministic event emission, event
spoofing prevention, event receipt linkage, event-as-authority rejection, actor
isolation, mailbox bounds, monotonic logical time, actor cost enforcement,
deterministic scheduler plans, read/write conflict handling, priority
tie-breaks, starvation prevention, fee/reputation cap enforcement,
deterministic storage roots, snapshot/export integrity, bounded storage
pagination, storage rent/deposit enforcement, UTF-8 memo validation, memo
immutability, memo index safety, memo byte-fee enforcement, non-authoritative
index output, index rebuildability, stale-cache rejection, deterministic
`LOAD_SCORE`, route/shard assignment determinism, load poisoning rejection, shard
starvation prevention, and deterministic routing economic bounds. These records
must not collapse back to generic module-level attack descriptions.

Every campaign setup record contains the deterministic campaign id, git commit,
branch, dirty status, Go version, OS, test command set, fuzz seed list, target
modules, enabled runtime modes, enabled simulator modes, and stop conditions.

Every scenario generator record contains the scenario family id, flow covered,
state transition covered, attack surface covered, invariant targets, execution
status, and explicit requirements to preserve deterministic seed and exact step
list for replay.

Every transaction mutator record contains the mutator id, mutation type, target
modules, flow covered, state transition covered, attack surface covered,
invariant targets, expected rejection path, execution status, and explicit
requirements to preserve mutation metadata and deterministic replay seed.

## Chaos Mode

Chaos mode injects reproducible failures to detect nondeterministic consensus
bugs and recovery gaps.

Required injections:

- validator delay;
- validator crash;
- partial mempool loss;
- mempool corruption;
- delayed block propagation in simulator;
- inconsistent local validator state;
- partial execution-zone failure;
- shard crash simulation;
- queue backlog spike;
- state snapshot interruption;
- export/import replay under load.

Rules:

- chaos seeds must be reproducible;
- injected faults must be recorded in reports;
- fault timing must be deterministic in replay mode;
- chaos mode must never hide invariant failures.

Implementation tasks:

- [ ] Add chaos event schema.
- [ ] Add deterministic fault scheduler.
- [ ] Add replay support for chaos runs.
- [ ] Add chaos-specific report fields.

## Exploit Reporter

If any invariant breaks, the pipeline must generate a report that an engineer
can replay and debug.

Required exploit fields:

- root cause summary;
- exploit path;
- minimal reproduction steps;
- seed;
- block height;
- transaction sequence;
- mutation sequence;
- affected modules;
- expected state;
- actual state;
- invariant that failed;
- severity score;
- fix recommendation.

Severity scale:

- Critical: unauthorized mint/burn, double spend, consensus divergence, fee
  bypass at scale, slashing bypass, deterministic app hash divergence.
- High: DEX drain, resolver hijack, routing manipulation, queue double-refund,
  validator-set inconsistency.
- Medium: bounded DoS, incorrect report/index state, non-critical accounting
  drift caught before finalization.
- Low: non-consensus UX/reporting issue with no fund-safety impact.

Implementation tasks:

- [ ] Add report data model.
- [ ] Add markdown report renderer.
- [ ] Add minimal reproduction reducer.
- [ ] Add state diff renderer.
- [ ] Add severity classifier.

## Output Files

Runtime outputs should live under `.work/aexs/` unless a future implementation
explicitly chooses a tracked report path.

### `TO_AUDIT.md`

Generated before a fuzz campaign.

Required content:

- full fuzz strategy;
- module breakdown;
- attack surface map;
- invariant list;
- attack scenarios;
- coverage plan;
- selected seeds;
- simulator mode;
- expected run duration.

### `AUDIT_RESULT.md`

Updated during and after a fuzz campaign.

Per-run content:

- scenario executed;
- mutation applied;
- attacker profile;
- result;
- invariant pass/fail;
- exploit found, if any;
- reproduction seed;
- affected modules.

Final content:

- security score from `0` to `100`;
- top vulnerabilities;
- system stability rating;
- economic risk rating;
- coverage matrix;
- mandatory invariant pass rate;
- red-team exploit summary;
- production-safe decision.

## Safe-System Criteria

Aetra can be marked safe for the tested scope only if:

- mandatory invariant pass rate is `100%`;
- no exploitable red-team path remains;
- no deterministic execution violation occurs;
- no economic loop is exploitable;
- no fee bypass is possible;
- no unauthorized mint/burn path is possible;
- no double spend is possible;
- no queue bounce/refund double-spend is possible;
- no domain/resolver hijack path is possible;
- every required coverage target is at or above `95%`;
- all generated exploit reports are triaged and either fixed or explicitly
  marked out of scope for the tested release.

## Implementation Checklist

1. Build scenario schema and seed handling.
2. Build stateless mutator for tx bytes, signatures, fees, denoms, memo, and
   addresses.
3. Build in-memory app runner and randomized genesis.
4. Add stateful multi-block fuzzing.
5. Add invariant registry and state snapshot comparator.
6. Add DEX, tokenfactory, fees, staking, and bank invariants.
7. Add identity, resolver, reputation, load, routing, and queue invariants.
8. Add AVM malformed input and gas-bound fuzzing.
9. Add red-team attack profiles.
10. Add chaos mode.
11. Add coverage matrix and audit threshold enforcement.
12. Add `TO_AUDIT.md` and `AUDIT_RESULT.md` renderers under `.work/aexs/`.
13. Add reproduction reducer and state diff output.
14. Wire the pipeline into security gates after the first stable implementation.

## One-Line Summary

This pipeline turns Aetra L1 into a continuously attacked simulation
environment where every module is treated as a hostile target under adversarial
fuzzing, invariant enforcement, deterministic replay, and exploit reporting.
