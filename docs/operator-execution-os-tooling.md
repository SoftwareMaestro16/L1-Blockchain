# Aetra Execution OS Operator Tooling

Phase 10 exposes modular execution prototypes through localnet profiles and
off-chain simulator diagnostics. These profiles are for local/testnet operator
validation only. They do not make production sharding, Mesh settlement, or
`.aet` identity keeper state live on mainnet.

## Profiles

| Profile | Purpose | Prototype gates |
| --- | --- | --- |
| `base` | Current chain only | none |
| `execution-os-sim` | Pure simulator commands plus load/routing genesis metadata | `load`, `routing` |
| `zones-prototype` | Feature-gated zones prototype metadata | `load`, `routing`, `zones` |
| `mesh-prototype` | Feature-gated Mesh prototype metadata | `load`, `routing`, `zones`, `mesh` |
| `identity-prototype` | `.aet` executable spec readiness | `load`, `routing`, `zones`, `identity-spec` |

## CLI

```powershell
build\aetrad.exe execution-os profiles
build\aetrad.exe execution-os smoke --profile execution-os-sim
build\aetrad.exe execution-os diagnostics --profile zones-prototype --genesis .localnet\node0\aetrad\config\genesis.json
```

The smoke command runs deterministic executable specs for:

- load score update;
- tx routing to an execution zone;
- load-driven shard activation;
- cross-zone Mesh message and receipt processing;
- `.aet` domain registration and resolution.

## Localnet

```powershell
.\scripts\localnet\init.ps1 -Profile zones-prototype
.\scripts\localnet\validate-genesis.ps1 -Profile zones-prototype
.\scripts\localnet\start.ps1 -Profile zones-prototype -NoInit
.\scripts\localnet\execution-os-diagnostics.ps1 -Profile zones-prototype -Json
```

The profile is written to `.localnet\profile.json`. Prototype profile data is
in genesis only and remains testnet-gated through `Enabled=true` and
`TestnetProfile=true`; `ProductionVersionGate` stays empty.

## Diagnostics

Diagnostics expose bounded operator state:

- current load score;
- active zones;
- active shard counts by zone;
- pending Mesh message count;
- replay marker count;
- Mesh receipt count;
- zone commitment roots.

Diagnostics do not copy keyring data, mnemonics, private validator keys,
`priv_validator_state.json`, node keys, raw secrets, or local key material.
`scripts/localnet/diagnostics.ps1` writes redacted bundles and includes
`execution-os.json` when the binary is available.

## Smoke Tests

```powershell
.\tests\e2e\execution_os_smoke.ps1 -SkipBuild
.\tests\e2e\zones_smoke.ps1 -SkipBuild
.\tests\e2e\mesh_smoke.ps1 -SkipBuild
.\tests\e2e\identity_smoke.ps1 -SkipBuild
```

Each smoke initializes a profile-specific localnet, validates genesis, starts
validators, reads execution OS diagnostics, restarts the localnet, and verifies
that profile diagnostics remain stable across restart.
