# Aetheris Roadmap V3: Economy, Execution OS, Identity, And Async VM

`ROADMAP_v3.md` is a local operator planning file and must not be committed or pushed. It is intentionally covered by `.gitignore`.

## Git Policy

- Use direct `git push` only when the operator explicitly asks for push.
- Do not create pull requests.
- Do not use GitHub compare/PR screens as the delivery path.
- Keep this file local. It is a working architecture plan, not a repository artifact.

## Product Definition

Aetheris is a production Layer 1 designed as:

> Execution OS + Financial Layer + Identity Layer + Async VM Blockchain

The chain must support:

- low-cost native payments in `AET` / `naet`;
- PoS staking and validator security;
- adaptive emission and burn;
- deterministic async contract execution;
- contract-owned assets;
- domain-based identity through `.aet`;
- reputation-aware anti-spam;
- smart contract runtime through AVM and gated CosmWasm;
- production-grade export/import, snapshots, state sync, audits, and observability.

## Non-Negotiable Invariants

- Native token: `AET`.
- Base denom: `naet`.
- `1 AET = 1,000,000,000 naet`.
- Native token supply is uncapped but bounded by adaptive inflation.
- Protocol fees are paid only in `naet`.
- User-created tokens cannot pay protocol fees.
- Orbitalis-era public address and denom formats are not public compatibility paths. Public L1 code, docs, scripts, and outputs should present the blockchain as Aetheris only; any legacy handling must stay isolated in explicit migration tooling.
- Raw addresses use `4:` plus 64 lowercase hex chars.
- Userfriendly addresses start with `AE`.
- Zero address is forbidden by default.
- Async execution must be deterministic across all validators.
- Contract standards must be testable without relying on manual operator judgment.

---

# Track 1: Production Economic Model

Goal: create a self-balancing token economy that keeps transactions cheap, avoids hyperinflation and deflationary collapse, and gives validators durable income.

## 1.1 Supply Model

Native supply variables:

- `S_t`: total circulating `AET` supply at time `t`.
- `M_t`: newly minted `AET` during period `t`.
- `B_t`: burned `AET` during period `t`.
- `S_{t+1} = S_t + M_t - B_t`.

Target policy:

- annual inflation range: `1.0%` to `5.0%`;
- default target inflation: `3.0%`;
- target staking participation: `67%`;
- lower inflation when staking participation is high;
- higher inflation when staking participation is low;
- burn grows with network activity, not with manual operator intervention.

Staking ratio:

```text
stake_ratio = bonded_naet / circulating_naet
```

Adaptive inflation formula:

```text
min_inflation = 0.01
max_inflation = 0.05
target_stake = 0.67
responsiveness = 0.08

raw_inflation =
  current_inflation + responsiveness * (target_stake - stake_ratio)

annual_inflation =
  clamp(raw_inflation, min_inflation, max_inflation)
```

Per-block mint:

```text
blocks_per_year = expected_blocks_per_year
mint_per_block = circulating_naet * annual_inflation / blocks_per_year
```

Mint distribution:

```text
validator_delegator_rewards = mint_per_block * 0.92
treasury_mint             = mint_per_block * 0.05
insurance_reserve_mint    = mint_per_block * 0.03
```

Requirements:

- `mint_per_block` must be deterministic integer math.
- Rounding remainder goes to treasury or is carried in a deterministic accumulator.
- Inflation params are governance-controlled but bounded by protocol min/max.
- Emergency governance cannot set inflation above hard-coded max without a software upgrade.

Tests:

- inflation decreases when stake ratio is above target;
- inflation increases when stake ratio is below target;
- inflation never exceeds `5%`;
- inflation never falls below `1%`;
- per-block mint is deterministic;
- supply changes only through mint and burn paths;
- export/import preserves mint accumulators.

## 1.2 Burn Model

Burn sources:

- base fee burn;
- domain auction finalization burn;
- token creation fee burn;
- contract deployment fee burn;
- storage rent burn if storage rent is enabled;
- slashing burn if policy chooses burn instead of redistribution.

Fee burn formula:

```text
base_burn = base_fee_paid * burn_ratio
```

Recommended starting params:

```text
normal_burn_ratio = 0.30
congested_burn_ratio = 0.40
treasury_ratio = 0.10
validator_fee_ratio = 1 - burn_ratio - treasury_ratio
```

Activity-linked burn:

```text
activity_burn = sum(base_fee_burn + deploy_burn + domain_burn + storage_burn)
```

Long-term balance target:

```text
net_supply_change = annual_mint - annual_burn
```

The system should aim for:

- positive but bounded net issuance when network is young;
- near-neutral issuance when network activity grows;
- never forcing deep deflation if validator income becomes unsafe.

Deflation guard:

```text
if annual_burn > annual_mint * 1.25:
  reduce burn_ratio gradually toward min_burn_ratio
```

Recommended bounds:

```text
min_burn_ratio = 0.10
max_burn_ratio = 0.50
```

Tests:

- burn reduces total supply;
- burn accounting cannot underflow;
- burn ratio stays within bounds;
- burn cannot drain fee collector incorrectly;
- high activity cannot force uncontrolled deflation.

## 1.3 Fee Model

Goal: fees stay low for normal users and increase only softly during overload.

Fee components:

```text
tx_fee =
  base_fee
  + gas_used * gas_price(load)
  + byte_fee * tx_size_bytes
  + memo_fee
  + storage_fee
  + priority_fee_optional
```

Network load:

```text
load = used_block_gas / target_block_gas
```

Gas price adjustment:

```text
target_load = 0.70
max_load_multiplier = 4.0
min_gas_price = protocol_min_gas_price
max_gas_price = protocol_max_gas_price
adjustment_speed = 0.125

next_base_gas_price =
  current_base_gas_price *
  (1 + adjustment_speed * (load - target_load))

base_gas_price =
  clamp(next_base_gas_price, min_gas_price, max_gas_price)
```

Soft congestion multiplier:

```text
if load <= target_load:
  congestion_multiplier = 1.0
else:
  congestion_multiplier = min(
    max_load_multiplier,
    1 + ((load - target_load) / (1 - target_load))^2
  )
```

Final gas price:

```text
gas_price(load) = base_gas_price * congestion_multiplier
```

Hard rule:

- high fees must not be the primary anti-spam mechanism.

Anti-spam should use:

- per-account rate limits;
- reputation-aware queue limits;
- proof-of-stake cost for high-throughput accounts;
- memo/storage byte limits;
- contract deploy limits;
- bounded async queue;
- account age throttles;
- validator mempool filters;
- deterministic execution quotas.

Fee distribution:

```text
validator_fee = tx_fee * validator_fee_ratio
burn_fee      = tx_fee * burn_ratio
treasury_fee  = tx_fee * treasury_ratio
```

Recommended starting values:

```text
validator_fee_ratio = 0.60
burn_ratio          = 0.30
treasury_ratio      = 0.10
```

Congestion adjustment:

```text
if load > 0.90:
  burn_ratio = min(max_burn_ratio, normal_burn_ratio + 0.10)
```

Tests:

- non-`naet` fees rejected;
- malformed fees rejected;
- missing fees rejected;
- fee cap enforced;
- load multiplier deterministic;
- fee distribution sums exactly to paid fee;
- rounding deterministic;
- fees remain bounded under stress profile.

## 1.4 Validator Income Model

Validator income:

```text
validator_income =
  validator_share_of_mint_rewards
  + validator_share_of_fee_rewards
  + validator_commission
```

Delegator income:

```text
delegator_income =
  delegator_share_of_mint_rewards
  + delegator_share_of_fee_rewards
  - validator_commission
```

Validator reward allocation:

```text
validator_reward_weight =
  validator_power / total_validator_power
```

Commission bounds:

```text
min_commission = 0.01
max_commission = 0.20
max_daily_commission_change = 0.01
```

Stability requirements:

- validator income does not depend only on tx fees;
- mint rewards provide baseline security budget;
- fees provide activity-linked upside;
- treasury/insurance reserve can fund public goods and emergency recovery;
- slashing must be large enough to make attacks economically irrational.

Security threshold:

```text
attack_cost >= bonded_value_controlled_by_attacker
```

Consensus safety target:

- less than `1/3` malicious voting power cannot halt safety;
- `1/3+` malicious voting power is a serious threat and should trigger monitoring;
- `2/3+` malicious voting power can finalize malicious state and must be economically infeasible.

Tasks:

- define min self-delegation;
- define validator commission bounds;
- define unbonding period;
- define redelegation limits;
- add downtime slashing;
- add double-sign slashing;
- add validator set concentration metrics;
- add top-N validator concentration alerts.

## 1.5 System Balance Controller

The economy should auto-balance using measured on-chain state:

Inputs:

- staking ratio;
- block gas load;
- tx count;
- failed tx rate;
- async queue depth;
- burn amount;
- mint amount;
- validator participation;
- validator concentration;
- treasury runway.

Outputs:

- adaptive inflation;
- dynamic gas price;
- burn ratio within bounds;
- queue limits;
- memo/storage cost;
- deploy cost;
- rate limits.

Equilibrium behavior:

```text
low staking -> inflation rises -> staking becomes more attractive
high staking -> inflation falls -> dilution decreases
high activity -> burn rises -> supply expansion slows
low activity -> burn falls -> mint rewards maintain validator security
high congestion -> soft fee multiplier + queue/rate limits
low congestion -> low fees
```

Safety bounds:

- hard inflation max prevents hyperinflation;
- hard burn bounds prevent deflationary collapse;
- fee caps prevent unusable transaction costs;
- anti-spam limits protect block capacity without extreme fees;
- treasury and insurance reserve protect long-term operations.

---

# Track 2: Module Architecture

## 2.1 CORE

### x/auth

Responsibilities:

- account model;
- signature verification;
- sequence/nonce replay protection;
- transaction validation;
- signer extraction;
- address validation;
- zero-address rejection.

Tasks:

- support account address `4:` and `AE...` formats through central codec;
- reject old public formats outside migration tooling;
- enforce sequence increment only after accepted tx;
- add wrong-chain and replay tests;
- expose account metadata for indexer.

### x/bank

Responsibilities:

- native `AET` transfers;
- `naet` balances;
- module account accounting;
- mint/burn permissions;
- multi-asset readiness for non-native tokens.

Tasks:

- keep `naet` as native denom;
- reject non-native denoms for protocol fees;
- support resolver-based transfer target lookup;
- include optional memo metadata;
- emit deterministic transfer events.

### x/staking

Responsibilities:

- validator set;
- delegation;
- unbonding;
- redelegation;
- power updates;
- staking reward eligibility.

Tasks:

- bond denom fixed to `naet`;
- parameter bounds;
- validator lifecycle tests;
- CometBFT validator update tests;
- restart and snapshot tests.

### x/slashing

Responsibilities:

- downtime penalties;
- double-sign penalties;
- jailing and tombstone behavior;
- security enforcement.

Tasks:

- define slash fractions;
- define missed block window;
- define recovery path for downtime;
- ensure slashed supply accounting is deterministic;
- decide burn vs redistribution policy for slashed stake.

### x/gov

Responsibilities:

- parameter updates;
- upgrade voting;
- treasury control;
- emergency controls.

Tasks:

- governance can change soft params only within hard bounds;
- hard economic bounds require software upgrade;
- add emergency halt/restart docs;
- add governance authority tests for fees, token params, domain params, reputation params, and VM params.

### x/distribution

Responsibilities:

- reward distribution;
- validator commission;
- delegator reward accounting;
- community pool/treasury integration.

Tasks:

- integrate mint rewards and fee rewards;
- deterministic rounding;
- reward withdrawal tests;
- validator commission bounds.

## 2.2 ECONOMY

### x/fee

Responsibilities:

- dynamic fee calculation;
- `naet`-only enforcement;
- fee distribution;
- congestion load tracking;
- fee params;
- anti-spam coordination with reputation and scheduler.

State:

- base gas price;
- target block gas;
- recent load EMA;
- fee distribution ratios;
- min/max fee bounds;
- memo byte price;
- storage byte price;
- deploy fee;

Messages:

- `MsgUpdateFeeParams` via governance authority only.

Queries:

- current fee params;
- current network load;
- estimated fee for tx;
- fee distribution stats.

### x/token

Responsibilities:

- native emission;
- burn accounting;
- supply controller;
- user token standard coordination.

State:

- current inflation;
- target stake ratio;
- mint accumulator;
- burn accumulator;
- annualized mint/burn stats;
- supply controller params.

Messages:

- governance update within hard bounds;
- internal mint per block;
- internal burn accounting.

Queries:

- supply;
- inflation;
- burn stats;
- staking ratio;
- net issuance.

## 2.3 INTEROP

### x/ibc

Responsibilities:

- cross-chain communication;
- packet lifecycle;
- asset transfer logic.

Rules:

- IBC assets cannot pay Aetheris protocol fees;
- bridged assets must be clearly namespaced;
- bridge/resolver/domain integrations must reject spoofed native `AET`.

### x/bridge

Responsibilities:

- external chain bridging;
- wrapped asset registry;
- proof/validator/oracle security model;
- bridge risk limits.

Tasks:

- cap bridge mint exposure;
- isolate bridge assets from native `naet`;
- add emergency pause;
- add proof verification tests;
- add bridge insolvency and replay tests.

## 2.4 APPLICATION

### x/dex

Responsibilities:

- AMM or hybrid exchange;
- liquidity pools;
- swap execution;
- LP accounting;
- fee integration.

Tasks:

- keep native DEX module until contract DEX is audited;
- reject zero-address participants;
- reject native denom spoofing;
- add reserve/module balance invariants;
- support future contract pool migration.

### x/identity

Responsibilities:

- `.aet` domain registry;
- domain auctions;
- domain ownership;
- domain NFT representation;
- resolver ownership permissions;
- expiry and renewal.

Domain is hybrid:

- NFT representation proves ownership in wallets and UIs;
- registry record is the source of truth.

Domain schema:

```text
DomainRecord {
  name: string
  tld: ".aet"
  owner: address
  resolver: address | null
  expiry: timestamp/block
  nft_item_id: string
  status: active | expired | auction
  created_at
  updated_at
}
```

Allowed names:

- lowercase `a-z`;
- digits `0-9`;
- selected safe symbols only if protocol explicitly enables them;
- no whitespace;
- no invisible Unicode;
- no mixed-script spoofing;
- normalized before validation.

Recommended production rule:

- start with `a-z`, `0-9`, `-`, `_`;
- add other symbols only after UI/security review.

### x/workflow

Responsibilities:

- multi-step transaction orchestration;
- conditional execution;
- module interaction flows;
- atomic workflows where synchronous execution is required.

Examples:

- resolver-based payment;
- domain auction finalization;
- token mint and wallet deploy;
- NFT mint and metadata attach;
- contract deployment plus first message.

---

# Track 3: `.aet` Domains And Resolver

Goal: create a safe domain identity layer where names can resolve to wallets, smart contracts, multisigs, or structured multi-address records.

## 3.1 Domain Lifecycle

Lifecycle:

```text
available -> auction -> active -> expired -> available/auction
```

Rules:

- if domain does not exist or is expired, registration starts an auction;
- auction duration: `24h`;
- highest valid bid wins;
- winner receives registry ownership and NFT representation;
- ownership duration: `365 days`;
- renewal extends expiry;
- non-renewed domains return to auction pool.

Auction start price by length:

```text
length 1-3: reserved or governance-only
length 4:   premium_start_price
length 5-6: high_start_price
length 7-10: medium_start_price
length 11+: low_start_price
```

Example starting params:

```text
premium_start_price = 10_000 AET
high_start_price    = 1_000 AET
medium_start_price  = 100 AET
low_start_price     = 10 AET
min_bid_increment   = 5%
auction_duration    = 24h
anti_snipe_window   = 10m
anti_snipe_extend   = 10m
max_extensions      = 6
registration_period = 365d
```

Fee distribution for domain auctions:

```text
domain_bid_final =
  burn:      40%
  treasury:  40%
  rewards:   20%
```

Renewal fee:

```text
renewal_fee = start_price(name_length) * renewal_discount
renewal_discount = 0.10
```

Tests:

- expired domain can enter auction;
- active domain cannot be auctioned;
- highest bidder wins;
- bid below increment rejected;
- anti-snipe extension bounded;
- winner becomes registry owner;
- NFT representation minted/updated;
- expiry set to 365 days;
- renewal extends expiry;
- failed auction finalization cannot steal bids.

## 3.2 Resolver

Module: `x/resolver` or identity submodule.

Responsibilities:

- domain to address resolution;
- reverse address to domain resolution;
- multi-address records;
- resolver ownership checks;
- fast query layer;
- event emission.

Resolver record:

```text
ResolverRecord {
  domain: "alice.aet"
  owner: address
  primary: address | null
  records: map<string, address>
  metadata: optional bounded metadata
  updated_at
}
```

Record keys:

- `wallet`;
- `contract`;
- `multisig`;
- `nft`;
- `dex`;
- custom keys with bounded length.

Payment routing:

```text
send AET to alice.aet:
  record = resolver.lookup("alice.aet")
  if record.primary == null:
    fail "domain not resolved"
  send funds to record.primary
```

Rules:

- owner can update resolver;
- delegated resolver manager can update only if owner grants permission;
- resolver target must be a valid non-zero address;
- unresolved domain transfer fails before funds move;
- expired domain cannot be used for new payment routing unless grace policy says otherwise;
- resolver updates emit events.

Advanced features:

- subdomains: `dex.alice.aet`, `nft.alice.aet`, `bot.alice.aet`;
- leasing: owner leases resolver rights for a period;
- DAO ownership: domain owner can be multisig/DAO contract;
- revenue sharing: auction or renewal revenue can split by policy;
- reverse resolution: address chooses primary domain if owner proves control;
- multi-resolver records for wallets, contracts, and app-specific endpoints.

Tests:

- set resolver;
- change resolver;
- reject zero resolver;
- reject unauthorized resolver update;
- send to resolved domain;
- send to unresolved domain fails;
- expired domain resolution fails or follows explicit grace rule;
- subdomain resolution;
- reverse resolution;
- domain NFT owner and registry owner stay consistent.

---

# Track 4: Reputation Layer

Module: `x/reputation`.

Goal: filter spam, prioritize honest users, and limit abuse without making base fees expensive.

## 4.1 Reputation State

State:

```text
ReputationRecord {
  account: address
  score: uint8 // 0..100
  age_score
  staking_score
  tx_success_score
  volume_score
  domain_score
  contract_score
  spam_penalty
  failed_tx_penalty
  slash_penalty
  last_updated
}
```

Score levels:

```text
0-20   restricted
20-50  new
50-80  normal
80-95  trusted
95-100 elite
```

Deterministic score:

```text
score =
  age_score
  + staking_time_score
  + tx_success_rate_score
  + bounded_volume_score
  + domain_reputation_score
  + contract_reputation_score
  - spam_penalty
  - failed_tx_penalty
  - slash_events_penalty
```

Bounds:

```text
score = clamp(score, 0, 100)
```

Decay:

```text
if inactive:
  score -= inactivity_decay_rate * inactive_epochs
```

Rules:

- score is based only on deterministic on-chain events;
- no direct reputation purchase;
- reputation staking may exist only as a bonded signal with slashing/risk;
- new accounts have progressive limits;
- contracts also have reputation;
- domain ownership can add bounded reputation, but cannot dominate score.

## 4.2 Reputation Usage

Anti-spam:

- low score means lower tx rate limit;
- low score means lower async queue quota;
- low score means higher memo/storage byte cost;
- low score means stricter contract deploy limits.

Execution priority:

- high score can improve queue priority within deterministic bounds;
- priority cannot bypass fees, signatures, or validation;
- validators must compute identical priority ordering.

Access control:

- token creation may require score threshold or deposit;
- contract deployment may require score threshold or deposit;
- DEX pool creation may require score threshold or deposit;
- domain auction spam can be rate-limited by score.

Tests:

- score updates deterministic;
- score cannot exceed 100;
- penalties apply;
- decay applies;
- low-score tx rate limited;
- high-score user does not bypass required fees;
- reputation cannot be bought directly;
- contract reputation updates on failed/successful executions.

---

# Track 5: Memo / Note System

Goal: allow optional human-readable text on transactions without letting metadata affect consensus execution logic or bloat state.

## 5.1 Core Fields

Transaction metadata:

```text
TxMetadata {
  memo: string optional
  memo_hash: bytes optional
  memo_visible: bool
}
```

Supported surfaces:

- native bank transfer;
- resolver/domain payment;
- token transfer;
- NFT transfer;
- SBT proof/revoke;
- contract call;
- domain auction bid;
- domain renewal;
- DEX swap/liquidity action.

Requirements:

- optional;
- UTF-8 only;
- max length default: `200` characters;
- max length governance configurable within hard protocol bound;
- immutable after block inclusion;
- stored as transaction metadata, not execution input;
- cannot change state transition result.

Validation:

```text
if memo == "":
  ok
else:
  require valid_utf8(memo)
  require char_count(memo) <= max_memo_chars
  require byte_len(memo) <= max_memo_bytes
  require no prohibited control chars
```

Recommended params:

```text
default_max_memo_chars = 200
hard_max_memo_chars    = 500
default_max_memo_bytes = 1024
```

## 5.2 Memo Economics

Recommended cost:

```text
memo_fee =
  memo_base_fee
  + memo_byte_fee * memo_bytes
  * reputation_multiplier(sender)
  * congestion_multiplier(load)
```

Reputation multiplier:

```text
score >= 80: 0.75
score >= 50: 1.00
score >= 20: 1.50
score < 20:  3.00
```

Rules:

- memo fee paid only in `naet`;
- memo fee can be zero for empty memo;
- memo size contributes to tx byte cost;
- low reputation can be rate-limited or delayed;
- memo cannot become cheap spam storage.

## 5.3 Storage And Indexing

Store:

- tx hash;
- sender;
- receiver if known;
- asset type;
- related domain if any;
- memo or memo hash depending on storage policy;
- block height;
- timestamp.

Indexes:

- by tx hash;
- by sender;
- by receiver;
- by domain;
- by contract;
- by asset;
- by event type.

Privacy option:

- allow full memo stored on-chain for transparency; or
- store memo hash on-chain and full memo in indexer only.

Production default:

- store bounded memo on-chain only if below configured byte limit;
- indexer may provide search;
- consensus does not depend on search index.

Events:

```text
EventMemoAttached {
  tx_hash
  from
  to
  domain
  memo_hash
  memo
}
```

Tests:

- empty memo accepted;
- valid memo accepted;
- invalid UTF-8 rejected;
- oversized memo rejected;
- memo cannot affect execution result;
- memo indexed by tx hash;
- memo indexed by domain;
- low reputation memo cost higher;
- memo event deterministic.

---

# Track 6: Execution OS

## 6.1 x/execution

Responsibilities:

- transaction orchestration;
- execution pipeline;
- module dispatch;
- async entrypoint;
- deterministic ordering;
- event collection;
- error handling.

Pipeline:

```text
CheckTx:
  decode
  validate signatures
  validate fees
  validate memo
  stateless checks

DeliverTx/FinalizeBlock:
  ante
  execution context
  module dispatch
  async enqueue if needed
  event emit
  state write
```

Tasks:

- define `ExecutionEnvelope`;
- include optional memo metadata;
- integrate resolver lookup;
- integrate reputation limits;
- integrate fee estimator;
- route VM calls;
- expose deterministic execution trace for tests.

## 6.2 x/vm

Responsibilities:

- AVM runtime;
- gated CosmWasm runtime;
- gas metering;
- sandbox;
- contract state access;
- host functions.

AVM minimum contract:

- counter contract;
- deploy;
- external call;
- internal call;
- bounced call;
- query/getter.

Tasks:

- bytecode/module format;
- ABI;
- storage ABI;
- message ABI;
- deterministic host functions;
- gas schedule;
- fuzz tests;
- adversarial tests.

## 6.3 x/messaging

Responsibilities:

- async calls between contracts;
- internal message envelope;
- bounce/refund behavior;
- outgoing message validation.

Message:

```text
Message {
  id
  source
  destination
  value_naet
  opcode
  query_id
  body
  bounce
  deadline
  gas_limit
  created_lt
}
```

Tests:

- message enqueue;
- message delivery;
- deterministic order;
- bounce on failure;
- expired message handling;
- refund cannot double-spend.

## 6.4 x/queue

Responsibilities:

- delayed execution;
- scheduled tasks;
- retry and failure handling;
- queue limits;
- queue observability.

Ordering:

```text
priority_key =
  scheduled_height,
  reputation_class,
  tx_height,
  tx_index,
  message_index,
  source_logical_time,
  sequence
```

Rules:

- priority must be deterministic;
- low reputation cannot starve forever;
- max per-block processing limit;
- max per-account queued messages;
- max per-contract queued messages.

## 6.5 x/events

Responsibilities:

- event-driven system;
- protocol events;
- indexer events;
- contract events;
- memo events;
- domain events;
- reputation events.

Event types:

- `EventTransfer`;
- `EventMemoAttached`;
- `EventDomainAuctionStarted`;
- `EventDomainResolved`;
- `EventContractMessageQueued`;
- `EventContractMessageProcessed`;
- `EventReputationUpdated`;
- `EventFeeDistributed`.

## 6.6 x/actors

Responsibilities:

- each contract behaves as an actor;
- actor state isolation;
- actor mailbox;
- actor message processing;
- actor lifecycle.

Actor state:

```text
Actor {
  address
  code_hash
  state_root
  logical_time
  mailbox_stats
  status
}
```

Rules:

- one actor state transition per delivered message;
- actor cannot mutate another actor state directly;
- all cross-actor effects go through messages;
- exported state includes actor state and mailbox.

## 6.7 x/scheduler

Responsibilities:

- parallel execution planning;
- conflict detection;
- deterministic batching;
- safe concurrent state access.

Initial version:

- sequential deterministic execution.

Production version:

- optimistic parallel execution with deterministic conflict resolution;
- DAG scheduler;
- read/write set tracking;
- fallback to sequential on conflict.

## 6.8 x/storage

Responsibilities:

- KV state engine;
- versioning;
- snapshots;
- state sync;
- contract storage;
- bounded iteration.

Tasks:

- define contract namespace;
- define storage key format;
- define max state size;
- define storage rent/deposit;
- export/import exact state;
- snapshot/state-sync tests.

---

# Track 7: Additional Modules

## x/compute

Purpose:

- measure CPU/compute usage separately from simple tx gas;
- price expensive computation;
- protect validators from CPU abuse.

State:

- compute unit schedule;
- per-op cost;
- per-contract compute stats;
- per-block compute budget.

Tests:

- expensive contract charged more;
- compute cap enforced;
- compute accounting deterministic.

## x/permissions

Purpose:

- ACL system for contracts and modules;
- resolver delegates;
- domain managers;
- contract extension permissions;
- governance-controlled permissions.

Rules:

- all permissions have owner, scope, expiry, and revocation path;
- permission checks are deterministic;
- no hidden superuser outside governance/emergency policy.

## x/indexer

Purpose:

- fast query layer;
- state search;
- event search;
- memo search;
- domain lookup;
- token/NFT discovery.

Rule:

- indexer must never be required for consensus.

## x/market

Purpose:

- market for compute, storage, and execution priority;
- bounded, deterministic, and non-extractive.

Rules:

- cannot replace base `naet` fee;
- cannot let wealthy users fully starve normal users;
- must be capped by scheduler fairness.

## x/scheduler-v2

Purpose:

- DAG execution engine;
- parallel tx scheduling;
- async actor mailbox planning.

Requirements:

- deterministic read/write set;
- deterministic conflict resolution;
- replayable schedule;
- identical result across validators.

---

# Track 8: Contract Standards

## AW-5 Wallet

Goal:

- contract wallet standard with replay protection, extensions, relayer compatibility, and bounded multi-send.

State:

- `seqno`;
- `wallet_id`;
- public key;
- owner;
- signature enabled flag;
- extensions map;
- recovery policy.

Tests:

- replay rejected;
- wrong wallet id rejected;
- expired command rejected;
- extension takeover rejected;
- multi-send bounded;
- relayer pays `naet`.

## AFT-44 Fungible Token

Goal:

- token master and token wallet standard for user-created assets.

Rules:

- token master controls metadata, minting, admin, supply;
- each holder has token wallet;
- token wallet stores per-holder balance;
- user tokens cannot pay protocol fees;
- native `AET` is not an AFT token.

## ANFT-66 NFT

Goal:

- NFT collection and NFT item standard.

Rules:

- collection is source of truth;
- item stores collection, owner, metadata;
- transfer requires owner;
- metadata and royalties bounded.

## ASBT-67 Soulbound Item

Goal:

- non-transferable identity/credential item.

Rules:

- owner immutable;
- transfer rejected;
- authority can revoke if policy allows;
- proof of ownership supported.

---

# Track 9: Security Risks And Controls

## Infinite Supply Risks

Risks:

- excessive inflation;
- validator reward dilution;
- weak token confidence;
- governance abuse.

Controls:

- hard inflation cap;
- target staking controller;
- public mint/burn telemetry;
- governance bounds;
- export/import supply invariants.

## Deflation Risks

Risks:

- burn exceeds mint for too long;
- validator income falls;
- users hoard instead of transacting;
- tx fees become politically hard to lower.

Controls:

- burn ratio floor/ceiling;
- validator baseline mint rewards;
- fee caps;
- deflation guard.

## Spam Risks

Risks:

- low-cost tx floods;
- memo spam;
- async queue flooding;
- contract deploy spam;
- domain auction spam.

Controls:

- reputation rate limits;
- per-account queue caps;
- per-contract queue caps;
- memo byte fees;
- deploy deposits;
- domain auction bid deposits;
- bounded block processing;
- scheduler fairness.

## Staking Attacks

Risks:

- stake concentration;
- validator cartel;
- long-range attack;
- validator downtime;
- double-sign.

Controls:

- unbonding period;
- slashing;
- tombstone for severe equivocation if enabled;
- validator concentration alerts;
- commission bounds;
- delegation transparency;
- snapshot/state-sync safety.

## Economic Attacks

Risks:

- fee market manipulation;
- fake volume to trigger burn;
- reputation farming;
- domain squatting;
- token metadata spoofing;
- bridge asset spoofing.

Controls:

- fee multiplier smoothing;
- bounded burn response;
- deterministic reputation decay;
- auction pricing and renewal;
- native token metadata reservation;
- bridged asset namespace isolation.

---

# Track 10: Test And Production Gates

## Unit Tests

- fee formulas;
- inflation formulas;
- burn accounting;
- staking ratio controller;
- reputation scoring;
- domain pricing;
- resolver validation;
- memo validation;
- token/NFT/SBT storage rules.

## Keeper Tests

- `x/fee`;
- `x/token`;
- `x/identity`;
- `x/resolver`;
- `x/reputation`;
- `x/execution`;
- `x/messaging`;
- `x/queue`;
- `x/events`;
- `x/actors`;
- `x/storage`.

## Integration Tests

- bank transfer with memo;
- resolver-based payment;
- domain auction to ownership;
- token creation and transfer;
- NFT mint and transfer;
- SBT mint and transfer rejection;
- async contract call;
- queue bounce/refund;
- reputation rate limit;
- dynamic fee under load.

## E2E Smoke

- 3-validator localnet;
- 5-validator localnet;
- staking lifecycle;
- fee distribution;
- domain lifecycle;
- resolver payment;
- AVM counter contract;
- AFT token transfer;
- ANFT mint/transfer;
- ASBT mint/prove/revoke;
- memo indexing;
- restart persistence;
- snapshot/state-sync.

## Security Gates

- `go test ./...`;
- `go vet ./...`;
- `buf lint`;
- deterministic execution gate;
- state export/import gate;
- govulncheck;
- gosec;
- gitleaks;
- CodeQL;
- dependency review;
- independent audit before production claim.

## Production Gate

Production cannot be claimed until:

- long-running public testnet has no untriaged consensus or fund-safety issues;
- validator set can upgrade safely;
- staking, fees, DEX, AVM, domains, reputation, memo, and contract standards have adversarial tests;
- state export/import is deterministic;
- snapshot/state-sync works;
- emergency governance and halt/restart process tested;
- audit findings are triaged.

---

# Immediate Build Order

1. Finish base-chain rename, address policy, and `naet` cleanup.
2. Finish base-chain safety and validation helpers.
3. Implement production fee formulas in `x/fee`.
4. Implement adaptive mint/burn controller in `x/token`.
5. Harden PoS/staking and distribution.
6. Add deterministic memo metadata support.
7. Add `x/reputation` with deterministic score only.
8. Add `.aet` domain registry and auction model.
9. Add resolver and resolver-based payment.
10. Build deterministic async queue without AVM.
11. Build minimal AVM with a counter contract.
12. Add actor model and messaging.
13. Implement AW-5 wallet.
14. Implement AFT-44 token master/wallet.
15. Implement ANFT-66 NFT collection/item.
16. Implement ASBT-67 soulbound item.
17. Add scheduler parallelism only after deterministic sequential async execution is stable.
18. Add compute/storage/market modules after baseline abuse controls exist.
19. Gate CosmWasm behind explicit config and tests.
20. Start partitioning/sharding simulator and spec only after async queue and AVM are audited.

---

# Final Economic Architecture Summary

Aetheris economy is controlled by four feedback loops:

```text
staking participation -> adaptive inflation -> validator/delegator rewards
network activity      -> burn             -> supply pressure reduction
network load          -> soft fees/queues -> congestion control
account behavior      -> reputation       -> anti-spam and priority
```

The intended long-term behavior:

- low usage: mint rewards keep validators paid;
- normal usage: fees stay low and burn offsets part of mint;
- high usage: burn rises, queue controls activate, fees rise softly but stay capped;
- low staking: inflation rises within cap to attract staking;
- high staking: inflation falls to reduce dilution;
- spam: rate limits, reputation, deposits, and queue caps absorb abuse before fees become punitive.

This creates a self-regulating production L1 model where `AET` has uncapped but bounded PoS supply, `naet` remains the only protocol fee asset, users keep cheap transactions, and validators have a durable security budget.
