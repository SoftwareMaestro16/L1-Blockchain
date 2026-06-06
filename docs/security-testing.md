> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
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
- Wrong fee denom, mixed fee denoms, malformed fee coins, and non-`FeeTx` ante inputs.
- Swap inputs targeting rounding edge cases.
- Duplicate DEX pairs, wrong LP denoms, reserve/module balance desync, LP supply mismatch, and excessive slippage minimums.
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

The prototype release gate lives in [Prototype Security And Determinism Audit Gate](security/prototype-audit-gate.md). It is the mandatory runnable checklist for V2 prototype release readiness.

The Cosmos-specific manual review lives in [Cosmos Security Audit Checklist](security/cosmos-security-checklist.md). It is mandatory for prototype changes touching app wiring, custom modules, proto/query surface, localnet scripts, or release artifacts.

The public-testnet security audit pack lives in [Security Audit Pack](security/security-audit-pack.md). It ties manual checklists, `govulncheck`, `gosec`, CodeQL, gitleaks, Dependency Review, Cosmos checks, contract checks, and high/critical triage into one release blocker.

The base-chain pre-contract gate lives in [Account And Transaction Safety](security/account-transaction-safety.md). AVM and CosmWasm must not mutate production state until signer, fee, address, zero-address, malformed tx, replay, consensus panic, event, scan, and determinism evidence is green.

The transaction lifecycle matrix lives in [Prototype Transaction Lifecycle Matrix](transaction-lifecycle-matrix.md). It traces prototype txs from actor and signer through state writes, events, verification queries, negative cases, and test evidence.

The tx event contract lives in [Prototype Tx Event Contract](event-contract.md). It defines stable custom event types and attributes for e2e evidence and future indexers without replacing state queries.

The explicit coverage matrix lives in [Prototype Test Pyramid](test-pyramid.md). It maps each module and working flow to unit, integration, adversarial, e2e, determinism, and benchmark coverage.
