# Genesis Export And Migration Contract

## Reliability Rules

- `types.GenesisState.Validate` is the single validation gate for imported module state.
- Keeper `InitGenesis` methods validate before writing to stores, so malformed genesis cannot partially initialize custom modules.
- Keeper `ExportGenesis` methods return errors and validate the exported state before returning it.
- AppModule `InitGenesis` and `ExportGenesis` may panic only because the Cosmos SDK module interface has no error return; these are startup/operator-export invariants, not transaction execution paths.
- Custom module migrations must be registered for every `ConsensusVersion` bump. No-op migrations are allowed only when they still validate current exportable state.

## Current Custom Module Versions

| Module | Consensus version | Registered migration |
| --- | --- | --- |
| `x/fees` | `2` | `1 -> 2`, validates protocol fee params and accounting state |
| `x/tokenfactory` | `2` | `1 -> 2`, validates factory denom metadata and params |
| `x/dex` | `2` | `1 -> 2`, validates pool IDs, pair uniqueness, reserves, LP supply, and params |

## Determinism And Scale

Custom module exports iterate over KVStore prefixes and rely on deterministic store key ordering. They do not iterate Go maps for exported state. DEX pair indexes are derived indexes and are not exported; importing pools rebuilds the pair index through keeper `SetPool`.

Full-state export is an operator action and can allocate memory proportional to exported module state. Large-state networks should run export on dedicated infrastructure and add streaming export helpers before mainnet-scale snapshots.

## Upgrade Discipline

Before merging a state-breaking change:

- increment the module `ConsensusVersion`;
- register a migration from the previous version;
- add a migration test that starts from the previous version map;
- add corrupted-state tests for any new fields or indexes;
- run `go test ./...`, `go vet ./...`, `buf lint`, `buf generate`, build, and localnet smoke.
