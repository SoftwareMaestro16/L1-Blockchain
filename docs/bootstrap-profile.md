# Orbitalis Local Bootstrap Profile

This document defines the reproducible local bootstrap contract for the prototype profile `orbitalis-local-1`.

The profile is structurally reproducible: the same commands create the same chain-id, validator count, account layout, balances, staking denom, custom module genesis, and endpoint layout. Validator keys, account addresses, gentx signatures, and `genesis_time` are generated per run and are not expected to be byte-identical across fresh initializations. Within one initialization run, every node must receive the same `genesis.json` hash.

## Profile

- Chain ID: `orbitalis-local-1`
- Base denom: `norb`
- Display denom: `ORB`
- Address prefixes: account `orb`, validator `orbvaloper`, consensus `orbvalcons`
- Default validator count: `3`
- Supported scale smoke profiles: `1`, `3`, and `5` validators
- Node directories: `.localnet\node<N>\orbitalisd`
- Test accounts: one generated key per validator, named `node0`, `node1`, ...
- Keyring backend: `test`
- Minimum gas prices: `0norb`
- Timeout commit: `1s`
- Default log level: `info`
- RPC, REST, and gRPC are enabled by default.

Each generated validator account starts with:

- `1000000000testtoken`
- `500000000norb`

Each validator gentx self-delegates:

- `100000000norb`

The default endpoint layout is formula-based and can be shifted with `-BaseP2PPort`, `-BaseRPCPort`, `-BaseRESTPort`, `-BaseGRPCPort`, and `-PortStride`:

- P2P: `26656 + (100 * node_index)`
- RPC: `26657 + (100 * node_index)`
- REST: `1317 + node_index`
- gRPC: `9090 + node_index`

Examples:

- node0: P2P `26656`, RPC `26657`, REST `1317`, gRPC `9090`
- node1: P2P `26756`, RPC `26757`, REST `1318`, gRPC `9091`
- node2: P2P `26856`, RPC `26857`, REST `1319`, gRPC `9092`

## Module Genesis

SDK module expectations:

- `x/bank` includes native token metadata for base `norb`, display `ORB`, exponent `9`.
- `x/staking` uses bond denom `norb`.
- `x/mint` uses mint denom `norb`.
- `x/genutil` contains one `MsgCreateValidator` gentx per validator.

Custom module expectations:

- `x/fees`: allowed fee denoms `["norb"]`, validator rewards ratio `0.98`, community pool ratio `0.02`.
- `x/tokenfactory`: starts with no factory denoms.
- `x/dex`: starts with `next_pool_id = 1` and no pools.

Module account permissions are intentionally narrow:

- `mint`: `minter`
- `bonded_tokens_pool`: `burner`, `staking`
- `not_bonded_tokens_pool`: `burner`, `staking`
- `gov`: `burner`
- `tokenfactory`: `minter`, `burner`
- `dex`: `minter`, `burner`
- fee collector, distribution, protocol pool, protocol pool escrow, and fees module accounts have no special permissions.

## Commands

Initialize the default 3-validator localnet:

```powershell
.\scripts\localnet\init.ps1
```

Initialize a 5-validator profile without copying scripts:

```powershell
.\scripts\localnet\init.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Custom local profiles can override chain ID, timeout, log level, endpoint base ports, and endpoint toggles:

```powershell
.\scripts\localnet\init.ps1 -ChainId orbitalis-local-1 -TimeoutCommit 1s -LogLevel info -BaseRPCPort 27657 -ValidatorCount 3
.\scripts\localnet\start.ps1 -BaseRPCPort 27657 -ValidatorCount 3 -Wait
```

Validate genesis and compare key fields:

```powershell
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\validate-genesis.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

If validation fails on an existing localnet after module genesis changes, reset and regenerate the ignored node homes:

```powershell
.\scripts\localnet\reset.ps1
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
```

Equivalent single-node manual validation:

```powershell
build\orbitalisd.exe genesis validate-genesis .localnet\node0\orbitalisd\config\genesis.json --home .localnet\node0\orbitalisd
```

Start, stop, and reset:

```powershell
.\scripts\localnet\start.ps1
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
```

Run a smoke test against the default or 5-validator profile:

```powershell
.\tests\e2e\localnet_smoke.ps1
.\tests\e2e\localnet_smoke.ps1 -ValidatorCount 5
```

The smoke test validates RPC readiness, block height, CometBFT validator set size, peer count, REST `/blocks/latest`, gRPC TCP availability, `query block`, a `bank send` transaction, stop/start chain progress, and negative cases for invalid validator count, missing binary, timeout, and occupied port. The 5-validator run is the heavier profile and can be run with `-OutputDir .localnet-5` if the default 3-validator localnet should be preserved.

Run the proof-of-stake smoke flow documented in [pos-smoke-flow.md](pos-smoke-flow.md):

```powershell
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

The PoS smoke validates staking params, bonded validators, delegation from a funded local account, delegation query state, slashing params, signing infos, and validator voting power updates.

Export state after the network has started:

```powershell
New-Item -ItemType Directory -Force .work\genesis | Out-Null
cmd /c "build\orbitalisd.exe export --home .localnet\node0\orbitalisd > .work\genesis\node0-export.json"
build\orbitalisd.exe genesis validate-genesis .work\genesis\node0-export.json --home .localnet\node0\orbitalisd
```

## Audit Notes

- Generated node homes are local artifacts and must stay out of git. Do not commit `.localnet`, `.localnet-*`, `key_seed.json`, keyring data, `priv_validator_key.json`, `priv_validator_state.json`, or `node_key.json`.
- `genesis.json` must not contain mnemonics, private validator keys, node keys, keyring secrets, wallet seeds, or private local paths.
- Localnet scripts must not print mnemonics by default. `testnet init-files` writes generated key seed files only under ignored node homes.
- `reset.ps1` and `init.ps1` verify the resolved output path is inside the repository and refuse to recursively delete the repository root or paths outside it.
- Initial stake is symmetric across validators in this prototype profile. Any non-symmetric validator distribution must be documented before use.
- Consensus-critical checks for this profile: deterministic app-state writes during init/export, valid `norb` bank supply and staking denom, positive validator self-delegations, authorized tokenfactory mint/burn only, DEX reserves absent at genesis, and no panics on malformed custom module genesis.

## Required Checks

```powershell
go test ./...
go vet ./...
buf lint
go build -o build/orbitalisd.exe ./cmd/l1d
.\scripts\localnet\validate-genesis.ps1
.\tests\e2e\localnet_smoke.ps1
.\tests\e2e\localnet_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Security baseline:

```powershell
go mod verify
gitleaks detect --source . --config .gitleaks.toml --redact --no-banner --verbose
```
