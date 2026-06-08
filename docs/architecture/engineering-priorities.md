# Engineering Priorities

Priority order:

```text
P0:
  consensus safety
  deterministic state
  staking correctness
  slashing correctness
  supply invariants
  export/import

P1:
  validator power cap
  fee burn/economics
  validator score
  nomination pool safety
  governance bounds

P2:
  AVM production hardening
  observability
  dashboards
  load tests
  public testnet docs

P3:
  advanced anti-cartel analytics
  AVM language research
  MEV policy
  encrypted mempool research
  higher validator cap experiments
```

Do not start P3 until P0 and P1 are stable.

## Execution Rule

P0 is the foundation. Consensus safety, deterministic state, staking correctness, slashing correctness, supply invariants, and export/import must be treated as launch blockers. A feature that weakens P0 cannot be justified by UX, speed, APR, or research value.

P1 is the core Aetra differentiation layer. Validator power cap, fee burn/economics, validator score, nomination pool safety, and governance bounds must stabilize before the network can claim the balanced BFT PoS model described by the architecture.

P2 prepares the public testnet surface. AVM production hardening, observability, dashboards, load tests, and public testnet docs should progress after P0/P1 are controlled enough that operators can test real flows without hiding consensus or economic instability.

P3 is research and advanced policy. Advanced anti-cartel analytics, AVM language research, MEV policy, encrypted mempool research, and higher validator cap experiments must not distract from P0/P1. P3 work may be documented as research, but implementation should not enter the production path until P0 and P1 are stable.

## Acceptance Gate

The implementation gate is `DefaultAetraEngineeringPrioritiesEvidence` in `app/params/engineering_priorities.go`.

Required behavior:

- missing P0 item fails readiness;
- missing P1 item fails readiness;
- missing P2 item fails readiness;
- missing P3 item fails readiness;
- duplicate or unexpected priority items fail readiness;
- P3 is rejected unless P0 and P1 are marked stable.
