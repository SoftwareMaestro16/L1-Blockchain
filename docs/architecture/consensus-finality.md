# Consensus, Block Production, and Finality

Aetra uses CometBFT BFT consensus and must keep block production aligned with
the validator-set size. The network deliberately avoids 1-2 second blocks
because that would increase validator networking pressure and push Aetra toward
a performance-first design that conflicts with the decentralization goal.

## Block Time Targets

Recommended active validator set policy:

- 100-128 validators: block time 5-6 seconds.
- 150-200 validators: block time 6 seconds.
- 250-300 validators: block time 7-8 seconds.

The configured target must stay inside the phase range. A 500+ validator set is
not a startup target.

## Finality Targets

Required finality targets:

- normal finality: 5-15 seconds.
- network stress finality: 20-90 seconds.
- worst acceptable target: <= 120 seconds.

CometBFT finality is expected after commit when at least 2/3 voting power is
healthy. Aetra must preserve liveness in degraded validator scenarios when
>= 2/3 voting power is healthy.

## Acceptance Criteria

Public testnet readiness requires:

- localnet with 100 validators must remain stable.
- localnet/load profile must demonstrate block production under configured
  block time.
- degraded validator scenarios must preserve liveness when >= 2/3 voting power
  is healthy.
- finality measurements must be included in testnet reports.

The executable policy lives in `app/params`:

- `DefaultNetworkProfile` defines the validator-set growth phases.
- `BlockTimeTargetRange` returns the configured block-time range by active
  validator count.
- `ValidateConsensusFinalityReport` validates localnet, degraded-scenario, and
  testnet-report acceptance evidence.

## Vote Extensions

Vote extensions may be used only for small deterministic-adjacent data:

- validator telemetry summary.
- oracle-like future extensions.
- encrypted mempool shares if implemented later.

Rules:

- keep vote extensions small.
- verify signatures before trusting extension data. In the ABCI handler this
  means rejecting extension verification requests that do not include a
  validator address from the signed vote path.
- avoid large payloads that hurt consensus latency.
- avoid non-deterministic validation.
- cover handlers with tests.

The current handler enforces explicit vote extension kinds, a maximum encoded
extension size, a maximum data size, deterministic data validation, non-empty
validator address, and tamper rejection.
