> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Aetra Local Bootstrap Profile

This document defines the reproducible local bootstrap contract for the prototype profile `aetra-local-1`.

The profile is structurally reproducible: the same commands create the same chain-id, validator count, account layout, balances, staking denom, custom module genesis, and endpoint layout. Validator keys, account addresses, gentx signatures, and `genesis_time` are generated per run and are not expected to be byte-identical across fresh initializations. Within one initialization run, every node must receive the same `genesis.json` hash.

## Profile

- Chain ID: `aetra-local-1`
- Base denom: `naet`
- Display denom: `AET`
- User-facing address prefix: `AE` for account, validator, and consensus addresses
- Default validator count: `3`
- Supported scale smoke profiles: `1`, `3`, and `5` validators
- Node directories: `.localnet\node<N>\aetrad`
- Test accounts: one generated key per validator, named `node0`, `node1`, ...
- Keyring backend: `test`
- Minimum gas prices: `0naet`
- Timeout commit: `1s`
- Default log level: `info`
- RPC, REST, and gRPC are enabled by default.

Each generated validator account starts with:

- `1000000000testtoken`
- `500000000naet`

`testtoken` is a local bootstrap test asset for module and DEX experiments only. It is not the native token, staking denom, fee denom, or display unit for `AET`.

Each validator gentx self-delegates:

- `100000000naet`

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

- `x/bank` includes native token metadata for base `naet`, display `AET`, exponent `9`.
- `x/staking` uses bond denom `naet`.
- `x/mint` uses mint denom `naet`.
- `x/genutil` contains one `MsgCreateValidator` gentx per validator.

Custom module expectations:

- `x/fees`: allowed fee denoms `["naet"]`, validator rewards ratio `0.98`, community pool ratio `0.02`.
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

For the executable working-prototype contract, including required flows, gaps, blocker classification, and evidence links, see [prototype-contract.md](prototype-contract.md). For the full operator CLI runbook, including reusable endpoint variables, common tx flags, and troubleshooting, see [operator-commands.md](operator-commands.md).

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
.\scripts\localnet\init.ps1 -ChainId aetra-local-1 -TimeoutCommit 1s -LogLevel info -BaseRPCPort 27657 -ValidatorCount 3
.\scripts\localnet\start.ps1 -BaseRPCPort 27657 -ValidatorCount 3 -Wait
```

Validate genesis and compare key fields:

```powershell
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\validate-genesis.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

The validation script checks the full local profile: chain ID, node count, node monikers, endpoint ports, minimum gas price, identical genesis hash across nodes, native metadata, staking/mint/fee denom, initial account balances, gentx self-delegations, empty tokenfactory state, and DEX `next_pool_id`.

If validation fails on an existing localnet after module genesis changes, reset and regenerate the ignored node homes:

```powershell
.\scripts\localnet\reset.ps1
.\scripts\localnet\init.ps1
.\scripts\localnet\validate-genesis.ps1
```

Equivalent single-node manual validation:

```powershell
build\aetrad.exe genesis validate-genesis .localnet\node0\aetrad\config\genesis.json --home .localnet\node0\aetrad
```

Start, stop, and reset:

```powershell
.\scripts\localnet\start.ps1
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
```

Run a smoke test against the default or 5-validator profile:

```powershell
.\tests\e2e\prototype_acceptance.ps1
.\tests\e2e\prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\localnet_smoke.ps1
.\tests\e2e\localnet_smoke.ps1 -ValidatorCount 5
```

The prototype acceptance suite is the default from-zero proof that build, genesis, localnet startup, block production, bank tx, fee policy, tokenfactory, DEX, PoS delegation, and query surfaces work together. See [prototype-acceptance-suite.md](prototype-acceptance-suite.md). The targeted localnet smoke validates RPC readiness, block height, CometBFT validator set size, peer count, REST `/blocks/latest`, gRPC TCP availability, `query block`, a `bank send` transaction, stop/start chain progress, and negative cases for invalid validator count, missing binary, timeout, and occupied port. The 5-validator run is the heavier profile and can be run with `-OutputDir .localnet-5` if the default 3-validator localnet should be preserved.

Run the query surface and health checks documented in [query-surface.md](query-surface.md) and [observability.md](observability.md):

```powershell
.\scripts\localnet\health.ps1 -ValidatorCount 3
.\tests\e2e\query_surface_smoke.ps1
```

Run the proof-of-stake smoke flow documented in [pos-smoke-flow.md](pos-smoke-flow.md):

```powershell
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

The PoS smoke validates staking params, bonded validators, delegation from a funded local account, delegation query state, slashing params, signing infos, and validator voting power updates.

Run the native token lifecycle smoke documented in [native-token-lifecycle.md](native-token-lifecycle.md):

```powershell
.\tests\e2e\native_token_smoke.ps1
.\tests\e2e\native_token_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

The native token smoke validates bank metadata, `naet` supply and balances, staking/fees/mint denom consistency, and a `bank send` transaction that pays fees in `naet`.

Run the tokenfactory lifecycle smoke documented in [tokenfactory-lifecycle.md](tokenfactory-lifecycle.md):

```powershell
.\tests\e2e\tokenfactory_smoke.ps1
.\tests\e2e\tokenfactory_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

The tokenfactory smoke validates create/query/mint/burn/change-admin, bank balance and supply consistency, REST denom query, native-spoof rejection, duplicate denom rejection, burn-from mismatch, invalid admin rejection, old-admin rejection, and new-admin mint.

Run the fees ante policy smoke documented in [fees-ante-policy.md](fees-ante-policy.md):

```powershell
.\tests\e2e\fees_ante_smoke.ps1
.\tests\e2e\fees_ante_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

The fees ante smoke validates fee params, successful `naet` fees across bank/tokenfactory/DEX txs, wrong fee rejection, mixed-denom fee rejection, and explicit zero/empty fee behavior.

Run the DEX prototype smoke documented in [dex-e2e-flow.md](dex-e2e-flow.md):

```powershell
.\tests\e2e\dex_smoke.ps1
.\tests\e2e\dex_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

The DEX smoke creates a factory asset, creates a pool, adds liquidity, swaps exact input, removes liquidity, checks LP balances/reserves, and validates slippage and wrong-denom failures.

Export state after the network has started:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\export-genesis.ps1
.\scripts\localnet\export-genesis.ps1 -OutputDir .localnet-5 -NodeIndex 0
```

`export-genesis.ps1` writes under ignored `.work\genesis`, validates the exported genesis with `aetrad genesis validate-genesis`, checks the expected chain ID, and refuses to export while the localnet process for that output directory is still running.

## Audit Notes

- Generated node homes are local artifacts and must stay out of git. Do not commit `.localnet`, `.localnet-*`, `key_seed.json`, keyring data, `priv_validator_key.json`, `priv_validator_state.json`, or `node_key.json`.
- `genesis.json` must not contain mnemonics, private validator keys, node keys, keyring secrets, wallet seeds, or private local paths.
- Localnet scripts must not print mnemonics by default. `testnet init-files` writes generated key seed files only under ignored node homes.
- `reset.ps1` and `init.ps1` verify the resolved output path is inside the repository and refuse to recursively delete the repository root or paths outside it.
- Initial stake is symmetric across validators in this prototype profile. Any non-symmetric validator distribution must be documented before use.
- Consensus-critical checks for this profile: deterministic app-state writes during init/export, valid `naet` bank supply and staking denom, positive validator self-delegations, authorized tokenfactory mint/burn only, DEX reserves absent at genesis, and no panics on malformed custom module genesis.

## Required Checks

```powershell
go test ./...
go vet ./...
buf lint
go build -o build/aetrad.exe ./cmd/l1d
.\scripts\localnet\validate-genesis.ps1
.\scripts\localnet\export-genesis.ps1
.\tests\e2e\prototype_acceptance.ps1
.\scripts\security\prototype-audit.ps1 -Profile Fast
.\scripts\release\prototype-package.ps1 -Version prototype-local -TargetOS windows -TargetArch amd64
.\tests\e2e\prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\localnet_smoke.ps1
.\tests\e2e\localnet_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\pos_smoke.ps1
.\tests\e2e\pos_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\native_token_smoke.ps1
.\tests\e2e\native_token_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\fees_ante_smoke.ps1
.\tests\e2e\fees_ante_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
.\tests\e2e\dex_smoke.ps1
.\tests\e2e\dex_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Security baseline:

```powershell
.\scripts\security\prototype-audit.ps1 -Profile Full
```
