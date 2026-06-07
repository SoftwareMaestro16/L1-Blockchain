# Protocol Fees

Date: 2026-06-05

## Model

Aetra keeps SDK fee collection on the standard `fee_collector` module
account. The `x/fees` module enforces protocol fee policy before the SDK ante
handler deducts fees, then records deterministic accounting after successful
deduction.

Default v1 params:

- allowed fee denom: `naet`
- minimum fee: `1naet`
- base fee: `1naet`
- hard max fee cap: `1000naet`
- target block utilization: `50%`
- congestion threshold: `80%`
- max tx gas: `1,000,000`
- max block gas: `20,000,000`
- max block tx count: `5,000`
- sender rate limit: `25` txs per block, stake-weighted up to `250`
- validator rewards target: `distribution/validator_rewards`
- community pool target: `protocolpool/community_pool`
- split: `98%` validator rewards, `2%` community pool

The community pool split is synchronized into `x/distribution` as
`community_tax`, so the accounting policy and actual distribution policy stay
aligned.

## Low-Fee Dynamic Formula

Aetra does not use an Ethereum-style fee auction. The protocol computes a
single required fee from block utilization and rejects over-cap fees:

```text
utilization_bps = floor((block_gas_consumed + tx_gas_limit) * 10000 / max_block_gas)

if utilization_bps <= target_block_utilization_bps:
  required_fee = base_fee_amount
else:
  over = utilization_bps - target_block_utilization_bps
  room = 10000 - target_block_utilization_bps
  required_fee = base_fee_amount +
    ceil((max_fee_amount - base_fee_amount) * over^2 / room^2)

required_fee = min(required_fee, max_fee_amount)
```

This curve keeps normal traffic at the base fee, rises gradually as blocks fill,
and reaches the hard cap only at full utilization. A transaction paying more
than `max_fee_amount` is invalid. Paying more than the current required fee
does not improve protocol priority, so users are not pushed into an auction.

## Congestion Controls

Spam resistance is protocol-level rather than price-escalation-based:

- per-transaction gas limit rejects oversized transactions;
- per-block gas limit rejects transactions that would exceed block capacity;
- per-block tx count prevents tiny zero-work transaction floods;
- per-sender block counter rejects one sender after its allowance is exhausted;
- stake-aware formulas allow long-lived economic participants more throughput
  without raising fees for everyone;
- congestion state is derived from deterministic gas utilization, not validator
  discretion.

The sender counter stores one rolling `{height,count}` value per sender and
resets automatically when block height changes. It does not create one key per
sender per block.

## Memo Fees

Optional transaction memo text is metadata, not execution input. The memo fee is
paid only in `naet` and can be zero for an empty memo. Non-empty memo bytes
contribute to tx byte cost:

```text
memo_fee =
  memo_base_fee
  + memo_byte_fee * memo_bytes
  * reputation_multiplier(sender)
  * congestion_multiplier(load)
```

The default memo validation bound is `200` characters and `1024` bytes, with a
hard character cap of `500`. Low reputation can increase memo byte cost or delay
memo-heavy transactions; high reputation cannot bypass required protocol fees.
Memo bytes contribute to tx byte cost so memo text cannot become cheap spam
storage.

## Transaction Priority

Priority is deterministic and bounded:

```text
fee_credit = min(paid_fee, required_fee)
fee_score = fee_credit * fee_priority_weight_bps / required_fee
stake_score = floor(sender_stake / stake_tx_allowance_step_amount) *
  stake_priority_weight_bps
priority = fee_score + stake_score
```

The default weights are `10%` fee and `90%` stake. Fee overpayment above the
required amount gives no extra priority. This keeps legitimate users able to
transact during congestion while preventing fee-market revenue maximization from
becoming the scheduler.

## Edge Cases

- Missing, empty, malformed, non-positive, duplicate-denom, and non-`naet` fees
  are rejected before state mutation.
- Fees below the dynamic requirement are rejected.
- Fees above the hard cap are rejected even during full blocks.
- `gas_limit = 0`, `gas_limit > max_tx_gas`, and block gas overflow are rejected.
- Sender and block counters are not written during simulation.
- Height-0 genesis `MsgCreateValidator` transactions keep the existing bootstrap
  exemption.
- Validators remain compensated through staking rewards and inflation; high fees
  are not a security assumption.

## Attack Analysis

- Fee auction attack: ineffective because over-cap fees are invalid and
  over-required fees do not increase priority.
- Congestion price spiral: bounded by `max_fee_amount`.
- Single-sender spam: bounded by per-block sender allowance and tx/gas limits.
- Many-account spam: requires funding many accounts and still hits block tx/gas
  limits; future proposer policy can add staking-derived priority at mempool
  admission.
- Oversized tx DoS: bounded by `max_tx_gas`.
- Tiny tx flood: bounded by `max_block_txs` and sender rate limits.
- Validator censorship for higher fees: protocol priority ignores overpayment,
  and the accepted fee range is capped.

## Security Notes

- Zero-fee deliver/check transactions are rejected by `x/fees` unless
  simulation mode is used.
- Non-`naet` fee denoms are rejected even when the fee payer owns that token.
- User-created tokens, DEX LP tokens, NFT/SBT assets, display denom `AET`, and
  `testtoken` cannot pay base-chain protocol fees.
- Governance cannot redirect fee collection to arbitrary module accounts in v1;
  target fields are explicit and validated against fixed safe values.
- Accounting uses integer truncation for the community share and assigns the
  remainder to validator rewards, preserving total collected fees exactly.
- Accounting state must satisfy
  `total_collected == validator_rewards + community_pool` and only supports
  `naet` in v1.

## Queries

```powershell
build\aetrad.exe query fees params --node tcp://127.0.0.1:26657
build\aetrad.exe query fees network-load --node tcp://127.0.0.1:26657
build\aetrad.exe query fees estimate-fee 200000 --node tcp://127.0.0.1:26657
build\aetrad.exe query fees accounting --node tcp://127.0.0.1:26657
build\aetrad.exe query fees module-balances --node tcp://127.0.0.1:26657
```

REST:

```text
GET /l1/fees/v1/params
GET /l1/fees/v1/network_load
GET /l1/fees/v1/estimate_fee?gas_limit=200000
GET /l1/fees/v1/accounting
GET /l1/fees/v1/module_balances
```
