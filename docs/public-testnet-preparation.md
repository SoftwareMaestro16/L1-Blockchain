> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Public Testnet Preparation

This runbook is the Phase 16 gate before opening Aetra to external validators. It is not mainnet readiness.

The full public testnet and production gate ledger is
[Public Testnet And Production Gates](public-testnet-production-gates.md).

## Profiles

Run all local profiles before publishing testnet genesis:

```powershell
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile All
```

Individual profiles:

```powershell
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 3 -SkipBuild
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 5 -SkipBuild
.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile 10 -SkipBuild
```

The preflight runs full prototype acceptance, validates the requested validator count, exercises bank, fees, staking, tokenfactory, DEX, query surfaces, restart persistence, and asserts CosmWasm remains disabled unless explicitly gated. The 10-validator profile is the stress profile for public testnet readiness; it is expected to be slower and should run before advertising modular execution features.

## Localnet Hardening

Public testnet prep depends on these script rules:

- localnet output directories must stay under the repository workspace,
- start refuses occupied P2P, RPC, REST, gRPC, and pprof ports,
- `-CleanLogs` is explicit when old logs should be removed,
- snapshot and state-sync scripts resolve paths through localnet helpers,
- diagnostics must not package keyrings, validator private keys, mnemonics, or environment secrets.

## Faucet Plan

Use an off-chain faucet for public testnet. Do not add a faucet mint path to the chain for v1.

Minimum faucet rules:

- faucet wallet is a normal account with capped prefunded `naet`,
- rate limit by address and IP,
- one request per address per cooldown window,
- max grant per request is documented before launch,
- faucet txs pay `naet` fees and use normal `bank send`,
- faucet private key is stored outside the repository in a secret manager,
- faucet logs never print private keys, mnemonics, or full environment dumps.

Initial operator command:

```powershell
build\aetherisd.exe tx bank send <faucet-key> <recipient-aetheris-address> 1000000naet --chain-id <testnet-chain-id> --fees 1000000naet --node <rpc-url> --keyring-backend <secure-backend> -y
```

## Explorer And Indexer Plan

Minimum public testnet explorer requirements:

- CometBFT RPC endpoint for block, tx, validator, and status views,
- REST or gRPC endpoint for bank, staking, fees, tokenfactory, and DEX queries,
- event indexing for bank sends, staking delegation, tokenfactory create/mint/burn/admin, DEX pool/liquidity/swap, and future wasm events,
- indexer database credentials kept outside repo config,
- indexer lag alert when latest indexed height falls behind node height by the launch threshold,
- no dependency on indexer availability for validator liveness.

The current event contract is documented in [Prototype Tx Event Contract](event-contract.md).

## Minimum Hardware

Development validator:

- 4 CPU cores,
- 8 GB RAM,
- 100 GB SSD,
- reliable broadband connection,
- Windows PowerShell for current scripts or Linux shell wrappers to be added before Linux-only operators join.

Public testnet validator:

- 4-8 CPU cores,
- 16 GB RAM,
- 200 GB SSD with growth monitoring,
- stable public P2P networking,
- separate monitoring host or process for alerting,
- time sync enabled.

Do not run public validators with localnet `--keyring-backend test` keys.

## Snapshot And State-Sync Plan

Snapshot creation on a trusted node:

```powershell
.\scripts\localnet\snapshot.ps1 -OutputDir .localnet -NodeIndex 0 -Height <height> -ArchivePath .work\snapshots\aetheris-testnet-<height>.tar
```

State sync configuration on a joining node:

```powershell
.\scripts\localnet\statesync.ps1 -OutputDir .localnet -TargetNodeIndex 2 -TrustHeight <height> -TrustHash <hash> -ResetData
```

Public testnet publishing requirements:

- publish snapshot height, hash, archive checksum, and source validator identity,
- publish at least two RPC servers for state sync,
- never publish validator private key files or keyrings in snapshots,
- keep one recent snapshot and one older fallback snapshot until the next upgrade.

## CosmWasm Test Contract

CosmWasm remains disabled by default. If a testnet config explicitly enables the wasm gate, run:

```powershell
.\tests\e2e\cosmwasm_smoke.ps1 -EnableWasm -ContractWasm .\artifacts\cw_template.wasm
```

The contract upload and instantiate flow is documented in [CosmWasm Readiness](security/cosmwasm-readiness.md).

If async contracts are enabled in a public testnet config, run the async
execution smoke and the AFT-44/ANFT-66/ASBT-67 standard tests before opening the
network:

```powershell
.\.work\tools\go1.25.11\go\bin\go.exe test ./x/aetherisvm/async ./x/aetherisvm/standards/...
```

## Rollback And Restart Procedure

Use restart only for operational failures where the chain state is valid:

1. Announce the target height, UTC window, affected RPC endpoints, and expected
   validator action.
2. Stop nonessential load generators, faucet jobs, and indexer writes.
3. Stop validators cleanly. Preserve `data`, `config\priv_validator_key.json`,
   `config\node_key.json`, genesis, and logs.
4. Apply only the announced config or binary change.
5. Restart one seed validator first, then the rest of the validator set.
6. Confirm block production, app hash agreement across at least three nodes,
   peer count, and public RPC health.
7. Resume faucet and indexer jobs after finalized blocks are visible.

Use rollback only when the new binary, config, genesis, snapshot, or state-sync
announcement is proven bad:

1. Freeze faucet/indexer writes and publish the rollback reason.
2. Restore the previous signed binary/config or corrected snapshot trust data.
3. Do not delete validator state unless the incident response owner explicitly
   requires a data reset and the validator has backed up evidence.
4. Run `go test ./...`, `go vet ./...`, `buf lint`, and the public-testnet
   preflight profile that matches the affected validator count before
   re-opening external joins.

## Launch Checklist

- `go test -p=1 ./...` passes.
- `go vet -p=1 ./...` passes.
- `buf lint` passes.
- Security scans pass or findings are triaged.
- Deterministic execution gate passes.
- `.\scripts\testnet\public-testnet-preflight.ps1 -ValidatorProfile All` passes.
- Genesis validates on every seed validator.
- At least one fresh validator follows [Validator Onboarding](validator-onboarding.md) and reaches the validator set.
- Faucet dry-run sends `naet` to a new address.
- Explorer/indexer follows node height and shows tx details.
- Snapshot and state-sync instructions are tested.
- Incident response contacts and severity rules are published.
- Faucet plan is implemented or explicitly deferred.
- Explorer/indexer plan is implemented or explicitly deferred.
- If CosmWasm or async contracts are enabled, simple contract deployment plus
  token/NFT/SBT smoke tests pass first.
