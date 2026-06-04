# Protocol Fees

Date: 2026-06-04

## Model

Orbitalis keeps SDK fee collection on the standard `fee_collector` module account. The `x/fees` module enforces protocol fee policy before the SDK ante handler deducts fees, then records deterministic accounting after successful deduction.

Default v1 params:

- allowed fee denom: `norb`
- minimum fee: `1norb`
- validator rewards target: `distribution/validator_rewards`
- community pool target: `protocolpool/community_pool`
- split: `98%` validator rewards, `2%` community pool

The community pool split is synchronized into `x/distribution` as `community_tax`, so the accounting policy and actual distribution policy stay aligned.

## Security Notes

- Zero-fee deliver/check transactions are rejected by `x/fees` unless simulation mode is used.
- Non-`norb` fee denoms are rejected.
- Governance cannot redirect fee collection to arbitrary module accounts in v1; target fields are explicit and validated against fixed safe values.
- Accounting uses integer truncation for the community share and assigns the remainder to validator rewards, preserving total collected fees exactly.
- Accounting state must satisfy `total_collected == validator_rewards + community_pool` and only supports `norb` in v1.

## Queries

```powershell
build\orbitalisd.exe query fees params --node tcp://127.0.0.1:26657
build\orbitalisd.exe query fees accounting --node tcp://127.0.0.1:26657
build\orbitalisd.exe query fees module-balances --node tcp://127.0.0.1:26657
```

REST:

```text
GET /l1/fees/v1/params
GET /l1/fees/v1/accounting
GET /l1/fees/v1/module_balances
```
