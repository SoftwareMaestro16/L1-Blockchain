# Security Requirements

Aetra security requirements are launch gates. A feature that violates consensus determinism or economic invariants must not be enabled on public testnet or mainnet, even if it passes normal functional tests.

## Consensus Safety

Required:

- deterministic state transitions;
- no non-deterministic external calls in consensus path;
- no wall-clock dependency in app state transitions except consensus-provided block time;
- no floating point accounting;
- no unordered map iteration affecting state;
- deterministic serialization;
- export/import equality tests;
- app hash stability tests.

Additional consensus rules:

- invalid tx, invalid genesis, invalid query, and malformed evidence paths must return errors instead of panicking;
- consensus code must not depend on local filesystem state, environment variables, RPC calls, HTTP calls, DNS, or node-local telemetry;
- block time may be used only through the consensus-provided block header/context;
- maps may be used internally only when outputs are canonicalized before state writes, hashing, serialization, events, or validator updates;
- integer or fixed-point math must be used for fees, rewards, stake, voting power, slashing, rent, burn, and treasury accounting;
- deterministic serialization must be stable across restarts and export/import.

Required tests:

- deterministic genesis serialization;
- deterministic export/import equality;
- app hash stability after restart;
- app hash stability after replaying the same blocks;
- static determinism gate for wall-clock, randomness, goroutines, `select`, floating point, external calls, unordered map state effects, and platform-dependent serialization;
- cross-architecture determinism review before mainnet.

## Economic Safety

Required:

- no unbounded mint;
- no unauthorized module account mint/burn;
- supply invariants;
- fee split invariants;
- delegation share invariants;
- reward distribution invariants;
- slashing cannot underflow stake;
- jailed validators cannot receive active validator rewards incorrectly.

Additional economic rules:

- every mint path must have authority checks, bounded amount checks, module-account checks, and event emission after bank success;
- every burn path must verify source authority and must not mutate module state before bank movement succeeds;
- total supply must reconcile after mint, burn, fee collection, reward distribution, slashing, treasury routing, and export/import;
- fee split accounting must satisfy `fees == burned + validators/delegators + treasury + remainder`;
- delegation share accounting must preserve stake/share ratios through delegate, redelegate, unbond, slash, export, and import;
- reward distribution must use fixed-point math and deterministic rounding;
- slashing must clamp or reject penalties that exceed available slashable stake;
- jailed, tombstoned, inactive, or unbonded validators must not receive active validator rewards.

Required tests:

- mint cap and authority tests;
- unauthorized module mint/burn tests;
- bank supply invariant tests;
- fee split invariant tests;
- delegation share invariant tests;
- reward distribution invariant tests;
- slashing underflow tests;
- jailed validator reward exclusion tests;
- export/import supply stability tests.

## Implementation Contract

The implementation gate is `app/params/security_requirements.go`.

Required catalog properties:

- `DefaultConsensusSafetyRequirements` must pass every consensus safety gate;
- `DefaultEconomicSafetyRequirements` must pass every economic safety gate;
- missing consensus determinism controls must fail the report;
- missing economic invariant controls must fail the report;
- `ValidateSlashingDoesNotUnderflowStake` must reject slash amounts above stake;
- `ValidateActiveValidatorRewardEligibility` must reject rewards for jailed, tombstoned, or inactive validators.

This gate does not replace keeper-level invariant tests. It prevents the architecture from drifting away from the required security model while module-specific tests prove the implementation details.
