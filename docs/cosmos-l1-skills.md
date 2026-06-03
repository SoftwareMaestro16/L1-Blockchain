# COSMOS_L1_SKILLS

This knowledge pack is consulted before Cosmos SDK L1 design or implementation work in this repository.

## Architecture

- A Cosmos SDK L1 is a deterministic state machine driven by CometBFT through ABCI.
- `app.go` is the composition root for stores, keepers, module order, hooks, and upgrades.
- Modules own isolated KVStore state and expose Msg and Query protobuf services.
- AppHash safety depends on deterministic state transitions, deterministic encoding, and identical ordered inputs.

## Message Flow

- External flow: client signs tx, node checks it, CometBFT orders it, BaseApp executes it, state commits.
- Internal flow: AnteHandler, MsgServiceRouter, MsgServer, Keeper, KVStore.
- MsgServer validates and orchestrates; keepers own state logic.
- Message execution in a tx is atomic.

## Keeper Rules

- Keepers are the only write path to module state.
- Cross-module access uses explicit keeper interfaces injected at app construction.
- QueryServer methods are read-only.
- Avoid global mutable state and hidden cross-module writes.

## Module Design

- One module should own one bounded domain.
- Keep the Msg surface minimal because messages are public APIs.
- Params are governance-controlled configuration.
- State key layout must be planned early because migrations are required for changes.

## Ignite CLI

- Use Ignite to scaffold the chain and boilerplate, then review generated app wiring before implementation.
- Use `--skip-git` inside this existing repository.
- Use `--require-registration` when scaffolding modules after the app exists.

## Common Errors

- Hardcoded consensus constants.
- Keeper bypasses or direct foreign store access.
- Query methods that mutate state.
- Floating-point math in consensus logic.
- Local timestamps, external APIs, randomness, or unordered map iteration in state transitions.
- Missing genesis validation and missing migration path.

## Best Practices

- Pin exact dependency versions in `go.mod`.
- Keep files below 500 lines by splitting keeper, msg, query, genesis, and params logic.
- Test keeper logic first, MsgServer authorization second, and full app flows third.
- Keep BlockSTM disabled until deterministic parallel execution has chain-specific test coverage.
- Run the engineering governance loop before code changes and record the outcome in the final implementation summary.
