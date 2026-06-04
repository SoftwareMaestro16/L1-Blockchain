# Repository Structure

## Target Tree

```text
.
  app/
    app.go
    keepers.go
    modules.go
    services.go
    module_accounts.go
    store_keys.go
  cmd/
    l1d/
      main.go
  proto/
    l1/
      tokenfactory/
      dex/
      fees/
  x/
    tokenfactory/
      keeper/
      types/
      module.go
    dex/
      keeper/
      types/
      module.go
    fees/
      keeper/
      types/
      module.go
  tests/
    unit/
    integration/
    adversarial/
  docs/
```

## Ownership Rules

- `app/` wires keepers, stores, module order, hooks, and upgrade handlers.
- `app/module_accounts.go` owns module account permissions and blocked address policy.
- `app/store_keys.go` owns KV store key inventory and store key accessors.
- `cmd/l1d/` owns the Orbitalis daemon entrypoint; build artifacts are named `orbitalisd`.
- `proto/` owns public wire contracts for Msg, Query, state, and genesis.
- `x/<module>/keeper/` owns state access and business logic.
- `x/<module>/types/` owns generated types, constants, errors, and expected keeper interfaces.
- `tests/` owns cross-module, adversarial, and determinism test suites.

## Git Hygiene

- Commit source, docs, proto, tests, and CI config.
- Do not commit local node homes, validator keys, mnemonics, generated chain data, research caches, or private configuration.
- Keep commits PR-sized and independently reviewable.
