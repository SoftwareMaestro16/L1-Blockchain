# Local Funding Flow

This flow is local-only prototype tooling. It funds accounts by sending `naet` from a genesis-funded localnet key, usually `node0`. It does not mint tokens, does not add a faucet module, and must not be used as a production funding path.

## Safety Rules

- Only local chain IDs are accepted. The default is `aetra-local-1`.
- The script reads CometBFT `/status` from `127.0.0.1:<RPCPort>` and refuses to run if the network does not match `-ChainId`.
- The sender key is read from an ignored localnet home with `--keyring-backend test`.
- No mnemonics, private keys, or keyring files are printed.
- Funding is a normal signed `tx bank send`, so sequence, fees, and insufficient funds are handled by the chain.

## Single Recipient

```powershell
$NODE1 = build\aetrad.exe keys show node1 -a --home .localnet\node1\aetrad --keyring-backend test
.\scripts\localnet\fund.ps1 `
  -OutputDir .localnet `
  -Binary build\aetrad.exe `
  -ChainId aetra-local-1 `
  -RPCPort 26657 `
  -Recipients @($NODE1) `
  -Amount 1000000naet
```

## Multiple Recipients And Amounts

Use `-Recipients` when every account receives the same `-Amount`, and `-Transfers` for explicit `address=amount` entries.

```powershell
$NODE1 = build\aetrad.exe keys show node1 -a --home .localnet\node1\aetrad --keyring-backend test
$NODE2 = build\aetrad.exe keys show node2 -a --home .localnet\node2\aetrad --keyring-backend test

.\scripts\localnet\fund.ps1 `
  -OutputDir .localnet `
  -Binary build\aetrad.exe `
  -ChainId aetra-local-1 `
  -RPCPort 26657 `
  -Recipients @($NODE1) `
  -Transfers @("$NODE2=2500000naet") `
  -Amount 1000000naet `
  -Json
```

Repeated runs are allowed; each run sends another bank transaction and verifies the recipient balance delta.

## Negative Checks

The e2e funding smoke covers:

- successful funding of multiple local accounts,
- non-local chain-id rejection,
- chain-id mismatch against RPC status,
- missing local funder key failure.

Run it with:

```powershell
.\tests\e2e\funding_smoke.ps1
```
