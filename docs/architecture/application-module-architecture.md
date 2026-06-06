> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Application Module Architecture

Track 2.4 defines the application layer as native modules and future contract
systems that sit above core account, bank, staking, fee, and safety rules.
Application modules must not bypass signer validation, zero-address rejection,
native `naet` fee policy, denom validation, genesis validation, or deterministic
event contracts.

## x/dex

`x/dex` remains the native DEX module until contract DEX pools and routers are
implemented, tested, audited, and migrated explicitly. The native DEX is the
current source of truth for AMM pool state, liquidity pool custody, swap
execution, LP accounting, and fee integration.

Required native DEX behavior:

- pool creator, liquidity provider, withdrawer, swap trader, and swap recipient
  are valid non-zero user addresses;
- native denom spoofing through pool denoms, factory denoms, LP denoms, or
  display metadata is rejected;
- reserves match the DEX module account balances;
- LP supply matches pool shares;
- swap math preserves constant-product constraints within fee and rounding
  policy;
- slippage bounds are checked before state mutation.

Future contract pool migration must be explicit. A migration plan must define
pool snapshot format, reserve custody transfer, LP token migration, router path
compatibility, invariant checks, rollback/pause procedure, and audit evidence.

## x/identity

`x/identity` is the planned `.aet` domain registry. The registry record is the
source of truth; NFT representation proves ownership for wallets and UIs but
does not replace registry ownership checks.

DomainRecord:

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

Identity responsibilities:

- `.aet` domain registry;
- domain auctions;
- domain ownership;
- domain NFT representation;
- resolver ownership permissions;
- expiry and renewal.

Initial domain names are normalized before validation and then restricted to
lowercase ASCII `a-z`, digits `0-9`, `-`, and `_`. Names with whitespace,
invisible Unicode, mixed-script spoofing, non-ASCII characters, unsupported
symbols, or a TLD other than `.aet` are rejected. Additional symbols require a
separate UI/security review before protocol enablement.

Owner and resolver addresses use central address validation and reject the zero
address by default. Resolver is optional, but a present resolver must be a valid
non-zero address.

### Domain Lifecycle

Domain lifecycle:

```text
available -> auction -> active -> expired -> available/auction
```

If a domain does not exist or is expired, registration starts an auction. An
active domain cannot be auctioned. A running auction cannot be restarted. The
default auction duration is `24h`; ownership duration is `365 days`.

Auction start price is deterministic and denominated in native base `naet`:

```text
length 1-3: reserved or governance-only
length 4:   premium_start_price = 10_000 AET
length 5-6: high_start_price    = 1_000 AET
length 7-10: medium_start_price = 100 AET
length 11+: low_start_price     = 10 AET
```

The default bid increment is `5%`. Bids inside the last `10m` extend the auction
by `10m`, capped at `6` extensions. The highest valid bid wins. Finalization
sets registry ownership, assigns or updates NFT representation, and sets expiry
to `365 days` after finalization. Failed or premature finalization cannot assign
ownership or steal bids.

Domain auction proceeds split deterministically:

```text
domain_bid_final =
  burn:      40%
  treasury:  40%
  rewards:   20%
```

Renewal extends expiry by `365 days`. The renewal fee is:

```text
renewal_fee = start_price(name_length) * 0.10
```

### Resolver

The resolver is implemented as the identity resolver submodule until a separate
`x/resolver` keeper is justified. Resolver records map domains to wallets,
contracts, multisigs, NFTs, DEX endpoints, or bounded custom keys.

ResolverRecord:

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

Built-in resolver keys are:

- `wallet`;
- `contract`;
- `multisig`;
- `nft`;
- `dex`.

Custom resolver keys are lowercase ASCII and bounded by protocol length limits.
Resolver targets must be valid non-zero addresses. Resolver metadata is bounded.

Payment routing:

```text
send AET to alice.aet:
  record = resolver.lookup("alice.aet")
  if record.primary == null:
    fail "domain not resolved"
  send funds to record.primary
```

Resolver rules:

- registry owner can update resolver records;
- delegated resolver manager can update only granted keys before grant expiry;
- unresolved domain transfer fails before funds move;
- expired domain resolution fails unless a future explicit grace policy is
  added;
- resolver updates emit deterministic events;
- domain NFT owner and registry owner must stay consistent at keeper wiring.

Advanced resolver features are tracked as explicit extensions: subdomains such
as `dex.alice.aet`, leasing, DAO ownership, revenue sharing policy, reverse
resolution, and multi-resolver records for wallets, contracts, and app-specific
endpoints. A subdomain resolves through the base registry domain owner, so
`dex.alice.aet` is controlled by the owner of `alice.aet`.

## x/workflow

`x/workflow` is the planned orchestration module for bounded multi-step
transactions where synchronous execution is explicitly required.

Workflow responsibilities:

- multi-step transaction orchestration;
- conditional execution;
- module interaction flows;
- atomic workflows where synchronous execution is required.

Initial workflow examples:

- resolver-based payment;
- domain auction finalization;
- token mint and wallet deploy;
- NFT mint and metadata attach;
- contract deployment plus first message.

Workflow execution must be bounded by maximum step count and payload size.
Workflow authority is a valid non-zero address. A workflow must never use
orchestration to bypass module-level authorization, `naet` fee checks,
zero-address rejection, replay protection, or per-module invariants.

## Test Boundary

Track 2.4 production code starts with pure validation packages:

- `x/identity/types`: domain name and record validation;
- `x/workflow/types`: bounded workflow and step validation;
- `x/dex`: native DEX keeper and invariant tests remain authoritative until a
  contract DEX exists.

Future keeper/proto additions must preserve these validation rules and extend
the test matrix before enabling new state transitions.
