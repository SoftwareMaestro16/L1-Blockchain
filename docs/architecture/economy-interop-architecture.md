> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Economy And Interop Module Architecture

Date: 2026-06-05

This document turns Roadmap V3 Track 2.2 and 2.3 into repository-visible
module boundaries. New economy and interop modules must not bypass the base
chain invariants: protocol fees are paid only in `naet`, native `AET` metadata
cannot be spoofed, zero addresses are invalid by default, and hard economic
bounds require software upgrade to relax.

## ECONOMY

### `x/fees`

Current module name: `x/fees`. Roadmap shorthand `x/fee` maps to this module.

Responsibilities:

- dynamic fee calculation;
- `naet`-only enforcement;
- fee distribution and deterministic accounting;
- congestion load tracking;
- fee params;
- anti-spam coordination with future reputation and scheduler modules.

State:

- allowed fee denom list, fixed to `["naet"]` in v1;
- base fee amount;
- hard max fee amount;
- target block utilization;
- congestion threshold;
- max tx gas;
- max block gas;
- max block tx count;
- sender rate limits;
- stake-weighted allowance and priority weights;
- fee distribution ratios;
- protocol fee accounting totals.

Queries:

- current fee params;
- current network load;
- estimated fee for tx gas limit;
- fee distribution accounting;
- module balance visibility for fee collector and accounting modules.

Message surface:

- `MsgUpdateParams` via governance authority only.

Rules:

- fee estimation is bounded by the same hard cap used by ante validation;
- fee overpayment above the current requirement does not increase priority;
- user-created tokens, IBC assets, wrapped assets, LP denoms, NFT/SBT assets,
  and display denom `AET` cannot pay protocol fees;
- spam controls rely on gas/rate/quota limits before fee escalation.
- future reputation scores are deterministic anti-spam inputs, not a direct
  reputation purchase or fee auction.

### `x/reputation`

Status: future deterministic anti-spam and scheduler signal. It is currently a
pure scoring package only and must not mutate chain state until keeper wiring is
explicitly designed.

Responsibilities:

- reputation state for accounts and contracts;
- score calculation from deterministic on-chain events;
- inactivity decay;
- progressive limits for new, normal, trusted, and elite accounts;
- bounded domain and contract reputation signals;
- spam, failed transaction, and slash penalties.

Rules:

- no direct reputation purchase;
- score is clamped to `0..100`;
- domain ownership can help but cannot dominate score;
- reputation staking, if added, is a bonded and slashable signal;
- reputation may feed rate limits, queue limits, and scheduling weight, but it
  cannot bypass native `naet` fee validation or authorization.
- low reputation raises memo/storage byte cost and tightens contract deploy
  limits before any fee-market escalation;
- high reputation may improve deterministic queue priority only within bounded
  weights shared by all validators.

### `x/token`

Status: future native economy controller. Do not confuse it with the existing
`x/tokenfactory` user-asset module or AFT-44 contract token standard.

Responsibilities:

- native emission;
- burn accounting;
- supply controller;
- coordination with user token standards without allowing user tokens to pay
  protocol fees.

Planned state:

- current inflation;
- target stake ratio;
- mint accumulator;
- burn accumulator;
- annualized mint and burn stats;
- supply controller params.

Planned messages:

- governance update within hard bounds;
- internal mint per block;
- internal burn accounting.

Planned queries:

- total native supply;
- current inflation;
- burn stats;
- staking ratio;
- net issuance.

Rules:

- native supply is uncapped but bounded by hard inflation caps;
- adaptive mint/burn math must use deterministic integer arithmetic;
- export/import must preserve mint and burn accumulators exactly;
- hard inflation and burn bounds cannot be relaxed by governance alone.

## INTEROP

### `x/ibc`

Responsibilities:

- cross-chain communication;
- packet lifecycle;
- asset transfer logic.

Rules:

- IBC assets cannot pay Aetra protocol fees;
- bridged and IBC assets must be clearly namespaced;
- bridge, resolver, domain, and indexer integrations must reject native `AET`
  or `naet` spoofing;
- IBC transfer metadata must not be interpreted as native token metadata unless
  an explicit governance-approved registry says so.

### `x/bridge`

Status: future module. It is not active consensus code in the current app.

Responsibilities:

- external chain bridging;
- wrapped asset registry;
- proof, validator, relayer, or oracle security model;
- bridge risk limits.

Tasks before activation:

- cap bridge mint exposure;
- isolate bridge assets from native `naet`;
- add emergency pause;
- add proof verification tests;
- add bridge insolvency and replay tests;
- document validator/oracle trust assumptions;
- run adversarial tests for spoofed native metadata, replayed withdrawals,
  proof reuse, and insolvent wrapped-asset supply.

Rules:

- bridge assets cannot become fee denoms;
- bridge mint/burn accounting must be deterministic and export/import safe;
- bridge emergency pause cannot seize native `naet` or bypass account
  signatures;
- public docs must distinguish native `AET` from wrapped or bridged assets.
