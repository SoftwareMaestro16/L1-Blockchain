> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Aetra Test Matrix

This matrix maps production protocol areas to their primary regression coverage.
Test fixtures belong under `tests`, localnet helper libraries under
`scripts/localnet/lib`, and release or smoke evidence under `docs` or `.work`.

| Module / Area | Unit Tests | Keeper Tests | Integration Tests | E2E Smoke Tests | Adversarial Tests |
| --- | --- | --- | --- | --- | --- |
| `app` genesis, export/import, upgrades | `app/*_test.go`, `app/params/*_test.go`, `app/addressing/*_test.go`, `app/wasmconfig/*_test.go` | n/a | `tests/integration/*_test.go` | `tests/e2e/localnet_smoke.ps1`, `tests/e2e/export_import_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | `tests/adversarial/custom_modules_test.go` |
| `x/fees` | `x/fees/types/*_test.go` | `x/fees/keeper/*_test.go` | `tests/integration/tx_lifecycle_test.go` | `tests/e2e/fees_ante_smoke.ps1`, `tests/e2e/mempool_negative_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | `tests/adversarial/custom_modules_test.go` |
| `x/tokenfactory` | `x/tokenfactory/types/*_test.go` | `x/tokenfactory/keeper/*_test.go` | `tests/integration/tx_lifecycle_test.go` | `tests/e2e/tokenfactory_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | `tests/adversarial/custom_modules_test.go` |
| `x/dex` | `x/dex/types/*_test.go`, `tests/scripts/dex_direction_doc_test.ps1` | `x/dex/keeper/*_test.go` | `app/export_import_test.go` | `tests/e2e/dex_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | `tests/adversarial/custom_modules_test.go`, `docs/architecture/dex-direction.md` |
| `x/aetravm/standards/aft` | `x/aetravm/standards/aft/*_test.go`, `tests/scripts/aft44_standard_doc_test.ps1` | n/a until VM wiring | n/a until VM wiring | n/a until VM wiring | `docs/standards/aft-44.md` |
| `x/aetravm/standards/anft` | `x/aetravm/standards/anft/*_test.go`, `tests/scripts/anft66_asbt67_standard_doc_test.ps1` | n/a until VM wiring | n/a until VM wiring | n/a until VM wiring | `docs/standards/anft-66-asbt-67.md` |
| `x/aetravm/standards/aw` | `x/aetravm/standards/aw/*_test.go`, `tests/scripts/aw5_standard_doc_test.ps1` | n/a until VM wiring | n/a until VM wiring | n/a until VM wiring | `docs/standards/aw-5.md` |
| `x/aetravm/async` | `x/aetravm/async/*_test.go`, `tests/scripts/async_contract_execution_doc_test.ps1` | n/a until VM wiring | n/a until VM wiring | n/a until VM wiring | `docs/architecture/async-smart-contract-execution.md` |
| `x/aetravm/avm` | `x/aetravm/avm/*_test.go`, `tests/scripts/avm_doc_test.ps1` | n/a until VM wiring | n/a until VM wiring | n/a until VM wiring | `docs/architecture/avm.md`; fuzz seed `FuzzDecodeModuleRejectsMalformedWithoutPanic`; keeper/adversarial audit required before enablement |
| `x/sharding/sim` | `x/sharding/sim/*_test.go`, `tests/scripts/sharding_rd_doc_test.ps1` | n/a until consensus-safe prototype | n/a until consensus-safe prototype | n/a until experimental testnet | `docs/architecture/sharding-rd.md`; duplicate receipt, missing receipt, invalid proof, stale header, wrong destination, replay, equivocation, data unavailable; `BenchmarkRoutingTableLookup`, `BenchmarkCrossShardProofVerification`, `BenchmarkShardSplitMerge`, `BenchmarkShardedStateExportImport` |
| VM direction and CosmWasm gate | `app/wasmconfig/*_test.go`, `tests/scripts/vm_direction_doc_test.ps1` | n/a until VM wiring | n/a until VM wiring | `tests/e2e/cosmwasm_smoke.ps1` disabled-by-default mode | `docs/architecture/vm-direction.md`, `docs/security/cosmwasm-readiness.md` |
| Security audit pack | `tests/scripts/security_audit_pack_doc_test.ps1`, `tests/scripts/cosmos_security_checklist_doc_test.ps1` | n/a | n/a | Security Gate workflow: `govulncheck`, `gosec`, CodeQL, gitleaks, dependency review | `docs/security/security-audit-pack.md`, `docs/security/manual-audit-checklist.md`, `docs/security/security-triage-policy.md`, `docs/security/prototype-audit-gate.md` |
| Base-chain safety before contracts | `tests/scripts/base_chain_safety_doc_test.ps1`, `app/addressing/*_test.go`, `app/params/*_test.go` | `x/fees/keeper/ante_test.go`, custom module keeper tests | `tests/integration/tx_lifecycle_test.go`, `tests/integration/pos_lifecycle_test.go` | `tests/e2e/mempool_negative_smoke.ps1`, `tests/e2e/adversarial_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | `tests/adversarial/custom_modules_test.go`, `docs/security/account-transaction-safety.md`, `scripts/security/determinism-gate.ps1` |
| Performance and speed | `app/addressing/*_benchmark_test.go`, `x/fees/types/*_test.go`, `x/aetravm/standards/*/*_test.go`, `tests/scripts/performance_speed_doc_test.ps1` | `x/dex/keeper/*_test.go`, `x/tokenfactory/keeper/*_test.go` | n/a | localnet `startup-timing.json`, CI `GITHUB_STEP_SUMMARY` smoke timing | `docs/performance-speed.md`, benchmark before/after records |
| Public testnet preparation | `tests/scripts/public_testnet_preparation_doc_test.ps1` | n/a | `scripts/testnet/public-testnet-preflight.ps1` profiles 3, 5, and 10 | `tests/e2e/prototype_acceptance.ps1`, `tests/e2e/cosmwasm_smoke.ps1` when gated, localnet 3/5/10 profiles | `docs/public-testnet-preparation.md`, `docs/validator-onboarding.md`, `docs/testnet-incident-response.md` |
| Public testnet and production gates | `tests/scripts/public_testnet_production_gates_doc_test.ps1` | n/a | public preflight profile All plus Phase 12 modular execution gate | long-running public testnet before production | `docs/public-testnet-production-gates.md`, `docs/security/prototype-audit-gate.md` |
| PoS staking flow | `app/pos_test.go` | SDK staking keeper coverage plus app wiring tests | `tests/integration/pos_lifecycle_test.go` | `tests/e2e/pos_smoke.ps1`, `tests/e2e/prototype_acceptance.ps1` | `docs/security/pos-staking-correctness.md` |
| CLI / localnet scripts | `cmd/l1d/cmd/*_test.go`, `tests/scripts/*_test.ps1` | n/a | n/a | `tests/e2e/localnet_smoke.ps1`, `tests/e2e/query_surface_smoke.ps1` | `tests/e2e/adversarial_smoke.ps1` |
| Observability | `observability/*_test.go` | n/a | n/a | `tests/scripts/observability_scripts_test.ps1`, `scripts/localnet/health.ps1` | `docs/security/determinism-gate.md` |

## Native Account, Staking, Rent, Proof, And Event Matrix

This matrix is the acceptance checklist for native wallet/account, auth policy,
official liquid staking, stake reputation, storage rent, proofs, and events.
Every item must map to a Go test, PowerShell static check, integration test, or
documented e2e smoke before the feature is considered production-complete.

### Wallet Tests

- Activate account success.
- Activate duplicate rejected.
- Address mismatch from `derive(pubkey)` rejected.
- Private key not serialized/exported.
- Seed phrase not serialized/exported.
- `AE...` roundtrip.
- `4:...` roundtrip.
- `AE... <-> 4:...` roundtrip.
- Sequence starts deterministic.
- Account number starts deterministic.
- Export/import preserves account.
- Lazy migration preserves existing account.
- Unsupported version rejected safely.
- Auth policy update requires authorization.
- Frozen account cannot spend.
- Archived account keeps minimal required state.
- Closed account allowed only with zero obligations.

### Auth Policy Tests

- Single-key success.
- Multisig insufficient signatures rejected.
- Threshold exact threshold accepted.
- Weighted threshold deterministic.
- Two-device policy requires device signature for protected action.
- Spending limit allows small transfer.
- Spending limit rejects large transfer.
- Timelock blocks early recovery.
- Recovery rotates key without changing address.
- Auth policy never serializes private key/seed/2FA secrets.

### Staking Tests

- User deposits AET into official liquid staking contract and receives pool shares.
- Normal user staking deposit does not include validator selection.
- User receives rewards from pool share after validators earn rewards.
- Low pool deposit below `MinPoolDeposit` rejected.
- Validator below `MinValidatorStake` rejected.
- Validator entry below `1_000_000 AET` rejected.
- Solo validator below `1_000_000 AET` self-stake rejected.
- Pool-backed validator below `400_000 AET` self-stake rejected.
- Pool-backed validator with more than `600_000 AET` nominator stake toward minimum entry rejected.
- Direct user delegation to validator rejected.
- `MaxValidatorCount` enforced.
- Validator count below `100` rejected for production/mainnet params unless explicit testnet override is active.
- Validator count above `300` rejected.
- `TargetValidatorCount` deterministic behavior.
- Allocation engine weights validators by reputation, uptime, commission, limits, stake efficiency, slashing risk, and network load.
- Allocation engine excludes jailed/slashed validators from new positive allocation.
- Pool injects aggregated stake into multiple validators.
- Pool rebalance deterministic and bounded.
- Pool unbonding delay of `18 days` enforced.
- Early pool unbonding release rejected.
- Reward index deterministic rounding.
- Pool-user income equals gross staking rewards minus validator commission, pool fee, and slashing losses.
- Validator income equals self-stake rewards plus commission/operator bonus minus infrastructure cost and slashing/jail losses.
- 300,000 AET pool user illustrative annual math fixture at 14.4% gross yield, 10% validator commission, and 1% pool fee returns `38,491.2 AET`.
- 300,000 AET validator illustrative annual math fixture with 300,000 AET pool allocation at 14.4% gross yield and 10% commission returns `47,520 AET` before infrastructure cost.
- Staking economics tests verify the same formulas for multiple configurable gross yield inputs, not only 14.4%.
- Claim rewards updates only caller pool-share state.
- No full scan path in reward calculation.
- Million-user style scalability test.
- Export/import preserves pools/shares/allocations/rewards/unbondings.
- Unauthorized params update rejected.
- Jailed validators receive no positive validator bonus.

### Reputation Tests

- Stake-time reputation increases with pool share exposure times duration.
- No stake-time means no reputation.
- Reputation claim deterministic.
- Slashed/jailed validator cannot receive bonus.
- Export/import preserves accumulator.
- Reputation is account-owned and non-transferable.

### Storage Rent Tests

- Active wallet pays tiny lazy rent.
- Unactivated address pays no rent.
- Empty/no-state address pays no rent.
- Rent is computed from `code_bytes + data_bytes`.
- Rent increases with elapsed seconds of storage duration.
- Rent is collected automatically during account tx or contract/pool action.
- Rent is deducted from the balance of the account/contract/pool/module/protocol payer that owns or funds the state.
- Wallet tx effective fee includes gas fee, storage rent delta, and unpaid storage debt.
- Wallet with enough balance pays debt and remains `active`.
- Wallet with insufficient balance becomes `frozen`.
- Zero balance undeployed/no-state address has zero rent and zero debt.
- Zero balance active wallet/contract with persistent state accrues debt and preserves state.
- System/module accounts are not storage-rent exempt and are charged through protocol accounting.
- System storage reserve runway is computed deterministically.
- System storage reserve warning threshold emits alert.
- System storage reserve critical threshold triggers deterministic top-up from fee collector/treasury.
- System rent top-up executes before user-account freeze logic.
- Protocol-critical system account cannot enter `frozen`, `archived`, or `deleted` status because of storage rent.
- Underfunded system rent raises invariant/alert but consensus-critical modules continue executing.
- Empty or undeployed addresses are the only addresses that do not accrue storage rent.
- Validator operator wallet pays like a normal wallet.
- Validator consensus-critical records are not frozen by rent.
- Official liquid staking pool contract has stable `AE...` and `4:...` addresses.
- Official liquid staking pool pays storage rent from pool fee/reserve/treasury policy.
- Pool share, allocation, reward index, and unbonding records contribute to pool storage accounting.
- Official pool cannot be deleted while shares, active stake, unbonding, pending rewards, allocation records, or required proofs exist.
- Official pool enters `frozen_limited` instead of trapping funds when rent debt exceeds threshold.
- `frozen_limited` pool rejects new deposits but allows claims, unbond requests, matured withdrawals, and governance/top-up recovery.
- Contract storage accrues rent.
- Normal frozen contract rejects execute and new state writes.
- Normal frozen contract allows top-up, debt payment, unfreeze, read-only query, and proof query.
- Frozen wallet preserves balance, state, sequence, auth/recovery policy, ownership, reputation, pool shares, domains, and pending rewards.
- Frozen contract preserves balance, code, data, admin/owner, storage records, and proof keys.
- Domain records accrue rent or are protocol-accounted.
- Debt freezes wallet.
- Payment unfreezes wallet.
- Top-up plus debt payment can recover frozen wallet/contract to `active`.
- Frozen status never wipes state, resets sequence, deletes code/data, or deletes ownership.
- `frozen`, `archived`, and `deleted` have distinct validated semantics.
- Wallet cannot be automatically deleted for storage debt.
- Wallet can be archived only when no balance, stake, pool shares, domains, pending rewards, ownership obligations, or required reputation state remain.
- Close rejected while obligations exist.
- Export/import preserves rent state.

### Proof Tests

- Pool deposit proof metadata stable.
- Pool share proof metadata stable.
- Pool allocation proof metadata stable.
- Reward claim proof metadata stable.
- Reputation proof metadata stable.
- State key deterministic.
- Proof query returns height/key/root/path metadata.
- Proofs do not require scanning all users.

### Event Tests

- `AccountActivated` stable.
- `PoolStakeDeposited` stable.
- `PoolSharesMinted` stable.
- `PoolAllocationUpdated` stable.
- `PoolUnbondingRequested` stable.
- `PoolUnbondingCompleted` stable.
- `PoolRewardsClaimed` stable.
- `StakeReputationClaimed` stable.
- `ValidatorRegistered` stable.
- `ValidatorUpdated` stable.
- `UnbondingStarted` stable.
- `UnbondingCompleted` stable.
- Events never include private key, seed phrase, or 2FA secrets.
