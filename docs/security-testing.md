# Security And Testing Strategy

## Threat Model

Assume malicious users, malicious validators, malformed transactions, mempool spam, governance abuse, corrupted imports, and adversarial liquidity operations.

Primary attack classes:
- Replay attempts and nonce misuse.
- Unauthorized token mint or burn.
- Malformed Protobuf and invalid address encoding.
- Double-spend attempts through cross-module flows.
- AMM invariant manipulation.
- Governance parameter abuse.
- Unbounded loops and state bloat DoS.
- Nondeterministic state transitions.

## Unit Tests

Per module:
- Keeper happy paths and failures.
- MsgServer validation and authorization.
- QueryServer read-only behavior.
- Genesis validation and import/export round trips.
- Params validation and governance authority checks.

## Adversarial Tests

Current attacker models and implemented adversarial/e2e coverage are documented in [Adversarial And E2E Coverage](adversarial-e2e-coverage.md) and [Custom Module Attacker Model](../tests/adversarial/ATTACKER_MODEL.md).

Required cases:
- Invalid signer and unauthorized admin calls.
- Duplicate or replayed messages where applicable.
- Malformed denoms, addresses, metadata, pool IDs, and fee weights.
- Swap inputs targeting rounding edge cases.
- Attempts to create zero-liquidity or insolvent pools.
- Governance param values that disable safety bounds.

## Integration Tests

Required flows:
- Full tx lifecycle from signing through state query.
- Tokenfactory denom creation into bank supply.
- DEX pool creation using tokenfactory assets.
- Protocol fee accrual through swaps.
- Governance-controlled param update.
- App and custom module genesis export/import round trips.
- Module migrations from previous consensus version maps.

## Determinism Tests

Required checks:
- Same genesis and tx sequence produce the same exported state.
- Repeated block execution inputs produce the same app hash.
- No test depends on local wall clock, map iteration order, external APIs, or floating-point math.

## Review Loop

After each implementation increment:
- Architecture self-review.
- Security audit notes.
- Scalability and gas analysis.
- Monolith risk check.
- Missing invariant list.
- Next refactor proposal.

The detailed governance loop lives in [Engineering Governance](engineering-governance.md) and is mandatory before code changes, after tests, and before commits.

Genesis and upgrade rules live in [Genesis Export And Migration Contract](genesis-migrations.md).
