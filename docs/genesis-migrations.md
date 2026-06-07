> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Genesis Export And Migration Contract

## Reliability Rules

- `types.GenesisState.Validate` is the single validation gate for imported module state.
- Keeper `InitGenesis` methods validate before writing to stores, so malformed genesis cannot partially initialize custom modules.
- Keeper `ExportGenesis` methods return errors and validate the exported state before returning it.
- AppModule `InitGenesis` and `ExportGenesis` may panic only because the Cosmos SDK module interface has no error return; these are startup/operator-export invariants, not transaction execution paths.
- Custom module migrations must be registered for every `ConsensusVersion` bump. No-op migrations are allowed only when they still validate current exportable state.
- App-level genesis policy runs before module `InitGenesis` and rejects
  Aetra-specific invariants that the SDK cannot infer, including duplicate
  auth accounts, zero auth accounts, invalid bank balances, duplicate balances,
  staking denom drift away from `naet`, mint denom drift away from `naet`, and
  fee denom drift away from `naet`.
- Contract-standard executable specs validate token masters, token wallets, NFT
  collections/items, SBT items, and wallet contracts before VM wiring.
- Contract-standard executable specs validate NFT collections/items.
- Async VM exported state validates contract accounts, duplicate contract
  addresses, message envelopes, queued message sequences, inbox/outbox views,
  and queue `next_sequence` before import.

## Acceptance Chain

The required deterministic chain is:

```text
DefaultGenesis -> InitChain/InitGenesis -> ExportAppStateAndValidators -> ValidateGenesis
```

Acceptance tests assert that:

- default genesis validates before init;
- init and export are panic-free on valid default state;
- repeated export of the same state is byte-identical;
- exported state validates through the module manager;
- duplicate auth accounts, malformed account `Any` values, duplicate balances, malformed balance addresses, supply mismatch, staking denom drift, and fee denom drift are rejected before they can become committed state.
- malformed contract state, duplicate contract addresses, malformed async
  queued messages, duplicate queue sequences, and queue sequence drift are
  rejected by the async executable specification.

## Current Custom Module Versions

| Module | Consensus version | Registered migration |
| --- | --- | --- |
| `x/fees` | `2` | `1 -> 2`, validates protocol fee params and accounting state |
| `x/tokenfactory` | `2` | `1 -> 2`, validates factory denom metadata and params |
| `x/dex` | `2` | `1 -> 2`, validates pool IDs, pair uniqueness, reserves, LP supply, and params |
| `x/aetravm/async` | n/a | executable spec only; import validates exported queue state before runtime wiring |
| `x/aetravm/standards/*` | n/a | executable specs only; standard states validate independently from VM choice |

## Legacy Format Isolation

Old Orbitalis public formats (`ORB`, `norb`, `orb1`, and raw `0:` addresses)
must not be accepted by public validation paths. If historical data migration is
ever required, it must live in explicit migration-only tooling with:

- an upgrade name;
- a bounded input set;
- old-format parsing isolated from `app/addressing` public validators;
- output normalized to Aetra `naet`, `AET`, raw `4:`, and userfriendly
  `AE...` formats;
- regression tests proving old formats still fail normal genesis, tx, query, and
  contract-standard validation.

## Determinism And Scale

Custom module exports iterate over KVStore prefixes and rely on deterministic store key ordering. They do not iterate Go maps for exported state. DEX pair indexes are derived indexes and are not exported; importing pools rebuilds the pair index through keeper `SetPool`.

Full-state export is an operator action and can allocate memory proportional to exported module state. Large-state networks should run export on dedicated infrastructure and add streaming export helpers before mainnet-scale snapshots.

## Upgrade Discipline

Before merging a state-breaking change:

- increment the module `ConsensusVersion`;
- register a migration from the previous version;
- add a migration test that starts from the previous version map;
- add corrupted-state tests for any new fields or indexes;
- update the upgrade handler only through an explicit upgrade name and a reviewed version-map migration path;
- never mutate store layout without a migration that validates existing exportable state before and after the migration;
- run `go test ./...`, `go vet ./...`, `buf lint`, generated-protobuf verification, build, and localnet smoke.
