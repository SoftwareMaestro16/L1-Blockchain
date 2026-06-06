> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# System Architecture

## Version Baseline

- Cosmos SDK: target `v0.54.x`, pinned exactly during scaffold. Current observed latest release is `v0.54.3`.
- Consensus: CometBFT through the Cosmos SDK application wiring.
- Serialization: Protobuf for consensus-critical transactions, state values, and genesis ingestion.
- Storage: Cosmos SDK store services and module KVStores. Application code must not depend directly on a database backend.
- Experimental features: BlockSTM and related performance features stay disabled until explicit load, determinism, and adversarial testing justify activation.

## Chain Model

The chain is a sovereign account-based proof-of-stake L1. CometBFT orders blocks and drives ABCI calls. The Cosmos SDK application defines deterministic transaction execution, module state, and the AppHash commitment.

Target execution flow:

```text
Client
  -> Tx bytes
  -> CometBFT CheckTx
  -> Mempool
  -> PrepareProposal
  -> ProcessProposal
  -> FinalizeBlock
  -> AnteHandler
  -> MsgServiceRouter
  -> x/<module> MsgServer
  -> Keeper
  -> KVStore
  -> Commit AppHash
```

## Core Modules

- `x/auth`: accounts, signing, sequence and replay protection.
- `x/bank`: balances and token transfers.
- `x/staking`: validators and delegation.
- `x/slashing`: downtime and double-sign accountability.
- `x/gov`: on-chain governance and authority for params.
- `x/mint`: native inflation.
- `x/distribution`: staking and reward distribution.
- `x/upgrade`: coordinated state migrations and binary upgrades.

## Custom Modules

- `x/tokenfactory`: controlled custom denom creation, admin rights, minting, burning, and metadata.
- `x/dex`: constant-product AMM with deterministic integer math.
- `x/fees`: configurable protocol fee collection and distribution.
- `x/bridge`: future interoperability module, disabled until a complete trust and verification model exists.

## Determinism Rules

- Use block context time, never local system time.
- Use integer math only in consensus state transitions.
- Do not call external APIs from consensus logic.
- Do not iterate maps when order affects writes, events, or results.
- Do not use global mutable state.
- Keep all state writes behind module keepers.
