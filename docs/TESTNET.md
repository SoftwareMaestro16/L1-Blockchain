# Aetra Testnet Launch Scope

## Overview

This document defines the **testnet kernel**: the minimal, stable set of functionality that constitutes a runnable public testnet. Any functionality not listed here is out-of-scope for testnet launch and should be treated as future or prototype.

## Testnet Kernel

### 1. Core Blockchain Infrastructure
- **Cosmos SDK + CometBFT node** (`aetrad` binary)
  - Standard consensus, mempool, ABCI
  - Deterministic execution
  - Block lifecycle management

### 2. Address & Wallet Compatibility
- **AWCE-1 wallet compatibility layer**
  - User-facing addresses start with `AE` prefix
  - Raw/internal addresses use `4:` prefix
  - Validator addresses use `4:` prefix (not aevaloper/aevalcons)
  - Private keys and seed phrases never stored on-chain

### 3. Native Balance Layer
- **Bank module** for native token balances
  - Standard fungible token transfers
  - Fee payment in native denom
  - Custom token creation via AVM contracts only (not native module)

### 4. Account & Security
- **Native account/auth module**
- **Freeze functionality** (where already wired)
- **Storage rent** (where already wired)
- **Delegator protection**

### 5. Staking (Pool-Based Only)
- **Nominator pool staking system**
  - Official liquid staking pools only
  - **No direct user delegation to validators**
  - Minimum 10 AET per deposit
  - No validator selection by users
  - Validators selected by pool operators/governance

### 6. AVM Smart Contracts
- **Aetra VM** (`x/aetravm`)
  - WASM/EVM contract execution
  - Contract standards: AFT, ANFT, ADex (as AVM standards, not native modules)
  - Contract upload, instantiate, execute, query

### 7. Fee & Economy
- **Dynamic fee market**
  - Deterministic congestion-based pricing
  - Reputation-weighted priority
  - Burn, treasury, validator rewards distribution
- **Fee collector module**
- **Burn module**
- **Treasury module**
- **Emissions module**

## Out of Scope (Not Launching)

### Native Application Asset Modules
The following are **NOT** part of the testnet kernel:

| Module | Status | Target |
|--------|--------|--------|
| `x/tokenfactory` | NOT launching | AVM contract standard |
| `x/dex` | NOT launching | AVM contract standard |
| `x/nft` | NOT launching | AVM contract standard (ANFT) |
| `x/market` | Deprecated | AVM market contract |

### Prototype/Future Modules
| Module | Status |
|--------|--------|
| `x/aetra-economics` | Prototype - in-memory state |
| `x/aetra-staking-policy` | Prototype - in-memory state |
| `x/aetra-validator-score` | Prototype - in-memory state |
| `x/awce-1` (if exists) | Future standard |

## User Staking Guide (Correct)

```
# Correct: Deposit to official liquid staking pool
aetrad tx nominator-pool deposit-to-official-liquid-staking \
  --pool-id official-pool-1 \
  --amount 10aet \
  --from my-wallet

# WRONG: Direct delegation (disabled)
aetrad tx staking delegate [validator-addr] 10aet --from my-wallet
# This will fail - direct delegation to validators is disabled
```

## Testnet Kernel Verification

Run the launch scope test:
```bash
go test ./docs/... -run TestTestnetKernel
```

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-06-09 | Initial testnet kernel definition |