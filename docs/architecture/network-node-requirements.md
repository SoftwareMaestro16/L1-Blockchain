# Network and Node Requirements

Aetra validator infrastructure must stay in the medium hardware class. The network should not require Solana-style machines for normal validating, but it must still require enough resources to protect liveness, deterministic execution, state sync, snapshots, AVM execution, and stable CometBFT voting.

## Hardware Target

Initial public testnet validator recommendation:

```text
CPU: 4-8 modern cores
RAM: 16-32 GB
Storage: NVMe SSD
Network: stable 100 Mbps+, low packet loss
OS: Linux recommended, Windows local tooling supported for development
```

Mainnet requirements should be finalized after load testing. The genesis/testnet profile is intentionally moderate; mainnet numbers must be confirmed with AVM execution benchmarks, state growth benchmarks, 100-300 validator localnet runs, snapshot creation tests, and degraded-network finality reports.

Acceptance rules:

- public testnet validators must satisfy the medium hardware profile;
- public validator operations should use Linux;
- Windows local tooling is supported for development and local smoke tests, not as the recommended public validator OS;
- hardware requirements must not be raised toward performance-first requirements unless governance accepts the decentralization tradeoff;
- mainnet hardware requirements must not be declared final before load testing is complete.

## Sync and State Management

Required node capabilities:

- state sync support;
- snapshots;
- pruning profiles;
- archive node profile;
- export/import reliability;
- restart safety;
- deterministic app hash across restarts;
- documented validator setup;
- documented sentry architecture.

Pruning profiles:

```text
default:
  normal validator profile

nothing:
  archive node profile, preserves historical state

everything:
  aggressive pruning, acceptable for low-disk development or non-critical nodes only

custom:
  operator-defined pruning for documented launch profiles
```

State sync must be backed by published RPC servers, trust height, trust hash, and trust period values. Snapshot production must be tested before public testnet launch, and snapshot metadata must be published in a form that operators can verify.

Export/import reliability is a launch gate. Aetra must be able to export genesis, re-import it, restart, and produce the same deterministic app hash across restarts for the same state. Any module that touches staking, slashing, AVM state, fee burn, treasury accounting, validator score, or custom economics must keep export/import safe state.

## Sentry Architecture

Public validators should not expose their validator node directly to the open P2P network.

Recommended topology:

```text
public peers <-> sentry nodes <-> private validator node
```

Rules:

- validator node keeps `priv_validator_key.json` private and never copies it to sentries;
- sentry nodes expose P2P and optionally RPC, but do not hold consensus signing keys;
- validator node connects only to trusted sentries through persistent peers or private peering;
- firewall rules restrict inbound access to the validator node;
- public RPC/indexer traffic should be served by non-validator infrastructure;
- sentries should be geographically and provider diversified when possible;
- restart and upgrade runbooks must preserve `priv_validator_state.json` to avoid double-sign risk.

The sentry architecture is mandatory documentation for public testnet and mainnet readiness because it reduces denial-of-service risk against validators and makes regulatory or operator pressure harder to focus on a single exposed machine.

## Test Requirements

Implementation must include tests for:

- public testnet hardware bounds;
- rejection of weak hardware and performance-first extreme profiles;
- Linux public validator recommendation;
- Windows local development support;
- mainnet load-testing gate;
- state sync support readiness;
- snapshot readiness;
- pruning profile coverage;
- archive node profile;
- export/import reliability;
- restart safety;
- deterministic app hash across restarts;
- documented validator setup;
- documented sentry architecture.
