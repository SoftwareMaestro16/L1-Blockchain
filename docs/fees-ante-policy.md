> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Fees And Ante Policy

This document defines the prototype fee policy for Aetra.

## Contract

- Allowed fee denom: `naet`
- Display token: `AET`, display metadata only
- Prototype example fee: `1000000naet`
- Default localnet minimum gas price: `0naet`
- Protocol `min_fee_amount`: `1`
- Protocol `min_fee_amount` v1 cap: `1000000000000000000`
- V1 allowed fee denom list size: exactly one denom
- Fee split params: validator rewards `0.98`, community pool `0.02`

`naet` is the only accepted fee denom because it is also the base bank, staking, and mint denom. Factory denoms, LP denoms, `testtoken`, and display denom `AET` must not be used for transaction fees in the prototype.

## Ante Behavior

The `x/fees` ante decorator wraps the base Cosmos SDK ante handler. It enforces denom and minimum amount policy before the SDK signature and fee deduction chain continues.

Accepted by `x/fees` policy:

- `--fees 1000000naet`

Rejected by `x/fees` policy:

- empty fee lists
- zero native fee coins
- fees below `min_fee_amount`
- `--fees 1000testtoken`
- `--fees 1000naet,1testtoken`
- malformed fee coins
- malformed fee lists such as duplicate denom entries
- transactions that do not expose the SDK `FeeTx` interface

The localnet default `minimum-gas-prices = "0naet"` is a validator mempool setting. Aetra protocol fee policy is stricter: delivered transactions must include at least `1naet` unless they are height-0 genesis create-validator transactions.

For public testnets, keep protocol `min_fee_amount >= 1` and set validator min-gas-prices independently for mempool filtering. Localnet keeps validator min-gas-prices at `0naet` only so development commands are not filtered before protocol ante can be exercised.

## One-Command Smoke

Run the default 3-validator flow:

```powershell
.\tests\e2e\fees_ante_smoke.ps1
```

Run the 5-validator profile:

```powershell
.\tests\e2e\fees_ante_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```

Expected result:

- CLI `query fees params` returns only `naet`
- REST `/l1/fees/v1/params` returns only `naet` when the local REST gateway is healthy
- bank send with `1000000naet` fee succeeds
- tokenfactory create-denom with `1000000naet` fee succeeds
- DEX create-pool with `1000000naet` fee succeeds
- bank send, tokenfactory tx, and DEX tx with `testtoken` fee are rejected
- mixed `naet,testtoken` fees are rejected
- zero and empty fee txs are rejected by protocol fee policy

Recovery:

```powershell
.\scripts\localnet\stop.ps1
.\scripts\localnet\reset.ps1
.\tests\e2e\fees_ante_smoke.ps1
```

## Manual CLI Flow

Query fee params:

```powershell
build\aetrad.exe query fees params --node tcp://127.0.0.1:26657 --output json
Invoke-RestMethod http://127.0.0.1:1317/l1/fees/v1/params
```

Expected params:

```json
{
  "params": {
    "allowed_fee_denoms": ["naet"],
    "validator_rewards_ratio": "0.98",
    "community_pool_ratio": "0.02"
  }
}
```

Accepted fee:

```powershell
build\aetrad.exe tx bank send node0 <AE-address> 1000naet --home .localnet\node0\aetrad --chain-id aetra-local-1 --keyring-backend test --fees 1000000naet --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
```

Rejected fee:

```powershell
build\aetrad.exe tx bank send node0 <AE-address> 1000naet --home .localnet\node0\aetrad --chain-id aetra-local-1 --keyring-backend test --fees 1000testtoken --yes --broadcast-mode sync --node tcp://127.0.0.1:26657 --output json
```

Expected rejection log includes:

```text
fee denom testtoken not accepted; use naet
```

## Audit Notes

- Ante policy executes before the wrapped SDK ante handler.
- Non-`FeeTx` transactions are rejected, so callers cannot bypass denom checks with a custom tx type.
- Fee denom validation is deterministic and bounded. V1 params allow exactly one denom: `naet`.
- Fee params are loaded once per tx; malformed, empty, zero, below-minimum, and wrong-denom fee lists are rejected before any wrapped ante handler can mutate state.
- Protocol fee accounting is recorded only after wrapped SDK ante success, so invalid signer, wrong chain ID, stale sequence, insufficient fee funds, and malformed tx failures do not update fee accounting.
- Empty allowed-denom lists, duplicate denoms, and multi-denom params are rejected by params validation.
- `min_fee_amount` is bounded to a positive integer no greater than `1000000000000000000`.
- `MsgUpdateParams` requires the governance module authority and validates params before writing state.
- Wrong fee denoms return a stable error message without logging keys, mnemonics, env vars, or local paths.
- This module only enforces fee denom policy. Protocol fee distribution/accounting remains separate future work.

## Required Checks

```powershell
go test ./x/fees/...
go test ./...
go vet ./...
buf lint
go build -o build/aetrad.exe ./cmd/l1d
.\tests\e2e\fees_ante_smoke.ps1
.\tests\e2e\fees_ante_smoke.ps1 -OutputDir .localnet-5 -ValidatorCount 5
```
