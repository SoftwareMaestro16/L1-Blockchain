# L1 Blockchain

Sovereign Cosmos SDK Layer 1 blockchain workspace.

This is not an L2, not a sidechain, and not dependent on Ethereum execution. The target stack is Go, Cosmos SDK, CometBFT, Protobuf, and module-isolated KVStore state.

## Current Stage

The repository is in architecture bootstrap. The first implementation increment defines the target chain architecture, module boundaries, genesis and params model, security testing strategy, and PR-sized scaffold plan.

No node binary or Cosmos application code has been scaffolded yet.

## Engineering Rules

- Follow the [engineering governance standard](docs/engineering-governance.md) before implementation, refactoring, testing, or debugging.
- Every feature must be small enough for a focused commit.
- Use Conventional Commit messages.
- Keep module state deterministic and isolated.
- Put configurable values in genesis or module params, not hardcoded logic.
- Keep MsgServer orchestration thin and place state logic behind keeper methods.
- Treat queries as read-only.
- Run checks, commit, and push to `origin/main` by default after completed work.

## Documents

- [System architecture](docs/architecture.md)
- [Repository structure](docs/repo-structure.md)
- [Module boundaries](docs/module-boundaries.md)
- [Genesis and params model](docs/genesis-params.md)
- [Security and testing strategy](docs/security-testing.md)
- [Engineering governance](docs/engineering-governance.md)
- [Bootstrap implementation plan](docs/bootstrap-plan.md)
- [COSMOS_L1_SKILLS](docs/cosmos-l1-skills.md)

## Privacy Guard

This repository tracks source and documentation only. Local research caches, node data, wallets, keys, mnemonics, validator keys, and private configuration are intentionally ignored.
