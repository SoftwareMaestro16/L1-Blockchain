# Aetra Architecture and Implementation Specification

## 1. Цель документа

Этот документ описывает техническое задание для реализации блокчейна Aetra как balanced BFT PoS Layer 1 на базе Cosmos SDK и CometBFT.

Aetra не должна быть сетью, которая гонится за максимальным TPS любой ценой. Основная цель - построить доверенную, умеренно быструю, хорошо децентрализованную и экономически устойчивую BFT-сеть с поддержкой смарт-контрактов, строгим objective slashing и защитой от концентрации stake у отдельных валидаторов или групп.

Документ должен использоваться как рабочее ТЗ для разработки, тестирования, аудита и подготовки public testnet/mainnet.

## 2. Архитектурная философия

Aetra выбирает баланс между:

- децентрализацией;
- скоростью и финальностью;
- безопасностью;
- доступностью валидаторов;
- экономической эффективностью.

Главный принцип:

```text
Aetra is not maximum speed.
Aetra is not maximum APR.
Aetra is not maximum simplicity.

Aetra is a trusted, moderately fast, economically balanced BFT PoS L1.
```

Сеть должна быть явно быстрее Ethereum по практической финальности, но не должна превращаться в high-hardware сеть уровня Solana, TON или других performance-first систем.

## 3. Базовая модель сети

Целевая модель:

```text
Consensus: CometBFT BFT
Staking: PoS + delegation + nomination pools
VM: AVM first, AVM only at genesis
Non-AVM VMs: post-genesis compatibility research only
Validator set: 100-300 active validators
Recommended genesis: 100-128 validators
Growth phase: 150-200 validators
Mature cap: 250-300 validators
Block time: 5-8 seconds
Normal finality: 5-15 seconds
Stress finality: 20-90 seconds
Worst acceptable finality target: <= 120 seconds
Hardware profile: medium, no Solana-style requirements
Inflation: dynamic 2-5 percent normal range
APR target: 5-8 percent for delegators
Fee model: burn + validator/delegator rewards + treasury
```

Финальность в CometBFT наступает после commit блока, если есть 2/3+ voting power. Aetra должна проектироваться так, чтобы нормальная финальность была в диапазоне 5-15 секунд, а degraded network finality не выходила за 120 секунд при здоровом majority наборе валидаторов.

## 4. Validator Set Policy

### 4.1 Целевые размеры validator set

Реализация должна поддерживать постепенный рост active validator set:

```text
Phase 1, genesis / early testnet:
  active validators: 100-128
  target block time: 5-6 seconds
  normal finality: 5-10 seconds

Phase 2, stable public testnet:
  active validators: 150-200
  target block time: 6 seconds
  normal finality: 6-12 seconds

Phase 3, mature network:
  active validators: 250-300
  target block time: 7-8 seconds
  normal finality: 8-15 seconds
```

Не запускать mainnet сразу с 300 валидаторами, если нет достаточного числа надежных операторов. Слабые валидаторы могут ухудшить liveness, синхронизацию и UX сильнее, чем улучшить децентрализацию.

### 4.2 Причины не использовать маленький validator set

Validator set меньше 80 не подходит как долгосрочная модель, потому что:

- выше cartel risk;
- проще давление на ключевых операторов;
- слабее восприятие доверия;
- меньше географическое и организационное разнообразие.

### 4.3 Причины не использовать слишком большой validator set на старте

Validator set 500+ не должен быть начальной целью, потому что:

- растет consensus overhead;
- усложняется синхронизация;
- растет latency;
- повышается риск слабых операторов;
- сложнее контролировать качество инфраструктуры.

## 5. Staking and Delegation Model

### 5.1 Требования к staking

Нужно реализовать или настроить:

- staking denom: native Aetra unit, например `naet`;
- validator self-bond;
- validator bond requirement;
- delegation;
- redelegation;
- unbonding;
- nomination pools;
- slashing inheritance for delegators;
- commission limits;
- validator metadata and identity profile.

### 5.2 Self-bond and validator bond

Вход должен быть умеренным, но не символическим:

```text
minimum self-bond:
  medium level, enough to prove skin in the game

minimum validator bond:
  higher than minimum self-bond

unbonding period:
  14-21 days
```

Точные числовые значения должны быть параметрами genesis/governance, а не hardcoded constants.

### 5.3 Nomination pools

Nomination pools должны позволять пользователям участвовать в staking без запуска валидаторской инфраструктуры.

Требования:

- pool accounting must be deterministic;
- pool shares must be precisely tracked;
- pool withdrawals must respect unbonding period;
- pool operator must not be able to steal principal;
- pool commission must be bounded;
- pool delegation must inherit validator slashing risk;
- pool state must be export/import safe.

## 6. Anti-Centralization Model

### 6.1 Validator power cap

Ключевая особенность Aetra - ограничение effective voting power валидатора.

```text
effective_power = min(raw_bonded_power, validator_power_cap)
```

Рекомендуемые значения:

```text
<= 150 active validators:
  validator_power_cap: 3.0 percent

151-250 active validators:
  validator_power_cap: 2.5 percent

> 250 active validators:
  validator_power_cap: 2.0 percent
```

Если валидатор собрал stake выше cap:

- excess stake не увеличивает voting power;
- excess stake получает сниженные rewards или нулевые rewards;
- делегаторам должны быть доступны предупреждения через query/API/UI;
- новые делегации к overloaded validator должны быть экономически менее выгодны.

Пример:

```text
Validator raw stake: 5 percent
Validator cap: 3 percent

Effective voting power: 3 percent
Normal rewards: first 3 percent
Reduced/no rewards: excess 2 percent
```

### 6.2 Concentration targets

Система должна мониторить концентрацию:

```text
top-10 voting power target: < 25 percent
top-20 voting power target: < 40 percent
top-33 voting power target: < 50 percent
single validator cap: 2-3 percent
```

Если концентрация превышает target, сеть не должна аварийно ломать staking. Вместо этого должны включаться экономические сигналы:

- lower reward multiplier for over-concentrated validators;
- delegation warnings;
- protocol metrics;
- governance alert;
- optional parameter adjustment proposal.

### 6.3 Anti-cartel controls

Power cap не решает проблему полностью, потому что один оператор может создать несколько валидаторов. Поэтому нужны дополнительные меры:

- commission floor: 3-5 percent;
- max commission: 15-20 percent;
- max daily commission change: 0.5-1 percent;
- optional validator identity registry without mandatory KYC;
- validator metadata transparency;
- public concentration metrics;
- self-bond ratio visibility;
- operator correlation warnings where evidence is public and objective.

Нельзя делать жесткий KYC как базовое требование консенсуса. Aetra должна снижать риск картеля через прозрачность, лимиты и экономические стимулы, а не через централизованное разрешение на участие.

## 7. Slashing and Validator Accountability

### 7.1 Главный принцип slashing

Slashing должен применяться только за объективные, криптографически доказуемые нарушения.

```text
slash only for objective, cryptographically provable faults
```

Нельзя штрафовать за спорные, субъективные или внешне интерпретируемые действия. Иначе slashing может стать атакующим инструментом против честных валидаторов.

### 7.2 Double-sign

Требования:

```text
double-sign:
  slash: 5-10 percent
  jail: immediate
  tombstone: permanent
  consensus key reuse: forbidden
```

Нужно использовать стандартные возможности Cosmos SDK `x/slashing` и `x/evidence`, но параметры должны быть явно заданы и покрыты тестами.

### 7.3 Downtime

Downtime должен наказываться прогрессивно:

```text
first offense:
  slash: 0.05-0.1 percent
  jail: 1-6 hours

repeat offense:
  slash: 0.25-0.5 percent
  jail: 24 hours

chronic downtime:
  slash: up to 1 percent
  jail: longer period
  optional governance/reputation flag
```

Implementation note: если стандартный `x/slashing` недостаточен для progressive downtime, добавить кастомный модуль поверх стандартного slashing state, не ломая evidence flow CometBFT.

### 7.4 Invalid proposal

Invalid proposal не должен автоматически приводить к slashing, если нарушение не доказуемо.

Требования:

- invalid proposal must be rejected deterministically;
- repeated objective invalid proposals may cause jail/slash;
- all proposal validation must be deterministic;
- `ProcessProposal` must avoid non-deterministic external inputs;
- tests must cover deterministic accept/reject behavior.

### 7.5 Timestamp manipulation

Для timestamp policy:

- reject block if timestamp outside allowed consensus/application bounds;
- use CometBFT-compatible timestamp rules;
- avoid custom time logic that can break consensus;
- slash only if signed evidence is objective and reproducible;
- repeated signed violations may trigger jail/slash.

### 7.6 Height manipulation

Height is consensus-controlled. A single validator must not be able to manipulate height.

Practical risks to cover:

- double-sign at the same height;
- equivocation;
- invalid proposal for a height;
- non-deterministic app validation;
- evidence expiration edge cases;
- unbonding and evidence timing.

## 8. Economics

### 8.1 Target economy

High APR must not be the selling point of Aetra. The chain should avoid an inflation trap.

Recommended parameters:

```text
target bonded ratio: 55-65 percent
inflation min: 1.5-2 percent
inflation normal: 3-4 percent
inflation max: 5-6 percent
delegator APR target: 4-7 percent
validator net APR target: 6-9 percent
```

APR approximation:

```text
staking_apr ~= inflation / bonded_ratio
```

Examples:

```text
3 percent inflation / 60 percent bonded = ~5.0 percent APR
4 percent inflation / 60 percent bonded = ~6.7 percent APR
5 percent inflation / 60 percent bonded = ~8.3 percent APR
```

### 8.2 Fee split

Recommended fee split:

```text
30-60 percent fees -> burn
20-40 percent fees -> validators/delegators
10-20 percent fees -> protocol treasury
```

The split must be governance-configurable within safe bounds.

### 8.3 Burn and supply behavior

Requirements:

- fee burn must be visible in events and queries;
- total burned amount must be queryable;
- supply accounting must be deterministic;
- burn must not break module account invariants;
- high activity periods may partially offset inflation;
- network may become near neutral-supply or temporarily deflationary if fees are high enough.

### 8.4 Reward smoothing

Rewards should be smoothed to avoid extreme short-term variance:

- epoch-based reward distribution;
- bounded reward changes per epoch;
- clear commission calculation;
- delegation reward accounting tests;
- validator reward accounting tests;
- export/import safe reward state.

## 9. Custom Aetra Modules

The implementation should include or finalize three core custom modules.

### 9.1 `x/aetra-staking-policy`

Purpose:

- enforce validator power cap;
- apply delegation overflow rules;
- enforce commission floor/max;
- enforce max commission change;
- calculate anti-concentration reward multipliers;
- expose validator concentration queries.

Required messages:

- update staking policy params through governance authority;
- optionally register validator identity metadata;
- optionally acknowledge concentration warning for validator/operator UI.

Required queries:

- params;
- validator effective power;
- validator raw stake vs effective stake;
- validator overflow stake;
- top-N concentration;
- validator reward multiplier;
- delegation warning status.

Required tests:

- unit tests for cap calculation;
- unit tests for cap transition at 150/250 validators;
- property/fuzz tests for effective power never exceeding cap;
- integration tests with staking keeper;
- export/import tests;
- genesis validation tests;
- deterministic query tests.

### 9.2 `x/aetra-validator-score`

Purpose:

- track validator uptime score;
- track governance participation score;
- track slash history;
- track decentralization score;
- expose public validator metrics;
- support wallet/explorer ranking without centralizing consensus.

Score must not directly override consensus unless explicitly approved by governance and tested. The first version should be informational plus reward modifier only where objective.

Required metrics:

- uptime window;
- missed blocks;
- jail history;
- slash history;
- self-bond ratio;
- commission history;
- governance participation;
- concentration status;
- optional identity metadata completeness.

Required tests:

- uptime accounting tests;
- missed block window tests;
- score update epoch tests;
- slash history tests;
- governance participation tests;
- export/import tests;
- deterministic scoring tests.

### 9.3 `x/aetra-economics`

Purpose:

- dynamic inflation;
- target bonded ratio logic;
- fee burn split;
- treasury split;
- validator/delegator reward split;
- reward smoothing;
- protocol treasury accounting.

Required params:

- inflation min;
- inflation max;
- target bonded ratio;
- burn percentage min/max/current;
- validator reward percentage;
- treasury percentage;
- reward smoothing window;
- APR target bounds;

Required queries:

- current inflation;
- current bonded ratio;
- estimated APR;
- fee split params;
- burned supply;
- treasury balance;
- epoch reward summary.

Required tests:

- inflation curve tests;
- bounded inflation tests;
- fee split accounting tests;
- burn accounting tests;
- treasury accounting tests;
- APR estimate tests;
- epoch reward tests;
- export/import tests;
- invariant tests for supply.

## 10. Smart Contract VM

### 10.1 Genesis requirement

Aetra should use AVM first and AVM only for genesis.

Current implementation direction:

- AVM is the default and production-target smart-contract runtime;
- AVM modules, docs, smoke tests, and malicious-contract tests are the canonical launch path;
- CosmWasm is optional compatibility research only and remains disabled unless an explicit governance/feature gate enables it after security, gas, storage-rent, and export/import review;
- default app wiring must not add the CosmWasm store key, genesis state, Msg service, Query service, or CLI surface;
- token, NFT, and DEX behavior remains contract-routed under the AVM-first policy and must not be reintroduced as native asset modules.

Reasons:

- native fit with Aetra accounts, storage rent, pool capability hooks, and proof metadata;
- deterministic execution model;
- strict resource accounting under one VM surface at genesis;
- lower attack surface than launching multiple contract VMs simultaneously;
- easier performance control;
- easier audit scope for initial network.

### 10.2 Non-AVM VM policy

CosmWasm, EVM, and other non-AVM runtimes are not genesis requirements and must not be wired into the genesis app path.

Before any post-genesis compatibility runtime or adapter can be considered:

- complete security review;
- complete gas model review;
- benchmark state growth;
- test interaction with fee burn;
- test interaction with staking rewards;
- test export/import;
- run adversarial contract tests.

### 10.3 AVM requirements

Required:

- instantiate/execute/query/migrate support;
- governance or permissioned code upload policy for early testnet;
- deterministic gas limits;
- contract size limits;
- contract storage rent or storage pricing;
- events compatible with indexers;
- localnet smoke tests;
- malicious contract tests;
- export/import tests with active contracts.

## 11. Consensus, Block Production, and Finality

### 11.1 Block time targets

Recommended:

```text
100-128 validators:
  block time: 5-6 seconds

150-200 validators:
  block time: 6 seconds

250-300 validators:
  block time: 7-8 seconds
```

Avoid 1-2 second blocks. That would increase pressure on validator networking and push Aetra toward a performance-first design that conflicts with the decentralization goal.

### 11.2 Finality targets

Required:

```text
normal finality: 5-15 seconds
network stress finality: 20-90 seconds
worst acceptable target: <= 120 seconds
```

Acceptance criteria:

- localnet with 100 validators must remain stable;
- localnet/load profile must demonstrate block production under configured block time;
- degraded validator scenarios must preserve liveness when >= 2/3 voting power is healthy;
- finality measurements must be included in testnet reports.

### 11.3 Vote extensions

Vote extensions may be used only for small deterministic-adjacent data such as:

- validator telemetry summary;
- oracle-like future extensions;
- encrypted mempool shares if implemented later.

Rules:

- keep vote extensions small;
- verify signatures before trusting extension data;
- avoid large payloads that hurt consensus latency;
- avoid non-deterministic validation;
- cover handlers with tests.

## 12. Network and Node Requirements

### 12.1 Hardware target

Target validator hardware should be medium, not extreme.

Initial recommendation for public testnet:

```text
CPU: 4-8 modern cores
RAM: 16-32 GB
Storage: NVMe SSD
Network: stable 100 Mbps+, low packet loss
OS: Linux recommended, Windows local tooling supported for development
```

Mainnet requirements should be finalized after load testing.

### 12.2 Sync and state management

Required:

- state sync support;
- snapshots;
- pruning profiles;
- archive node profile;
- export/import reliability;
- restart safety;
- deterministic app hash across restarts;
- documented validator setup;
- documented sentry architecture.

## 13. Governance and Parameters

Governance must control network parameters but within safety bounds.

Governance-controlled params:

- validator set size;
- validator power cap;
- commission floor/max;
- max commission change;
- inflation min/max;
- target bonded ratio;
- fee split;
- slashing fractions;
- downtime windows;
- AVM contract upload policy;
- treasury spend policy.

Safety requirements:

- params must have min/max validation;
- unsafe params must be rejected at proposal execution;
- genesis validation must reject invalid params;
- parameter changes must emit events;
- critical changes should use longer voting period or higher quorum.

## 14. Observability and Public Metrics

Required metrics:

- block time;
- finality latency;
- missed blocks;
- validator uptime;
- validator concentration;
- top-10/top-20/top-33 voting power;
- inflation;
- bonded ratio;
- estimated APR;
- burned fees;
- treasury balance;
- slashing events;
- jail/unjail events;
- contract execution gas;
- failed tx reasons;
- node sync status.

Required surfaces:

- CLI queries;
- gRPC queries;
- REST queries where applicable;
- Prometheus metrics;
- explorer/indexer compatibility events;
- public testnet dashboards.

## 15. Security Requirements

### 15.1 Consensus safety

Required:

- deterministic state transitions;
- no non-deterministic external calls in consensus path;
- no wall-clock dependency in app state transitions except consensus-provided block time;
- no floating point accounting;
- no unordered map iteration affecting state;
- deterministic serialization;
- export/import equality tests;
- app hash stability tests.

### 15.2 Economic safety

Required:

- no unbounded mint;
- no unauthorized module account mint/burn;
- supply invariants;
- fee split invariants;
- delegation share invariants;
- reward distribution invariants;
- slashing cannot underflow stake;
- jailed validators cannot receive active validator rewards incorrectly.

### 15.3 Permission safety

Required:

- module account permissions validated at startup;
- reserved addresses cannot sign user txs;
- blocked addresses cannot receive normal user funds unless explicitly allowed;
- governance authority checked;
- params authority checked;
- keeper wiring tests.

## 16. Testing Requirements

Every implementation task must include tests. A feature is not complete without tests.

### 16.1 Required test layers

Unit tests:

- keeper logic;
- params validation;
- math and accounting;
- cap calculation;
- slashing policy;
- reward split;
- inflation curve;
- score calculation.

Integration tests:

- staking + custom staking policy;
- slashing + validator score;
- distribution + economics;
- fee collector + burn + treasury;
- nomination pool + delegation + unbonding;
- governance param updates;
- AVM tx flow.

E2E/localnet tests:

- node startup;
- validator creation;
- delegation;
- redelegation;
- unbonding;
- downtime scenario;
- double-sign evidence scenario where feasible;
- fee burn scenario;
- AVM instantiate/execute/query;
- export/import;
- restart;
- state sync/snapshot where feasible.

Adversarial tests:

- concentration attack simulation;
- validator overflow stake simulation;
- commission manipulation attempt;
- invalid params proposal;
- malformed evidence;
- jailed validator reward attempt;
- module account abuse attempt;
- contract gas exhaustion;
- contract storage abuse.

Performance tests:

- 100 validator localnet/profile;
- 150-200 validator simulation/profile if feasible;
- block time under load;
- finality latency measurement;
- mempool pressure;
- AVM execution load;
- state growth profile.

### 16.2 Test acceptance rule

No module should be considered production-ready unless:

- unit tests pass;
- integration tests pass;
- genesis validation tests pass;
- export/import tests pass;
- deterministic restart tests pass;
- adversarial tests for the relevant module pass;
- CI runs the critical subset automatically.

## 17. Implementation Phases

### Phase 0 - Baseline audit

Tasks:

- inspect current Cosmos SDK and CometBFT versions;
- document current app module graph;
- identify existing modules overlapping with `aetra-staking-policy`, `aetra-validator-score`, and `aetra-economics`;
- decide which modules are renamed, reused, or wrapped;
- verify current `naet` staking denom;
- verify fee collector, burn, treasury, emissions, mint authority wiring;
- verify current localnet scripts and test coverage.

Deliverables:

- module inventory;
- gap analysis;
- risk list;
- updated implementation checklist.

Tests:

- current full unit test run;
- current integration test run;
- current localnet smoke test;
- current export/import test.

### Phase 1 - Staking policy and validator cap

Tasks:

- implement effective voting power cap;
- implement overflow stake accounting;
- implement commission floor/max/max-change policy;
- add concentration metrics;
- add queries for validator raw/effective/overflow stake;
- add governance params with validation;
- wire module into app lifecycle.

Tests:

- cap math unit tests;
- validator set transition tests;
- concentration query tests;
- commission bounds tests;
- integration tests with staking;
- export/import tests;
- invariant tests.

Acceptance:

- no validator can exceed configured effective power cap;
- excess stake does not increase voting power;
- params cannot be set outside safe bounds;
- state remains deterministic after export/import.

### Phase 2 - Economics and fee split

Tasks:

- implement dynamic inflation bounds;
- implement target bonded ratio logic;
- implement fee split to burn/rewards/treasury;
- implement reward smoothing;
- expose APR estimate query;
- expose burned supply and treasury accounting queries;
- add governance param controls.

Tests:

- inflation curve tests;
- bonded ratio tests;
- fee split tests;
- burn accounting tests;
- treasury accounting tests;
- APR query tests;
- supply invariant tests;
- export/import tests.

Acceptance:

- inflation remains within configured bounds;
- fee split sums to 100 percent;
- burned fees reduce spendable/module-held supply according to chain accounting rules;
- treasury receives correct amount;
- rewards are deterministic.

### Phase 3 - Validator score and accountability

Tasks:

- implement uptime score;
- implement slash history;
- implement governance participation score;
- implement decentralization score;
- implement public validator metrics queries;
- integrate score with reward modifier only for objective inputs.

Tests:

- uptime window tests;
- missed block tests;
- slash history tests;
- governance participation tests;
- score determinism tests;
- reward modifier tests;
- export/import tests.

Acceptance:

- score is deterministic;
- score cannot be manipulated through subjective inputs;
- score is queryable for explorers and wallets;
- score does not break consensus safety.

### Phase 4 - Slashing hardening

Tasks:

- configure double-sign slash fraction and tombstone behavior;
- configure downtime windows and jail duration;
- implement progressive downtime if not covered by standard module;
- add timestamp/proposal violation policy where objective;
- document evidence lifecycle and unbonding interaction.

Tests:

- double-sign evidence tests where feasible;
- downtime tests;
- jail/unjail tests;
- progressive downtime tests;
- slashing accounting tests;
- delegator loss tests;
- tombstone tests;
- evidence expiry tests.

Acceptance:

- double-sign leads to severe slash and permanent tombstone;
- downtime penalties are bounded and progressive;
- no subjective slashing path exists;
- slashing cannot underflow stake or corrupt shares.

### Phase 5 - AVM integration

Tasks:

- finalize AVM module wiring;
- define code upload policy;
- define contract gas limits;
- define contract size limits;
- integrate storage rent or storage pricing;
- expose contract events for indexers;
- document contract developer flow.

Tests:

- instantiate/execute/query tests;
- migration tests;
- gas limit tests;
- storage limit/rent tests;
- malicious contract tests;
- export/import tests with contracts;
- localnet AVM smoke test.

Acceptance:

- contracts are deterministic;
- contract gas is bounded;
- malicious contracts cannot halt chain;
- contract state survives export/import.

### Phase 6 - Finality and performance profile

Tasks:

- configure block time targets;
- configure block size/gas limits;
- profile 100 validator localnet;
- profile 150-200 validator scenario if feasible;
- estimate 250-300 validator operational requirements;
- measure finality under load;
- measure finality under partial validator failure.

Tests:

- localnet load profile;
- mempool pressure test;
- block time measurement;
- finality measurement;
- validator failure scenario;
- restart scenario;
- state sync/snapshot scenario.

Acceptance:

- normal finality remains within target;
- stressed finality remains below 120 seconds in healthy majority scenario;
- node requirements remain medium-level;
- no excessive consensus payloads.

### Phase 7 - Public testnet readiness

Tasks:

- write validator setup docs;
- write sentry architecture docs;
- write monitoring docs;
- publish genesis parameter explanation;
- publish economic model explanation;
- publish slashing risk explanation;
- publish delegation and nomination pool guide;
- publish AVM developer guide;
- prepare public dashboards;
- prepare incident response process.

Tests:

- clean node bootstrap from docs;
- validator join from docs;
- snapshot restore from docs;
- state sync from docs;
- tx flow smoke tests;
- governance proposal smoke tests;
- public RPC/gRPC/REST smoke tests.

Acceptance:

- new validator can join using docs;
- public endpoints are observable;
- network can recover from node restarts;
- core economic and staking flows work end to end.

## 18. Mainnet Readiness Criteria

Aetra should not be considered mainnet-ready until all criteria are met:

- validator set policy implemented and tested;
- effective power cap implemented and tested;
- anti-concentration rewards implemented and tested;
- dynamic inflation implemented and tested;
- fee burn/treasury/reward split implemented and tested;
- slashing configured and tested;
- AVM integrated and tested;
- export/import stable;
- state sync/snapshots stable;
- public testnet has run long enough to observe validator behavior;
- load tests demonstrate finality target;
- security audit completed;
- critical findings fixed;
- docs complete for validators, delegators, and contract developers.

## 19. Non-Goals for Initial Release

Initial release should not attempt:

- PoH;
- Solana-level TPS;
- 1-second blocks;
- mandatory KYC validator admission;
- any non-AVM VM at genesis;
- subjective slashing;
- unlimited validator set;
- unbounded contract execution;
- high inflation APR marketing.

## 20. Final Target Formula

```text
Aetra =
  CometBFT BFT PoS
  + Cosmos SDK
  + AVM-only genesis smart contracts
  + 100-300 active validators over time
  + 5-8 second block time
  + <= 120 second worst acceptable finality target
  + strict objective slashing
  + validator effective power cap
  + anti-concentration rewards
  + dynamic low/moderate inflation
  + fee burn
  + protocol treasury
  + mandatory tests for every feature
```

The most important product decision: Aetra should be a chain people can trust, not a chain optimized only for speed or short-term APR.

## 21. Detailed Engineering Scope

Этот раздел переводит архитектурную идею в конкретный engineering backlog. Любой пункт ниже должен быть реализован как production-grade функциональность, а не как placeholder.

Общее правило для всех задач:

```text
feature = code + params + genesis validation + queries + events + tests + docs
```

Если у фичи есть только код, но нет тестов, query surface, genesis validation или acceptance criteria, фича не считается завершенной.

### 21.1 Core chain configuration

Tasks:

- зафиксировать chain id naming policy для devnet, testnet, mainnet;
- зафиксировать staking denom `naet`;
- зафиксировать display denom `AET`;
- проверить coin metadata для `naet/AET`;
- проверить address prefix и reserved system address policy;
- проверить module account permissions;
- проверить blocked address policy;
- проверить mint authority;
- проверить burn authority;
- проверить fee collector authority;
- проверить treasury authority;
- проверить genesis validation для всех Aetra модулей;
- проверить app export/import with all modules enabled.

Expected deliverables:

- `app` wiring review;
- genesis params table;
- module accounts table;
- authority matrix;
- CLI command matrix;
- query matrix;
- event matrix;
- tests for startup validation.

Required tests:

- app boots with default genesis;
- app rejects invalid denom metadata;
- app rejects missing module accounts;
- app rejects duplicate reserved addresses;
- app rejects unsafe module account permissions;
- export/import preserves app hash where expected;
- simulation or integration test covers module initialization order.

### 21.2 Consensus parameter policy

Tasks:

- define target block time range;
- define max block bytes;
- define max block gas;
- define evidence max age by blocks;
- define evidence max age by duration;
- define validator public key types;
- define CometBFT timeout profile for 100, 200, 300 validators;
- define snapshot interval;
- define state sync parameters;
- define pruning profiles.

Recommended initial values must be conservative. Do not maximize block size early.

Example target policy:

```text
block_time_target:
  100 validators: 5-6s
  200 validators: 6s
  300 validators: 7-8s

max_block_gas:
  start conservative
  increase only after load tests

max_block_bytes:
  keep below values that increase propagation delay
  change only through governance after testnet evidence
```

Required tests:

- localnet remains stable under configured timeout profile;
- oversized blocks are rejected;
- invalid consensus params are rejected;
- governance cannot set unsafe block gas/bytes outside bounds;
- evidence remains valid through configured evidence period.

## 22. Module Specification: `x/aetra-staking-policy`

Цель: контролировать effective voting power, delegation overflow, commission policy и anti-concentration incentives.

This module is the central anti-centralization module of Aetra.

### 22.1 Responsibilities

The module must:

- calculate raw validator stake;
- calculate effective validator stake;
- calculate overflow stake;
- enforce or expose effective voting power cap;
- calculate reward multiplier for overflow stake;
- expose delegation concentration warnings;
- enforce commission floor;
- enforce max commission;
- enforce max commission change rate;
- expose top-N concentration metrics;
- validate governance param changes;
- emit events for cap/overflow/commission policy changes;
- remain deterministic and export/import safe.

### 22.2 State

Suggested state:

```text
Params:
  MaxValidatorsSoftTarget
  ValidatorPowerCapBps
  ValidatorPowerCapSchedule
  OverflowRewardMultiplierBps
  CommissionFloorBps
  CommissionMaxBps
  CommissionMaxDailyChangeBps
  Top10TargetBps
  Top20TargetBps
  Top33TargetBps
  MinSelfBond
  MinValidatorBond
  WarningThresholdBps

ValidatorPolicy:
  OperatorAddress
  RawBondedTokens
  EffectiveBondedTokens
  OverflowBondedTokens
  EffectivePowerBps
  IsOverCap
  RewardMultiplierBps
  LastCommissionChangeTime
  LastCommissionRateBps

ConcentrationSnapshot:
  Height
  BondedRatio
  ActiveValidators
  Top10Bps
  Top20Bps
  Top33Bps
  NakamotoCoefficientEstimate
```

All decimal values should use integer basis points or SDK decimal types consistently. Avoid floating point.

### 22.3 Parameter rules

Parameter validation:

```text
ValidatorPowerCapBps:
  min: 100      # 1%
  max: 500      # 5%
  recommended: 200-300

OverflowRewardMultiplierBps:
  min: 0
  max: 10000
  recommended: 0-3000 for overflow zone

CommissionFloorBps:
  min: 0
  max: 1000
  recommended: 300-500

CommissionMaxBps:
  min: CommissionFloorBps
  max: 3000
  recommended: 1500-2000

CommissionMaxDailyChangeBps:
  min: 1
  max: 500
  recommended: 50-100
```

Governance must not be able to set:

- cap below 1%;
- max commission below floor;
- overflow multiplier above normal reward multiplier;
- invalid top-N targets;
- zero active validator target;
- negative or overflowing math values.

### 22.4 Effective power calculation

The implementation must define whether cap affects:

1. only reward calculation;
2. actual CometBFT voting power;
3. both.

Recommended staged approach:

```text
Stage 1:
  cap affects rewards and delegation warnings
  low consensus risk

Stage 2:
  cap affects effective staking power used for validator updates
  requires deeper integration and heavy tests
```

If Stage 2 is implemented, the staking keeper integration must ensure:

- validator updates sent to CometBFT use capped power;
- total voting power remains consistent;
- no validator can exceed cap;
- delegation and unbonding shares remain correct;
- slashing can still slash the underlying raw stake;
- evidence handling remains correct.

### 22.5 Messages

Required governance-only or authority-only messages:

```text
MsgUpdateStakingPolicyParams
MsgUpdateValidatorPowerCapSchedule
MsgSetCommissionPolicy
```

Optional validator messages:

```text
MsgRegisterValidatorIdentity
MsgUpdateValidatorIdentity
MsgAcknowledgeOverCapWarning
```

All messages must:

- validate authority;
- validate signer;
- reject malformed addresses;
- reject invalid params;
- emit events;
- be covered by tests.

### 22.6 Queries

Required queries:

```text
Query/Params
Query/ValidatorPolicy
Query/ValidatorEffectivePower
Query/ValidatorOverflow
Query/TopNConcentration
Query/DelegationWarning
Query/CommissionPolicy
Query/ConcentrationSnapshot
Query/NakamotoCoefficient
```

Query responses must be stable and indexer-friendly.

### 22.7 Events

Required events:

```text
aetra.staking_policy.params_updated
aetra.staking_policy.validator_over_cap
aetra.staking_policy.validator_back_under_cap
aetra.staking_policy.commission_rejected
aetra.staking_policy.concentration_snapshot
aetra.staking_policy.reward_multiplier_changed
```

### 22.8 Invariants

Required invariants:

- effective power never exceeds configured cap;
- overflow stake is never negative;
- raw stake = effective stake + overflow stake for capped calculation;
- commission floor <= commission <= commission max;
- commission change <= max daily change;
- top-N calculations do not exceed 100%;
- state export/import preserves policy state.

### 22.9 Tests

Required tests:

- cap math for 100 validators;
- cap math for 150 validators;
- cap math for 250 validators;
- cap math for 300 validators;
- validator crossing cap upward;
- validator crossing cap downward;
- delegation to over-cap validator;
- redelegation from over-cap validator;
- unbonding from over-cap validator;
- slashing over-cap validator;
- commission below floor rejected;
- commission above max rejected;
- commission daily jump rejected;
- governance param update accepted within bounds;
- governance param update rejected outside bounds;
- export/import with over-cap validators;
- deterministic concentration snapshot.

## 23. Module Specification: `x/aetra-economics`

Цель: low/moderate inflation, fee burn, treasury allocation, reward smoothing и transparent APR model.

### 23.1 Responsibilities

The module must:

- calculate dynamic inflation;
- track bonded ratio;
- estimate staking APR;
- split fees;
- burn configured fee share;
- send configured share to distribution/rewards;
- send configured share to treasury;
- smooth reward changes;
- expose economic metrics;
- protect supply invariants.

### 23.2 State

Suggested state:

```text
Params:
  InflationMinBps
  InflationMaxBps
  InflationChangeRateBps
  TargetBondedRatioBps
  BurnFeeShareBps
  RewardFeeShareBps
  TreasuryFeeShareBps
  RewardSmoothingEpochs
  AprTargetMinBps
  AprTargetMaxBps

EpochEconomics:
  EpochNumber
  StartHeight
  EndHeight
  BondedRatioBps
  InflationBps
  EstimatedAprBps
  FeesCollected
  FeesBurned
  FeesToRewards
  FeesToTreasury
  MintedRewards

SupplyStats:
  TotalMinted
  TotalBurned
  NetIssuance
```

### 23.3 Inflation curve

Inflation should respond to bonded ratio:

```text
if bonded_ratio < target:
  increase inflation gradually

if bonded_ratio > target:
  decrease inflation gradually
```

Hard requirements:

- inflation never below min;
- inflation never above max;
- inflation change per epoch bounded;
- no floating point;
- no per-block instability;
- all calculations deterministic.

### 23.4 Fee split rules

Fee split must always sum to 100%.

Recommended initial range:

```text
BurnFeeShareBps: 3000-6000
RewardFeeShareBps: 2000-4000
TreasuryFeeShareBps: 1000-2000
```

Example:

```text
50% burn
35% validators/delegators
15% treasury
```

The module must reject fee split params if:

- sum != 10000 bps;
- any share is negative;
- burn share exceeds max governance bound;
- treasury share exceeds max governance bound;
- rewards share is zero unless explicitly permitted by emergency governance.

### 23.5 APR query

APR query must clearly distinguish:

- inflation-only APR;
- fee-adjusted APR;
- validator commission impact;
- estimated delegator APR;
- estimated validator gross APR;
- estimated validator net APR.

APR must be labeled as estimate, not guaranteed return.

### 23.6 Tests

Required tests:

- inflation increases when bonded ratio below target;
- inflation decreases when bonded ratio above target;
- inflation remains within min/max;
- inflation change rate bounded;
- fee split exact accounting;
- burn accounting;
- treasury accounting;
- rewards accounting;
- APR estimate math;
- zero-fee block handling;
- high-fee block handling;
- export/import economics state;
- supply invariant after many epochs;
- governance invalid params rejected.

## 24. Module Specification: `x/aetra-validator-score`

Цель: public accountability without subjective consensus control.

### 24.1 Responsibilities

The module must:

- track validator uptime;
- track missed block windows;
- track jail history;
- track slashing history;
- track commission behavior;
- track self-bond ratio;
- track governance participation;
- track concentration status;
- produce public score;
- expose explorer-friendly queries.

Score must not become a subjective censorship mechanism. It should be informational first and reward-affecting only when based on objective chain data.

### 24.2 State

Suggested state:

```text
Params:
  UptimeWindow
  UptimeWeightBps
  SlashHistoryWeightBps
  GovernanceWeightBps
  SelfBondWeightBps
  ConcentrationWeightBps
  MinScore
  MaxScore
  RewardModifierEnabled
  MaxRewardPenaltyBps

ValidatorScore:
  OperatorAddress
  Score
  UptimeScore
  SlashScore
  GovernanceScore
  SelfBondScore
  ConcentrationScore
  MissedBlocks
  SignedBlocks
  JailCount
  SlashCount
  LastUpdatedHeight
```

### 24.3 Score requirements

Score must be:

- deterministic;
- based only on chain state;
- explainable;
- queryable;
- bounded;
- export/import safe;
- resistant to overflow/underflow.

### 24.4 Tests

Required tests:

- perfect uptime score;
- partial uptime score;
- missed block penalty;
- jail penalty;
- slash penalty;
- governance participation score;
- concentration penalty;
- reward modifier bounded;
- score cannot go below min;
- score cannot exceed max;
- export/import;
- deterministic recomputation.

## 25. Slashing Implementation Details

### 25.1 Standard slashing integration

Use Cosmos SDK `x/slashing` and CometBFT evidence for base faults:

- double-sign;
- liveness/downtime;
- tombstone;
- jail/unjail.

Custom logic should wrap or extend standard behavior only where necessary. Do not fork core slashing logic unless there is no safer option.

### 25.2 Progressive downtime design

If progressive downtime is implemented, it should track repeated liveness faults:

```text
DowntimeOffense:
  ValidatorConsAddr
  OffenseCount
  FirstOffenseTime
  LastOffenseTime
  LastSlashFraction
  CurrentJailDuration
```

Rules:

- offense count decays after long clean period;
- repeated downtime increases penalty;
- maximum penalty is capped;
- delegators inherit validator downtime risk;
- validator can query own downtime status;
- unjail does not erase slash history immediately.

### 25.3 Evidence and unbonding

Tests must cover:

- evidence submitted while validator bonded;
- evidence submitted while validator unbonding;
- evidence submitted after unbonding but before evidence expiration;
- evidence submitted after expiration;
- slashing delegators who were bonded at infraction height;
- tombstone cap behavior.

### 25.4 Invalid proposal policy

Invalid proposal handling must be conservative:

- reject objectively invalid proposals;
- do not slash unless there is signed, reproducible evidence;
- do not depend on local wall clock;
- do not depend on external APIs;
- do not make `ProcessProposal` fragile.

Required tests:

- invalid tx proposal rejected;
- oversized proposal rejected;
- malformed special tx rejected;
- valid proposal accepted;
- same proposal accepted/rejected identically by all validators in test harness.

## 26. Nomination Pool Detailed Specification

Nomination pools are important for accessibility, but they introduce accounting and centralization risks.

### 26.1 Pool model

Each pool should have:

```text
Pool:
  PoolId
  OperatorAddress
  ValidatorAddress
  TotalBonded
  TotalShares
  CommissionBps
  Status
  CreatedHeight
  UnbondingEntries
```

Delegator state:

```text
PoolDelegation:
  DelegatorAddress
  PoolId
  Shares
  PrincipalEstimate
  RewardsAccrued
```

### 26.2 Pool requirements

Required:

- users deposit native staking denom;
- pool mints shares deterministically;
- pool delegates to validator;
- pool distributes rewards pro-rata;
- pool commission bounded;
- pool withdrawal follows unbonding period;
- pool slashing reduces share value;
- pool operator cannot withdraw user principal;
- pool cannot bypass validator power cap;
- pool must expose risk warnings.

### 26.3 Pool tests

Required tests:

- first deposit share price;
- subsequent deposit share price;
- reward distribution;
- commission deduction;
- partial withdrawal;
- full withdrawal;
- slashing pool validator;
- jailed validator;
- redelegation if allowed;
- pool operator abuse attempt;
- export/import with active unbonding entries;
- rounding dust handling.

## 27. Governance Specification

Governance must be powerful enough to tune the network, but not powerful enough to accidentally destroy it through invalid params.

### 27.1 Governance-controlled modules

Governance may control:

- staking policy params;
- economics params;
- validator score params;
- slashing params within bounds;
- AVM contract upload policy;
- treasury spend;
- validator set growth schedule;
- block gas/size within safe bounds.

### 27.2 Param safety bounds

Every param must define:

- type;
- default value;
- min value;
- max value;
- authority;
- whether change is immediate or epoch-delayed;
- event emitted on change;
- tests for invalid update.

Critical params should apply only at epoch boundary to avoid surprising mid-block behavior.

### 27.3 Governance tests

Required tests:

- valid param proposal executes;
- invalid param proposal rejected;
- unauthorized authority rejected;
- emergency unsafe value rejected;
- epoch-delayed param activation;
- event emitted;
- query reflects new params;
- export/import after param change.

## 28. AVM Detailed Requirements

### 28.1 Contract permissions

Define launch policy:

```text
early testnet:
  permissioned code upload or governance-gated upload

later testnet:
  permissionless upload with strong fees/deposits

mainnet:
  policy decided after security review
```

### 28.2 Gas and storage

Required:

- max AVM code size;
- max instantiate gas;
- max execute gas per tx;
- max query gas;
- storage rent or storage pricing;
- contract upload fee;
- contract migration authority rules;
- pinned code policy if used.

### 28.3 Contract security tests

Required tests:

- infinite loop contract hits gas limit;
- large storage write bounded;
- failed contract does not corrupt state;
- contract cannot access reserved module funds;
- migration authorization enforced;
- reply/submessage behavior deterministic;
- event emission stable;
- export/import with contracts;
- contract query does not mutate state.

## 29. Threat Model

### 29.1 Validator cartel

Threat:

- several validators coordinate censorship or governance capture.

Controls:

- 100-300 validator target;
- validator power cap;
- top-N monitoring;
- commission floor;
- identity transparency;
- governance participation metrics;
- delegation warnings.

Tests/simulations:

- top-10 concentration simulation;
- split-identity validator simulation;
- delegation overflow simulation;
- governance capture threshold analysis.

### 29.2 Stake centralization through rewards

Threat:

- large validators grow faster because delegators chase apparent safety/APR.

Controls:

- overflow rewards reduced;
- over-cap warnings;
- commission floor;
- concentration metrics;
- reward multiplier based on cap.

Tests:

- rewards for over-cap validator lower than normal;
- delegator APR estimate reflects overflow penalty;
- cap changes do not create accounting corruption.

### 29.3 Downtime and weak operators

Threat:

- too many low-quality validators reduce liveness.

Controls:

- minimum self-bond;
- validator score;
- downtime slashing;
- jail;
- public metrics;
- gradual validator set growth.

Tests:

- liveness with < 1/3 voting power offline;
- halt behavior with > 1/3 offline documented;
- recovery after validators return;
- downtime penalties applied.

### 29.4 Governance attack

Threat:

- malicious proposal changes economics, slashing, cap, or VM params dangerously.

Controls:

- param bounds;
- delayed activation;
- emergency review window for critical params;
- explicit authority checks;
- event monitoring.

Tests:

- malicious param proposal rejected;
- out-of-range values rejected;
- authority spoofing rejected;
- delayed activation works.

### 29.5 Contract attack

Threat:

- malicious AVM contract consumes gas/storage, exploits permissions, or causes state bloat.

Controls:

- gas limits;
- storage pricing;
- upload policy;
- migration controls;
- contract size limit;
- malicious contract test suite.

Tests:

- gas exhaustion;
- storage abuse;
- unauthorized migration;
- invalid instantiate;
- export/import with malicious-but-contained contract state.

## 30. API, CLI, Query, and Event Surface

Every module must expose enough surface for validators, wallets, explorers, dashboards and monitoring.

### 30.1 CLI requirements

Required command categories:

```text
aetrad query aetra-staking-policy ...
aetrad query aetra-economics ...
aetrad query aetra-validator-score ...
aetrad tx aetra-staking-policy ...
aetrad tx aetra-economics ...
aetrad tx aetra-validator-score ...
```

Commands must support:

- json output;
- height query where applicable;
- pagination where applicable;
- clear errors;
- examples in docs.

### 30.2 gRPC/REST requirements

Every query must have:

- protobuf definition;
- gRPC service;
- REST gateway mapping if project supports it;
- response examples;
- tests where feasible.

### 30.3 Events

Events must be emitted for:

- validator cap crossing;
- delegation overflow;
- reward multiplier change;
- fee burn;
- treasury allocation;
- inflation update;
- APR estimate update by epoch;
- validator score update;
- downtime offense;
- slash event;
- jail/unjail;
- governance param activation.

Events should include stable attributes:

```text
validator
delegator
amount
denom
height
epoch
old_value
new_value
reason
module
```

## 31. Data Migration and Upgrade Strategy

Aetra is expected to evolve. Upgrade safety is part of the architecture.

### 31.1 Upgrade requirements

Every new module or state-breaking change must include:

- store key decision;
- genesis import/export;
- migration handler;
- version map update;
- upgrade test;
- rollback notes where possible;
- operator instructions.

### 31.2 Migration tests

Required tests:

- old genesis imports into new binary;
- migration initializes params;
- migration preserves balances;
- migration preserves staking state;
- migration preserves slashing state;
- migration preserves contract state if applicable;
- app hash after migration is deterministic.

## 32. Repository-Level Work Breakdown

This section maps work to likely repository areas. Exact paths may change, but responsibilities should remain.

### 32.1 `proto/`

Tasks:

- define protobuf messages for new modules;
- define query services;
- define tx services;
- define genesis messages;
- define params messages;
- run code generation;
- add proto breaking-change checks if available.

Tests:

- generated code compiles;
- proto lint passes if configured;
- query/tx service registration tested.

### 32.2 `x/`

Tasks:

- implement keepers;
- implement message servers;
- implement query servers;
- implement genesis;
- implement params validation;
- implement invariants;
- implement hooks where needed;
- implement events;
- implement module interfaces.

Tests:

- keeper unit tests;
- msg server tests;
- query server tests;
- genesis tests;
- invariant tests;
- fuzz/property tests for math.

### 32.3 `app/`

Tasks:

- wire keepers;
- wire modules;
- wire module account permissions;
- wire begin/end/preblock order;
- wire simulation manager if used;
- wire API routes;
- wire AutoCLI if used;
- validate startup.

Tests:

- app startup;
- module account permissions;
- begin/end order;
- export/import;
- deterministic restart;
- API service registration.

### 32.4 `tests/`

Tasks:

- integration test suites;
- e2e localnet smoke tests;
- adversarial tests;
- load profile tests;
- documentation path tests;
- CI scripts.

Required:

- tests must be runnable from documented commands;
- Windows PowerShell local scripts should remain usable if current project supports them;
- Linux CI path should remain primary for production confidence.

## 33. Definition of Done

No task is complete until:

- code is implemented;
- params are validated;
- genesis import/export works;
- query surface exists;
- events exist where operationally relevant;
- unit tests pass;
- integration tests pass;
- e2e/localnet test exists for user-facing flow;
- docs describe operator/user impact;
- failure modes are documented;
- security implications are reviewed.

For consensus/economics/staking changes, also required:

- adversarial tests;
- invariant tests;
- export/import test;
- deterministic restart test;
- migration test if state changed.

## 34. Acceptance Test Matrix

Minimum acceptance matrix before public testnet:

```text
Base node:
  boot single node
  boot multi-validator localnet
  restart
  export/import
  state sync or snapshot restore

Staking:
  create validator
  delegate
  redelegate
  unbond
  withdraw rewards
  validator commission update

Anti-centralization:
  validator reaches cap
  validator exceeds cap
  excess stake reward penalty applied
  top-N concentration query works
  commission floor enforced

Slashing:
  downtime tracked
  downtime jail
  double-sign evidence path where feasible
  tombstone behavior
  delegator slash accounting

Economics:
  inflation update
  fee burn
  treasury allocation
  rewards allocation
  APR query
  supply invariant

AVM:
  upload code
  instantiate
  execute
  query
  migrate if enabled
  gas exhaustion contained

Governance:
  valid param proposal
  invalid param proposal
  treasury proposal
  delayed critical param activation

Observability:
  Prometheus metrics
  CLI queries
  gRPC queries
  events indexable
```

## 35. Engineering Priorities

Priority order:

```text
P0:
  consensus safety
  deterministic state
  staking correctness
  slashing correctness
  supply invariants
  export/import

P1:
  validator power cap
  fee burn/economics
  validator score
  nomination pool safety
  governance bounds

P2:
  AVM production hardening
  observability
  dashboards
  load tests
  public testnet docs

P3:
  advanced anti-cartel analytics
  non-AVM compatibility research
  MEV policy
  encrypted mempool research
  higher validator cap experiments
```

Do not start P3 until P0 and P1 are stable.

## 36. Concrete Near-Term Task List

Near-term implementation should be split into small pull requests:

1. Audit existing validator/economics modules and map them to this spec.
2. Add missing params validation for validator power cap and commission policy.
3. Add effective power and overflow stake queries.
4. Add concentration snapshot query.
5. Add cap math tests for 100/150/200/300 validator scenarios.
6. Add fee split accounting tests.
7. Add inflation bounds tests.
8. Add supply invariant tests.
9. Add validator score state and query tests.
10. Add progressive downtime design or document why standard slashing is enough for v1.
11. Add nomination pool accounting tests.
12. Add AVM smoke and malicious contract tests.
13. Add public testnet finality measurement script.
14. Add documentation for validators and delegators.
15. Add CI gate for critical unit/integration tests.

Each PR must state:

- what consensus/economics behavior changes;
- what params are added or changed;
- what tests were added;
- what migration risk exists;
- whether public docs need updates.
