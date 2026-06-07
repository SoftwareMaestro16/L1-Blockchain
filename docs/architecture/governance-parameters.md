# Governance and Parameters

Aetra governance controls network parameters, but governance must not be able to execute unsafe values. Governance is a control plane with safety bounds, not a bypass around protocol invariants.

## 27. Governance Specification

Governance must be powerful enough to tune the network, but not powerful enough
to accidentally destroy it through invalid params.

## 27.1 Governance-Controlled Modules

Governance may control:

- staking policy params;
- economics params;
- validator score params;
- slashing params within bounds;
- CosmWasm upload policy;
- treasury spend;
- validator set growth schedule;
- block gas/size within safe bounds.

The implementation catalog maps those modules to parameter categories:

- `staking_policy`: validator set size, validator power cap, commission
  floor/max, max commission change;
- `economics`: inflation min/max, target bonded ratio, fee split;
- `validator_score`: validator score policy;
- `slashing`: double-sign slash, downtime slash, downtime window;
- `vm`: CosmWasm upload policy;
- `treasury`: treasury spend policy;
- `validator_set_growth`: validator set growth schedule;
- `consensus`: block gas limit and block max bytes.

Each parameter must have an explicit key, value type, category, min/max or enum validation, genesis validation, proposal execution validation, and event emission.

## 27.2 Param Safety Bounds

Every param must define:

- type;
- default value;
- min value;
- max value;
- authority;
- whether change is immediate or epoch-delayed;
- event emitted on change;
- tests for invalid update.

Critical params should apply only at epoch boundary to avoid surprising
mid-block behavior.

Safety requirements:

- params must have min/max validation;
- unsafe params must be rejected at proposal execution;
- genesis validation must reject invalid params;
- parameter changes must emit events;
- critical changes should use longer voting period or higher quorum.

Critical changes include validator set size, validator power cap, inflation min/max, target bonded ratio, fee split, slashing fractions, downtime windows, CosmWasm upload policy, and treasury spend policy. These changes can affect consensus safety, validator economics, slashing risk, VM attack surface, or treasury custody, so they require the critical governance path.

Recommended baseline:

```text
normal voting period: 10,000 blocks
critical voting period: 20,000 blocks
normal quorum: 40%
critical quorum: 50%
```

## Proposal Execution Rules

Proposal execution must:

- load the current parameter spec;
- verify that the parameter key is known;
- validate integer params against min/max bounds;
- validate enum params against allowed policy values;
- reject values outside safety bounds before writing state;
- emit a deterministic event containing parameter key, previous value, new value, proposal id, and executor;
- refuse critical changes that did not use the critical voting period or critical quorum.

Unsafe parameter values must fail closed. A passed governance proposal is not enough to write invalid state.

## Genesis Validation Rules

Genesis validation must:

- require all genesis-required governance params;
- reject unknown parameter keys;
- reject duplicate keys;
- reject values outside min/max bounds;
- reject enum values outside allowed policy values;
- validate cross-parameter relationships before launch.

The goal is to prevent a network from booting with an unsafe validator set size, unsafe fee split, unsafe slashing fraction, open CosmWasm upload mode, or treasury spend policy that bypasses governance.

## Event Requirements

Every parameter change must emit events. Events are required for:

- explorer visibility;
- validator monitoring;
- off-chain risk alerts;
- governance audit trails;
- post-upgrade incident analysis.

Event content must be deterministic and indexer-friendly. Events must not rely on external time, external APIs, or non-deterministic metadata.

## Parameter Catalog

The implementation catalog is `DefaultGovernanceParameterSpecs` in `app/params/governance_parameters.go`.

Required catalog properties:

- every governed parameter is bounded;
- every governed parameter is checked during genesis validation;
- every governed parameter change emits events;
- every governed parameter has authority metadata;
- every governed parameter has default value metadata;
- every governed parameter declares immediate or epoch-delayed application;
- every governed parameter has invalid update tests;
- critical parameters require longer voting period or higher quorum;
- critical parameters apply at epoch boundary;
- enum policies are closed lists, not free-form strings.
