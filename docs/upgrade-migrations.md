> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Upgrade Dry-Run And Migration Checklist

Prototype upgrades are consensus-critical. A migration must be deterministic, bounded, and covered before it is scheduled on any public network.

## Current Sanity

The current `UpgradeName` handler is a no-op migration pattern: it delegates to `ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)`. App tests verify:

- the stored module version map includes every app module,
- custom modules `tokenfactory`, `dex`, and `fees` have current consensus
  version `2`,
- missing module versions and impossible future versions are rejected before migration,
- the registered no-op upgrade handler can run from the current version map,
- export after the dry-run upgrade produces valid genesis.

## Future Migration Checklist

Before adding a real migration:

1. Define an upgrade name and expected from/to module versions.
2. Keep store upgrades explicit and minimal: added, renamed, or deleted stores must be documented.
3. Validate the pre-migration version map includes every module touched by the migration; pass explicitly allowed new modules only when a migration intentionally adds a module.
4. Reject or fail loudly on impossible/corrupted state rather than silently rewriting it.
5. Use deterministic iteration order only; no wall time, randomness, goroutines, external APIs, or unordered map writes.
6. Bound all loops by existing state collections and document expected cardinality.
7. Preserve module-account and bank accounting invariants.
8. Export after migration and validate exported genesis.
9. Add regression tests for malformed old state and the expected migrated state.
10. Run `go test ./...`, `go vet ./...`, `buf lint`, proto generated verification if proto changed, and the prototype audit gate.

## Legacy Format Rule

Old Orbitalis public data formats (`ORB`, `norb`, `orb1`, and raw `0:`
addresses) must not be reintroduced through normal upgrade handlers. Migration
support for historical data is allowed only inside explicit migration-only
tools or named upgrade handlers, and the output must be normalized to Aetra
formats before being accepted by normal app validation.

Migration output must be normalized to Aetra formats.

## Security Review

Use the Cosmos security audit checklist for:

- migration panic paths,
- missing module version entries,
- invalid store loader definitions,
- corrupted custom module genesis after export,
- nondeterministic state writes,
- module account balance/supply mismatch.

Any Critical/High migration finding blocks release until it has a regression test or a documented owner decision.
