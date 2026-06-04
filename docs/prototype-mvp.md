# Orbitalis Prototype MVP

This document defines the minimum contract for calling Orbitalis a working L1
prototype. It is intentionally narrower than a testnet or mainnet launch plan:
the goal is a buildable, runnable, observable local chain with representative
transactions across the custom modules and explicit gates for consensus,
security, and operator recovery.

## MVP Definition

Orbitalis is considered a working prototype when a new engineer can:

- build `orbitalisd`;
- initialize and start the local validator network;
- observe block production through CometBFT RPC, REST, and gRPC;
- submit and verify a native bank transfer;
- create, mint, burn, and query a tokenfactory denom;
- create and query a DEX pool;
- execute and verify a DEX swap;
- query fee policy params and prove the ante handler accepts only `norb` fees;
- stop and reset the local network without leaving committed artifacts in git;
- run the required build, test, proto, smoke, and security checks listed below.

The prototype is not mainnet-ready. It is an acceptance baseline for engineering
iterations and prototype releases.

## Terminology

- `prototype`: local or internal build proving the L1 can run, accept
  representative transactions, and expose state through operator interfaces.
- `localnet`: local validator network created under `.localnet` by
  `scripts/localnet/*.ps1`; the default supported profile is 3 validators.
- `testnet`: public or semi-public network with external validators, faucet,
  governance process, persistent infrastructure, and upgrade policy. This is
  out of scope for the prototype MVP unless a separate testnet plan says
  otherwise.
- `mainnet-ready`: production security, economics, monitoring, validator
  onboarding, incident response, and upgrade process. This is out of scope.

The `orbitalisd testnet` command is a Cosmos SDK local testnet helper. In this
repository, `scripts/localnet` is the supported operator entrypoint for the MVP.

## Supported Scope

- Cosmos SDK `v0.54.3` and CometBFT `v0.39.3`.
- Binary name: `orbitalisd`.
- Chain ID: `orbitalis-local-1` for the supported localnet profile.
- Account prefix: `orb`.
- Validator prefix: `orbvaloper`.
- Consensus prefix: `orbvalcons`.
- Native base denom: `norb`.
- Display denom: `ORB`.
- Token exponent: `1 ORB = 1,000,000,000 norb`.
- Custom modules: `x/tokenfactory`, `x/dex`, `x/fees`.
- Core modules required for the prototype: auth, bank, staking, slashing,
  distribution, mint, gov, feegrant, authz, evidence, upgrade, consensus,
  epochs, protocolpool.

## Out Of Scope

- Public validator onboarding.
- Production governance economics.
- IBC and external bridge support.
- Exchange-grade DEX features such as routing, order books, MEV controls, price
  oracles, external indexers, or market-maker tooling.
- Production observability stack, explorer, faucet, and alerting.
- High-TPS claims or benchmark marketing targets.
- Production key custody or hardware security module integration.
- Redis, PostgreSQL, or other external databases for consensus, mempool, or
  state.

## Operator Prerequisites

- Windows PowerShell for the current localnet scripts.
- Go `1.25.x` on `PATH`, or the repository-local toolchain at
  `.work\tools\go1.25.11\go\bin`.
- `buf` available on `PATH` for proto linting.
- Security tools available on `PATH` for release checks: `govulncheck`,
  `gosec`, and `gitleaks`.
- Free local ports for node0: P2P `26656`, RPC `26657`, gRPC `9090`, REST
  `1317`.
- Free local ports for node1: P2P `26756`, RPC `26757`, gRPC `9091`, REST
  `1318`.
- Free local ports for node2: P2P `26856`, RPC `26857`, gRPC `9092`, REST
  `1319`.

Minimum local hardware assumptions:

- 1-node dev profile: 2 CPU cores, 4 GB RAM, 2 GB free disk.
- 3-validator BFT smoke profile: 4 CPU cores, 8 GB RAM, 5 GB free disk.
- 5-validator config-scaling profile: 4-8 CPU cores, 16 GB RAM, 10 GB free disk.

The MVP does not require dedicated hardware. It must not hang or require manual
process cleanup in the supported 3-validator profile.

## Network Profiles

| Profile | Purpose | Acceptance |
| --- | --- | --- |
| 1 validator | Fast local development | The binary can initialize and produce blocks with one validator. If no script exists, the command sequence must be documented before prototype release. |
| 3 validators | Default BFT smoke | `scripts/localnet/init.ps1`, `start.ps1`, `stop.ps1`, `reset.ps1`, and `tests/e2e/localnet_smoke.ps1` pass. |
| 5 validators | Config-scaling check | The testnet init path can generate five validators and the network reaches height 3 without port or config collisions. Script support may be added in a later PR. |

## Test Accounts

The supported localnet creates validator keys named `node0`, `node1`, and
`node2` in test keyrings under each node home:

```powershell
$NODE0_HOME = ".localnet\node0\orbitalisd"
$NODE1_HOME = ".localnet\node1\orbitalisd"
$NODE2_HOME = ".localnet\node2\orbitalisd"

$NODE0_ADDR = (& build\orbitalisd.exe keys show node0 -a --home $NODE0_HOME --keyring-backend test)
$NODE1_ADDR = (& build\orbitalisd.exe keys show node1 -a --home $NODE1_HOME --keyring-backend test)
$NODE2_ADDR = (& build\orbitalisd.exe keys show node2 -a --home $NODE2_HOME --keyring-backend test)
```

Localnet key material is for development only. Files under `.localnet`, node
homes, key seeds, process IDs, and logs must never be committed.

## Endpoints

Default node0 endpoints:

- CometBFT RPC: `tcp://127.0.0.1:26657`
- REST: `http://127.0.0.1:1317`
- gRPC: `127.0.0.1:9090`

The localnet uses `--commit-timeout 1s`. Expected local block time is roughly
1-3 seconds on a healthy developer machine.

## Common Command Variables

Run the flow commands from the repository root:

```powershell
$BINARY = "build\orbitalisd.exe"
$CHAIN_ID = "orbitalis-local-1"
$NODE = "tcp://127.0.0.1:26657"
$NODE0_HOME = ".localnet\node0\orbitalisd"
$NODE1_HOME = ".localnet\node1\orbitalisd"
$NODE0_ADDR = (& $BINARY keys show node0 -a --home $NODE0_HOME --keyring-backend test)
$NODE1_ADDR = (& $BINARY keys show node1 -a --home $NODE1_HOME --keyring-backend test)
$TX_FLAGS = @("--home", $NODE0_HOME, "--node", $NODE, "--chain-id", $CHAIN_ID, "--keyring-backend", "test", "--fees", "1000000norb", "--yes")
$QUERY_FLAGS = @("--node", $NODE, "--output", "json")
```

## Prototype Flows

### Build `orbitalisd`

Command:

```powershell
go build -o build\orbitalisd.exe ./cmd/l1d
```

Expected result:

- `build\orbitalisd.exe` exists.
- `build\orbitalisd.exe version` exits successfully.

Recovery if it fails:

- Run `go env GOVERSION` and confirm Go `1.25.x`.
- Run `go mod verify`.
- If generated protobuf types are stale, run `buf generate` and review the diff
  before committing generated code.

### Initialize Localnet

Command:

```powershell
.\scripts\localnet\init.ps1
```

Expected result:

- `.localnet` is recreated.
- Three node homes exist under `.localnet\node0`, `.localnet\node1`, and
  `.localnet\node2`.
- The script prints the node0, node1, and node2 ports.
- Genesis chain ID is `orbitalis-local-1`.

Recovery if it fails:

- Stop stale processes with `.\scripts\localnet\stop.ps1`.
- Remove stale data with `.\scripts\localnet\reset.ps1`.
- Check that ports `26656`, `26657`, `26756`, `26757`, `26856`, and `26857`
  are free.

### Start Localnet

Command:

```powershell
.\scripts\localnet\start.ps1
```

Expected result:

- Three hidden `orbitalisd start` processes are launched.
- PID files are written under `.localnet\pids`.
- Logs are written under `.localnet\logs`.
- Node0 RPC responds on `tcp://127.0.0.1:26657`.

Recovery if it fails:

- Inspect `.localnet\logs\node*.err.log`.
- Confirm `minimum-gas-prices = "0norb"` in each node `app.toml`.
- Run `.\scripts\localnet\reset.ps1` and initialize again if config files are
  partially written.

### Query Latest Block

Command:

```powershell
& $BINARY query block --node $NODE --output json
```

Expected result:

- JSON output contains a block header and a positive height.
- Re-running the command after a few seconds returns a higher height.

Recovery if it fails:

- Confirm localnet processes are running.
- Inspect node0 stderr log.
- Restart with `.\scripts\localnet\stop.ps1` and `.\scripts\localnet\start.ps1`.

### Observe REST And gRPC

REST command:

```powershell
Invoke-RestMethod "http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/blocks/latest"
```

gRPC command, if `grpcurl` is installed:

```powershell
grpcurl -plaintext 127.0.0.1:9090 cosmos.base.tendermint.v1beta1.Service.GetLatestBlock
```

Expected result:

- REST returns latest block data.
- gRPC returns latest block data or a typed service response.

Recovery if it fails:

- Confirm REST and gRPC are enabled in `.localnet\node0\orbitalisd\config\app.toml`.
- Verify ports `1317` and `9090` are not occupied by another process.
- Use CometBFT RPC as the primary liveness signal while debugging.

### Send Native Bank Transaction

Command:

```powershell
& $BINARY tx bank send node0 $NODE1_ADDR 1000000000norb @TX_FLAGS
& $BINARY query bank balances $NODE1_ADDR @QUERY_FLAGS
```

Expected result:

- The tx command returns a transaction hash and code `0`.
- Node1 balance includes the transferred `norb`.

Recovery if it fails:

- Confirm `$NODE1_ADDR` was read from the node1 keyring.
- Confirm node0 has enough `norb` with
  `& $BINARY query bank balances $NODE0_ADDR @QUERY_FLAGS`.
- Check fee denom: the MVP accepts `norb` only.

### Create Tokenfactory Denom

Command:

```powershell
& $BINARY tx tokenfactory create-denom gold --from node0 @TX_FLAGS
$FACTORY_DENOM = "factory/$NODE0_ADDR/gold"
& $BINARY query tokenfactory denom $FACTORY_DENOM @QUERY_FLAGS
```

Expected result:

- The tx returns code `0`.
- The denom query returns `factory/$NODE0_ADDR/gold` with admin `$NODE0_ADDR`.

Recovery if it fails:

- Confirm `--from node0` uses the same home as `$NODE0_HOME`.
- Use a new subdenom if `gold` already exists after a previous run.
- Check address prefix and denom length validation errors.

### Mint And Burn Tokenfactory Supply

Command:

```powershell
& $BINARY tx tokenfactory mint "200000000$FACTORY_DENOM" $NODE0_ADDR --from node0 @TX_FLAGS
& $BINARY query bank balances $NODE0_ADDR @QUERY_FLAGS
& $BINARY tx tokenfactory burn "1000000$FACTORY_DENOM" $NODE0_ADDR --from node0 @TX_FLAGS
& $BINARY query bank balances $NODE0_ADDR @QUERY_FLAGS
```

Expected result:

- Mint tx returns code `0`.
- Node0 balance includes the factory denom.
- Burn tx returns code `0`.
- Node0 factory balance decreases by the burned amount.

Recovery if it fails:

- Confirm `$FACTORY_DENOM` matches the creator address exactly.
- Confirm node0 is the denom admin.
- For burn, `burn_from_address` must match the sender in v1.
- Confirm the amount is positive and uses the full factory denom.

### Create DEX Pool

Command:

```powershell
& $BINARY tx dex create-pool "100000000$FACTORY_DENOM" 100000000norb --from node0 @TX_FLAGS
& $BINARY query dex pool 1 @QUERY_FLAGS
```

Expected result:

- The tx returns code `0`.
- Pool `1` exists.
- The pool has two canonical denoms, positive reserves, positive total shares,
  and LP denom `lp/1`.

Recovery if it fails:

- Confirm node0 has enough `norb` and factory balance.
- Confirm the pool uses two different denoms.
- Use `query dex pools` if pool ID `1` was already created in the current
  localnet state.

### Swap Through DEX Pool

Command:

```powershell
& $BINARY tx dex swap-exact-in 1 "1000000$FACTORY_DENOM" norb 1 --from node0 @TX_FLAGS
& $BINARY query dex pool 1 @QUERY_FLAGS
& $BINARY query bank balances $NODE0_ADDR @QUERY_FLAGS
```

Expected result:

- The tx returns code `0`.
- Pool reserves change deterministically.
- Node0 receives a positive amount of `norb`.

Recovery if it fails:

- Confirm pool `1` contains `$FACTORY_DENOM` and `norb`.
- Lower `min-out` only for local debugging; production flows must use a
  slippage-aware value.
- Query the pool before and after the swap to check reserve direction.

### Query Fee Params

Command:

```powershell
& $BINARY query fees params @QUERY_FLAGS
```

Expected result:

- Output shows `norb` as the only allowed fee denom.
- Validator/community fee split params are present and sum to `1`.

Recovery if it fails:

- Confirm the `fees` module was included in genesis.
- Confirm node0 RPC is reachable.
- Check `x/fees/types/genesis.go` default params if genesis export is invalid.

### Stop And Reset Localnet

Commands:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
```

Expected result:

- `stop.ps1` removes PID files and stops node processes.
- `reset.ps1` removes `.localnet`.

Recovery if it fails:

- Manually inspect remaining `orbitalisd` processes.
- Remove `.localnet` only after confirming no localnet process still owns it.
- Verify `.localnet` remains ignored by git.

## Required Checks

Run these before a prototype release PR is pushed:

```powershell
go test ./...
go vet ./...
buf lint
go build -o build\orbitalisd.exe ./cmd/l1d
.\tests\e2e\localnet_smoke.ps1
```

Security baseline:

```powershell
go mod verify
New-Item -ItemType Directory -Force .work\security | Out-Null
govulncheck -scan=package -format json ./... > .work\security\govulncheck.json
go run .\scripts\security\govulncheck_triage.go -triage security\govulncheck-triage.json < .work\security\govulncheck.json
gosec -conf .gosec.json -exclude-generated -severity high -confidence medium ./...
gitleaks detect --source . --config .gitleaks.toml --redact --no-banner --verbose
```

Custom module tx coverage:

- `x/tokenfactory`: create denom, mint, burn, query denom.
- `x/dex`: create pool, query pool, swap, query updated pool and account
  balances.
- `x/fees`: query params and submit at least one successful tx using `norb`
  fees; submit or simulate one negative-path tx using a non-`norb` fee denom
  when the test harness can do so without polluting localnet state.

## Consensus And Security Gates

These gates are mandatory for prototype acceptance:

| Area | Minimum audit gate |
| --- | --- |
| Genesis | Genesis validation must reject malformed module state; exported genesis must preserve bank supply, staking state, tokenfactory denoms, DEX pools, and fees params. |
| BeginBlocker / EndBlocker | No nondeterministic state writes, unbounded iteration over attacker-controlled state, goroutines, wall-clock reads, random values, floating-point math, or panic-prone unchecked operations in consensus paths. |
| Vote extensions | If vote extensions are enabled, `VerifyVoteExtension` must be deterministic, bounded, and panic-free. The current dummy random payload is acceptable only as prototype-local behavior and is MUST FIX before public testnet or mainnet readiness. |
| Ante handler | Fee validation must be deterministic and must reject unsupported fee denoms before message execution. |
| Bank supply | Tokenfactory and DEX mint/burn paths must check all bank keeper errors and keep custom bookkeeping synchronized with `x/bank`. |
| Staking power | Validator genesis, staking denom, and power reduction must be deterministic and match `norb` policy. |
| DEX reserves | Pool reserves and LP shares must remain positive, canonical, and fully backed after create-pool, add-liquidity, remove-liquidity, swap, genesis import, and export. |
| Tokenfactory admin rights | Only denom admins may mint, burn, or transfer admin rights; malformed admin addresses and denoms must fail cleanly. |
| Malformed input | Msg, Query, and Genesis handlers must return errors for malformed input instead of panicking. |
| Secrets and logs | Mnemonics, key seeds, private validator keys, node keys, and localnet logs must not be committed or printed in prototype release artifacts. |

Manual review must include the Cosmos-specific sweep:

- No map iteration whose order affects state transitions.
- No platform-dependent `int`, `uint`, `float32`, or `float64` arithmetic in
  consensus-critical accounting.
- No `time.Now`, `rand`, goroutines, or multi-channel `select` in message
  handlers, BeginBlocker, EndBlocker, InitGenesis, or ExportGenesis.
- Any vote extension handler must be separately reviewed because it can affect
  consensus even when it does not write application state.
- All bank keeper sends, mints, burns, and module-account transfers check and
  propagate errors.
- All Msg signers match the authority checked by the keeper.

## Blocker Classification

MUST FIX before a prototype release:

- `orbitalisd` does not build.
- `go test ./...`, `go vet ./...`, or `buf lint` fails without documented,
  approved triage.
- The 3-validator localnet cannot reach height 3 or cannot restart cleanly.
- Any required custom-module tx flow fails.
- RPC, REST, and gRPC are not observable on node0.
- Any consensus nondeterminism, ABCI panic, unauthorized mint/burn, bank supply
  mismatch, DEX reserve mismatch, staking power mismatch, or secret leak is
  found.
- `STEP.md`, `STEP_V2.md`, localnet data, keys, logs, or security raw outputs
  are staged for commit.

SHOULD FIX before a prototype release:

- Missing command examples for supported flows.
- Missing negative-path tests for malformed denoms, addresses, pool IDs, fee
  denoms, or slippage.
- Incomplete 1-validator or 5-validator operator scripts.
- Non-blocking security tooling triage gaps with known owner and issue.
- Documentation terminology drift between prototype, localnet, testnet, and
  mainnet-ready.

NICE TO HAVE after MVP acceptance:

- Explorer, faucet, dashboards, and public status page.
- Persistent non-local infrastructure automation.
- Rich performance benchmarks and load tests.
- IBC, bridge, and cross-chain relayer plans.
- Advanced DEX routing, price impact UI, or market-maker tooling.

## GitHub Workflow

Implement MVP-definition changes in branch `docs/prototype-mvp-definition`.
Before pushing:

- Stage only tracked MVP docs and explicitly related small fixes.
- Do not stage `STEP.md` or `STEP_V2.md`.
- Do not stage `.localnet`, node homes, logs, key material, or `.work` outputs.
- Run the required checks or document any unavailable local tool.
- Include the security checklist above in the PR description.

The PR is acceptable when a reviewer can run the documented commands and decide
whether each result is successful, recoverable, or a release blocker.
