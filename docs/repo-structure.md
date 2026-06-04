# Repository Structure

## Target Tree

```text
.
  app/
    app.go
    module_manager.go
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
- `cmd/l1d/` owns the Orbitalis daemon entrypoint; build artifacts are named `orbitalisd`.
- `proto/` owns public wire contracts for Msg, Query, state, and genesis.
- `x/<module>/keeper/` owns state access and business logic.
- `x/<module>/types/` owns generated types, constants, errors, and expected keeper interfaces.
- `tests/` owns cross-module, adversarial, and determinism test suites.

Proto source and generated Go must follow [proto-workflow.md](proto-workflow.md): edit `.proto`, run `buf lint`, verify generation into `.work\bufgen`, and update checked-in generated files only from approved buf output.

## Git Hygiene

- Commit source, docs, proto, tests, and CI config.
- Do not commit local node homes, validator keys, mnemonics, generated chain data, research caches, or private configuration.
- Keep commits PR-sized and independently reviewable.
