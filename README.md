# Aetra Blockchain

Aetra is a sovereign Cosmos SDK Layer 1 blockchain implemented in Go — a deterministic, account-based PoS chain with an embedded Aetra Virtual Machine (AVM) for smart contracts. Built for moderate hardware, pool-based staking, and governance-controlled economics.

| Property | Value |
|----------|-------|
| Native asset | **AET** (1 AET = 10⁹ naet) |
| Consensus | CometBFT (2–5s blocks) |
| VM | AVM v1 — stack-based, typed, deterministic |
| Staking | Pool-based, no direct user→validator choice |
| Fee target | ~0.01 AET per transfer (governance-adjustable) |
| Address format | User: `AE...` / Raw: `4:...` / Protocol: `-7:...` |

## Quick Start

```powershell
.\scripts\build-aetrad.ps1
.\scripts\localnet\init.ps1 -ChainId aetra-local-1 -ValidatorCount 3
.\scripts\localnet\start.ps1 -ChainId aetra-local-1
```

No Redis, no PostgreSQL — just the binary and CometBFT.

---

## Key Subsystems

### 1. Transaction Lifecycle

```mermaid
flowchart LR
  CLIENT["CLI / REST / gRPC / RPC"]
  CLIENT --> MEMPOOL["CometBFT mempool"]
  MEMPOOL --> PROPOSAL["Block proposal"]
  PROPOSAL --> PREBLOCK["PreBlock: upgrade + auth"]
  PREBLOCK --> ANTE["Ante: signatures, sequence, fee payer<br/>reserved address checks"]
  ANTE --> FEES["x/fees: naet fee admission<br/>dynamic formula * reputation<br/>storage rent side effects"]
  FEES --> ROUTER["SDK message router"]
  ROUTER --> SDK["SDK base: auth, bank, staking,<br/>slashing, distribution, mint, gov"]
  ROUTER --> AETRA["Aetra modules: economy,<br/>validator systems, AVM,<br/>identity-root, config"]
  SDK --> KV["KVStore commit"]
  AETRA --> KV
  KV --> EXPORT["Genesis export / import / restart"]
```

### 2. Staking & Validator Set

Users never choose a validator. All deposits go into the official nominator pool, which allocates to validators by deterministic weights:

```mermaid
flowchart TD
  USER["AE... wallet"]
  USER -- "deposit ≥10 AET" --> POOL["Nominator Pool<br/>(x/nominator-pool)"]
  POOL --> SHARES["Pool shares minted"]
  POOL --> ALLOC["Allocation Engine<br/>weighted by reputation<br/>+ performance + stake"]
  ALLOC --> VAL1["Validator A<br/>(score-weighted %)"]
  ALLOC --> VAL2["Validator B<br/>(score-weighted %)"]
  ALLOC --> VAL3["Validator C<br/>(score-weighted %)"]
  VAL1 --> CONSENSUS["CometBFT validator set<br/>(x/validator-election)"]
  VAL2 --> CONSENSUS
  VAL3 --> CONSENSUS
  CONSENSUS --> REWARDS["Block rewards + fees"]
  REWARDS --> POOL
  POOL --> USER_SHARE["User withdraws via pool shares"]
```

- `MsgDepositToStakingPool` has no validator field — rejected at validation
- `MsgDelegate` disabled for normal user path
- Validator selection via `ValidatorScore`, `DynamicCommission`, `StakeConcentration`

### 3. Fee Economy

Every transaction pays a deterministic fee that splits across protocol buckets:

```mermaid
flowchart LR
  TX["Transaction execution"] --> FEE["Required fee = base + gas + bytes + msgs<br/>+ congestion surcharge + reputation premium<br/>- reputation discount + storage rent"]
  FEE --> COLLECTOR["Fee Collector<br/>(x/fee-collector)"]
  COLLECTOR --> BURN["Burn (x/burn)<br/>supply reduction"]
  COLLECTOR --> VAL_REWARDS["Validator rewards<br/>(x/distribution)"]
  COLLECTOR --> TREASURY["Treasury (x/treasury)<br/>community funds"]
  COLLECTOR --> RESERVES["Reserves: delegator protection,<br/>validator insurance, storage rent"]
  EMISSIONS["Emissions policy<br/>(x/emissions)"] --> VAL_REWARDS
```

Fee formula parameters — including `min_tx_fee_naet`, `base_transfer_fee_naet`, `target_transfer_fee_naet`, `low_reputation_premium_cap_naet`, `high_reputation_discount_cap_naet`, congestion thresholds — are governance/genesis params. Neutral transfer target: `0.01 AET`.

### 4. AVM Smart Contract Execution

```mermaid
flowchart TD
  USER["User / Contract"] --> DEPLOY["Deploy: bytecode → ChunkMap<br/>storage rent check"]
  DEPLOY --> EXECUTE["Execute: stack VM + typed values<br/>gas metering per instruction"]
  EXECUTE --> STORAGE["Storage Phase: load state Chunks"]
  STORAGE --> COMPUTE["Compute Phase: VM execution<br/>stack ops, host calls"]
  COMPUTE --> ACTIONS["Action Phase: emit messages,<br/>events, receipts, proofs"]
  ACTIONS --> COMMIT["Finalization: new Chunk roots<br/>state root, receipt hash"]
  COMMIT --> RECEIPT["Receipt: exit code, gas used,<br/>state root, events hash"]
```

- Content-addressed immutable Chunks (≤2048 data bits, ≤8 refs)
- Typed values: uint/int 8–256, address, hash, coins, tuple, Chunk
- Deterministic: same code/state/message → same exit code, gas, receipt, root
- Get methods are read-only, no state mutation
- Storage rent enforced before execution

### 5. Reputation System

Identity reputation is a single unified score fed by on-chain signals, influencing fee premiums/discounts and allocation priority — but never blocking basic rights:

```mermaid
flowchart LR
  SIGNALS["Signal sources:<br/>stake time, tx success,<br/>contract interactions,<br/>spam events, lifecycles"]
  SIGNALS --> SCORE["IdentityReputation<br/>0–10000 bps<br/>neutral = 5000"]
  SCORE --> FEES["x/fees: bounded premium/discount<br/>Low score → extra naet fee<br/>High score → bounded discount"]
  SCORE --> PRIORITY["Queue priority boost<br/>Higher score → earlier execution"]
  SCORE --> ALLOCATION["Pool allocation weight<br/>Higher score → better yield"]
```

- Reputation is a soft weighting signal, not a permission gate
- Low reputation cannot block transfers, staking, token/NFT creation, or contract deploy
- All effect caps are governance params

---

## Addresses

- **User-friendly**: `AE...` (Bech32-like, user-facing everywhere)
- **Raw internal**: `4:<64 hex chars>` (256-bit high-entropy)
- **Protocol core**: `-7:<64 hex chars>` (non-receivable system addresses)
- Zero addresses rejected by default

Key system accounts: `AETMint`, `AETBurn`, `AETFeeCollector`, `AETTreasury`, `AETStorageRent`, `AETDelegatorProtection`, `AETValidatorInsurance`, `AETReporterRewards`.

---

## Native Modules

| Category | Modules |
|----------|---------|
| SDK base | `auth`, `bank`, `staking`, `slashing`, `evidence`, `distribution`, `mint`, `gov`, `upgrade`, `consensus`, `epochs`, `authz`, `feegrant` |
| Config & authority | `x/config`, `x/config-voting`, `x/constitution`, `x/system-registry` |
| Economy | `x/fees`, `x/fee-collector`, `x/treasury`, `x/burn`, `x/emissions`, `x/mint-authority` |
| Validator systems | `x/validator-registry`, `x/validator-election`, `x/nominator-pool`, `x/validator-insurance`, `x/delegator-protection`, `x/reputation`, `x/performance`, `x/dynamic-commission`, `x/stake-concentration` |
| Execution | `x/scheduler`, `x/avm-scheduler`, `x/actor-registry`, `x/storage-rent` |
| Identity | `x/identity-root` |
| AVM | `x/aetravm`, `x/contracts`, `x/vm` |
| Cross-chain | `x/bridge-hub`, `x/cross-chain-registry`, `x/sharding-coordinator` |

No native token/NFT/DEX modules — application assets belong in AVM contracts (AFT-44, ANFT-66).

---

## Build & Run

```powershell
# Build
.\scripts\build-aetrad.ps1          # → build\aetrad.exe

# Local 3-validator network
.\scripts\localnet\init.ps1 -ChainId aetra-local-1 -ValidatorCount 3
.\scripts\localnet\start.ps1 -ChainId aetra-local-1

# Validate genesis
.\scripts\localnet\validate-genesis.ps1

# Monitor health
.\scripts\localnet\health.ps1
.\scripts\localnet\wait-height.ps1 -Height 10

# Export & restart
.\scripts\localnet\export-genesis.ps1 -Output genesis-export.json
.\scripts\localnet\reset.ps1
.\scripts\localnet\init.ps1 -ChainId aetra-local-1 -ValidatorCount 3
.\scripts\localnet\start.ps1 -ChainId aetra-local-1
```

Also available: `scripts/localnet/diagnostics.ps1`, `statesync.ps1`, `snapshot.ps1`, `stress-profile.ps1`.

---

## Common Commands

```powershell
build\aetrad.exe version --long --output json
build\aetrad.exe status --node tcp://127.0.0.1:26657
build\aetrad.exe query block --node tcp://127.0.0.1:26657
build\aetrad.exe query bank total-supply-of naet --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query staking validators --node tcp://127.0.0.1:26657 --output json
build\aetrad.exe query fees params --grpc-addr 127.0.0.1:9090 --grpc-insecure --output json
```

---

## Validator Info

For operator guides see [docs/VALIDATOR.md](docs/VALIDATOR.md), [docs/TESTNET.md](docs/TESTNET.md), and [docs/COSMOVISOR.md](docs/COSMOVISOR.md).

---

## Token

| Field | Value |
|-------|-------|
| Name | Aetra |
| Symbol | AET |
| Base denom | `naet` |
| Conversion | `1 AET = 1,000,000,000 naet` |
| Staking denom | `naet` |
| Fee denom | `naet` |
| Supply | Governance-capped emissions + validator rewards |

---

## Security

Deterministic genesis validation, export/import roundtrip tests, zero-address rejection, reserved system address checks, native fee validation, bounded dynamic fees, reputation-based fee adjustments, module-account wiring invariants, blocked-address policy, and localnet smoke tests.
