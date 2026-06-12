# Launch Module Inventory

This document classifies every `x/*` module for testnet launch purposes.
Canonical source of truth is `app/launch_module_inventory.json`; this markdown
must mirror its launch scope and contract-only asset model.
Token/NFT/DEX behavior remains contract-only through AVM contracts, not native
application modules.

## Classification Legend

| Classification | Description |
|---------------|-------------|
| `launch_core` | Required for testnet launch - consensus critical |
| `launch_support` | Supports launch but not consensus-critical |
| `future_avm_standard` | AVM contract standards (not native modules) |
| `prototype_only` | In-memory state, prototype only |
| `disabled` | Not wired, should not be activated |

## Launch Core Modules

| Module | Consensus State | KV-Backed | Export/Import | Invariants | Notes |
|--------|---------------|-----------|---------------|------------|-------|
| `bank` | Yes | Yes | Full | Yes | Native token balances |
| `staking` | Yes | Yes | Full | Yes | Validator bonds only (no direct delegation) |
| `nominator-pool` | Yes | Yes | Full | Yes | Official liquid staking pools |
| `single-nominator-pool` | Yes | Yes | Full | Yes | Alternative pool model |
| `auth` | Yes | Yes | Full | Yes | Account authentication |
| `gov` | Yes | Yes | Full | Yes | On-chain governance |
| `slashing` | Yes | Yes | Full | Yes | Validator slashing |
| `mint` | Yes | Yes | Full | Yes | Token minting |
| `distr` | Yes | Yes | Full | Yes | Distribution |
| `evidence` | Yes | Yes | Full | Yes | Double-sign evidence |
| `aetravm` | Yes | Yes | Full | Yes | AVM contract runtime |
| `aetracore` | Yes | Yes | Full | Yes | Core protocol state |
| `validator-registry` | Yes | Yes | Full | Yes | Validator metadata |
| `validator-election` | Yes | Yes | Full | Yes | Validator selection |
| `validator-insurance` | Yes | Yes | Full | Yes | Validator insurance fund |

## Launch Support Modules

| Module | Consensus State | KV-Backed | Export/Import | Invariants | Notes |
|--------|---------------|-----------|---------------|------------|-------|
| `fees` | Yes | Yes | Full | Yes | Fee market |
| `fee-collector` | Yes | Yes | Full | Yes | Fee collection |
| `burn` | Yes | Yes | Full | Yes | Token burn |
| `treasury` | Yes | Yes | Full | Yes | Treasury management |
| `emissions` | Yes | Yes | Full | Yes | Emission distribution |
| `constitution` | Yes | Yes | Full | Yes | Protocol constitution |
| `reputation` | Yes | Yes | Full | Yes | Validator reputation |
| `delegator-protection` | Yes | Yes | Full | Yes | Delegator safety |
| `storage-rent` | Yes | Yes | Full | Yes | Storage rent |
| `dynamic-commission` | Yes | Yes | Full | Yes | Dynamic commission |
| `config` | Yes | Yes | Full | Yes | Configuration |
| `config-voting` | Yes | Yes | Full | Yes | Config governance |
| `epoch` | Yes | Yes | Full | Yes | Epoch tracking |
| `events` | Yes | Yes | Full | Yes | Event emission |
| `messages` | Yes | Yes | Full | Yes | Message routing |
| `payments` | Yes | Yes | Full | Yes | Payment channels |
| `compute` | Yes | Yes | Full | Yes | Compute resources |
| `storage` | Yes | Yes | Full | Yes | Storage resources |
| `scheduler` | Yes | Yes | Full | Yes | Task scheduling |
| `mesh` | Yes | Yes | Full | Yes | Mesh networking |
| `zones` | Yes | Yes | Full | Yes | Zone management |
| `routing` | Yes | Yes | Full | Yes | Message routing |
| `identity` | Yes | Yes | Full | Yes | Identity management |
| `identity-root` | Yes | Yes | Full | Yes | Identity root |
| `proofregistry` | Yes | Yes | Full | Yes | Proof registration |
| `native-account` | Yes | Yes | Full | Yes | Native accounts |
| `permissions` | Yes | Yes | Full | Yes | Permission system |
| `system-registry` | Yes | Yes | Full | Yes | System registry |
| `actor-registry` | Yes | Yes | Full | Yes | Actor registry |
| `actors` | Yes | Yes | Full | Yes | Actor management |
| `contracts` | Yes | Yes | Full | Yes | Contract management |
| `execution` | Yes | Yes | Full | Yes | Execution engine |
| `avm-scheduler` | Yes | Yes | Full | Yes | AVM scheduling |
| `cross-chain-registry` | Yes | Yes | Full | Yes | Cross-chain |
| `bridge-hub` | Yes | Yes | Full | Yes | Bridge hub |
| `queue` | Yes | Yes | Full | Yes | Message queue |
| `services` | Yes | Yes | Full | Yes | Service registry |
| `taskgroups` | Yes | Yes | Full | Yes | Task groups |
| `workflow` | Yes | Yes | Full | Yes | Workflow engine |
| `messaging` | Yes | Yes | Full | Yes | Messaging |
| `networking` | Yes | Yes | Full | Yes | Networking |
| `reporter` | Yes | Yes | Full | Yes | Reporting |
| `sharding` | Yes | Yes | Full | Yes | Sharding |
| `sharding-coordinator` | Yes | Yes | Full | Yes | Sharding coordination |
| `load` | Yes | Yes | Full | Yes | Load balancing |
| `performance` | Yes | Yes | Full | Yes | Performance metrics |
| `mint-authority` | Yes | Yes | Full | Yes | Mint authority |
| `validator-economy` | Yes | Yes | Full | Yes | Validator economics |
| `vm` | Yes | Yes | Full | Yes | VM interface |

## Future AVM Standards (Not Native Modules)

AVM contract standards are deployed as contracts.

## Launch Support (KV-Backed Calculation Surface)

| Module | State | Notes |
|--------|-------|-------|
| `aetra-economics` | KV-backed runtime | Governance-owned economic policy calculations and query surface |
| `aetra-staking-policy` | KV-backed runtime | Pool/validator allocation policy calculations and query surface |
| `aetra-validator-score` | KV-backed runtime | Deterministic validator score calculation surface |

**Note**: These modules are wired into app as `launch_support` calculation/query surfaces. Historical in-memory prototype wording is obsolete; current inventory marks them with `owns_consensus_state: true` and `kv_backed_runtime_mutations: true`.

## Prototype Only Modules (Not Wired)

| Module | State | Notes |
|--------|-------|-------|
| `epoch` | Prototype | Prototype epoch helper; launch uses SDK epochs module |
| `events` | Prototype | Shared event helper types only |
| `indexer` | Prototype | Off-chain/indexing projection helpers only |
| `pos` | Prototype | PoS specification helpers; launch uses SDK staking |
| `services` | Prototype | Prototype service-interface module |
| `taskgroups` | Prototype | Prototype task-group helper package |
| `validator-economy` | Prototype | Validator economy helper/spec package |

## Disabled Modules (Not Wired)

| Module | Status | Notes |
|--------|--------|-------|


## App Wiring Test

Run the launch profile validation:
```bash
go test ./docs/... -run TestLaunchModuleInventory
```

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-06-09 | Initial launch module inventory |
