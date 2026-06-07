> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# PoS And Staking Correctness

This document records the Phase 6 Aetra PoS and staking acceptance policy.
The detailed protocol-level slashing design is defined in
[Aetra Slashing System](slashing-system.md).

## Production Staking Policy

The production staking track must keep the following policy explicit and
testable before public testnet:

- `bond_denom = "naet"` is the only valid staking denom.
- validator self-delegation must be positive, in `naet`, and bounded by
  genesis and message validation.
- commission max rate, commission max change rate, and minimum commission must
  be governance-bounded and validated.
- unbonding time, max validators, max entries, historical entries, and
  redelegation limits must be bounded.
- jailed validators must not remain active producers until unjail rules pass.
- validator edit, jail, unjail, unbond, redelegation, and removal paths must
  preserve staking, slashing, distribution, and bank module invariants.

Parameter validation must cover staking, slashing, distribution, mint, and fee params together because `naet` issuance, rewards, fee collection, and validator power are coupled.

## Native Staking Denom

Aetra staking uses only `naet` as the bond denom. Validator creation,
delegation, unbonding, redelegation, rewards, slashing, and localnet genesis
tests must not rely on `stake`, `uatom`, display denom `AET`, factory denoms,
LP denoms, or `testtoken`.

## Uncapped PoS Supply

AET has no fixed max supply. The base denom `naet` is issued through the
configured Cosmos SDK mint inflation and reward policy. In v1, mint params must
remain deterministic and governance-bounded:

- `mint_denom = "naet"`
- `max_supply = 0`, which means uncapped supply in SDK mint params
- inflation rate change, min inflation, max inflation, goal bonded, and blocks
  per year must pass SDK mint param validation
- reward and distribution accounting must be deterministic and test-covered

## Validator Income Model

Validator income must not depend only on transaction fees. The production
security budget is modeled as:

```text
validator_income =
  validator_share_of_mint_rewards
  + validator_share_of_fee_rewards
  + validator_commission

delegator_income =
  delegator_share_of_mint_rewards
  + delegator_share_of_fee_rewards
  - validator_commission

validator_reward_weight =
  validator_power / total_validator_power
```

The implementation keeps the arithmetic in native integer `naet` units and
uses basis points for ratios. The hard policy bounds are:

```text
min_commission = 100 bps    # 1%
max_commission = 2000 bps   # 20%
max_daily_commission_change = 100 bps # 1%
```

Mint rewards provide baseline validator/delegator income. Fee rewards add
activity-linked upside, but fee revenue is not a security assumption. The
treasury and future insurance reserve are separate public-goods and emergency
recovery tracks.

The security threshold remains:

```text
attack_cost >= bonded_value_controlled_by_attacker
```

Operational monitoring must flag validator concentration:

- `< 1/3` malicious voting power should not break consensus safety;
- `1/3+` malicious voting power is a serious halt/liveness risk;
- `2/3+` malicious voting power can finalize malicious state and must be
  economically infeasible.

`app/params/economy.go` contains the deterministic income formulas used by
tests and future keeper integration.

## System Balance Controller

Aetra economy controls four feedback loops:

```text
low staking -> inflation rises -> staking becomes more attractive
high staking -> inflation falls -> dilution decreases
high activity -> burn rises -> supply expansion slows
low activity -> burn falls -> mint rewards maintain validator security
high congestion -> soft fee multiplier + queue/rate limits
low congestion -> low fees
```

The controller inputs are staking ratio, block load, async queue depth, failed
tx rate, annual mint, and annual burn. The deterministic outputs are:

- adaptive inflation, clamped to `1%..5%`;
- burn ratio, clamped to `10%..50%`;
- validator fee ratio after treasury and burn shares;
- congestion, queue-limit, and rate-limit flags.

Deflation guard:

```text
if annual_burn > annual_mint * 1.25:
  burn_ratio moves down toward min_burn_ratio
```

The controller is deliberately bounded. Governance can tune soft params only
inside hard-coded safety ranges; raising the hard inflation cap or disabling
fee caps requires a software upgrade.

## Validator Lifecycle Coverage

Unit tests in `app/pos_test.go` cover:

- validator creation with `naet` self-delegation
- delegation increasing validator tokens and consensus power
- unbonding entries and delayed balance return
- redelegation entries between bonded validators
- invalid delegation denom, funds, and address rejection
- slashing params and downtime missed-block bitmap persistence
- delegator reward withdrawal through the distribution module
- `naet` mint params, uncapped max supply, and bounded inflation params

Integration tests in `tests/integration/pos_lifecycle_test.go` cover:

- signed staking tx delivery through ante and `FinalizeBlock`
- validator-set updates returned to CometBFT after staking power changes
- staking delegation state surviving export/import restart

The next production-hardening test additions are:

- validator edit, jail, unjail, and inactive validator removal
- unbonding completion after the unbonding period
- redelegation completion and redelegation limit rejection
- downtime jailing behavior with missed block windows
- rewards and commission accounting against distribution module balances
- staking pool invariants for bonded, unbonded, and not-bonded funds

## Localnet Acceptance

The localnet acceptance path is:

```powershell
.\scripts\build-aetrad.ps1
.\tests\e2e\pos_smoke.ps1 -Binary .\build\aetrad.exe
.\tests\e2e\pos_smoke.ps1 -Binary .\build\aetrad.exe -OutputDir .localnet-5 -ValidatorCount 5
```

`tests/e2e/pos_smoke.ps1` validates a 3-validator local network by default and
supports a 5-validator profile:

- initializes and validates localnet genesis
- starts the network and waits for blocks
- checks CometBFT validator count and bonded staking validators
- confirms staking bond denom `naet`
- checks slashing params and signing info
- submits a delegation and confirms total voting power increases
- confirms invalid delegation paths fail
- confirms replayed signed tx fails

`scripts/localnet/validate-genesis.ps1` is the genesis-specific guard for local
profiles. It asserts matching genesis hashes across nodes, exactly
`ValidatorCount` gentxs and bank balances, `MsgCreateValidator`
self-delegation in `naet`, `staking/mint/fees` denom consistency, and empty
tokenfactory and DEX genesis state.

## Production Localnet Profiles

The localnet profile matrix for staking hardening is:

| Profile | Validators | Purpose | Required before |
| --- | ---: | --- | --- |
| Dev | 1 | fast local module and CLI iteration | every developer smoke |
| Smoke | 3 | normal validator-set and peer behavior | public testnet candidate |
| Rehearsal | 5 | public-testnet validator topology rehearsal | public testnet launch |
| Stress | 10 | staking, query, and localnet script scale evidence | production candidate |
| Long run | 20 | long-running validator lifecycle and restart evidence | production candidate |

3-validator and 5-validator profiles are mandatory before public testnet. The
10-validator and 20-validator profiles are production-readiness gates and may
run in scheduled or manual environments.

## Restart, State-Sync, Snapshot, And Chaos

Production staking readiness requires repeatable recovery drills:

- stop/start after delegation;
- stop/start after unbonding;
- stop/start after slashing state changes;
- state-sync join after staking changes;
- snapshot restore after staking changes;
- kill one validator in a 3-validator profile;
- kill two validators in a 5-validator profile;
- restart a validator with stale local state and verify safe failure;
- corrupt local node data outside consensus state and verify diagnostics.

State-sync and snapshot restore must produce the same staking state root for
the restored height. Recovery evidence must not include validator keys,
mnemonics, node keys, or keyring contents.

## Invariants And Benchmarks

Required staking invariants:

- bonded, unbonded, and not-bonded pools match staking keeper state;
- validator tokens and delegator shares are internally consistent;
- distribution outstanding rewards match module balances;
- total `naet` supply changes only through mint, reward, slash, and explicit
  burn paths;
- validator updates returned by `FinalizeBlock` match staking keeper power.

Required benchmarks:

- create validator;
- delegate;
- redelegate;
- unbond;
- validator set update;
- reward withdrawal;
- slashing missed-block window scan.

## Acceptance

- Local 3-validator network produces blocks.
- Local 5-validator genesis validates with the same staking and denom policy.
- Validator power updates after staking transactions and propagates through
  `FinalizeBlock` validator updates.
- Export/import restart preserves staking state and consensus progress.
- 3-validator and 5-validator long-run smokes produce blocks without staking
  invariant drift.
- State-sync and snapshot restore preserve staking state roots.
- No staking path accepts non-`naet` bond or fee denoms.
